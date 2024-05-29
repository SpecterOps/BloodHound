package api_test

import (
	"context"
	"crypto/sha256"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/src/api"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"net/http"
	"testing"
	"time"
)

func Test_NewRequestSignature(t *testing.T) {
	t.Run("returns error on context timeout", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		goCtx, cancel := context.WithDeadline(context.Background(), time.Now())
		defer cancel()
		time.Sleep(1 * time.Microsecond)
		_, err = api.NewRequestSignature(goCtx, sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "context deadline exceeded")
	})

	t.Run("returns error on empty hmac signature", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		goCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		defer cancel()
		_, err = api.NewRequestSignature(goCtx, nil, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.Error(t, err)
		require.Contains(t, err.Error(), "hasher must not be nil")
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		req, err := http.NewRequest(http.MethodGet, "http://teapotsrus.dev", nil)
		require.NoError(t, err)

		req.Header.Add(headers.RequestDate.String(), time.Now().Format(time.RFC3339))

		goCtx, cancel := context.WithDeadline(context.Background(), time.Now().Add(5*time.Second))
		defer cancel()
		signature, err := api.NewRequestSignature(goCtx, sha256.New, "token", time.Now().Format(time.RFC3339), req.Method, req.RequestURI, nil)
		require.Nil(t, err)
		require.NotEmpty(t, signature)
	})
}
