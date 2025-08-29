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

func NewWrappedTextHandler(pipe io.Writer) slog.Handler {
	return slog.NewTextHandler(pipe, &slog.HandlerOptions{
		Level:       level.GetLevelVar(),
		ReplaceAttr: textReplaceAttr,
	})
}

func NewWrappedJSONHandler(pipe io.Writer) slog.Handler {
	return slog.NewJSONHandler(pipe, &slog.HandlerOptions{
		Level:       level.GetLevelVar(),
		ReplaceAttr: jsonReplaceAttr,
	})
}

func NewWrappedLogger(handler slog.Handler) *slog.Logger {
	return slog.New(&contextHandler{
		IDResolver: auth.NewIdentityResolver(),
		Handler:    handler,
	})
}

// Deprecated: use slog.SetDefault with NewWrappedTextHandler and NewWrappedLogger instead
func ConfigureDefaultText(pipe io.Writer) {
	handler := NewWrappedTextHandler(pipe)
	logger := NewWrappedLogger(handler)
	slog.SetDefault(logger)
}

// Deprecated: use slog.SetDefault with NewWrappedJSONHandler and NewWrappedLogger instead
func ConfigureDefaultJSON(pipe io.Writer) {
	handler := NewWrappedJSONHandler(pipe)
	logger := NewWrappedLogger(handler)
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
