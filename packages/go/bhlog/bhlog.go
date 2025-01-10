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

package bhlog

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"runtime"
	"strings"

	"github.com/specterops/bloodhound/bhlog/handlers"
	"github.com/specterops/bloodhound/bhlog/level"
	"github.com/specterops/bloodhound/src/auth"
)

func BaseHandler(pipe io.Writer, options *slog.HandlerOptions) slog.Handler {
	return slog.NewJSONHandler(pipe, options)
}

func TextHandler(pipe io.Writer, options *slog.HandlerOptions) slog.Handler {
	return slog.NewTextHandler(pipe, options)
}

func ConfigureDefault(text bool) {
	var (
		handler        slog.Handler
		pipe           = os.Stderr
		handlerOptions = &slog.HandlerOptions{Level: level.GetLevelVar(), ReplaceAttr: handlers.ReplaceMessageKey}
	)

	if text {
		handler = TextHandler(pipe, handlerOptions)
	} else {
		handler = BaseHandler(pipe, handlerOptions)
	}

	logger := slog.New(&handlers.ContextHandler{
		IDResolver: auth.NewIdentityResolver(),
		Handler:    handler,
	})

	slog.SetDefault(logger)
}

type stackFrame struct {
	File string `json:"file"`
	Line int    `json:"line"`
	Func string `json:"func"`
}

func GetCallStack() slog.Attr {
	var outputFrames []stackFrame

	pc := make([]uintptr, 25) // Arbitrarily only go to a call depth of 25
	n := runtime.Callers(1, pc)
	if n == 0 {
		return slog.Attr{}
	}
	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	for {
		frame, more := frames.Next()

		outputFrames = append(outputFrames, stackFrame{File: frame.File, Line: frame.Line, Func: frame.Function})

		if !more {
			break
		}
	}

	return slog.Any("stack", outputFrames)
}

var (
	levelErrorValue = slog.LevelError.String()
	levelWarnValue  = slog.LevelWarn.String()
	levelInfoValue  = slog.LevelInfo.String()
	levelDebugValue = slog.LevelDebug.String()
)

func ParseLevel(rawLevel string) (slog.Level, error) {
	switch strings.ToUpper(rawLevel) {
	case levelErrorValue:
		return slog.LevelError, nil
	case levelWarnValue:
		return slog.LevelWarn, nil
	case levelInfoValue:
		return slog.LevelInfo, nil
	case levelDebugValue:
		return slog.LevelDebug, nil
	default:
		return 0, fmt.Errorf("unknown log level: %s", rawLevel)
	}
}
