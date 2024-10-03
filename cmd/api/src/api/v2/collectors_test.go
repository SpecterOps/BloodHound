package v2_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/database/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_GetCollectorManifest(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = mocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
	)

	defer mockCtrl.Finish()

	t.Run("success", func(t *testing.T) {

		collectorType := v2.CollectorTypeAzurehound
		endpoint := fmt.Sprintf("/api/v2/collectors/%s", collectorType)

		req, err := http.NewRequest(http.MethodGet, endpoint, nil)
		require.Nil(t, err)

		req.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())

		router := mux.NewRouter()
		router.HandleFunc(endpoint, resources.GetCollectorManifest).Methods(http.MethodGet)

		response := httptest.NewRecorder()

		router.ServeHTTP(response, req)
		require.Equal(t, http.StatusBadRequest, response.Code)
		require.Contains(t, response.Body.String(), "Public cannot be true while user_ids is populated")
	})
}
