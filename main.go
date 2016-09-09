package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
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

func main() {
	var feedURL = flag.String("url", defaultFeedURL, "URL for feed list")
	flag.Parse()
	fmt.Printf("Starting importer from %s\n", *feedURL)

	err := DownloadFeedList(*feedURL)
	if err != nil {
		fmt.Print("Error downloading feed list", err)
	}
}
