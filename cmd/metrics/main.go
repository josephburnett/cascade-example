package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type metric struct {
	qps int
	end time.Time
}

var (
	mux     sync.Mutex
	metrics map[string]map[string]*metric
	qps     *prom.GaugeVec
)

func init() {
	metrics = make(map[string]map[string]*metric)
	qps = prom.NewGaugeVec(
		prom.GaugeOpts{
			Name:      "qps",
			Subsystem: "total",
		},
		[]string{"service"},
	)
	prom.MustRegister(qps)
}

func main() {
	ticker := time.NewTicker(time.Second)
	go func() {
		for {
			select {
			case <-ticker.C:
				reportMetrics()
			}
		}
	}()

	http.HandleFunc("/", httpHandler)
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	path := strings.Split(r.URL.Path, "/")
	if len(path) != 7 || path[1] != "service" || path[3] != "pod" || path[5] != "qps" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expecting /service/{}/pod/{}/qps/{}\n"))
		return
	}
	service, pod, qps := path[2], path[4], path[6]
	qpsInt, err := strconv.Atoi(qps)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}
	log.Printf("metric (%v, %v, %v)", service, pod, qps)
	writeMetric(service, pod, qpsInt)
}

func writeMetric(service, pod string, qps int) {
	mux.Lock()
	defer mux.Unlock()
	mService, ok := metrics[service]
	if !ok {
		mService = make(map[string]*metric)
		metrics[service] = mService
	}
	mPod, ok := mService[pod]
	if !ok {
		mPod = &metric{}
		mService[pod] = mPod
	}
	mPod.qps = qps
	mPod.end = time.Now()
}

func reportMetrics() {
	mux.Lock()
	defer mux.Unlock()
	now := time.Now()
	for service, mService := range metrics {
		totalQps := 0
		for pod, mPod := range mService {
			if mPod.end.Add(time.Minute).Before(now) {
				delete(mService, pod)
				continue
			}
			totalQps += mPod.qps
		}
		gauge := qps.WithLabelValues(service)
		gauge.Set(float64(totalQps))
		log.Printf("reporting (%v %v)", service, totalQps)
	}
}
