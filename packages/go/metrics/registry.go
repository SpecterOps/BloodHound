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
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type registry struct {
	lock               *sync.Mutex
	prometheusRegistry *prometheus.Registry
	counters           map[string]prometheus.Counter
	counterVecs        map[string]*prometheus.CounterVec
	gauges             map[string]prometheus.Gauge
}

func metricKey(name, namespace string, labels map[string]string) string {
	builder := strings.Builder{}

	builder.WriteString(namespace)
	builder.WriteString(name)

	for key, value := range labels {
		builder.WriteString(key)
		builder.WriteString(value)
	}

	return builder.String()
}

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

func (s *registry) CounterVec(name, namespace string, constLabels map[string]string, variableLabelNames []string) *prometheus.CounterVec {
	s.lock.Lock()
	defer s.lock.Unlock()

	key := metricKey(name, namespace, constLabels)

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
	globalRegistry *registry
)

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

func Counter(name, namespace string, labels map[string]string) prometheus.Counter {
	return globalRegistry.Counter(name, namespace, labels)
}

func CounterVec(name, namespace string, labels map[string]string, variableLabelNames []string) *prometheus.CounterVec {
	return globalRegistry.CounterVec(name, namespace, labels, variableLabelNames)
}

func Gauge(name, namespace string, labels map[string]string) prometheus.Gauge {
	return globalRegistry.Gauge(name, namespace, labels)
}

func Registerer() *prometheus.Registry {
	return globalRegistry.prometheusRegistry
}

func NewCounter(opts prometheus.CounterOpts) prometheus.Counter {
	return promauto.With(Registerer()).NewCounter(opts)
}

func NewGauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	return promauto.With(Registerer()).NewGauge(opts)
}

func Register(collector prometheus.Collector) error {
	return Registerer().Register(collector)
}

func Unregister(collector prometheus.Collector) {
	Registerer().Unregister(collector)
}
