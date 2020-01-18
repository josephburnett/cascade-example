package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	metricName := flag.String("metric-name", "foo", "custom metric name")
	metricValue := flag.Int64("metric-value", 0, "custom metric value")
	port := flag.Int64("port", 8080, "port to expose metrics on")
	flag.Parse()

	metric := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: *metricName,
			Help: "Custom metric",
		},
	)
	prometheus.MustRegister(metric)
	metric.Set(float64(*metricValue))

	http.Handle("/metrics", promhttp.Handler())
	log.Printf("Starting to listen on :%d", *port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	log.Fatal("Failed to start serving metrics: %v", err)
}
