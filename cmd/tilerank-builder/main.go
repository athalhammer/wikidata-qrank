// SPDX-License-Identifier: MIT

package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
)

var logger *log.Logger

func main() {
	logpath := filepath.Join("logs", "tilerank-builder.log")
	if err := os.MkdirAll("logs", os.ModePerm); err != nil {
		log.Fatal(err)
	}

	logfile, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer logfile.Close()
	logger = log.New(logfile, "", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	client := &http.Client{}
	weeks, err := GetAvailableWeeks(client)
	if err != nil {
		logger.Fatal(err)
		return
	}

	maxWeeks := 3 * 52 // 3 years
	if len(weeks) > maxWeeks {
		weeks = weeks[len(weeks)-maxWeeks:]
	}
	logger.Printf(
		"found %d weeks with OpenStreetMap tile logs, from %s to %s",
		len(weeks), weeks[0], weeks[len(weeks)-1])

	cachedir := "cache"
	for _, week := range weeks {
		if _, err := GetTileLogs(week, client, cachedir); err != nil {
			logger.Fatal(err)
			return
		}
	}
}