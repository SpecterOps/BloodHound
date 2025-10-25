// Copyright 2025 Specter Ops, Inc.
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

package graphify

import (
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/dawgs/graph"
)

var (
	// ingestThroughputGauge is registered withthe metrics daemon's Prometheus registry at startup
	ingestThroughputGauge *prometheus.GaugeVec
)

// InitializeIngestMetrics registers the ingestion throughput gauge with the Prometheus registry
func InitializeIngestMetrics(registerer prometheus.Registerer) error {
	ingestThroughputGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "bhe_ingest_throughput",
			Help: "Ingestion throughput in entities per second",
		},
		[]string{"entity_type", "stage"}, // "nodes" or "relationships", "processed" or "written"
	)

	return registerer.Register(ingestThroughputGauge)
}

// PublishIngestThroughput publishes ingestion throughput metrics to Prometheus
func PublishIngestThroughput(nodesProcessed, relsProcessed, nodesWritten, relsWritten int64, duration time.Duration) {
	if ingestThroughputGauge == nil || duration.Seconds() <= 0 {
		return
	}

	// Calculate throughput rates
	nodesProcessedPerSec := float64(nodesProcessed) / duration.Seconds()
	relsProcessedPerSec := float64(relsProcessed) / duration.Seconds()
	nodesWrittenPerSec := float64(nodesWritten) / duration.Seconds()
	relsWrittenPerSec := float64(relsWritten) / duration.Seconds()

	// Update gauges immediately
	ingestThroughputGauge.WithLabelValues("nodes", "processed").Set(nodesProcessedPerSec)
	ingestThroughputGauge.WithLabelValues("relationships", "processed").Set(relsProcessedPerSec)
	ingestThroughputGauge.WithLabelValues("nodes", "written").Set(nodesWrittenPerSec)
	ingestThroughputGauge.WithLabelValues("relationships", "written").Set(relsWrittenPerSec)
}

// IngestStats tracks the number of nodes and relationships processed during ingestion
type IngestStats struct {
	// Total entities processed (including deduplicated ones).
	// Consider this the number of elements present in the raw ingest payload.
	NodesProcessed         atomic.Int64
	RelationshipsProcessed atomic.Int64

	// Entities actually written to database (subset of processed)
	NodesWritten         atomic.Int64
	RelationshipsWritten atomic.Int64
}

func (s *IngestStats) Reset() {
	s.NodesProcessed.Store(0)
	s.RelationshipsProcessed.Store(0)
	s.NodesWritten.Store(0)
	s.RelationshipsWritten.Store(0)
}

func (s *IngestStats) GetCounts() (nodesProcessed, relsProcessed, nodesWritten, relsWritten int64) {
	return s.NodesProcessed.Load(), s.RelationshipsProcessed.Load(),
		s.NodesWritten.Load(), s.RelationshipsWritten.Load()
}

// countingBatchUpdater wraps a BatchUpdater and counts all node and relationship operations
type countingBatchUpdater struct {
	inner BatchUpdater
	stats *IngestStats
}

// NewCountingBatchUpdater creates a BatchUpdater wrapper that tracks operation counts
func NewCountingBatchUpdater(inner BatchUpdater, stats *IngestStats) BatchUpdater {
	return &countingBatchUpdater{
		inner: inner,
		stats: stats,
	}
}

func (s *countingBatchUpdater) UpdateNodeBy(update graph.NodeUpdate) error {
	if err := s.inner.UpdateNodeBy(update); err != nil {
		return err
	}
	// Track database writes
	s.stats.NodesWritten.Add(1)
	return nil
}

func (s *countingBatchUpdater) UpdateRelationshipBy(update graph.RelationshipUpdate) error {
	if err := s.inner.UpdateRelationshipBy(update); err != nil {
		return err
	}
	// Track database writes
	s.stats.RelationshipsWritten.Add(1)
	return nil
}

func (s *countingBatchUpdater) Nodes() graph.NodeQuery {
	return s.inner.Nodes()
}

func (s *countingBatchUpdater) Relationships() graph.RelationshipQuery {
	return s.inner.Relationships()
}
