package main

import (
	"go-yandex-practicum/internal/handler"
	"net/http"
)

type MemStorage struct {
	gauges   map[string]float64
	counters map[string]int64
}

func main() {
	r := handler.ConfigServerRouter()

	http.ListenAndServe(":8080", r)
}
