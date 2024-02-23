package v2_test

import (
	"net/http"
	"testing"

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/api/v2/apitest"
	taskerMocks "github.com/specterops/bloodhound/src/daemons/datapipe/mocks"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"go.uber.org/mock/gomock"
)

func TestDatabaseWipe(t *testing.T) {
	var (
		mockCtrl   = gomock.NewController(t)
		mockDB     = dbMocks.NewMockDatabase(mockCtrl)
		mockTasker = taskerMocks.NewMockTasker(mockCtrl)
		resources  = v2.Resources{DB: mockDB, TaskNotifier: mockTasker}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.HandleDatabaseWipe).
		Run([]apitest.Case{
			// {
			// 	Name: "JSON Malformed",
			// 	Test: func(output apitest.Output) {
			// 		apitest.StatusCode(output, http.StatusBadRequest)
			// 	},
			// },
			{
				Name: "asset group id must be provided if deleting asset group selectors",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{
						DeleteHighValueSelectors: true,
					})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "please provide an assetGroupId to delete")
				},
			},
		})
}
