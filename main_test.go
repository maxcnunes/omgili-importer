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
