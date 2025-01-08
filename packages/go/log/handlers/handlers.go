package handlers

import (
	"context"
	"log/slog"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
)

type ContextHandler struct {
	IdResolver auth.IdentityResolver

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
			if identity, err := h.IdResolver.GetIdentity(bhCtx.AuthCtx); err == nil {
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
