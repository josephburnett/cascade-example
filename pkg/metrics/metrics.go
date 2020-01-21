package metrics

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
)

type Reporter struct {
	mux sync.Mutex

	qps     *prom.GaugeVec
	latency *prom.GaugeVec

	successCount  float64
	overloadCount float64
	timeoutCount  float64
	failCount     float64

	successMillis  float64
	overloadMillis float64
	timeoutMillis  float64
	failMillis     float64
}

func NewReporter(service, pod string) *Reporter {
	r := &Reporter{
		qps: prom.NewGaugeVec(
			prom.GaugeOpts{
				Name:      "qps",
				Subsystem: service,
			},
			[]string{"code"},
		),
		latency: prom.NewGaugeVec(
			prom.GaugeOpts{
				Name:      "latency",
				Subsystem: service,
			},
			[]string{"code"},
		),
	}
	prom.MustRegister(r.qps)
	prom.MustRegister(r.latency)

	// Update metrics once per second.
	go func() {
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ticker.C:
				r.mux.Lock()

				// Report to local aggregating service.
				var totalQps float64
				for _, qps := range []float64{r.successCount, r.failCount, r.overloadCount, r.timeoutCount} {
					totalQps += qps
				}
				go reportQps(service, pod, totalQps)

				// Report to Stackdriver
				for _, m := range []struct {
					gauge  *prom.GaugeVec
					status int
					value  *float64
					over   float64
				}{{
					gauge:  r.qps,
					status: http.StatusOK,
					value:  &r.successCount,
					over:   1,
				}, {
					gauge:  r.qps,
					status: http.StatusServiceUnavailable,
					value:  &r.overloadCount,
					over:   1,
				}, {
					gauge:  r.qps,
					status: http.StatusGatewayTimeout,
					value:  &r.timeoutCount,
					over:   1,
				}, {
					gauge:  r.qps,
					status: http.StatusInternalServerError,
					value:  &r.failCount,
					over:   1,
				}, {
					gauge:  r.latency,
					status: http.StatusOK,
					value:  &r.successMillis,
					over:   r.successCount,
				}, {
					gauge:  r.latency,
					status: http.StatusServiceUnavailable,
					value:  &r.overloadMillis,
					over:   r.overloadCount,
				}, {
					gauge:  r.latency,
					status: http.StatusGatewayTimeout,
					value:  &r.timeoutMillis,
					over:   r.timeoutCount,
				}, {
					gauge:  r.latency,
					status: http.StatusInternalServerError,
					value:  &r.failMillis,
					over:   r.failCount,
				}} {
					g := m.gauge.WithLabelValues(strconv.Itoa(m.status))
					g.Set(*m.value)
					*m.value = 0
				}
				r.mux.Unlock()
			}
		}
	}()
	return r
}

func (r *Reporter) Success(d time.Duration) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.successCount++
	r.successMillis += float64(d.Milliseconds())
}

func (r *Reporter) Overload(d time.Duration) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.overloadCount++
	r.overloadMillis += float64(d.Milliseconds())
}

func (r *Reporter) Timeout(d time.Duration) {
	r.mux.Lock()
	defer r.mux.Unlock()
	r.timeoutCount++
	r.timeoutMillis += float64(d.Milliseconds())
}

func (r *Reporter) Failure(d time.Duration) {
	r.mux.Lock()
	defer r.mux.Lock()
	r.failCount++
	r.failMillis += float64(d.Milliseconds())
}

func reportQps(service, pod string, qps float64) {
	svc := "http://metrics.cascade-example.svc.cluster.local"
	url := svc + fmt.Sprintf("/service/%s/pod/%s/qps/%v", service, pod, int(qps))
	client := &http.Client{
		Timeout: time.Second * 10,
	}
	resp, err := client.Get(url)
	if err != nil {
		log.Printf("error reporting qps: %v", err.Error())
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		msg, _ := ioutil.ReadAll(resp.Body)
		log.Printf("non ok status reporting qps: %v (%v)", string(msg), resp.StatusCode)
	}
}
