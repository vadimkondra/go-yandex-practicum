package main

import (
	"flag"
	"go-yandex-practicum/internal/config"
	"log"
	"os"
	"strconv"
)

func ParseFlags() config.AgentConfig {

	cfg := config.AgentConfig{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.PollInterval, "p", 2, "polling interval for collecting metrics")
	flag.IntVar(&cfg.ReportInterval, "r", 10, "reporting interval for sending metrics to server")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.ServerAddress = envRunAddr
	}
	if envRunReportInterval := os.Getenv("REPORT_INTERVAL"); envRunReportInterval != "" {
		value, err := strconv.Atoi(envRunReportInterval)
		if err != nil {
			log.Fatal("invalid REPORT_INTERVAL:", err)
		}

		cfg.ReportInterval = value
	}
	if envRunPoolInterval := os.Getenv("POLL_INTERVAL"); envRunPoolInterval != "" {
		value, err := strconv.Atoi(envRunPoolInterval)
		if err != nil {
			log.Fatal("invalid POLL_INTERVAL:", err)
		}

		cfg.PollInterval = value
	}

	return cfg
}
