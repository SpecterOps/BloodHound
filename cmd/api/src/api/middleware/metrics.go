package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ApiInFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	ApiTotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "api_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	// duration is partitioned by the HTTP method and handler. It uses custom
	// buckets based on the expected request duration.
	ApiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"handler", "method"},
	)

	// apiResponseSize has no labels, making it a zero-dimensional
	// ObserverVec.
	ApiResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{},
	)
)

func RegisterApiMiddlewareMetrics(registry *prometheus.Registry) error {
	if err := registry.Register(ApiInFlightGauge); err != nil {
		return err
	} else if err = registry.Register(ApiTotalRequests); err != nil {
		return err
	} else if err = registry.Register(ApiRequestDuration); err != nil {
		return err
	} else if err = registry.Register(ApiResponseSize); err != nil {
		return err
	}
	return nil
}
