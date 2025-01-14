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
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/specterops/bloodhound/bhlog/level"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/src/auth"
	bhctx "github.com/specterops/bloodhound/src/ctx"
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

func ConfigureDefaultText() {
	var (
		handler = TextHandler(os.Stdout, &slog.HandlerOptions{
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

func ConfigureDefaultJSON() {
	var (
		handler = JsonHandler(os.Stdout, &slog.HandlerOptions{
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

type contextHandler struct {
	IDResolver auth.IdentityResolver

	slog.Handler
}

func (s contextHandler) handle(ctx context.Context, record slog.Record) error {
	if bhCtx, ok := ctx.Value(bhctx.ValueKey).(*bhctx.Context); ok {
		if bhCtx.RequestID != "" {
			record.Add(slog.String("request_id", bhCtx.RequestID))
		}

		if bhCtx.RequestIP != "" {
			record.Add(slog.String("request_ip", bhCtx.RequestIP))
		}

		if bhCtx.RemoteAddr != "" {
			record.Add(slog.String("remote_addr", bhCtx.RemoteAddr))
		}

		if bhCtx.AuthCtx.Authenticated() {
			if identity, err := s.IDResolver.GetIdentity(bhCtx.AuthCtx); err == nil {
				record.Add(slog.String(identity.Key, identity.ID.String()))
			}
		}
	}

	return s.Handler.Handle(ctx, record)
}

func textReplaceAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.MessageKey {
		attr.Key = bhlogMessageKey
	}

	return attr
}

func jsonReplaceAttr(_ []string, attr slog.Attr) slog.Attr {
	if attr.Key == slog.MessageKey {
		attr.Key = bhlogMessageKey
	}

	if attr.Key == measure.FieldElapsed && attr.Value.Kind() == slog.KindDuration {
		durationMilli := float64(attr.Value.Duration()) / float64(time.Millisecond)
		attr.Value = slog.Float64Value(durationMilli)
	}

	return attr
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
