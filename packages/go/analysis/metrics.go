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

package analysis

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

var (
	// pzNodeTagCounterVec counts tag_added and tag_removed events that occur during AGT analysis.
	pzNodeTagCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: model.Namespace,
		Subsystem: "analysis",
		Name:      "pzm_node_tags_total",
		Help:      "Total number of privilege zone tag additions and removals applied to graph nodes during analysis.",
	}, []string{
		"action",
	})
)

// RegisterAnalysisMetrics registers all analysis-subsystem Prometheus metrics
// with the provided registerer. Additional analysis metrics should be added here.
func RegisterAnalysisMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(pzNodeTagCounterVec); err != nil {
		return fmt.Errorf("failed to register analysis metrics: %w", err)
	}

	return nil
}
