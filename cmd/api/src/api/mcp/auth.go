package mcp

import (
	"context"
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
)

type contextKey string

const tokenCtxKey contextKey = "mcp-token"

type tokenCreds struct {
	ID  string
	Key string
}

func tokenFromHeader(header string) (tokenCreds, bool) {
	parts := strings.SplitN(header, ":", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return tokenCreds{}, false
	}
	return tokenCreds{ID: parts[0], Key: parts[1]}, true
}

func tokenFromContext(ctx context.Context) (tokenCreds, bool) {
	creds, ok := ctx.Value(tokenCtxKey).(tokenCreds)
	return creds, ok
}

func tokenFromRequest(r *http.Request) (tokenCreds, bool) {
	header := r.Header.Get("X-BH-Token")
	if header == "" {
		return tokenCreds{}, false
	}
	return tokenFromHeader(header)
}

func authMiddleware(db database.Database) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			creds, ok := tokenFromRequest(r)
			if !ok {
				http.Error(w, `{"errors":[{"message":"X-BH-Token header required (format: token_id:token_key)"}]}`, http.StatusUnauthorized)
				return
			}

			tokenID, err := uuid.FromString(creds.ID)
			if err != nil {
				http.Error(w, `{"errors":[{"message":"invalid token ID format"}]}`, http.StatusBadRequest)
				return
			}

			if _, err := db.GetAuthToken(r.Context(), tokenID); err != nil {
				http.Error(w, `{"errors":[{"message":"invalid or expired token"}]}`, http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), tokenCtxKey, creds)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
