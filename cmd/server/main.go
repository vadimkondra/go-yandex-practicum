package main

import (
	"go-yandex-practicum/internal/config"
	"go-yandex-practicum/internal/handler"
	"go-yandex-practicum/internal/service"
	"go-yandex-practicum/internal/store"
	"log"
	"net"
	"net/http"
)

func main() {
	cfg := ParseFlags()

	storage := InitStorage(cfg)

	defer func(storage store.Storage) {
		err := storage.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(storage)

	metricsService := service.NewMetricsService(storage)
	r := handler.ConfigServerRouter(metricsService, cfg)

	_, port, err := net.SplitHostPort(cfg.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + port

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func InitStorage(cfg config.ServerConfig) store.Storage {
	storage, err := store.NewStorage(cfg)

	if err != nil {
		log.Fatal(err)
	}

	service.SetStorage(storage)

	return storage
}
