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

var AppConfig config.ServerConfig

func main() {
	ParseFlags()

	storage := InitStorage()

	defer storage.Close()

	r := handler.ConfigServerRouter()

	_, port, err := net.SplitHostPort(AppConfig.ServerAddress)
	if err != nil {
		log.Fatal(err)
	}

	addr := ":" + port

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal(err)
	}
}

func InitStorage() store.Storage {
	storage, err := store.NewStorage(AppConfig)

	if err != nil {
		log.Fatal(err)
	}

	service.SetStorage(storage)

	return storage
}
