package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"

	"github.com/mholt/archiver"
)

const (
	defaultFeedURL      = "http://bitly.com/nuvi-plz"
	pathTemFeedListFile = "omgili-feed-list.html"
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
func ExtractFeedFileNames(pathTemFeedListFile string, chFiles chan string) error {
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
			chFiles <- matches[1]
		}
	}

	return scanner.Err()
}

// FindZIPFiles ...
func FindZIPFiles(pathZIPFiles string, chFiles chan string) error {
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
			chFiles <- fi.Name()
		}
		return nil
	}

	return filepath.Walk(pathZIPFiles, findZIPFile)
}

func main() {
	var feedURL = flag.String("url", defaultFeedURL, "URL for feed list")
	var downloadDisabled = flag.Bool("disable-download", false, "Disable downloads. Useful to run over pre fetched zip files")
	flag.Parse()
	fmt.Printf("Starting importer from %s\n", *feedURL)

	var err error
	var resp *DownloadResponse
	chFiles := make(chan string)

	if *downloadDisabled {
		go func() {
			if err := FindZIPFiles(".", chFiles); err != nil {
				fmt.Println("Error finding feed filenames", err)
				os.Exit(1)
			}
			close(chFiles)
		}()
	} else {
		resp, err = Download(*feedURL, pathTemFeedListFile)
		if err != nil {
			fmt.Println("Error downloading feed list", err)
			os.Exit(1)
		}

		go func() {
			if err := ExtractFeedFileNames(pathTemFeedListFile, chFiles); err != nil {
				fmt.Print("Error extracting feed filenames", err)
				os.Exit(1)
			}
			close(chFiles)
		}()
	}

	for filename := range chFiles {
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

		if !*downloadDisabled {
			fmt.Println("DOWNLOAD OK")
			os.Exit(1)
		}
	}
}
