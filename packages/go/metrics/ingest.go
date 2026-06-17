// Copyright 2026 Specter Ops, Inc.
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

// Package metrics declares cross-package BHCE lifecycle Prometheus metrics that
// are instrumented across multiple packages. Metrics are exposed through recording
// functions that accept typed label enums.
package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

const ingestSubsystem = "ingest"

// IngestSource represents the origin of an ingest task in the pipeline.
// This typed enum is used for Prometheus label values to ensure compile-time
// validation and prevent typos that would create stray time series.
type IngestSource string

const (
	// IngestSourceFile represents tasks created from manual file uploads via the web UI or API.
	IngestSourceFile IngestSource = "file"

	// IngestSourceClient represents tasks created from SharpHound, AzureHound, or OpenHound client ingests.
	IngestSourceClient IngestSource = "client"
)

var (
	// ingestTasksCreated tracks the total number of ingest tasks created and
	// persisted to the database. This counter is used for volume analytics and
	// trend analysis (e.g., "How many tasks were created this week?", "Is usage growing?").
	//
	// For operational queue monitoring, use:
	//   - bhe_ingest_tasks (gauge: current queue depth)
	//   - bh_ingest_task_queue_latency_seconds_count (summary: tasks processed)
	ingestTasksCreated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: model.Namespace,
			Subsystem: ingestSubsystem,
			Name:      "tasks_created_total",
			Help:      "Total number of ingest tasks created and saved to the database (for volume/trend analysis)",
		},
		[]string{"source"}, // "file" for manual file upload, "client" for client ingest
	)

	// ingestTaskQueueLatency tracks the time an ingest task spends waiting in the
	// database queue from creation (DB write) until it is picked up for processing.
	// This measures queue wait time, not processing time.
	//
	// The _count value represents tasks that have been PROCESSED (picked up from queue) since last restart.
	// For current queue depth, use the bhe_ingest_tasks gauge instead.
	ingestTaskQueueLatency = prometheus.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  model.Namespace,
			Subsystem:  ingestSubsystem,
			Name:       "task_queue_latency_seconds",
			Help:       "Queue wait time: duration from when an ingest task is saved to disk until it is picked up for processing",
			Objectives: map[float64]float64{0.5: 0.05, 0.9: 0.01, 0.95: 0.005, 0.99: 0.001},
		},
		[]string{"source"}, // "file" for manual file upload, "client" for client ingest
	)
)

// RecordIngestTaskCreated increments the counter for ingest tasks created and saved to disk.
// This metric is used for volume tracking and trend analysis.
// The source parameter indicates the ingest source type (file upload or client ingest).
func RecordIngestTaskCreated(source IngestSource) {
	ingestTasksCreated.WithLabelValues(string(source)).Inc()
}

// RecordIngestTaskQueueLatency records the queue wait time from when an ingest task was saved to disk
// until it was picked up for processing. This measures time waiting in queue, not processing time.
// The source parameter indicates the ingest source type (file upload or client ingest).
func RecordIngestTaskQueueLatency(taskCreatedAt time.Time, source IngestSource) {
	ingestTaskQueueLatency.WithLabelValues(string(source)).Observe(time.Since(taskCreatedAt).Seconds())
}

// RegisterIngestMetrics registers all ingest-subsystem Prometheus metrics with the provided registerer.
func RegisterIngestMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(ingestTasksCreated); err != nil {
		return fmt.Errorf("failed to register ingest tasks created counter: %w", err)
	}

	if err := registerer.Register(ingestTaskQueueLatency); err != nil {
		return fmt.Errorf("failed to register ingest task queue latency summary: %w", err)
	}

	return nil
}
