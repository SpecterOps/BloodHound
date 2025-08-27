// Copyright 2025 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// --------------------------------------------------
// TODO: replace with stdlib version once released
// imported from in-progress MultiHandler implementation at https://go-review.googlesource.com/c/go/+/692237
// --------------------------------------------------

package bhlog

import (
	"context"
	"errors"
	"log/slog"
)

// NewMultiHandler creates a [MultiHandler] with the given Handlers.
func NewMultiHandler(handlers ...slog.Handler) *MultiHandler {
	h := make([]slog.Handler, len(handlers))
	copy(h, handlers)
	return &MultiHandler{multi: h}
}

// MultiHandler is a [slog.Handler] that invokes all the given Handlers.
// Its Enable method reports whether any of the handlers' Enabled methods return true.
// Its Handle, WithAttr and WithGroup methods call the corresponding method on each of the enabled handlers.
type MultiHandler struct {
	multi []slog.Handler
}

func (h *MultiHandler) Enabled(ctx context.Context, l slog.Level) bool {
	for i := range h.multi {
		if h.multi[i].Enabled(ctx, l) {
			return true
		}
	}
	return false
}

func (h *MultiHandler) Handle(ctx context.Context, r slog.Record) error {
	var errs []error
	for i := range h.multi {
		if h.multi[i].Enabled(ctx, r.Level) {
			if err := h.multi[i].Handle(ctx, r.Clone()); err != nil {
				errs = append(errs, err)
			}
		}
	}
	return errors.Join(errs...)
}

func (h *MultiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.multi))
	for i := range h.multi {
		handlers = append(handlers, h.multi[i].WithAttrs(attrs))
	}
	return &MultiHandler{multi: handlers}
}

func (h *MultiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, 0, len(h.multi))
	for i := range h.multi {
		handlers = append(handlers, h.multi[i].WithGroup(name))
	}
	return &MultiHandler{multi: handlers}
}
