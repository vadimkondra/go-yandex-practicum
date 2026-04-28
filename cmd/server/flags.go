package main

import (
	"flag"
	"os"
	"strconv"
)

func ParseFlags() {
	flag.StringVar(&AppConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&AppConfig.StoreInterval, "i", 300, "interval in seconds between metrics store")
	flag.StringVar(&AppConfig.FileStorePath, "f", "./metric-data", "path to store data")
	flag.BoolVar(&AppConfig.Restore, "r", false, "restore metric data")
	flag.StringVar(&AppConfig.DatabaseDsn, "d", "", "database dsn")

	flag.Parse()

	if envRunAddr := os.Getenv("ADDRESS"); envRunAddr != "" {
		AppConfig.ServerAddress = envRunAddr
	}

	if storeInterval := os.Getenv("STORE_INTERVAL"); storeInterval != "" {
		parsedStoreInterval, err := strconv.Atoi(storeInterval)

		if err == nil {
			AppConfig.StoreInterval = parsedStoreInterval
		}
	}

	if filePath := os.Getenv("FILE_STORAGE_PATH"); filePath != "" {
		AppConfig.FileStorePath = filePath
	}

	if restore := os.Getenv("RESTORE"); restore != "" {
		parsedRestore, err := strconv.ParseBool(restore)

		if err == nil {
			AppConfig.Restore = parsedRestore
		}
	}

	if dataBaseDsn := os.Getenv("DATABASE_DSN"); dataBaseDsn != "" {
		AppConfig.DatabaseDsn = dataBaseDsn
	}

}
