// Copyright 2026 Specter Ops, Inc.
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

package handlers_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// newNodeRequestWithID builds a GET request whose mux path variables carry the
// supplied raw node id, mirroring what the router does at runtime.
func newNodeRequestWithID(t *testing.T, rawID string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "/api/v2/nodes/"+rawID, nil)
	require.NoError(t, err)
	return mux.SetURLVars(request, map[string]string{handlers.URIPathVariableNodeID: rawID})
}

// int32Ptr is a helper function that returns a pointer to an int32 value.
func int32Ptr(v int32) *int32 {
	return &v
}

func TestHandlers_GetNodeByID(t *testing.T) {
	var (
		nodeID = int64(9876543210)
		node   = services.Node{
			ID: nodeID,
			Kinds: []services.Kind{
				{ID: int32Ptr(1), Name: "User"},
				{ID: int32Ptr(2), Name: "Group"},
			},
			Properties: map[string]any{"name": "admin"},
			KindInfos: []services.KindInfo{
				{
					InfoKey:    "overview",
					Title:      "Overview",
					Position:   0,
					NodeKindID: int32Ptr(1),
					Content:    json.RawMessage(`{"markdown":{"content":"one"}}`),
				},
			},
		}
	)

	tests := []struct {
		name       string
		rawID      string
		setupMock  func(graphDBMock *mocks.MockGraphDB)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name:  "returns 200 with the node view on success",
			rawID: "9876543210",
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().GetNode(mock.Anything, nodeID, true).Return(node, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.NodeView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, nodeID, envelope.Data.NodeID)
				assert.Len(t, envelope.Data.Kinds, 2)
				require.NotNil(t, envelope.Data.Kinds[0].NodeKindID)
				assert.Equal(t, int32(1), *envelope.Data.Kinds[0].NodeKindID)
				assert.Equal(t, "User", envelope.Data.Kinds[0].Name)
				require.NotNil(t, envelope.Data.Kinds[1].NodeKindID)
				assert.Equal(t, int32(2), *envelope.Data.Kinds[1].NodeKindID)
				assert.Equal(t, "Group", envelope.Data.Kinds[1].Name)
				assert.Equal(t, "admin", envelope.Data.Properties["name"])
				require.Len(t, envelope.Data.KindInfos, 1)
				assert.Equal(t, "one", envelope.Data.KindInfos[0].Markdown.Content)
				assert.NotContains(t, string(body), `\"markdown\"`)
			},
		},
		{
			name:       "returns 400 when the id is malformed",
			rawID:      "not-a-number",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "returns 404 when the node is not found",
			rawID: "9876543210",
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().GetNode(mock.Anything, nodeID, true).Return(services.Node{}, services.ErrNodeNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name:  "returns 404 when the kind is not found",
			rawID: "9876543210",
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().GetNode(mock.Anything, nodeID, true).Return(services.Node{}, services.ErrKindNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				graphDBMock = mocks.NewMockGraphDB(t)
				handlerSet  = handlers.NewHandlersContainer(graphDBMock)
				recorder    = httptest.NewRecorder()
				request     = newNodeRequestWithID(t, tt.rawID)
			)

			if tt.setupMock != nil {
				tt.setupMock(graphDBMock)
			}

			handlerSet.GetNodeByID(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}
