package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"strings"
)

var (
	reg           = prometheus.NewPedanticRegistry()
	collectorList = []prometheus.Collector{}

	LastReq      prometheus.Gauge
	CountError   prometheus.Counter
	CountRequest prometheus.Counter
	ResponseTime prometheus.Histogram
)

func AddBasicCollector(prefix string) {
	prefix = strings.ReplaceAll(prefix, "-", "_")
	LastReq = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: prefix + "_last_request",
		Help: "The time of last income request",
	})

	CountError = prometheus.NewCounter(prometheus.CounterOpts{
		Name: prefix + "_error_count",
		Help: "The total number of request errors",
	})

	CountRequest = prometheus.NewCounter(prometheus.CounterOpts{
		Name: prefix + "_request_count",
		Help: "The total request count",
	})

	ResponseTime = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    prefix + "_response_time",
		Help:    "Response time in ms",
		Buckets: []float64{1, 10, 100, 1000, 10000},
	})

	collectorList = append(collectorList,
		prometheus.NewGoCollector(),
		LastReq,
		CountError,
		CountRequest,
		ResponseTime,
	)
}

func AddCollector(collector ...prometheus.Collector) {
	collectorList = append(collectorList, collector...)
}

// Metrics prometheus metrics
func Metrics() http.Handler {
	reg.MustRegister(collectorList...)
	return promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
}
