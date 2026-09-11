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
package graphify

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

var (
	ingestThroughputGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: model.Namespace,
			Name:      "ingest_throughput",
			Help:      "Ingestion throughput in entities per second",
		},
		[]string{"entity_type", "stage"}, // "nodes" or "relationships", "processed" or "written"
	)
)

// RegisterIngestMetrics registers the ingestion throughput gauge with the provided Prometheus registerer.
func RegisterIngestMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(ingestThroughputGauge); err != nil {
		return fmt.Errorf("failed to register ingest throughput gauge: %w", err)
	}

	return nil
}
