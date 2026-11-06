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
package analysis

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestRecordAnalysisStepDuration(t *testing.T) {
	analysisStepDuration.Reset()
	analysisStepLastDuration.Reset()

	registry := prometheus.NewRegistry()
	require.NoError(t, RegisterAnalysisMetrics(registry))

	RecordAnalysisStepDuration(analysisPipelineStep{name: "unnamed"}.String(), 5*time.Second)
	RecordAnalysisStepDuration(analysisPipelineStep{analysisStep: model.AnalysisStepTagging()}.String(), 8*time.Minute)
	RecordAnalysisStepDuration(analysisPipelineStep{analysisStep: model.AnalysisStepGenerateFindings()}.String(), 1*time.Hour+12*time.Minute)

	RecordAnalysisStepDuration(analysisPipelineStep{name: "twice"}.String(), 11*time.Second)
	RecordAnalysisStepDuration(analysisPipelineStep{name: "twice"}.String(), 45*time.Second)

	const metric1 = "bh_analysis_step_duration_seconds"
	const metric2 = "bh_analysis_step_last_duration_seconds"
	const expected = `
		# HELP bh_analysis_step_duration_seconds How long each analysis step takes to complete
		# TYPE bh_analysis_step_duration_seconds histogram
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="30"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="60"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="300"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="600"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="1800"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="3600"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="7200"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="14400"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="28800"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="43200"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="86400"} 1
		bh_analysis_step_duration_seconds_bucket{step="unnamed",le="+Inf"} 1
		bh_analysis_step_duration_seconds_sum{step="unnamed"} 5
		bh_analysis_step_duration_seconds_count{step="unnamed"} 1

		bh_analysis_step_duration_seconds_bucket{step="tagging",le="30"} 0
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="60"} 0
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="300"} 0
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="600"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="1800"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="3600"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="7200"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="14400"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="28800"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="43200"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="86400"} 1
		bh_analysis_step_duration_seconds_bucket{step="tagging",le="+Inf"} 1
		bh_analysis_step_duration_seconds_sum{step="tagging"} 480
		bh_analysis_step_duration_seconds_count{step="tagging"} 1

		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="30"} 0
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="60"} 0
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="300"} 0
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="600"} 0
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="1800"} 0
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="3600"} 0
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="7200"} 1
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="14400"} 1
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="28800"} 1
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="43200"} 1
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="86400"} 1
		bh_analysis_step_duration_seconds_bucket{step="generate_findings",le="+Inf"} 1
		bh_analysis_step_duration_seconds_sum{step="generate_findings"} 4320
		bh_analysis_step_duration_seconds_count{step="generate_findings"} 1

		bh_analysis_step_duration_seconds_bucket{step="twice",le="30"} 1
		bh_analysis_step_duration_seconds_bucket{step="twice",le="60"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="300"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="600"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="1800"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="3600"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="7200"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="14400"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="28800"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="43200"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="86400"} 2
		bh_analysis_step_duration_seconds_bucket{step="twice",le="+Inf"} 2
		bh_analysis_step_duration_seconds_sum{step="twice"} 56
		bh_analysis_step_duration_seconds_count{step="twice"} 2

		# HELP bh_analysis_step_last_duration_seconds How long the most recent analysis step run took to complete
		# TYPE bh_analysis_step_last_duration_seconds gauge
		bh_analysis_step_last_duration_seconds{step="unnamed"} 5
		bh_analysis_step_last_duration_seconds{step="tagging"} 480
		bh_analysis_step_last_duration_seconds{step="generate_findings"} 4320
		bh_analysis_step_last_duration_seconds{step="twice"} 45
	`
	// verify metric name because GatherAndCompare will not error if the filtered metricName is mismatched
	require.Contains(t, expected, metric1, "mismatch between metric and expected")
	// require.Contains(t, expected, metric2, "mismatch between metric and expected")
	require.NoError(t, testutil.GatherAndCompare(registry, strings.NewReader(expected), metric1, metric2))
}
