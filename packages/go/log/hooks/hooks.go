package hooks

import (
	"context"
	"github.com/specterops/bloodhound/src/ctx"
	"log/slog"
)

type ContextHandler struct {
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

		if !bhCtx.AuthCtx.Session.UserID.IsNil() {
			r.Add("user_id", bhCtx.AuthCtx.Session.UserID)
		}
	}

	return h.Handler.Handle(c, r)
}
