package metrics

import (
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

	successMillis  float64
	overloadMillis float64
	timeoutMillis  float64
}

func NewReporter(service string) *Reporter {
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
