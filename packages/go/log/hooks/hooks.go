package hooks

import (
	"github.com/rs/zerolog"
	"github.com/specterops/bloodhound/src/ctx"
)

type BhCtx struct{}

func (s BhCtx) Run(e *zerolog.Event, level zerolog.Level, msg string) {
	c := e.GetCtx()

	if bhCtx, ok := c.Value(ctx.ValueKey).(*ctx.Context); ok {
		if bhCtx.RequestID != "" {
			e.Str("request_id", bhCtx.RequestID)
		}

		if bhCtx.RequestIP != "" {
			e.Str("request_ip", bhCtx.RequestIP)
		}

		if !bhCtx.AuthCtx.Session.UserID.IsNil() {
			e.Str("user_id", bhCtx.AuthCtx.Session.UserID.String())
		}
	}
}
