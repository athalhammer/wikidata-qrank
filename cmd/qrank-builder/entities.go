// SPDX-FileCopyrightText: 2022 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/andybalholm/brotli"
	"github.com/lanrat/extsort"
)


// findEntitiesDump now returns the latest date and the sitelinks directory path
func findEntitiesDump(dumpsPath string) (time.Time, string, error) {
	sitelinksDir := filepath.Join(dumpsPath, "sitelinks")
	entries, err := os.ReadDir(sitelinksDir)
	if err != nil {
		return time.Time{}, "", err
	}
	var latest time.Time
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Expect format: <site>-YYYYMMDD.site.links
		parts := strings.Split(name, "-")
		if len(parts) < 2 {
			continue
		}
		datePart := strings.Split(parts[len(parts)-1], ".")[0]
		t, err := time.Parse("20060102", datePart)
		if err == nil && t.After(latest) {
			latest = t
		}
	}
	if latest.IsZero() {
		return time.Time{}, "", fmt.Errorf("no sitelinks files found")
	}
	return latest, sitelinksDir, nil
}

func processEntities(testRun bool, sitelinksDir string, date time.Time, outDir string, ctx context.Context) (string, error) {
	year, month, day := date.Year(), date.Month(), date.Day()
	sitelinksPath := filepath.Join(
		outDir,
		fmt.Sprintf("sitelinks-%04d%02d%02d.br", year, month, day))
	_, err := os.Stat(sitelinksPath)
	if err == nil {
		logger.Printf("using cached sitelinks: %s", sitelinksPath)
		return sitelinksPath, nil // use pre-existing file
	}
	if !os.IsNotExist(err) {
		return "", err
	}

	logger.Printf("processing sitelinks files for %04d-%02d-%02d", year, month, day)
	start := time.Now()

	tmpSitelinksPath := sitelinksPath + ".tmp"
	tmpSitelinksFile, err := os.Create(tmpSitelinksPath)
	if err != nil {
		return "", err
	}
	defer tmpSitelinksFile.Close()

	sitelinksWriter := brotli.NewWriterLevel(tmpSitelinksFile, 6)
	defer sitelinksWriter.Close()

	ch := make(chan string, 10000)
	config := extsort.DefaultConfig()
	config.ChunkSize = 8 * 1024 * 1024 / 16 // 8 MiB, 16 Bytes/line avg
	config.NumWorkers = runtime.NumCPU()
	sorter, outChan, errChan := extsort.Strings(ch, config)
	g, subCtx := errgroup.WithContext(ctx)
	g.Go(func() error {
		return readSitelinksFiles(testRun, sitelinksDir, ch, subCtx)
	})
	g.Go(func() error {
		sorter.Sort(subCtx)
		if err := writeSitelinks(outChan, sitelinksWriter, subCtx); err != nil {
			return err
		}
		return nil
	})
	if err := g.Wait(); err != nil {
		return "", err
	}
	if err := <-errChan; err != nil {
		return "", err
	}

	if err := sitelinksWriter.Close(); err != nil {
	}
	if err := tmpSitelinksFile.Sync(); err != nil {
	}
	if err := tmpSitelinksFile.Close(); err != nil {
	}
	if err := os.Rename(tmpSitelinksPath, sitelinksPath); err != nil {
		return "", err
	}
	logger.Printf("built sitelinks for %04d-%02d-%02d in %.1fs",
		year, month, day, time.Since(start).Seconds())
	return sitelinksPath, nil
}


// Reads all .site.links files in the sitelinksDir and emits lines in the format: <site>/<id> <entity>
func readSitelinksFiles(testRun bool, sitelinksDir string, sitelinks chan<- string, ctx context.Context) error {
	defer close(sitelinks)
	entries, err := os.ReadDir(sitelinksDir)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".site.links") {
			continue
		}
		rawSite := entry.Name()
		// Transform: enwiki-20251201.site.links -> en.wikipedia
		site := rawSite
		if dash := strings.Index(site, "-"); dash != -1 {
			site = site[:dash]
		}
		if strings.HasSuffix(site, "wiki") {
			site = site[:len(site)-4] + ".wikipedia"
		}
		sitePath := filepath.Join(sitelinksDir, rawSite)
		f, err := os.Open(sitePath)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			fields := strings.Fields(line)
			if len(fields) != 2 {
				continue
			}
			entity := fields[0]
			id := fields[1]
			out := fmt.Sprintf("%s/%s %s", site, id, entity)
			select {
			case sitelinks <- out:
			case <-ctx.Done():
				f.Close()
				return ctx.Err()
			}
		}
		f.Close()
	}
	return nil
}


func writeSitelinks(ch <-chan string, w io.Writer, ctx context.Context) error {
	for {
		select {
		case line, ok := <-ch:
			if !ok { // channel closed, end of input
				return nil
			}
			if _, err := w.Write([]byte(line)); err != nil {
				return err
			}
			if _, err := w.Write([]byte{'\n'}); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}
