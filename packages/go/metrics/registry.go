// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package metrics

import (
	"slices"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// registry provides a thread-safe wrapper around Prometheus metrics registration.
type registry struct {
	lock               *sync.Mutex
	prometheusRegistry *prometheus.Registry
	counters           map[string]prometheus.Counter
	counterVecs        map[string]*prometheus.CounterVec
	gauges             map[string]prometheus.Gauge
}

// metricKey generates a unique key for a metric based on name, namespace, and labels.
// It sorts label keys alphabetically to ensure consistent ordering across invocations.
func metricKey(name, namespace string, labels map[string]string) string {
	builder := strings.Builder{}

	builder.WriteString(namespace)
	builder.WriteRune('|')
	builder.WriteString(name)
	builder.WriteRune('|')

	sortedLabelKeys := make([]string, 0, len(labels))

	for labelKey := range labels {
		sortedLabelKeys = append(sortedLabelKeys, labelKey)
	}

	slices.Sort(sortedLabelKeys)

	for idx, key := range sortedLabelKeys {
		if idx > 0 {
			builder.WriteRune('|')
		}

		builder.WriteString(key)
		builder.WriteString(labels[key])
	}

	return builder.String()
}

// metricVecKey extends metricKey with support for variable label names used in CounterVec.
// It appends "vec" at the end of the generated key to distinguish it from regular metrics.
func metricVecKey(name, namespace string, labels map[string]string, variableLabels []string) string {
	builder := strings.Builder{}
	builder.WriteString(metricKey(name, namespace, labels))
	builder.WriteRune('|')

	sortedVariableLabels := slices.Clone(variableLabels)
	slices.Sort(sortedVariableLabels)

	for idx, label := range sortedVariableLabels {
		if idx > 0 {
			builder.WriteRune('|')
		}

		builder.WriteString(label)
	}

	builder.WriteString("vec")
	return builder.String()
}

// Counter retrieves or creates a counter metric with the given name, namespace, and constant labels.
// If a matching counter already exists, it returns the existing one; otherwise, it registers a new one.
func (s *registry) Counter(name, namespace string, constLabels map[string]string) prometheus.Counter {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := metricKey(name, namespace, constLabels)

	if counter, hasCounter := s.counters[key]; hasCounter {
		return counter
	} else {
		newCounter := promauto.With(s.prometheusRegistry).NewCounter(prometheus.CounterOpts{
			Name:        name,
			Namespace:   namespace,
			ConstLabels: constLabels,
		})

		s.counters[key] = newCounter
		newCounter.Add(0)

		return newCounter
	}
}

// CounterVec retrieves or creates a vectorized counter metric with the given name, namespace, constant labels, and variable label names.
// It uses metricVecKey to generate the unique identifier for the vector.
func (s *registry) CounterVec(name, namespace string, constLabels map[string]string, variableLabelNames []string) *prometheus.CounterVec {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := metricVecKey(name, namespace, constLabels, variableLabelNames)

	if counterVec, hasCounter := s.counterVecs[key]; hasCounter {
		return counterVec
	} else {
		newCounterVec := promauto.With(s.prometheusRegistry).NewCounterVec(prometheus.CounterOpts{
			Name:        name,
			Namespace:   namespace,
			ConstLabels: constLabels,
		}, variableLabelNames)

		s.counterVecs[key] = newCounterVec
		return newCounterVec
	}
}

// Gauge retrieves or creates a gauge metric with the given name, namespace, and constant labels.
// If a matching gauge already exists, it returns the existing one; otherwise, it registers a new one.
func (s *registry) Gauge(name, namespace string, constLabels map[string]string) prometheus.Gauge {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := metricKey(name, namespace, constLabels)

	if gauge, hasGauge := s.gauges[key]; hasGauge {
		return gauge
	} else {
		newGauge := promauto.With(s.prometheusRegistry).NewGauge(prometheus.GaugeOpts{
			Name:        name,
			Namespace:   namespace,
			ConstLabels: constLabels,
		})

		s.gauges[key] = newGauge
		newGauge.Set(0)

		return newGauge
	}
}

var (
	globalRegistry *registry // Global singleton registry instance.
)

// init initializes the global registry with default collectors for Go and process statistics.
func init() {
	prometheusRegistry := prometheus.NewRegistry()

	// Default collectors for Golang and process stats. This will panic on failure to register.
	prometheusRegistry.MustRegister(
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
	)

	globalRegistry = &registry{
		lock:               &sync.Mutex{},
		prometheusRegistry: prometheusRegistry,
		counters:           map[string]prometheus.Counter{},
		counterVecs:        map[string]*prometheus.CounterVec{},
		gauges:             map[string]prometheus.Gauge{},
	}
}

// Counter retrieves or creates a counter metric.
func Counter(name, namespace string, labels map[string]string) prometheus.Counter {
	return globalRegistry.Counter(name, namespace, labels)
}

// CounterVec retrieves or creates a counter vec.
func CounterVec(name, namespace string, labels map[string]string, variableLabelNames []string) *prometheus.CounterVec {
	return globalRegistry.CounterVec(name, namespace, labels, variableLabelNames)
}

// Gauge retrieves or creates a gauge metric.
func Gauge(name, namespace string, labels map[string]string) prometheus.Gauge {
	return globalRegistry.Gauge(name, namespace, labels)
}

// Registerer returns the underlying Prometheus registry instance for direct access.
func Registerer() *prometheus.Registry {
	return globalRegistry.prometheusRegistry
}

// Register adds a collector to the global registry.
func Register(collector prometheus.Collector) error {
	return Registerer().Register(collector)
}

// Unregister removes a collector from the global registry.
func Unregister(collector prometheus.Collector) {
	Registerer().Unregister(collector)
}
