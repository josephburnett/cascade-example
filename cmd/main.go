package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"math"
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
	limit        = flag.Int("ops-limit", 1, "OPS limit of the service after which it returns 503 overload.")
	dependencies = flag.String("dependencies", "", "Comma-separated list of downstream service dependencies.")

	reporter *metrics.Reporter
	limiter  *rate.Limiter
	services []string
)

func init() {
	flag.Parse()
	reporter = metrics.NewReporter(*serviceName)
	limiter = rate.NewLimiter(*limit)
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
		output += fmt.Sprintf("%v %v )", time.Since(start), status)
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
	// burnCpu(*opWeight)
	// output += fmt.Sprintf("%v ", *opWeight)
	p := largestPrime(8000000) // ~0.1 cpu/sec
	output += fmt.Sprintf("%v ", p)

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
			resp.Body.Close()
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

// Primes less than or equal to N will be generated
// Algorithm from https://stackoverflow.com/a/21854246
func allPrimes(N int) []int {

	var x, y, n int
	nsqrt := math.Sqrt(float64(N))

	is_prime := make([]bool, N)

	for x = 1; float64(x) <= nsqrt; x++ {
		for y = 1; float64(y) <= nsqrt; y++ {
			n = 4*(x*x) + y*y
			if n <= N && (n%12 == 1 || n%12 == 5) {
				is_prime[n] = !is_prime[n]
			}
			n = 3*(x*x) + y*y
			if n <= N && n%12 == 7 {
				is_prime[n] = !is_prime[n]
			}
			n = 3*(x*x) - y*y
			if x > y && n <= N && n%12 == 11 {
				is_prime[n] = !is_prime[n]
			}
		}
	}

	for n = 5; float64(n) <= nsqrt; n++ {
		if is_prime[n] {
			for y = n * n; y < N; y += n * n {
				is_prime[y] = false
			}
		}
	}

	is_prime[2] = true
	is_prime[3] = true

	primes := make([]int, 0, 1270606)
	for x = 0; x < len(is_prime)-1; x++ {
		if is_prime[x] {
			primes = append(primes, x)
		}
	}

	// primes is now a slice that contains all primes numbers up to N
	return primes
}

func largestPrime(max int) int {
	p := allPrimes(max)
	if len(p) > 0 {
		return p[len(p)-1]
	} else {
		return 0
	}
}
