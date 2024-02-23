package v2_test

import (
	"errors"
	"net/http"
	"testing"

	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
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
		mockGraph  = graph_mocks.NewMockDatabase(mockCtrl)
		resources  = v2.Resources{DB: mockDB, TaskNotifier: mockTasker, Graph: mockGraph}
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.HandleDatabaseWipe).
		Run([]apitest.Case{
			{
				Name: "JSON Malformed",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "JSON malformed")

				},
			},
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
			{
				Name: "failure creating an intent audit log",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{})
				},
				Setup: func() {
					mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(errors.New("oopsy! "))
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "failure creating an intent audit log")
				},
			},
			{
				Name: "failed fetching nodes during attempt to delete collected graph data",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{DeleteCollectedGraphData: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedFetchNodesToDelete := mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedFetchNodesToDelete, successfulAuditLogFailure)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "we encountered an error while deleting collected graph data")
				},
			},
			{
				Name: "failed batch operation to delete nodes during attempt to delete collected graph data",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{DeleteCollectedGraphData: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulFetchNodesToDelete := mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedBatchDelete := mockGraph.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulFetchNodesToDelete, failedBatchDelete, successfulAuditLogFailure)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "we encountered an error while deleting collected graph data")
				},
			},
			{
				Name: "succesful deletion of collected graph data kicks of analysis",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{DeleteCollectedGraphData: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulFetchNodesToDelete := mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulBatchDelete := mockGraph.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					sucsessfulAnalysisKickoff := mockTasker.EXPECT().RequestAnalysis().Times(1)
					successfulAuditLogSuccess := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulFetchNodesToDelete, successfulBatchDelete, sucsessfulAnalysisKickoff, successfulAuditLogSuccess)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
			{
				Name: "failed deletion of high value selectors",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{DeleteHighValueSelectors: true, AssetGroupId: 1})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedAssetGroupSelectorsDelete := mockDB.EXPECT().DeleteAssetGroupSelectorsForAssetGroup(gomock.Any(), gomock.Any()).Return(errors.New("oopsy1")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedAssetGroupSelectorsDelete, successfulAuditLogFailure)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "we encountered an error while deleting high value selectors")
				},
			},
			{
				Name: "successful deletion of high value selectors",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{DeleteHighValueSelectors: true, AssetGroupId: 1})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulAssetGroupSelectorsDelete := mockDB.EXPECT().DeleteAssetGroupSelectorsForAssetGroup(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulAnalysisKickoff := mockTasker.EXPECT().RequestAnalysis().Times(1)
					successfulAuditLogSuccess := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulAssetGroupSelectorsDelete, successfulAnalysisKickoff, successfulAuditLogSuccess)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
			{
				Name: "failed deletion of file ingest history",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseManagement{DeleteFileIngestHistory: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedFileIngestHistoryDelete := mockDB.EXPECT().DeleteAllFileUploads().Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedFileIngestHistoryDelete, successfulAuditLogFailure)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "we encountered an error while deleting file ingest history")
				},
			},
		})
}
