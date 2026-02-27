// SPDX-FileCopyrightText: 2022 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFindEntitiesDump(t *testing.T) {
	dumpsDir := t.TempDir()
	sitelinksDir := filepath.Join(dumpsDir, "sitelinks")
	if err := os.MkdirAll(sitelinksDir, 0755); err != nil {
		t.Error(err)
		return
	}

	// Create a sitelinks file with format: <site>-YYYYMMDD.site.links
	sitelinksFile := filepath.Join(sitelinksDir, "enwiki-20250215.site.links")
	if f, err := os.Create(sitelinksFile); err == nil {
		f.Close()
	} else {
		t.Error(err)
		return
	}

	// Create an older file to ensure we pick the latest one
	oldFile := filepath.Join(sitelinksDir, "enwiki-20250201.site.links")
	if f, err := os.Create(oldFile); err == nil {
		f.Close()
	} else {
		t.Error(err)
		return
	}

	expectedDate := "2025-02-15"
	date, path, err := findEntitiesDump(dumpsDir)
	if err != nil {
		t.Error(err)
		return
	}

	if d := date.Format("2006-01-02"); d != expectedDate {
		t.Errorf("expected %s, got %s", expectedDate, d)
	}

	if path != sitelinksDir {
		t.Errorf("expected %q, got %q", sitelinksDir, path)
	}
}
