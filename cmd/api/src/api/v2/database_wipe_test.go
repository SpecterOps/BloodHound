// Copyright 2024 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

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
				Name: "endpoint returns a 400 error if the request body is empty",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "please select something to delete")
				},
			},
			{
				Name: "failure creating an intent audit log",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteCollectedGraphData: true})
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
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteCollectedGraphData: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedFetchNodesToDelete := mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedFetchNodesToDelete, successfulAuditLogFailure)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "We encountered an error while deleting collected graph data")
				},
			},
			{
				Name: "failed batch operation to delete nodes during attempt to delete collected graph data",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteCollectedGraphData: true})
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
					apitest.BodyContains(output, "We encountered an error while deleting collected graph data")
				},
			},
			{
				Name: "succesful deletion of collected graph data kicks of analysis",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteCollectedGraphData: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulFetchNodesToDelete := mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulBatchDelete := mockGraph.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulAuditLogSuccess := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					sucsessfulAnalysisKickoff := mockTasker.EXPECT().RequestAnalysis().Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulFetchNodesToDelete, successfulBatchDelete, successfulAuditLogSuccess, sucsessfulAnalysisKickoff)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
			{
				Name: "failed deletion of high value selectors",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteAssetGroupSelectors: []int{1}})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedAssetGroupSelectorsDelete := mockDB.EXPECT().DeleteAssetGroupSelectorsForAssetGroups(gomock.Any(), gomock.Any()).Return(errors.New("oopsy1")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedAssetGroupSelectorsDelete, successfulAuditLogFailure)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "We encountered an error while deleting custom high value selectors")
				},
			},
			{
				Name: "successful deletion of high value selectors",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteAssetGroupSelectors: []int{1}})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulAssetGroupSelectorsDelete := mockDB.EXPECT().DeleteAssetGroupSelectorsForAssetGroups(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulAuditLogSuccess := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulAnalysisKickoff := mockTasker.EXPECT().RequestAnalysis().Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulAssetGroupSelectorsDelete, successfulAuditLogSuccess, successfulAnalysisKickoff)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
			{
				Name: "failed deletion of file ingest history",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteFileIngestHistory: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedFileIngestHistoryDelete := mockDB.EXPECT().DeleteAllFileUploads(gomock.Any()).Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedFileIngestHistoryDelete, successfulAuditLogFailure)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "We encountered an error while deleting file ingest history")
				},
			},
			{
				Name: "successful deletion of file ingest history",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteFileIngestHistory: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulFileIngestHistoryDelete := mockDB.EXPECT().DeleteAllFileUploads(gomock.Any()).Return(nil).Times(1)
					successfulAuditLogSuccess := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulFileIngestHistoryDelete, successfulAuditLogSuccess)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
			{
				Name: "failed deletion of data quality history",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteDataQualityHistory: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					failedDataQualityHistoryDelete := mockDB.EXPECT().DeleteAllDataQuality(gomock.Any()).Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedDataQualityHistoryDelete, successfulAuditLogFailure)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "We encountered an error while deleting data quality history")
				},
			},
			{
				Name: "succesful deletion of data quality history",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteDataQualityHistory: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulDataQualityHistoryDelete := mockDB.EXPECT().DeleteAllDataQuality(gomock.Any()).Return(nil).Times(1)
					successfulAuditLogSuccess := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulDataQualityHistoryDelete, successfulAuditLogSuccess)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
			{
				Name: "correctly forms the error message when multiple delete operations fail",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteDataQualityHistory: true, DeleteFileIngestHistory: true})
				},
				Setup: func() {
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					failedFileIngestHistoryDelete := mockDB.EXPECT().DeleteAllFileUploads(gomock.Any()).Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogFileHistoryFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					failedDataQualityHistoryDelete := mockDB.EXPECT().DeleteAllDataQuality(gomock.Any()).Return(errors.New("oopsy!")).Times(1)
					successfulAuditLogDataQualityHistoryFailure := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, failedFileIngestHistoryDelete, successfulAuditLogFileHistoryFailure, failedDataQualityHistoryDelete, successfulAuditLogDataQualityHistoryFailure)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusInternalServerError)
					apitest.BodyContains(output, "We encountered an error while deleting file ingest history, data quality history")
				},
			},
			{
				Name: "handler produces one `AuditLogIntent` and 4 `AuditLogSuccess` entries if all four data types are deleted successfully",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{
						DeleteCollectedGraphData:  true,
						DeleteAssetGroupSelectors: []int{1},
						DeleteFileIngestHistory:   true,
						DeleteDataQualityHistory:  true,
					})
				},
				Setup: func() {
					// intent
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					// collected graph data operations
					fetchNodesToDelete := mockGraph.EXPECT().ReadTransaction(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					batchDelete := mockGraph.EXPECT().BatchOperation(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					nodesDeletedAuditLog := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					gomock.InOrder(successfulAuditLogIntent, fetchNodesToDelete, batchDelete, nodesDeletedAuditLog)

					// high value selector operations
					assetGroupSelectorsDelete := mockDB.EXPECT().DeleteAssetGroupSelectorsForAssetGroups(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					assetGroupSelectorsAuditLog := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					// analysis kickoff
					analysisKickoff := mockTasker.EXPECT().RequestAnalysis().Times(1)

					gomock.InOrder(assetGroupSelectorsDelete, assetGroupSelectorsAuditLog, analysisKickoff)

					// file ingest history operations
					deleteFileHistory := mockDB.EXPECT().DeleteAllFileUploads(gomock.Any()).Return(nil).Times(1)
					fileHistoryAuditLog := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					gomock.InOrder(deleteFileHistory, fileHistoryAuditLog)

					// data quality history operations
					deleteDataQuality := mockDB.EXPECT().DeleteAllDataQuality(gomock.Any()).Return(nil).Times(1)
					dataQualityAuditLog := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					gomock.InOrder(deleteDataQuality, dataQualityAuditLog)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
		})
}
