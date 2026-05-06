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
	flag.StringVar(&cfg.Key, "k", "", "hash key")
	flag.IntVar(&cfg.RateLimit, "l", 100, "rate limit for sending metrics to server")
	flag.Parse()

	if envRunAddr, ok := os.LookupEnv("ADDRESS"); ok && envRunAddr != "" {
		cfg.ServerAddress = envRunAddr
	}
	if envRunReportInterval, ok := os.LookupEnv("REPORT_INTERVAL"); ok && envRunReportInterval != "" {
		value, err := strconv.Atoi(envRunReportInterval)
		if err != nil {
			log.Fatal("invalid REPORT_INTERVAL:", err)
		}

		cfg.ReportInterval = value
	}
	if envRunPollInterval, ok := os.LookupEnv("POLL_INTERVAL"); ok && envRunPollInterval != "" {
		value, err := strconv.Atoi(envRunPollInterval)
		if err != nil {
			log.Fatal("invalid POLL_INTERVAL:", err)
		}

		cfg.PollInterval = value
	}

	if envRunKey, ok := os.LookupEnv("KEY"); ok && envRunKey != "" {
		cfg.Key = envRunKey
	}

	if envRateLimit, ok := os.LookupEnv("RATE_LIMIT"); ok && envRateLimit != "" {
		value, err := strconv.Atoi(envRateLimit)
		if err == nil {
			cfg.RateLimit = value
		} else {
			log.Fatal("invalid RATE_LIMIT:", err)
		}
	}

	if cfg.RateLimit <= 0 {
		cfg.RateLimit = 1
	}

	return cfg
}
