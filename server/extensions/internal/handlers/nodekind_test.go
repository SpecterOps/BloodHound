// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
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
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers"
	"github.com/specterops/bloodhound/server/extensions/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/extensions/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newNodeKindRequestWithID builds a GET request whose mux path variables carry the
// supplied raw node kind id, mirroring what the router does at runtime.
func newNodeKindRequestWithID(t *testing.T, rawID string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "/api/v2/node-kinds/"+rawID, nil)
	require.NoError(t, err)
	return mux.SetURLVars(request, map[string]string{handlers.URIPathVariableNodeKindID: rawID})
}

// assertErrorBody returns an assertBody func that exhaustively verifies the standard
// responses.ErrorWrapper payload: status echo, populated timestamp, and exactly one
// error detail carrying an empty context and the expected message.
func assertErrorBody(wantStatus int, wantMessage string) func(t *testing.T, body []byte) {
	return func(t *testing.T, body []byte) {
		var errorWrapper responses.ErrorWrapper
		require.NoError(t, json.Unmarshal(body, &errorWrapper))

		assert.Equal(t, wantStatus, errorWrapper.HTTPStatus)
		assert.NotZero(t, errorWrapper.Timestamp)
		require.Len(t, errorWrapper.Errors, 1)
		assert.Empty(t, errorWrapper.Errors[0].Context)
		assert.Equal(t, wantMessage, errorWrapper.Errors[0].Message)
	}
}

func TestHandlers_GetNodeKindByID(t *testing.T) {
	var (
		nodeKindID = int32(42)
		nodeKind   = services.NodeKind{
			ID: nodeKindID, Name: "User", DisplayName: "User", IsDisplayKind: true, Icon: "user", Color: "#fff",
			Info: []services.KindInfo{
				{InfoKey: "overview", Title: "Overview", Position: 0, Content: json.RawMessage(`{"markdown":{"content":"one"}}`)},
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
			name:  "success_-_returns_node_kind_view_with_info",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				extensionsMock.EXPECT().GetNodeKind(mock.Anything, nodeKindID).Return(nodeKind, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.NodeKindView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))

				assert.Equal(t, handlers.NodeKindView{
					NodeKindID:    nodeKindID,
					Name:          "User",
					DisplayName:   "User",
					Description:   "",
					IsDisplayKind: true,
					Icon:          "user",
					Color:         "#fff",
					Info: map[string]handlers.KindInfoView{
						"overview": {
							Title:    "Overview",
							Position: 0,
							Markdown: handlers.MarkdownView{Content: "one"},
						},
					},
				}, envelope.Data)
				assert.NotContains(t, string(body), `\"markdown\"`)
			},
		},
		{
			name:  "success_-_returns_empty_markdown_when_info_content_is_malformed",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				degraded := services.NodeKind{
					ID: nodeKindID, Name: "User",
					Info: []services.KindInfo{{InfoKey: "overview", Title: "Overview", Position: 0, Content: json.RawMessage(`not-json`)}},
				}
				extensionsMock.EXPECT().GetNodeKind(mock.Anything, nodeKindID).Return(degraded, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.NodeKindView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))

				assert.Equal(t, handlers.NodeKindView{
					NodeKindID: nodeKindID,
					Name:       "User",
					Info: map[string]handlers.KindInfoView{
						"overview": {
							Title:    "Overview",
							Position: 0,
							Markdown: handlers.MarkdownView{Content: ""},
						},
					},
				}, envelope.Data)
			},
		},
		{
			name:       "error_-_returns_400_when_id_is_malformed",
			rawID:      "not-a-number",
			wantStatus: http.StatusBadRequest,
			assertBody: assertErrorBody(http.StatusBadRequest, "node kind id is malformed"),
		},
		{
			name:  "error_-_returns_404_when_node_kind_not_found",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				extensionsMock.EXPECT().GetNodeKind(mock.Anything, nodeKindID).Return(services.NodeKind{}, services.ErrNodeKindNotFound)
			},
			wantStatus: http.StatusNotFound,
			assertBody: assertErrorBody(http.StatusNotFound, "node kind not found"),
		},
		{
			name:  "error_-_returns_500_on_unexpected_service_error",
			rawID: "42",
			setupMock: func(extensionsMock *mocks.MockExtensions) {
				extensionsMock.EXPECT().GetNodeKind(mock.Anything, nodeKindID).Return(services.NodeKind{}, errors.New("boom"))
			},
			wantStatus: http.StatusInternalServerError,
			assertBody: assertErrorBody(http.StatusInternalServerError, "an internal error has occurred that is preventing the service from servicing this request"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				extensionsMock = mocks.NewMockExtensions(t)
				handlerSet     = handlers.NewHandlersContainer(extensionsMock)
				recorder       = httptest.NewRecorder()
				request        = newNodeKindRequestWithID(t, tt.rawID)
			)

			if tt.setupMock != nil {
				tt.setupMock(extensionsMock)
			}

			handlerSet.GetNodeKindByID(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			require.NotNil(t, tt.assertBody, "every case must assert the response body")
			tt.assertBody(t, recorder.Body.Bytes())
		})
	}
}
