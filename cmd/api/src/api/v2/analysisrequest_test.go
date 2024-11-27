package v2_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	v2 "github.com/specterops/bloodhound/src/api/v2"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils/test"
	"go.uber.org/mock/gomock"
)

func TestResources_GetAnalysisRequest(t *testing.T) {
	var (
		mockCtrl  = gomock.NewController(t)
		mockDB    = dbMocks.NewMockDatabase(mockCtrl)
		resources = v2.Resources{DB: mockDB}
		url       = "api/v2/analysis/status"
	)
	defer mockCtrl.Finish()

	t.Run("success getting analysis", func(t *testing.T) {
		analysisRequest := model.AnalysisRequest{
			RequestedAt: time.Now(),
			RequestedBy: "test",
			RequestType: model.AnalysisRequestType("test-type"),
		}

		mockDB.EXPECT().GetAnalysisRequest(gomock.Any()).Return(analysisRequest, nil)

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL(url).
			OnHandlerFunc(resources.GetAnalysisRequest).
			Require().
			ResponseJSONBody(analysisRequest).
			ResponseStatusCode(http.StatusOK)
	})

	t.Run("error getting analysis", func(t *testing.T) {
		mockDB.EXPECT().GetAnalysisRequest(gomock.Any()).Return(model.AnalysisRequest{}, fmt.Errorf("an error"))

		test.Request(t).
			WithMethod(http.MethodGet).
			WithURL(url).
			OnHandlerFunc(resources.GetAnalysisRequest).
			Require().
			ResponseStatusCode(http.StatusInternalServerError)
	})
}
