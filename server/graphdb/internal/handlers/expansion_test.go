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
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers"
	"github.com/specterops/bloodhound/server/graphdb/internal/handlers/mocks"
	"github.com/specterops/bloodhound/server/graphdb/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func setupUserCtx(user model.User) context.Context {
	return context.WithValue(context.Background(), bhctx.ValueKey, &bhctx.Context{
		AuthCtx: auth.Context{
			PermissionOverrides: auth.PermissionOverrides{},
			Owner:               user,
			Session:             model.UserSession{},
		},
	})
}

func newExpandGraphRequest(t *testing.T, payload any) *http.Request {
	t.Helper()

	jsonPayload, err := json.Marshal(payload)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/v2/graphs/expand", bytes.NewReader(jsonPayload))
	request.Header.Set(headers.ContentType.String(), "application/json")
	return request.WithContext(setupUserCtx(model.User{AllEnvironments: true}))
}

func TestHandlers_ExpandGraph(t *testing.T) {
	var (
		nodeID = int64(42)
		user   = model.User{AllEnvironments: true}
	)

	tests := []struct {
		name       string
		request    *http.Request
		setupMock  func(graphDBMock *mocks.MockGraphDB)
		wantStatus int
		assertBody func(t *testing.T, body []byte)
	}{
		{
			name: "returns expansion response with property keys",
			request: newExpandGraphRequest(t, handlers.GraphExpansionPayload{
				NodeID:            &nodeID,
				Direction:         services.GraphExpansionDirectionOutbound,
				Limit:             1,
				IncludeProperties: true,
			}),
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().ExpandGraph(mock.Anything, services.GraphExpansionRequest{
					NodeID:            nodeID,
					Direction:         services.GraphExpansionDirectionOutbound,
					Limit:             1,
					IncludeProperties: true,
					User:              user,
				}).Return(services.GraphExpansion{
					Graph: model.UnifiedGraph{
						Nodes: map[string]model.UnifiedNode{
							"1": {Label: "one", Properties: map[string]any{"node_key": "value"}},
							"2": {Label: "two", Properties: map[string]any{"node_key": "value"}},
						},
						Edges: []model.UnifiedEdge{
							{Source: "1", Target: "2", Properties: map[string]any{"edge_key": "value"}},
						},
					},
					Limit:     1,
					Truncated: true,
				}, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				assert.JSONEq(t, `{"data":{"node_keys":["node_key"],"edge_keys":["edge_key"],"nodes":{"1":{"label":"one","kind":"","kinds":null,"objectId":"","isTierZero":false,"isOwnedObject":false,"lastSeen":"0001-01-01T00:00:00Z","properties":{"node_key":"value"}},"2":{"label":"two","kind":"","kinds":null,"objectId":"","isTierZero":false,"isOwnedObject":false,"lastSeen":"0001-01-01T00:00:00Z","properties":{"node_key":"value"}}},"edges":[{"id":"","source":"1","target":"2","label":"","kind":"","lastSeen":"0001-01-01T00:00:00Z","properties":{"edge_key":"value"}}],"literals":null,"limit":1,"truncated":true}}`, string(body))
			},
		},
		{
			name:       "returns bad request when node_id is missing",
			request:    newExpandGraphRequest(t, map[string]any{"direction": services.GraphExpansionDirectionOutbound}),
			wantStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "node_id is required")
			},
		},
		{
			name:       "returns bad request when node_id is negative",
			request:    newExpandGraphRequest(t, map[string]any{"node_id": -1, "direction": services.GraphExpansionDirectionOutbound}),
			wantStatus: http.StatusBadRequest,
			assertBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), "node_id must be greater than or equal to 0")
			},
		},
		{
			name: "returns not found when the anchor node does not exist",
			request: newExpandGraphRequest(t, handlers.GraphExpansionPayload{
				NodeID:    &nodeID,
				Direction: services.GraphExpansionDirectionOutbound,
			}),
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().ExpandGraph(mock.Anything, mock.Anything).Return(services.GraphExpansion{}, services.ErrNodeNotFound)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "returns forbidden when the anchor node is inaccessible",
			request: newExpandGraphRequest(t, handlers.GraphExpansionPayload{
				NodeID:    &nodeID,
				Direction: services.GraphExpansionDirectionOutbound,
			}),
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().ExpandGraph(mock.Anything, mock.Anything).Return(services.GraphExpansion{}, services.ErrNodeAccessDenied)
			},
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				graphDBMock = mocks.NewMockGraphDB(t)
				handlerSet  = handlers.NewHandlersContainer(graphDBMock)
				recorder    = httptest.NewRecorder()
			)

			if tt.setupMock != nil {
				tt.setupMock(graphDBMock)
			}

			handlerSet.ExpandGraph(recorder, tt.request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}
