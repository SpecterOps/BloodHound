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

	registry := prometheus.NewRegistry()
	require.NoError(t, RegisterAnalysisMetrics(registry))

	RecordAnalysisStepDuration(analysisPipelineStep{name: "unnamed"}.String(), 5*time.Second)
	RecordAnalysisStepDuration(analysisPipelineStep{analysisStep: model.AnalysisStepTagging()}.String(), 8*time.Minute)
	RecordAnalysisStepDuration(analysisPipelineStep{analysisStep: model.AnalysisStepGenerateFindings()}.String(), 1*time.Hour+12*time.Minute)

	const metric = "bh_analysis_step_duration_seconds"
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
	`
	// verify metric name because GatherAndCompare will not error if the filtered metricName is mismatched
	require.Contains(t, expected, metric, "mismatch between metric and expected")
	require.NoError(t, testutil.GatherAndCompare(registry, strings.NewReader(expected), metric))
}
