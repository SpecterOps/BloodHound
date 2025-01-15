package bhlog

import (
	"context"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/src/auth"
	bhctx "github.com/specterops/bloodhound/src/ctx"
)

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
