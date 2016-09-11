package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"

	"gopkg.in/redis.v4"

	"github.com/mholt/archiver"
)

const (
	defaultFeedURL        = "http://bitly.com/nuvi-plz"
	pathTemFeedListFile   = "omgili-feed-list.html"
	redisListNewsXML      = "news_xml"
	redisBaseIndexNewsXML = "news_xml_index_"
)

// DownloadResponse ...
type DownloadResponse struct {
	Path string
	URL  string
}

// Download ...
func Download(url string, outputPath string) (*DownloadResponse, error) {
	out, err := os.Create(outputPath)
	if err != nil {
		return nil, err
	}
	defer out.Close()

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, errors.New(resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return nil, err
	}

	return &DownloadResponse{
		Path: outputPath,
		URL:  resp.Request.URL.String(),
	}, nil
}

// ExtractFeedFileNames read the feed list HTML and extract the feed filenames
func ExtractFeedFileNames(pathTemFeedListFile string, chZIPFiles chan string) error {
	regexFeedFile, _ := regexp.Compile(`href="(.*\.zip)"`)

	file, err := os.Open(pathTemFeedListFile)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		matches := regexFeedFile.FindStringSubmatch(scanner.Text())
		if len(matches) == 2 {
			chZIPFiles <- matches[1]
		}
	}

	return scanner.Err()
}
func redisNewClient(address string, password string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:     address,
		Password: password,
		DB:       db,
	})
}

// FindZIPFiles ...
func FindZIPFiles(pathZIPFiles string, chZIPFiles chan string) error {
	findZIPFile := func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() && fp != pathZIPFiles {
			return filepath.SkipDir
		}
		matched, err := regexp.MatchString(`^.*\.zip$`, fi.Name())
		if err != nil {
			return err
		}
		if matched {
			chZIPFiles <- fi.Name()
		}
		return nil
	}

	return filepath.Walk(pathZIPFiles, findZIPFile)
}

// FindXMLFiles ...
func FindXMLFiles(pathXMLFiles string, chXMLFiles chan string) error {
	findXMLFile := func(fp string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() && fp != pathXMLFiles {
			return filepath.SkipDir
		}
		matched, err := regexp.MatchString(`^.*\.xml$`, fi.Name())
		if err != nil {
			return err
		}
		if matched {
			chXMLFiles <- fi.Name()
		}
		return nil
	}

	return filepath.Walk(pathXMLFiles, findXMLFile)
}

// SetOrPushNewsToList ...
func SetOrPushNewsToList(client *redis.Client, listName string, hash string, content []byte) error {
	var index int64
	indexKey := redisBaseIndexNewsXML + hash

	val, err := client.Get(indexKey).Result()
	if err == redis.Nil {
		index = -1
	} else if err != nil {
		return err
	} else {
		index, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return err
		}
	}

	if index > -1 {
		fmt.Println("Updating news", index, hash)
		return client.LSet(listName, index, content).Err()
	}

	fmt.Println("Pushing news", index, hash)
	rPush := client.RPush(listName, content)
	if rPush.Err() != nil {
		return err
	}

	index = rPush.Val()

	return client.Set(indexKey, index, 0).Err()
}

func main() {
	var feedURL = flag.String("url", defaultFeedURL, "URL for feed list")
	var redisAddress = flag.String("redis-address", "localhost:6379", "Redis address")
	var redisPassword = flag.String("redis-password", "", "Redis password")
	var redisDB = flag.Int("redis-database", 0, "Redis database")
	var downloadDisabled = flag.Bool("disable-download", false, "Disable downloads. Useful to run over pre fetched zip files")
	flag.Parse()
	fmt.Printf("Starting importer from %s\n", *feedURL)

	client := redisNewClient(*redisAddress, *redisPassword, *redisDB)
	var err error
	var resp *DownloadResponse
	chZIPFiles := make(chan string)

	if *downloadDisabled {
		go func() {
			if err := FindZIPFiles(".", chZIPFiles); err != nil {
				fmt.Println("Error finding ZIP feed filenames", err)
				os.Exit(1)
			}
			close(chZIPFiles)
		}()
	} else {
		resp, err = Download(*feedURL, pathTemFeedListFile)
		if err != nil {
			fmt.Println("Error downloading feed list", err)
			os.Exit(1)
		}

		go func() {
			if err := ExtractFeedFileNames(pathTemFeedListFile, chZIPFiles); err != nil {
				fmt.Print("Error extracting feed filenames", err)
				os.Exit(1)
			}
			close(chZIPFiles)
		}()
	}

	for filename := range chZIPFiles {
		if !*downloadDisabled {
			zipURL := resp.URL + filename
			fmt.Printf("Downloading file %s\n", zipURL)
			_, err := Download(zipURL, filename)
			if err != nil {
				fmt.Println("Error downloading ZIP feed data", err)
				os.Exit(1)
			}
		}

		fmt.Println("Extracting", filename)
		err = archiver.Unzip(filename, ".")
		if err != nil {
			fmt.Println("Error extracting ZIP feed data", err)
			os.Exit(1)
		}

		chXMLFiles := make(chan string)
		go func() {
			if err := FindXMLFiles(".", chXMLFiles); err != nil {
				fmt.Println("Error finding feed XML filenames", err)
				os.Exit(1)
			}
			close(chXMLFiles)
		}()

		for xmlFileName := range chXMLFiles {
			content, err := ioutil.ReadFile(xmlFileName)
			if err != nil {
				fmt.Println("Error reading XML", err)
				os.Exit(1)
			}

			err = SetOrPushNewsToList(client, redisListNewsXML, xmlFileName, content)
			if err != nil {
				fmt.Println("Error saving XML into the Redis DB", err)
				os.Exit(1)
			}
		}

		if !*downloadDisabled {
			fmt.Println("DOWNLOAD OK")
			os.Exit(1)
		}
	}
}
