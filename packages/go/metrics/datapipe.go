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

package metrics

import (
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

const datapipeSubsystem = "datapipe"

// OptimizeStoragePipelineStage is the typed label value enum for the
// optimize-storage duration gauge.
type OptimizeStoragePipelineStage string

const (
	OptimizeStoragePipelineStageBoot     OptimizeStoragePipelineStage = "boot"
	OptimizeStoragePipelineStageAnalysis OptimizeStoragePipelineStage = "analysis"
)

var (
	datapipeStatus = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: model.Namespace,
			Subsystem: datapipeSubsystem,
			Name:      "status",
			Help:      "Current datapipe status.",
		},
		[]string{"status"},
	)

	optimizeStorageDuration = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: model.Namespace,
		Subsystem: datapipeSubsystem,
		Name:      "optimize_duration",
		Help:      "Duration in seconds of the most recent OptimizeStorage call, labeled by pipeline stage (boot, analysis).",
	}, []string{"pipeline_stage"})

	optimizeStoragePipelineStages = []OptimizeStoragePipelineStage{
		OptimizeStoragePipelineStageBoot,
		OptimizeStoragePipelineStageAnalysis,
	}
)

func RecordDatapipeStatus(status model.DatapipeStatus) {
	for _, knownStatus := range model.AllDatapipeStatuses() {
		value := 0.0
		if knownStatus == status {
			value = 1.0
		}

		datapipeStatus.WithLabelValues(string(knownStatus)).Set(value)
	}
}

// RegisterDatapipeMetrics registers the datapipe Prometheus metrics with the
// provided registerer.
func RegisterDatapipeMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(datapipeStatus); err != nil {
		return fmt.Errorf("failed to register datapipe status gauge: %w", err)
	}

	return nil
}

// RecordOptimizeStorageDuration records the duration of an OptimizeStorage
// call for the given pipeline stage.
func RecordOptimizeStorageDuration(stage OptimizeStoragePipelineStage, duration time.Duration) {
	optimizeStorageDuration.WithLabelValues(string(stage)).Set(duration.Seconds())
}

func RegisterOptimizeStorageMetrics(registerer prometheus.Registerer) error {
	if err := registerer.Register(optimizeStorageDuration); err != nil {
		return fmt.Errorf("failed to register datapipe optimize storage duration gauge: %w", err)
	}

	// Initialize each known stage series so the metric is present at /metrics
	// before the first OptimizeStorage call.
	for _, stage := range optimizeStoragePipelineStages {
		optimizeStorageDuration.WithLabelValues(string(stage))
	}

	return nil
}
