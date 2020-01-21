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
	"github.com/josephburnett/cascade-example/pkg/rate"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	serviceName  = flag.String("service-name", "default", "The name of this service.")
	opWeight     = flag.Duration("op-weight-millicore-duration", 0, "Weight in millicores per time unit of each request.")
	generateOps  = flag.Int("generate-ops", 0, "Operations per second to generate.")
	dependencies = flag.String("dependencies", "", "Comma-separated list of downstream service dependencies.")

	reporter *metrics.Reporter
	limiter  *rate.Limiter
	services []string
)

func init() {
	flag.Parse()
	reporter = metrics.NewReporter(*serviceName)
	limiter = rate.NewLimiter(1)
	services = strings.Split(*dependencies, ",")
}

func main() {

	if *generateOps > 0 {
		log.Printf("Generating ops %v\n", *generateOps)
		intervalMillis := 1000 / *generateOps
		ticker := time.NewTicker(time.Duration(intervalMillis) * time.Millisecond)
		for i := 0; i < 100; i++ {
			go gen(ticker.C)
		}
	}

	http.HandleFunc("/", httpHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

func op() (string, int) {

	// Metrics and logs
	start := time.Now()
	report := reporter.Success
	status := http.StatusOK
	output := "( " + *serviceName + " "
	done := func() {
		output += fmt.Sprintf("%v )", status)
		report(time.Since(start))
	}

	// Circuit breaker
	if ok := limiter.In(); !ok {
		report = reporter.Overload
		status = http.StatusServiceUnavailable
		done()
		return output, status
	}
	defer limiter.Out()

	// Do some work
	burnCpu(*opWeight)
	output += fmt.Sprintf("%v ", *opWeight)

	// Call some other services
	failure := false
	for _, s := range services {
		if s == "" {
			continue
		}
		svc := fmt.Sprintf("http://%v.cascade-example.svc.cluster.local", s)
		client := &http.Client{
			Timeout: time.Second * 10,
		}
		resp, err := client.Get(svc)
		switch {
		case err != nil:
			failure = true
			output += fmt.Sprintf("%q ", err.Error())
		case resp.StatusCode != http.StatusOK:
			failure = true
			output += fmt.Sprintf("(%v %v)", s, resp.StatusCode)
		default:
			o, _ := ioutil.ReadAll(resp.Body)
			resp.Body.Close()
			output += fmt.Sprintf("%v ", string(o))
		}
	}

	// Error if any dependencies failed.
	if failure {
		report = reporter.Failure
		status = http.StatusInternalServerError
	}

	done()
	return output, status
}

func gen(ch <-chan time.Time) {
	for {
		select {
		case <-ch:
			go func() {
				output, _ := op()
				log.Println(output)
			}()
		}
	}
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	output, status := op()
	w.WriteHeader(status)
	w.Write([]byte(output))
	log.Println(output)
}

func burnCpu(d time.Duration) {
	if d == 0 {
		return
	}
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
