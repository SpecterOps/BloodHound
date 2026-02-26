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
package trace

import (
	"context"
	"log/slog"
	"sync/atomic"
	"time"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/metrics"
)

type measureCtxKey struct{}

var (
	nextContextID = &atomic.Uint64{}
)

func combineArgs(args ...any) []any {
	var all []any

	for _, arg := range args {
		switch typedArg := arg.(type) {
		case []any:
			all = append(all, typedArg...)
		case any:
			all = append(all, typedArg)
		}
	}

	return all
}

type Trace struct {
	ID        uint64
	Started   time.Time
	Level     slog.Level
	Namespace string
	Component string
}

func newContext(level slog.Level, namespace, component string) *Trace {
	return &Trace{
		ID:        nextContextID.Add(1),
		Started:   time.Now(),
		Level:     level,
		Namespace: namespace,
		Component: component,
	}
}

func withContext(ctx context.Context, newMeasureCtx *Trace) context.Context {
	return context.WithValue(ctx, measureCtxKey{}, newMeasureCtx)
}

func fromContext(ctx context.Context) (*Trace, bool) {
	if measureCtx := ctx.Value(measureCtxKey{}); measureCtx != nil {
		typedMeasureCtx, typeOK := measureCtx.(*Trace)
		return typedMeasureCtx, typeOK
	}

	return nil, false
}

func Context(ctx context.Context, level slog.Level, namespace, component string) context.Context {
	return withContext(ctx, newContext(level, namespace, component))
}

func Function(ctx context.Context, function string, startArgs ...any) func(args ...any) {
	var (
		level                 = slog.LevelInfo
		then                  = time.Now()
		traceCtx, hasTraceCtx = fromContext(ctx)
		commonArgs            []any
	)

	commonArgs = combineArgs([]any{
		attr.Scope("process"),
		attr.Function(function),
	}, startArgs)

	if hasTraceCtx {
		level = traceCtx.Level

		commonArgs = combineArgs(commonArgs, []any{
			attr.Namespace(traceCtx.Namespace),
			attr.Measurement(traceCtx.ID),
		})
	}

	slog.Log(ctx, level, "Function Trace", combineArgs(
		commonArgs,
		attr.Enter(),
		startArgs,
	)...)

	return func(exitArgs ...any) {
		elapsed := time.Since(then)

		if hasTraceCtx {
			metrics.Counter("function_trace", traceCtx.Namespace, map[string]string{
				"fn": function,
			}).Add(elapsed.Seconds())
		}

		slog.Log(ctx, level, "Function Trace", combineArgs(
			commonArgs,
			attr.Exit(),
			startArgs,
			attr.Elapsed(elapsed),
			exitArgs,
		)...)
	}
}

func Method(ctx context.Context, receiver, function string, startArgs ...any) func(args ...any) {
	return Function(ctx, receiver+"."+function, startArgs...)
}
