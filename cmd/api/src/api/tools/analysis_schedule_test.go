package tools_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/golang/mock/gomock"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/api/tools"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestToolContainer_GetScheduledAnalysisConfiguration_Errors(t *testing.T) {
	t.Run("returns error on invalid rrule", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "abc123",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: wrong")
		}
	})

	t.Run("returns error on rrule with count", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "FREQ=DAILY;INTERVAL=1;COUNT=3",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: count/until not supported")
		}
	})

	t.Run("returns error on rrule with until", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "FREQ=DAILY;INTERVAL=1;UNTIL=20240930T000000Z",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: count/until not supported")
		}
	})

	t.Run("returns error on rrule without dtstart", func(t *testing.T) {
		var (
			ctrl     = gomock.NewController(t)
			handlers = tools.ToolContainer{}
		)

		defer ctrl.Finish()

		endpoint := "/analysis/schedule"
		requestCtx := context.WithValue(context.Background(), ctx.ValueKey, &ctx.Context{})
		scheduledAnalysisRequest := tools.ScheduledAnalysisConfiguration{
			Enabled: true,
			RRule:   "RRULE:FREQ=DAILY;INTERVAL=1",
		}

		reqBody, _ := json.Marshal(scheduledAnalysisRequest)

		if req, err := http.NewRequestWithContext(requestCtx, http.MethodPut, endpoint, bytes.NewBuffer(reqBody)); err != nil {
			t.Fatal(err)
		} else {
			req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

			router := mux.NewRouter()
			router.HandleFunc(endpoint, handlers.SetScheduledAnalysisConfiguration).Methods(http.MethodPut)

			response := httptest.NewRecorder()
			router.ServeHTTP(response, req)

			require.Equal(t, http.StatusBadRequest, response.Code)
			require.Contains(t, response.Body.String(), "invalid rrule specified: dtstart is required")
		}
	})
}
