package main

import (
	"flag"
	"go-yandex-practicum/internal/config"
	"os"
	"strconv"
)

func ParseFlags() config.ServerConfig {
	serverConfig := config.ServerConfig{}

	flag.StringVar(&serverConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&serverConfig.StoreInterval, "i", 300, "interval in seconds between metrics store")
	flag.StringVar(&serverConfig.FileStorePath, "f", "./metric-data", "path to store data")
	flag.BoolVar(&serverConfig.Restore, "r", false, "restore metric data")
	flag.StringVar(&serverConfig.DatabaseDSN, "d", "", "database dsn")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		serverConfig.ServerAddress = envRunAddr
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		parsedStoreInterval, err := strconv.Atoi(storeInterval)

		if err == nil {
			serverConfig.StoreInterval = parsedStoreInterval
		}
	}

	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		serverConfig.FileStorePath = filePath
	}

	if restore := os.Getenv("RESTORE"); restore != "" {
		parsedRestore, err := strconv.ParseBool(restore)

		if err == nil {
			serverConfig.Restore = parsedRestore
		}
	}

	if dataBaseDsn := os.Getenv("DATABASE_DSN"); dataBaseDsn != "" {
		serverConfig.DatabaseDSN = dataBaseDsn
	}

	return serverConfig
}
