package main

import (
	"flag"
	"log"
	"os"
	"strconv"
)

func ParseFlags() {
	flag.StringVar(&AppConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&AppConfig.PollInterval, "p", 2, "polling interval for collecting metrics")
	flag.IntVar(&AppConfig.ReportInterval, "r", 10, "reporting interval for sending metrics to server")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		AppConfig.ServerAddress = envRunAddr
	}
	if envRunReportInterval := os.Getenv("REPORT_INTERVAL"); envRunReportInterval != "" {
		value, err := strconv.Atoi(envRunReportInterval)
		if err != nil {
			log.Fatal("invalid REPORT_INTERVAL:", err)
		}

		AppConfig.ReportInterval = value
	}
	if envRunPoolInterval := os.Getenv("POLL_INTERVAL"); envRunPoolInterval != "" {
		value, err := strconv.Atoi(envRunPoolInterval)
		if err != nil {
			log.Fatal("invalid POLL_INTERVAL:", err)
		}

		AppConfig.PollInterval = value
	}
}
