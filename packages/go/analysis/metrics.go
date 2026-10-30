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
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

const analysisSubsystem = "analysis"

var (
	// pzNodeTagCounterVec counts tag_added and tag_removed events that occur during AGT analysis.
	// The "position" label is the tier position for tiers (e.g. "1", "2") and the tag type for
	// labels and owned tags ("label", "owned").
	pzNodeTagCounterVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: model.Namespace,
		Subsystem: analysisSubsystem,
		Name:      "pzm_node_tags_total",
		Help:      "Total number of privilege zone tag additions and removals applied to graph nodes during analysis.",
	}, []string{
		"action",
		"position",
	})

	analysisStepDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: model.Namespace,
		Subsystem: analysisSubsystem,
		Name:      "step_duration_seconds",
		Help:      "How long each analysis step takes to complete",
		Buckets: []float64{
			(30 * time.Second).Seconds(),
			(1 * time.Minute).Seconds(),
			(5 * time.Minute).Seconds(),
			(10 * time.Minute).Seconds(),
			(30 * time.Minute).Seconds(),
			(1 * time.Hour).Seconds(),
			(2 * time.Hour).Seconds(),
			(4 * time.Hour).Seconds(),
			(8 * time.Hour).Seconds(),
			(12 * time.Hour).Seconds(),
			(24 * time.Hour).Seconds(),
		},
	}, []string{"step"})
)

func RecordAnalysisStepDuration(step string, duration time.Duration) {
	analysisStepDuration.WithLabelValues(step).Observe(duration.Seconds())
}

// tagToPosition returns the value used for the "position" label on pzNodeTagCounterVec.
// Tiers use their numeric position (e.g. "1", "2"). Labels and owned tags are grouped
// under their type name since they have no meaningful position.
func tagToPosition(tag model.AssetGroupTag) string {
	switch tag.Type {
	case model.AssetGroupTagTypeTier:
		if !tag.Position.Valid {
			return "unknown"
		}
		return strconv.Itoa(int(tag.Position.Int32))
	case model.AssetGroupTagTypeLabel:
		return "label"
	case model.AssetGroupTagTypeOwned:
		return "owned"
	default:
		return "unknown"
	}
}

// RegisterAnalysisMetrics registers all analysis-subsystem Prometheus metrics
// with the provided registerer. Additional analysis metrics should be added here.
func RegisterAnalysisMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(pzNodeTagCounterVec); err != nil {
		return fmt.Errorf("failed to register analysis metrics: %w", err)
	}

	if err := registerer.Register(analysisStepDuration); err != nil {
		return fmt.Errorf("failed to register analysis step duration metric: %w", err)
	}

	return nil
}
