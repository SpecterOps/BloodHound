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

package measure_test

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/stretchr/testify/require"
)

func captureLogs(t *testing.T) *bytes.Buffer {
	t.Helper()

	var (
		buffer         = &bytes.Buffer{}
		previousLogger = slog.Default()
	)

	slog.SetDefault(slog.New(slog.NewJSONHandler(buffer, &slog.HandlerOptions{Level: slog.LevelDebug})))
	t.Cleanup(func() {
		slog.SetDefault(previousLogger)
	})

	return buffer
}

func decodeLogLines(t *testing.T, buffer *bytes.Buffer) []map[string]any {
	t.Helper()

	var logLines []map[string]any

	for _, rawLine := range strings.Split(strings.TrimSpace(buffer.String()), "\n") {
		if rawLine == "" {
			continue
		}

		var logLine map[string]any
		require.NoError(t, json.Unmarshal([]byte(rawLine), &logLine))
		logLines = append(logLines, logLine)
	}

	return logLines
}

func TestContextLogAndMeasureWithResults(t *testing.T) {
	var buffer = captureLogs(t)

	measureDone := measure.ContextLogAndMeasureWithResults(context.Background(), slog.LevelInfo, "measured operation", slog.String("input", "value"))
	measureDone(slog.Int("num_results", 42))

	logLines := decodeLogLines(t, buffer)
	require.Len(t, logLines, 2)

	startLine, completionLine := logLines[0], logLines[1]

	require.Equal(t, "measured operation", startLine["msg"])
	require.Equal(t, "value", startLine["input"])
	require.NotContains(t, startLine, "num_results")
	require.NotContains(t, startLine, measure.FieldElapsed)

	require.Equal(t, "measured operation", completionLine["msg"])
	require.Equal(t, "value", completionLine["input"])
	require.EqualValues(t, 42, completionLine["num_results"])
	require.Contains(t, completionLine, measure.FieldElapsed)
	require.Equal(t, startLine[measure.FieldMeasurementID], completionLine[measure.FieldMeasurementID])
}

func TestContextMeasureWithThresholdAndResults_BelowThreshold(t *testing.T) {
	var buffer = captureLogs(t)

	measureDone := measure.ContextMeasureWithThresholdAndResults(context.Background(), slog.LevelInfo, "fast operation")
	measureDone(slog.Int("num_results", 1))

	require.Empty(t, decodeLogLines(t, buffer))
}
