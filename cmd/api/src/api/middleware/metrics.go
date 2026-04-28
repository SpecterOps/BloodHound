package middleware

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Label cardinality guidance for the Prometheus metrics below.
//
// Every label must have a bounded set of values. Never use raw request paths
// (r.URL.Path, r.RequestURI) or client-supplied input as label values: paths
// embed unique IDs (UUIDs, user names, search terms) and produce unbounded
// cardinality.
//
// Use the matched gorilla/mux route template as the "handler" label, e.g.
// mux.CurrentRoute(r).GetPathTemplate() (falling back to "unmatched"), then
// curry it via MustCurryWith(prometheus.Labels{"handler": template}) before
// calling the promhttp.InstrumentHandler* helpers.
var (
	// ApiInFlightGauge is label-free: in-flight counts are only meaningful
	// globally, and labels would inflate cardinality.
	ApiInFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	// ApiTotalRequests is partitioned by response code and HTTP method (both
	// bounded). Do not add a path/URI label; use a templated "handler" label
	// as described above if a per-endpoint breakdown is required.
	ApiTotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	// ApiRequestDuration is partitioned by response code, HTTP method, and a
	// "handler" label populated from the matched mux route template. The
	// "handler" value must be curried via MustCurryWith (see guidance above)
	// before observation so concrete request paths never reach Prometheus.
	ApiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"code", "method", "handler"},
	)

	// ApiResponseSize is a zero-dimensional ObserverVec. If ever extended,
	// follow the same "handler"-label rules rather than adding a path label.
	ApiResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{200, 500, 900, 1500},
		},
		[]string{},
	)
)

// RegisterApiMiddlewareMetrics registers Prometheus metrics used by API middleware
// with the provided registry and returns an error if any registration fails.
func RegisterApiMiddlewareMetrics(registry prometheus.Registerer) error {
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
