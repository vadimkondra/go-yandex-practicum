package main

import (
	"flag"
	"go-yandex-practicum/internal/config"
	"os"
	"strconv"
)

func ParseFlags() config.ServerConfig {
	cfg := config.ServerConfig{}

	flag.StringVar(&cfg.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&cfg.StoreInterval, "i", 300, "interval in seconds between metrics store")
	flag.StringVar(&cfg.FileStorePath, "f", "./metric-data", "path to store data")
	flag.BoolVar(&cfg.Restore, "r", false, "restore metric data")
	flag.StringVar(&cfg.DatabaseDSN, "d", "", "database dsn")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		cfg.ServerAddress = envRunAddr
	}

	if dataBaseDsn := os.Getenv("DATABASE_DSN"); dataBaseDsn != "" {
		cfg.DatabaseDSN = dataBaseDsn
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		parsedStoreInterval, err := strconv.Atoi(storeInterval)

		if err == nil {
			cfg.StoreInterval = parsedStoreInterval
		}
	}

	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		cfg.FileStorePath = filePath
	}

	if restore := os.Getenv("RESTORE"); restore != "" {
		parsedRestore, err := strconv.ParseBool(restore)

		if err == nil {
			cfg.Restore = parsedRestore
		}
	}

	return cfg
}
