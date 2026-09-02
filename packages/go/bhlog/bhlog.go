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
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/packages/go/bhlog/level"
)

const (
	bhlogMessageKey = "message"
)

func JsonHandler(pipe io.Writer, options *slog.HandlerOptions) slog.Handler {
	return slog.NewJSONHandler(pipe, options)
}

func TextHandler(pipe io.Writer, options *slog.HandlerOptions) slog.Handler {
	return slog.NewTextHandler(pipe, options)
}

func ConfigureDefaultText(pipe io.Writer) {
	var (
		handler = TextHandler(pipe, &slog.HandlerOptions{
			Level:       level.GetLevelVar(),
			ReplaceAttr: textReplaceAttr,
		})

		logger = slog.New(&contextHandler{
			IDResolver: auth.NewIdentityResolver(),
			Handler:    handler,
		})
	)

	slog.SetDefault(logger)
}

func ConfigureDefaultJSON(pipe io.Writer) {
	var (
		handler = JsonHandler(pipe, &slog.HandlerOptions{
			Level:       level.GetLevelVar(),
			ReplaceAttr: jsonReplaceAttr,
		})

		logger = slog.New(&contextHandler{
			IDResolver: auth.NewIdentityResolver(),
			Handler:    handler,
		})
	)

	slog.SetDefault(logger)
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
