// SPDX-FileCopyrightText: 2022 Sascha Brawer <sascha@brawer.ch>
// SPDX-License-Identifier: MIT

package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var logger *log.Logger

func main() {
	var dumps = flag.String("dumps", "/public/dumps/public", "path to Wikimedia dumps")
	var testRun = flag.Bool("testRun", false, "if true, we process only a small fraction of the data; used for testing")
	flag.Parse()

	workdir, _ := os.Getwd()
	logPath := filepath.Join("logs", "qrank-builder.log")
	fmt.Printf("logs written to %s in workdir=%s", logPath, workdir)
	fmt.Fprintf(os.Stderr, "logs written to %s in workdir=%s", logPath, workdir)
	logfile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)
	logger.Printf("qrank-builder starting up")

	if err := computeQRank(*dumps, *testRun); err != nil {
		logger.Printf("ComputeQRank failed: %v", err)
		log.Fatal(err)
		return
	}

	logger.Printf("qrank-builder exiting")
}

func computeQRank(dumpsPath string, testRun bool) error {
	ctx := context.Background()
	outDir := "cache"
	if testRun {
		outDir = "cache-testrun"
	}

	if err := os.MkdirAll(outDir, 0755); err != nil {
		return err
	}

	if err := CleanupCache(outDir); err != nil {
		return err
	}

	edate, _, err := findEntitiesDump(dumpsPath)
	if err != nil {
		return err
	}

	// Use the month before the sitelinks date for pageviews, since pageviews
	// data typically lags behind sitelinks data
	pvDate := edate.AddDate(0, -1, 0)
	pageviews, err := processPageviews(testRun, dumpsPath, pvDate, outDir, ctx)
	if err != nil {
		return err
	}

	sitelinksDir := filepath.Join(dumpsPath, "sitelinks")
	sitelinks, err := processEntities(testRun, sitelinksDir, edate, outDir, ctx)
	if err != nil {
		return err
	}

	qviews, err := buildQViews(testRun, edate, sitelinks, pageviews, outDir, ctx)
	if err != nil {
		return err
	}

	_, err = buildQRank(edate, qviews, outDir, ctx)
	if err != nil {
		return err
	}

	return nil
}
