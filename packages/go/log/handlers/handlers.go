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

package handlers

import (
	"context"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
)

var lvl = new(slog.LevelVar)

type ContextHandler struct {
	IDResolver auth.IdentityResolver

	slog.Handler
}

func (h ContextHandler) Handle(c context.Context, r slog.Record) error {
	if bhCtx, ok := c.Value(ctx.ValueKey).(*ctx.Context); ok {
		if bhCtx.RequestID != "" {
			r.Add(slog.String("request_id", bhCtx.RequestID))
		}

		if bhCtx.RequestIP != "" {
			r.Add(slog.String("request_ip", bhCtx.RequestIP))
		}

		if bhCtx.RemoteAddr != "" {
			r.Add(slog.String("remote_addr", bhCtx.RemoteAddr))
		}

		if bhCtx.AuthCtx.Authenticated() {
			if identity, err := h.IDResolver.GetIdentity(bhCtx.AuthCtx); err == nil {
				r.Add(slog.String(identity.Key, identity.ID.String()))
			}
		}
	}

	return h.Handler.Handle(c, r)
}

func ReplaceAttr(_ []string, a slog.Attr) slog.Attr {
	if a.Key == slog.MessageKey {
		a.Key = "message"
	}

	return a
}

func NewDefaultLogger() *slog.Logger {
	return slog.New(&ContextHandler{IDResolver: auth.NewIdentityResolver(), Handler: slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: lvl, ReplaceAttr: ReplaceAttr})})
}

func SetGlobalLevel(level slog.Level) {
	lvl.Set(level)
}

func GlobalLevel() slog.Level {
	return lvl.Level()
}
