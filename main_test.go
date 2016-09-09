package main

import (
	"flag"
	"os"
	"testing"
)

var isIntegratoinEnabled = flag.Bool("integration", false, "Enable integration tests")

func TestDownloadFeedList(t *testing.T) {
	if !*isIntegratoinEnabled {
		t.Skip("Itegration tests disabled")
	}

	var tests = []struct {
		url     string
		success bool
	}{
		{defaultFeedURL, true},
		{"http://bitly.com/nuvi-plz_not_valid", false},
	}

	for _, tt := range tests {
		err := DownloadFeedList(tt.url)
		if !tt.success && err == nil {
			t.Error("Expected to return an error for invalid URLs")
		} else if tt.success {
			if err != nil {
				t.Error(err)
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
