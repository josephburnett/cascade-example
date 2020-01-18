package main

import (
	"flag"
	"net/http"
	"strings"
	"time"

	"github.com/josephburnett/cascade-example/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	serviceName := flag.String("service-name", "default", "The name of this service.")
	requestWeightMillicoreSeconds := flag.Int("request-weight-millicore-seconds", 200, "Weight in millicores of each request.")
	dependencies := flag.String("dependencies", "", "Comma-separated list of downstream service dependencies.")
	flag.Parse()

	requestHandler := newHandler(*serviceName, *requestWeightMillicoreSeconds, *dependencies)

	http.HandleFunc("/", requestHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

func newHandler(serviceName string, weight int, dependencies string) func(http.ResponseWriter, *http.Request) {
	reporter := metrics.NewReporter(serviceName)
	services := strings.Split(dependencies, ",")
	return func(w http.ResponseWriter, r *http.Request) {

		// Do some work
		burnCpu(weight)

		// Call some other services
		for _, _ = range services {
			// TODO: make service calls
		}
		reporter.Success(0)
	}
}

func burnCpu(millicoreSeconds int) {
	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-done:
				return
			default:
			}
		}
	}()
	time.Sleep(time.Duration(millicoreSeconds*1000000) * time.Millisecond)
	close(done)
}
