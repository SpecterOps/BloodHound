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

package measure

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"
)

const (
	FieldElapsed       = "elapsed"
	FieldMeasurementID = "measurement_id"
)

var (
	logMeasurePairCounter = atomic.Uint64{}
	measureThreshold      = time.Second
)

// ContextMeasure will measure the duration a function takes under a given context and generate the associated log message at a given slog.Level.
// This function should be used to output critical application behavior which must be logged for troubleshooting purposes.
// If your log message may be noisy or should only output when taking an extended duration, use `ContextMeasureWithThreshold` instead.
func ContextMeasure(ctx context.Context, level slog.Level, msg string, args ...slog.Attr) func() {
	then := time.Now()

	return func() {
		elapsed := time.Since(then)
		args = append(args, slog.Duration(FieldElapsed, elapsed))
		slog.LogAttrs(ctx, level, msg, args...)
	}
}

// ContextMeasureWithThreshold will measure the duration a function takes under a given context and generate the associated log message at a given slog.Level.
// This function will only output the specified log message if the function requires >1 second to complete.
// If your log message indicates critical application behavior and must be logged, use `ContextMeasure` instead.
func ContextMeasureWithThreshold(ctx context.Context, level slog.Level, msg string, args ...slog.Attr) func() {
	then := time.Now()

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, slog.Duration(FieldElapsed, elapsed))
			slog.LogAttrs(ctx, level, msg, args...)
		}
	}
}

// Measure will measure the duration a function takes and generate the associated log message at a given slog.Level.
// This function should be used to output critical application behavior which must be logged for troubleshooting purposes.
// If your log message may be noisy or should only output when taking an extended duration, use `MeasureWithThreshold` instead.
func Measure(level slog.Level, msg string, args ...slog.Attr) func() {
	then := time.Now()

	return func() {
		elapsed := time.Since(then)
		args = append(args, slog.Duration(FieldElapsed, elapsed))
		slog.LogAttrs(context.TODO(), level, msg, args...)
	}
}

// MeasureWithThreshold will measure the duration a function takes and generate the associated log message at a given slog.Level.
// This function will only output the specified log message if the function requires >1 second to complete.
// If your log message indicates critical application behavior and must be logged, use `Measure` instead.
func MeasureWithThreshold(level slog.Level, msg string, args ...slog.Attr) func() {
	then := time.Now()

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, slog.Duration(FieldElapsed, elapsed))
			slog.LogAttrs(context.TODO(), level, msg, args...)
		}
	}
}

// ContextLogAndMeasure will measure the duration an individual function call takes under a given context and generate the associated log message at a given slog.Level.
// This function should be used to output critical application behavior which must be logged for troubleshooting purposes.
// If your log message may be noisy or should only output when taking an extended duration, use `ContextLogAndMeasureWithThreshold` instead.
func ContextLogAndMeasure(ctx context.Context, level slog.Level, msg string, args ...slog.Attr) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, slog.Uint64(FieldMeasurementID, pairID))
	slog.LogAttrs(ctx, level, msg, args...)

	return func() {
		elapsed := time.Since(then)
		args = append(args, slog.Duration(FieldElapsed, elapsed))
		slog.LogAttrs(ctx, level, msg, args...)
	}
}

// ContextLogAndMeasureWithThreshold will measure the duration an individual function call takes under a given context and generate the associated log message at a given slog.Level.
// This function always logs the initial message; the elapsed log is emitted only if the function requires >1 second to complete.
// If your log message indicates critical application behavior and must be logged, use `measure.ContextLogAndMeasure` instead.
func ContextLogAndMeasureWithThreshold(ctx context.Context, level slog.Level, msg string, args ...slog.Attr) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, slog.Uint64(FieldMeasurementID, pairID))
	slog.LogAttrs(ctx, level, msg, args...)

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, slog.Duration(FieldElapsed, elapsed))
			slog.LogAttrs(ctx, level, msg, args...)
		}
	}
}

// LogAndMeasure will measure the duration an individual function call takes and generate the associated log message at a given slog.Level.
// This function should be used to output critical application behavior which must be logged for troubleshooting purposes.
// If your log message may be noisy or should only output when taking an extended duration, use `LogAndMeasureWithThreshold` instead.
func LogAndMeasure(level slog.Level, msg string, args ...slog.Attr) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, slog.Uint64(FieldMeasurementID, pairID))
	slog.LogAttrs(context.TODO(), level, msg, args...)

	return func() {
		elapsed := time.Since(then)
		args = append(args, slog.Duration(FieldElapsed, elapsed))
		slog.LogAttrs(context.TODO(), level, msg, args...)
	}
}

// LogAndMeasureWithThreshold will measure the duration an individual function call takes and generate the associated log message at a given slog.Level.
// This function always logs the initial message; the elapsed log is emitted only if the function requires >1 second to complete.
// If your log message indicates critical application behavior and must be logged, use `LogAndMeasure` instead.
func LogAndMeasureWithThreshold(level slog.Level, msg string, args ...slog.Attr) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, slog.Uint64(FieldMeasurementID, pairID))
	slog.LogAttrs(context.TODO(), level, msg, args...)

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, slog.Duration(FieldElapsed, elapsed))
			slog.LogAttrs(context.TODO(), level, msg, args...)
		}
	}
}
