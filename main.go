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

func main() {
	var feedURL = flag.String("url", defaultFeedURL, "URL for feed list")
	flag.Parse()
	fmt.Printf("Starting importer from %s\n", *feedURL)

	resp, err := Download(*feedURL, pathTemFeedListFile)
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
		zipURL := resp.URL + filename
		fmt.Printf("Downloading file %s\n", zipURL)
		_, err := Download(zipURL, filename)
		if err != nil {
			fmt.Print("Error downloading ZIP feed data", err)
			os.Exit(1)
		}
		fmt.Print("DOWNLOAD OK")
		os.Exit(1)
		// err := archiver.Unzip("input.zip", "output_folder")
	}
}
