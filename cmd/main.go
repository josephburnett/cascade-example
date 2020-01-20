package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/josephburnett/cascade-example/pkg/metrics"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serviceName  = flag.String("service-name", "default", "The name of this service.")
	weight       = flag.Duration("request-weight-millicore-duration", 200, "Weight in millicores per time unit of each request.")
	dependencies = flag.String("dependencies", "", "Comma-separated list of downstream service dependencies.")
)

func main() {
	flag.Parse()

	http.HandleFunc("/", newHandler())
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

func newHandler() func(http.ResponseWriter, *http.Request) {
	reporter := metrics.NewReporter(*serviceName)
	services := strings.Split(*dependencies, ",")
	return func(w http.ResponseWriter, r *http.Request) {

		start := time.Now()
		output := "( " + *serviceName + " "

		// Do some work
		burnCpu(*weight)
		output += fmt.Sprintf("%v ", *weight)

		// Call some other services
		failure := false
		for _, s := range services {
			if s == "" {
				continue
			}
			svc := fmt.Sprintf("http://%v.cascade-example.svc.cluster.local", s)
			resp, err := http.Get(svc)
			if err != nil {
				failure = true
				output += fmt.Sprintf("%q ", err.Error())
			} else {
				o, _ := ioutil.ReadAll(resp.Body)
				resp.Body.Close()
				output += fmt.Sprintf("%v ", string(o))
			}
		}

		if failure {
			reporter.Failure(time.Since(start))
			w.WriteHeader(http.StatusInternalServerError)
			output += "ERROR "
		} else {
			reporter.Success(time.Since(start))
			output += "OK "
			w.WriteHeader(http.StatusOK)
		}
		output += ")"
		log.Println(output)
		w.Write([]byte(output))
	}
}

func burnCpu(d time.Duration) {
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
	time.Sleep(d)
	close(done)
}
