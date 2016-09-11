package main

import (
	"flag"
	"os"
	"reflect"
	"testing"
)

var isIntegratoinEnabled = flag.Bool("integration", false, "Enable integration tests")

func TestDownload(t *testing.T) {
	if !*isIntegratoinEnabled {
		t.Skip("Itegration tests disabled")
	}

	var tests = []struct {
		url     string
		success bool
		resp    *DownloadResponse
	}{
		{defaultFeedURL, true, &DownloadResponse{
			Path: "omgili-feed-list.html",
			URL:  "http://feed.omgili.com/5Rh5AMTrc4Pv/mainstream/posts/"}},
		{"http://bitly.com/nuvi-plz_not_valid", false, nil},
	}

	for _, tt := range tests {
		resp, err := Download(tt.url, pathTemFeedListFile)
		if !tt.success && err == nil {
			t.Error("Expected to return an error for invalid URLs")
		} else if tt.success {
			if err != nil {
				t.Error(err)
			}

			if !reflect.DeepEqual(resp, tt.resp) {
				t.Errorf("Expected %#v to be equal to %#v", resp, tt.resp)
			}

			if _, err := os.Stat(pathTemFeedListFile); os.IsNotExist(err) {
				t.Errorf("Expected %s to have been download from %s", pathTemFeedListFile, defaultFeedURL)
			}
		}

	}
}

func TestExtractFeedFileNames(t *testing.T) {
	var tests = []struct {
		filename   string
		success    bool
		totalFound int
	}{
		{"fixtures/" + pathTemFeedListFile, true, 1159},
		{"fixtures/invalid_file", false, 0},
	}

	for _, tt := range tests {
		totalFound := 0
		chFiles := make(chan string)

		go func() {
			err := ExtractFeedFileNames(tt.filename, chFiles)
			if !tt.success && err == nil {
				t.Error("Expected to return an error for invalid files")
			}
			close(chFiles)
		}()

		for range chFiles {
			totalFound++
		}

		if totalFound != tt.totalFound {
			t.Errorf("Expected to find %d but got %d", tt.totalFound, totalFound)
		}
	}
}

func TestFindZIPFiles(t *testing.T) {
	var tests = []struct {
		path       string
		success    bool
		totalFound int
	}{
		{"fixtures/omgili-feeds.zip", true, 1},
		{"fixtures/invalid_file", false, 0},
	}

	for _, tt := range tests {
		totalFound := 0
		chFiles := make(chan string)

		go func() {
			err := FindZIPFiles(tt.path, chFiles)
			if !tt.success && err == nil {
				t.Error("Expected to return an error for invalid files")
			}
			close(chFiles)
		}()

		for range chFiles {
			totalFound++
		}

		if totalFound != tt.totalFound {
			t.Errorf("Expected to find %d but got %d", tt.totalFound, totalFound)
		}
	}
}
