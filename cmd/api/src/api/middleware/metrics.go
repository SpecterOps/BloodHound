// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
	namespace = "bh"
	// ApiInFlightGauge is label-free: in-flight counts are only meaningful
	// globally, and labels would inflate cardinality.
	ApiInFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: namespace,
		Subsystem: "api",
		Name:      "in_flight_requests",
		Help:      "A gauge of requests currently being served by the wrapped handler.",
	})

	// ApiTotalRequests is partitioned by response code and HTTP method (both
	// bounded). Do not add a path/URI label; use a templated "handler" label
	// as described above if a per-endpoint breakdown is required.
	ApiTotalRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "requests_total",
			Help:      "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	// ApiRequestDuration is partitioned by response code, HTTP method, and a
	// "handler" label populated from the matched mux route template. The
	// "handler" value must be curried via MustCurryWith (see guidance above)
	// before observation so concrete request paths never reach Prometheus.
	ApiRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "request_duration_seconds",
			Help:      "A histogram of latencies for requests.",
			Buckets:   []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"code", "method", "handler"},
	)

	// ApiResponseSize is a zero-dimensional ObserverVec. If ever extended,
	// follow the same "handler"-label rules rather than adding a path label.
	ApiResponseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: "api",
			Name:      "response_size_bytes",
			Help:      "A histogram of response sizes for requests.",
			Buckets:   []float64{200, 500, 900, 1500},
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
