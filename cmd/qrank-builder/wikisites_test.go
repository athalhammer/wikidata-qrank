// SPDX-FileCopyrightText: 2024 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestReadWikiSites(t *testing.T) {
	sites, err := ReadWikiSites(filepath.Join("testdata", "dumps"))
	if err != nil {
		t.Error(err)
	}

	tests := []struct{ key, domain, lastDumped string }{
		{"loginwiki", "login.wikimedia.org", "2024-05-01"},
		{"rmwiki", "rm.wikipedia.org", "2024-03-01"},
		{"wikidatawiki", "www.wikidata.org", "2024-04-01"},
	}
	for _, tc := range tests {
		site := (*sites)[tc.key]
		if site.Domain != tc.domain {
			t.Errorf(`got "%s", want "%s", for sites["%s"].Domain`, site.Domain, tc.domain, tc.key)
		}
		lastDumped := site.LastDumped.Format(time.DateOnly)
		if lastDumped != tc.lastDumped {
			t.Errorf(`got %s, want %s, for sites["%s"].LastDumped`, lastDumped, tc.lastDumped, tc.key)
		}
	}
}

func TestReadWikiSites_BadPath(t *testing.T) {
	_, err := ReadWikiSites(filepath.Join("testdata", "no-such-dir"))
	if !os.IsNotExist(err) {
		t.Errorf("want os.NotExists, got %v", err)
	}
}

// A fake HTTP transport that simulates a Wikimedia site for testing.
type FakeWikiSite struct {
	Broken bool
}

func (f *FakeWikiSite) RoundTrip(req *http.Request) (*http.Response, error) {
	header := make(http.Header)

	if f.Broken {
		header.Add("Content-Type", "text/plain")
		body := io.NopCloser(bytes.NewBufferString("Service Unavailable"))
		return &http.Response{StatusCode: 503, Body: body, Header: header}, nil
	}

	var project string
	switch req.URL.Hostname() {
	case "rm.wikipedia.org":
		project = "rmwiki"
	}

	var filename string
	if req.URL.Path == "/w/api.php" {
		if req.URL.RawQuery == "action=query&meta=siteinfo&siprop=interwikimap&format=json" {
			filename = "interwikimap.json"
		}
	}

	if project == "" || filename == "" {
		fmt.Printf("*** %q\n", req.URL.RawPath)
		fmt.Printf("*** %q\n", req.URL.RawQuery)
		return nil, fmt.Errorf("unexpected request: %s", req.URL.String())
	}

	path := filepath.Join("testdata", "fake_wikisite", project, filename)
	body, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(path, ".json") {
		header.Add("Content-Type", "application/json")
	}
	return &http.Response{StatusCode: 200, Body: body, Header: header}, nil
}
