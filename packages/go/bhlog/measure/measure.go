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

func ContextMeasure(ctx context.Context, level slog.Level, msg string, args ...any) func() {
	then := time.Now()

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(ctx, level, msg, args...)
		}
	}
}

func Measure(level slog.Level, msg string, args ...any) func() {
	then := time.Now()

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(context.TODO(), level, msg, args...)
		}
	}
}

func ContextLogAndMeasure(ctx context.Context, level slog.Level, msg string, args ...any) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, FieldMeasurementID, pairID)
	slog.Log(ctx, level, msg, args...)

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(ctx, level, msg, args...)
		}
	}
}

func LogAndMeasure(level slog.Level, msg string, args ...any) func() {
	var (
		pairID = logMeasurePairCounter.Add(1)
		then   = time.Now()
	)

	args = append(args, FieldMeasurementID, pairID)
	slog.Log(context.TODO(), level, msg, args...)

	return func() {
		if elapsed := time.Since(then); elapsed >= measureThreshold {
			args = append(args, FieldElapsed, elapsed)
			slog.Log(context.TODO(), level, msg, args...)
		}
	}
}
