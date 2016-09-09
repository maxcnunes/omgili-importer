package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
)

const (
	defaultFeedURL      = "http://bitly.com/nuvi-plz"
	pathTemFeedListFile = "omgili-feed-list.html"
)

// DownloadFeedList get the feed list from the omgili webpage and save the result to a local file
func DownloadFeedList(omgiliURL string) error {
	out, err := os.Create(pathTemFeedListFile)
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(omgiliURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New("Problem fetching feed list. Status Code: " + resp.Status)
	}

	_, err = io.Copy(out, resp.Body)
	return err
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

func main() {
	var feedURL = flag.String("url", defaultFeedURL, "URL for feed list")
	flag.Parse()
	fmt.Printf("Starting importer from %s\n", *feedURL)

	err := DownloadFeedList(*feedURL)
	if err != nil {
		fmt.Print("Error downloading feed list", err)
		os.Exit(1)
	}

	chFiles := make(chan string)
	go func() {
		if err := ExtractFeedFileNames(pathTemFeedListFile, chFiles); err != nil {
			fmt.Print("Error extracting feed filenames", err)
			os.Exit(1)
		}
		close(chFiles)
	}()

	for filename := range chFiles {
		fmt.Print(filename)
	}
}
