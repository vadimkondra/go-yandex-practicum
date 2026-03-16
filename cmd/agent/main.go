package main

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"time"

	"go-yandex-practicum/internal/config"
)

var AppConfig config.AgentConfig

type gauge float64
type counter int64

func main() {
	parseFlags()

	client := &http.Client{}
	metricsMap := make(map[string]gauge)

	pollTicker := time.NewTicker(time.Duration(AppConfig.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(AppConfig.ReportInterval) * time.Second)
	var pollCount counter = 0

	for {
		select {
		case <-pollTicker.C:
			// обновляем метрики runtime
			readMemStatMetrics(metricsMap)
			pollCount++

		case <-reportTicker.C:
			// отправляем метрики на сервер
			if pollCount > 0 {
				sendGaugeMetrics(client, metricsMap)
				sendCounterMetric(client, "PollCount", pollCount)
				pollCount = 0
			}
		}
	}
}

func parseFlags() {
	flag.StringVar(&AppConfig.ServerAddress, "a", "localhost:8080", "address and port to run server")
	flag.IntVar(&AppConfig.PollInterval, "p", 2, "polling interval for collecting metrics")
	flag.IntVar(&AppConfig.ReportInterval, "r", 10, "reporting interval for sending metrics to server")

	flag.Parse()
}

func buildUpdateMetricURL(metricType string, metricNm string, metricVal string) string {
	return "update/" + metricType + "/" + metricNm + "/" + metricVal
}

func sendRequest(client *http.Client, url string) {
	req, err := http.NewRequest(http.MethodPost, url, nil)

	if err != nil {
		panic(err)
	}

	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	fmt.Println("url:", url)
	fmt.Println("Status:", response.Status)
}

func sendGaugeMetrics(client *http.Client, metrics map[string]gauge) {
	for metricName, metricValue := range metrics {
		sendGaugeMetric(client, metricName, metricValue)
	}
}

func sendCounterMetric(client *http.Client, metricName string, metricValue counter) {
	url := "http://" + AppConfig.ServerAddress + "/" + buildUpdateMetricURL("counter", metricName, strconv.Itoa(int(metricValue)))
	sendRequest(client, url)
}

func sendGaugeMetric(client *http.Client, metricName string, metricValue gauge) {
	url := "http://" + AppConfig.ServerAddress + "/" + buildUpdateMetricURL("gauge", metricName, strconv.FormatFloat(float64(metricValue), 'f', -1, 64))
	sendRequest(client, url)
}

func readMemStatMetrics(metrics map[string]gauge) {
	var memStats runtime.MemStats

	runtime.ReadMemStats(&memStats)

	metrics["Alloc"] = gauge(memStats.Alloc)
	metrics["BuckHashSys"] = gauge(memStats.BuckHashSys)
	metrics["Frees"] = gauge(memStats.Frees)
	metrics["GCCPUFraction"] = gauge(memStats.GCCPUFraction)
	metrics["GCSys"] = gauge(memStats.GCSys)
	metrics["HeapAlloc"] = gauge(memStats.HeapAlloc)
	metrics["HeapIdle"] = gauge(memStats.HeapIdle)
	metrics["HeapInuse"] = gauge(memStats.HeapInuse)
	metrics["HeapObjects"] = gauge(memStats.HeapObjects)
	metrics["HeapReleased"] = gauge(memStats.HeapReleased)
	metrics["HeapSys"] = gauge(memStats.HeapSys)
	metrics["LastGC"] = gauge(memStats.LastGC)
	metrics["Lookups"] = gauge(memStats.Lookups)
	metrics["MCacheInuse"] = gauge(memStats.MCacheInuse)
	metrics["MCacheSys"] = gauge(memStats.MCacheSys)
	metrics["MSpanInuse"] = gauge(memStats.MSpanInuse)
	metrics["MSpanSys"] = gauge(memStats.MSpanSys)
	metrics["Mallocs"] = gauge(memStats.Mallocs)
	metrics["NextGC"] = gauge(memStats.NextGC)
	metrics["NumForcedGC"] = gauge(memStats.NumForcedGC)
	metrics["NumGC"] = gauge(memStats.NumGC)
	metrics["OtherSys"] = gauge(memStats.OtherSys)
	metrics["PauseTotalNs"] = gauge(memStats.PauseTotalNs)
	metrics["StackInuse"] = gauge(memStats.StackInuse)
	metrics["StackSys"] = gauge(memStats.StackSys)
	metrics["Sys"] = gauge(memStats.Sys)
	metrics["TotalAlloc"] = gauge(memStats.TotalAlloc)

	metrics["RandomValue"] = gauge(rand.Float64())
}
