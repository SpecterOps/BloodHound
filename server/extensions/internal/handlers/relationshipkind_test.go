// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package handlers_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func newRelationshipKindRequestWithID(t *testing.T, rawID string) *http.Request {
	t.Helper()

	request, err := http.NewRequest(http.MethodGet, "/api/v2/relationship-kinds/"+rawID, nil)
	require.NoError(t, err)

	return mux.SetURLVars(request, map[string]string{
		handlers.URIPathVariableRelationshipKindID: rawID,
	})
}

func TestHandlers_GetRelationshipKindByID(t *testing.T) {
	var (
		relationshipKindID = int32(42)
		relationshipKind   = services.RelationshipKind{
			ID:            relationshipKindID,
			KindID:        99,
			Name:          "MemberOf",
			Description:   "a membership relationship",
			IsTraversable: true,
			Extension: services.Extension{
				ID:          7,
				Name:        "TestExtension",
				DisplayName: "Test Extension",
				Namespace:   "TST",
				Version:     "1.0.0",
			},
		}
	)

	tests := []struct {
		name       string
		rawID      string
		setupMock  func(extensionsMock *mocks.MockExtensions)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:  "returns relationship kind view",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				extensionsMock.EXPECT().GetRelationshipKind(mock.Anything, relationshipKindID).Return(relationshipKind, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RelationshipKindView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, handlers.RelationshipKindView{
					RelationshipKindID: relationshipKindID,
					Name:               "MemberOf",
					Description:        "a membership relationship",
					IsTraversable:      true,
					Info:               map[string]handlers.KindInfoView{},
					Extension: handlers.ExtensionView{
						ExtensionID: 7,
						Name:        "TestExtension",
						DisplayName: "Test Extension",
						Namespace:   "TST",
						Version:     "1.0.0",
					},
				}, envelope.Data)
			},
		},
		{
			name:  "returns relationship kind view with info",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				withInfo := relationshipKind
				withInfo.Info = []services.KindInfo{
					{
						InfoKey:  "overview",
						Title:    "Overview",
						Position: 0,
						Content:  json.RawMessage(`{"markdown":{"content":"relationship overview"}}`),
					},
				}
				extensionsMock.EXPECT().GetRelationshipKind(mock.Anything, relationshipKindID).Return(withInfo, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RelationshipKindView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, handlers.KindInfoView{
					Title:    "Overview",
					Position: 0,
					Markdown: handlers.MarkdownView{Content: "relationship overview"},
				}, envelope.Data.Info["overview"])
			},
		},
		{
			name:  "returns relationship kind view when info markdown is malformed",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				withMalformedInfo := relationshipKind
				withMalformedInfo.Info = []services.KindInfo{
					{
						InfoKey:  "overview",
						Title:    "Overview",
						Position: 0,
						Content:  json.RawMessage(`not-json`),
					},
				}
				extensionsMock.EXPECT().GetRelationshipKind(mock.Anything, relationshipKindID).Return(withMalformedInfo, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RelationshipKindView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, handlers.KindInfoView{
					Title:    "Overview",
					Position: 0,
					Markdown: handlers.MarkdownView{},
				}, envelope.Data.Info["overview"])
			},
		},
		{
			name:       "returns bad request for a malformed ID",
			rawID:      "not-a-number",
			wantStatus: http.StatusBadRequest,
			assertBody: assertErrorBody(http.StatusBadRequest, "relationship kind id is malformed"),
		},
		{
			name:  "returns not found when relationship kind does not exist",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				extensionsMock.EXPECT().GetRelationshipKind(mock.Anything, relationshipKindID).Return(services.RelationshipKind{}, services.ErrRelationshipKindNotFound)
			},
			wantStatus: http.StatusNotFound,
			assertBody: assertErrorBody(http.StatusNotFound, "relationship kind not found"),
		},
		{
			name:  "returns internal server error for an unexpected service error",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				extensionsMock.EXPECT().GetRelationshipKind(mock.Anything, relationshipKindID).Return(services.RelationshipKind{}, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
			assertBody: assertErrorBody(http.StatusInternalServerError, "an internal error has occurred that is preventing the service from servicing this request"),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				extensionsMock = mocks.NewMockExtensions(t)
				handlerSet     = handlers.NewHandlersContainer(extensionsMock)
				recorder       = httptest.NewRecorder()
				request        = newRelationshipKindRequestWithID(t, testCase.rawID)
			)

			if testCase.setupMock != nil {
				testCase.setupMock(extensionsMock)
			}

			handlerSet.GetRelationshipKindByID(recorder, request)

			assert.Equal(t, testCase.wantStatus, recorder.Code)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
			require.NotNil(t, testCase.assertBody, "every case must assert the response body")
			testCase.assertBody(t, recorder.Body.Bytes())
		})
	}
}
