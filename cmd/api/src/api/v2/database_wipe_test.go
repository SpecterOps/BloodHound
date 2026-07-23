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
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"testing"

	"github.com/gofrs/uuid"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apitest"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	graph_mocks "github.com/specterops/bloodhound/cmd/api/src/vendormocks/dawgs/graph"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/dawgs/graph"
	"go.uber.org/mock/gomock"
)

func TestDatabaseWipe(t *testing.T) {
	var (
		mockCtrl              = gomock.NewController(t)
		mockDB                = dbMocks.NewMockDatabase(mockCtrl)
		mockGraph             = graph_mocks.NewMockDatabase(mockCtrl)
		resources             = v2.Resources{DB: mockDB, Graph: mockGraph}
		user                  = setupUser()
		userCtx               = setupUserCtx(user)
		capturedDeleteRequest model.AnalysisRequest
	)
	defer mockCtrl.Finish()

	apitest.NewHarness(t, resources.HandleDatabaseWipe).
		WithCommonRequest(func(input *apitest.Input) {
			apitest.SetContext(input, userCtx)
		}).
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
				Name: "endpoint returns a 400 error if the request body has mixed delete payload",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteCollectedGraphData: true, DeleteSourceKinds: []int{1, 2, 3}})
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
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
				Name: "deletion of collected graph data kicks off analysis",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteCollectedGraphData: true})
				},
				Setup: func() {
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), gomock.Any()).Return(appcfg.FeatureFlag{
						Enabled: true,
					}, nil)

					successfulRequestDeletion := mockDB.EXPECT().RequestCollectedGraphDataDeletion(gomock.Any(), gomock.Any()).Times(1)
					successfulAuditLogIntent := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					successfulAuditLogWipe := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					gomock.InOrder(successfulAuditLogIntent, successfulRequestDeletion, successfulAuditLogWipe)

				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
				},
			},
			{
				Name: "successful source kind deletion request",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteSourceKinds: []int{0, 42}})
				},
				Setup: func() {
					capturedDeleteRequest = model.AnalysisRequest{}

					mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), gomock.Any()).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return([]model.SourceKind{{ID: 42, Name: "Base"}}, nil)
					mockGraph.EXPECT().FetchKinds(gomock.Any()).Return(graph.Kinds{graph.StringKind("Base"), graph.StringKind("User")}, nil)
					mockDB.EXPECT().RequestCollectedGraphDataDeletion(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, request model.AnalysisRequest) error {
						capturedDeleteRequest = request
						return nil
					}).Times(1)
					mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
					if capturedDeleteRequest.RequestedBy != user.ID.String() {
						t.Fatalf("expected requested by %q, got %q", user.ID.String(), capturedDeleteRequest.RequestedBy)
					}
					if capturedDeleteRequest.RequestType != model.AnalysisRequestDeletion {
						t.Fatalf("expected request type %q, got %q", model.AnalysisRequestDeletion, capturedDeleteRequest.RequestType)
					}
					if !capturedDeleteRequest.DeleteSourcelessGraph {
						t.Fatal("expected delete sourceless graph to be true")
					}
					if !slices.Equal(capturedDeleteRequest.DeleteSourceKinds, []string{"Base"}) {
						t.Fatalf("expected delete source kinds [Base], got %v", capturedDeleteRequest.DeleteSourceKinds)
					}
				},
			},
			{
				Name: "successful relationship deletion request",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteRelationships: []string{"HasSession"}})
				},
				Setup: func() {
					capturedDeleteRequest = model.AnalysisRequest{}

					mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), gomock.Any()).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().RequestCollectedGraphDataDeletion(gomock.Any(), gomock.Any()).DoAndReturn(func(_ context.Context, request model.AnalysisRequest) error {
						capturedDeleteRequest = request
						return nil
					}).Times(1)
					mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusNoContent)
					if capturedDeleteRequest.RequestedBy != user.ID.String() {
						t.Fatalf("expected requested by %q, got %q", user.ID.String(), capturedDeleteRequest.RequestedBy)
					}
					if capturedDeleteRequest.RequestType != model.AnalysisRequestDeletion {
						t.Fatalf("expected request type %q, got %q", model.AnalysisRequestDeletion, capturedDeleteRequest.RequestType)
					}
					if !slices.Equal(capturedDeleteRequest.DeleteRelationships, []string{"HasSession"}) {
						t.Fatalf("expected delete relationships [HasSession], got %v", capturedDeleteRequest.DeleteRelationships)
					}
					if capturedDeleteRequest.DeleteAllGraph {
						t.Fatal("expected delete all graph to be false")
					}
				},
			},
			{
				Name: "source kind deletion request returns 400 for unknown source kind id",
				Input: func(input *apitest.Input) {
					apitest.SetHeader(input, headers.ContentType.String(), mediatypes.ApplicationJson.String())
					apitest.BodyStruct(input, v2.DatabaseWipe{DeleteSourceKinds: []int{99}})
				},
				Setup: func() {
					mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), gomock.Any()).Return(appcfg.FeatureFlag{Enabled: true}, nil)
					mockDB.EXPECT().GetSourceKinds(gomock.Any()).Return([]model.SourceKind{{ID: 42, Name: "Base"}}, nil)
					mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
				},
				Test: func(output apitest.Output) {
					apitest.StatusCode(output, http.StatusBadRequest)
					apitest.BodyContains(output, "failure building delete request: requested source kind id 99 not found")
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
					successfulAnalysisKickoff := mockDB.EXPECT().RequestAnalysis(gomock.Any(), uuid.UUID{}.String(), model.AnalysisModeFull).Times(1)

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
					failedFileIngestHistoryDelete := mockDB.EXPECT().DeleteAllIngestJobs(gomock.Any()).Return(errors.New("oopsy!")).Times(1)
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
					successfulFileIngestHistoryDelete := mockDB.EXPECT().DeleteAllIngestJobs(gomock.Any()).Return(nil).Times(1)
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
				Name: "successful deletion of data quality history",
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

					failedFileIngestHistoryDelete := mockDB.EXPECT().DeleteAllIngestJobs(gomock.Any()).Return(errors.New("oopsy!")).Times(1)
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
					mockDB.EXPECT().GetFlagByKey(gomock.Any(), gomock.Any()).Return(appcfg.FeatureFlag{
						Enabled: true,
					}, nil)
					successfulDeletionRequest := mockDB.EXPECT().RequestCollectedGraphDataDeletion(gomock.Any(), gomock.Any()).Times(1)
					nodesDeletedAuditLog := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					gomock.InOrder(successfulAuditLogIntent, successfulDeletionRequest, nodesDeletedAuditLog)

					// high value selector operations
					assetGroupSelectorsDelete := mockDB.EXPECT().DeleteAssetGroupSelectorsForAssetGroups(gomock.Any(), gomock.Any()).Return(nil).Times(1)
					assetGroupSelectorsAuditLog := mockDB.EXPECT().AppendAuditLog(gomock.Any(), gomock.Any()).Return(nil).Times(1)

					// analysis kickoff
					analysisKickoff := mockDB.EXPECT().RequestAnalysis(gomock.Any(), uuid.UUID{}.String(), model.AnalysisModeFull).Times(1)

					gomock.InOrder(assetGroupSelectorsDelete, assetGroupSelectorsAuditLog, analysisKickoff)

					// file ingest history operations
					deleteFileHistory := mockDB.EXPECT().DeleteAllIngestJobs(gomock.Any()).Return(nil).Times(1)
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

func TestResources_BuildDeleteRequest(t *testing.T) {
	t.Run("maps relationship delete request", func(t *testing.T) {
		var (
			testCtx   = context.Background()
			mockCtrl  = gomock.NewController(t)
			mockDB    = dbMocks.NewMockDatabase(mockCtrl)
			mockGraph = graph_mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB, Graph: mockGraph}
		)
		defer mockCtrl.Finish()

		deleteRequest, err := resources.BuildDeleteRequest(testCtx, "test-user", v2.DatabaseWipe{
			DeleteRelationships: []string{"HasSession"},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if deleteRequest.RequestedBy != "test-user" {
			t.Fatalf("expected requested by to be test-user, got %q", deleteRequest.RequestedBy)
		}

		if deleteRequest.RequestType != model.AnalysisRequestDeletion {
			t.Fatalf("expected request type %q, got %q", model.AnalysisRequestDeletion, deleteRequest.RequestType)
		}

		if !slices.Equal(deleteRequest.DeleteRelationships, []string{"HasSession"}) {
			t.Fatalf("expected delete relationships to contain HasSession, got %v", deleteRequest.DeleteRelationships)
		}
	})

	t.Run("maps source kinds and sourceless flag", func(t *testing.T) {
		var (
			testCtx   = context.Background()
			mockCtrl  = gomock.NewController(t)
			mockDB    = dbMocks.NewMockDatabase(mockCtrl)
			mockGraph = graph_mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB, Graph: mockGraph}
		)
		defer mockCtrl.Finish()

		mockDB.EXPECT().GetSourceKinds(testCtx).Return([]model.SourceKind{{ID: 42, Name: "Base"}}, nil)
		mockGraph.EXPECT().FetchKinds(testCtx).Return(graph.Kinds{graph.StringKind("Base"), graph.StringKind("User")}, nil)

		deleteRequest, err := resources.BuildDeleteRequest(testCtx, "test-user", v2.DatabaseWipe{
			DeleteSourceKinds: []int{0, 42},
		})
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if !deleteRequest.DeleteSourcelessGraph {
			t.Fatal("expected delete sourceless graph to be true")
		}

		if !slices.Equal(deleteRequest.DeleteSourceKinds, []string{"Base"}) {
			t.Fatalf("expected delete source kinds to contain Base, got %v", deleteRequest.DeleteSourceKinds)
		}
	})

	t.Run("rejects unknown source kind id", func(t *testing.T) {
		var (
			testCtx   = context.Background()
			mockCtrl  = gomock.NewController(t)
			mockDB    = dbMocks.NewMockDatabase(mockCtrl)
			mockGraph = graph_mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB, Graph: mockGraph}
		)
		defer mockCtrl.Finish()

		mockDB.EXPECT().GetSourceKinds(testCtx).Return([]model.SourceKind{{ID: 42, Name: "Base"}}, nil)

		_, err := resources.BuildDeleteRequest(testCtx, "test-user", v2.DatabaseWipe{
			DeleteSourceKinds: []int{99},
		})
		if err == nil {
			t.Fatal("expected an error for an unknown source kind id")
		}

		if !strings.Contains(err.Error(), "requested source kind id 99 not found") {
			t.Fatalf("expected unknown source kind error, got %v", err)
		}
	})

	t.Run("rejects deleteCollectedGraphData combined with source kinds", func(t *testing.T) {
		var (
			testCtx   = context.Background()
			mockCtrl  = gomock.NewController(t)
			mockDB    = dbMocks.NewMockDatabase(mockCtrl)
			mockGraph = graph_mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB, Graph: mockGraph}
		)
		defer mockCtrl.Finish()

		_, err := resources.BuildDeleteRequest(testCtx, "test-user", v2.DatabaseWipe{
			DeleteCollectedGraphData: true,
			DeleteSourceKinds:        []int{42},
		})
		if err == nil {
			t.Fatal("expected an error when deleteCollectedGraphData is combined with source kinds")
		}

		if !strings.Contains(err.Error(), "deleteCollectedGraphData may not be combined with deleteSourceKinds or deleteRelationships") {
			t.Fatalf("expected mutual exclusion error, got %v", err)
		}
	})

	t.Run("rejects deleteCollectedGraphData combined with relationships", func(t *testing.T) {
		var (
			testCtx   = context.Background()
			mockCtrl  = gomock.NewController(t)
			mockDB    = dbMocks.NewMockDatabase(mockCtrl)
			mockGraph = graph_mocks.NewMockDatabase(mockCtrl)
			resources = v2.Resources{DB: mockDB, Graph: mockGraph}
		)
		defer mockCtrl.Finish()

		_, err := resources.BuildDeleteRequest(testCtx, "test-user", v2.DatabaseWipe{
			DeleteCollectedGraphData: true,
			DeleteRelationships:      []string{"HasSession"},
		})
		if err == nil {
			t.Fatal("expected an error when deleteCollectedGraphData is combined with relationships")
		}

		if !strings.Contains(err.Error(), "deleteCollectedGraphData may not be combined with deleteSourceKinds or deleteRelationships") {
			t.Fatalf("expected mutual exclusion error, got %v", err)
		}
	})
}
