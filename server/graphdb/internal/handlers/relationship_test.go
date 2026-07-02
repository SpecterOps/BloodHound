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

// newRequestWithID builds a GET request whose mux path variables carry the
// supplied raw relationship id, mirroring what the router does at runtime.
func newRequestWithID(t *testing.T, rawID string) *http.Request {
	t.Helper()
	request, err := http.NewRequest(http.MethodGet, "/api/v2/relationships/"+rawID, nil)
	require.NoError(t, err)
	return mux.SetURLVars(request, map[string]string{handlers.URIPathVariableRelationshipID: rawID})
}

func TestHandlers_GetRelationshipByID(t *testing.T) {
	var (
		relationshipID = int64(1234567890)
		kindID         = int32(42)
		relationship   = services.Relationship{
			ID:           relationshipID,
			SourceNodeID: 100,
			TargetNodeID: 200,
			Kind:         services.Kind{ID: &kindID, Name: "MemberOf"},
			Properties:   map[string]any{"foo": "bar"},
		}
		nilKindRelationship = services.Relationship{
			ID:           relationshipID,
			SourceNodeID: 100,
			TargetNodeID: 200,
			Kind:         services.Kind{Name: "MemberOf"},
			Properties:   map[string]any{"foo": "bar"},
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
			name:  "returns 200 with the relationship view on success",
			rawID: "1234567890",
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().GetRelationship(mock.Anything, relationshipID).Return(relationship, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				var envelope struct {
					Data handlers.RelationshipView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Equal(t, relationshipID, envelope.Data.RelationshipID)
				assert.Equal(t, int64(100), envelope.Data.SourceNodeID)
				assert.Equal(t, int64(200), envelope.Data.TargetNodeID)
				require.NotNil(t, envelope.Data.Kind.RelationshipKindID)
				assert.Equal(t, int32(42), *envelope.Data.Kind.RelationshipKindID)
				assert.Equal(t, "MemberOf", envelope.Data.Kind.Name)
			},
		},
		{
			name:  "returns 200 with a null relationship kind id when the kind has no schema entry",
			rawID: "1234567890",
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().GetRelationship(mock.Anything, relationshipID).Return(nilKindRelationship, nil)
			},
			wantStatus: http.StatusOK,
			assertBody: func(t *testing.T, body []byte) {
				assert.Contains(t, string(body), `"relationship_kind_id":null`)

				var envelope struct {
					Data handlers.RelationshipView `json:"data"`
				}
				require.NoError(t, json.Unmarshal(body, &envelope))
				assert.Nil(t, envelope.Data.Kind.RelationshipKindID)
				assert.Equal(t, "MemberOf", envelope.Data.Kind.Name)
			},
		},
		{
			name:       "returns 400 when the id is malformed",
			rawID:      "not-a-number",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:  "returns 404 when the relationship is not found",
			rawID: "1234567890",
			setupMock: func(graphDBMock *mocks.MockGraphDB) {
				graphDBMock.EXPECT().GetRelationship(mock.Anything, relationshipID).Return(services.Relationship{}, services.ErrRelationshipNotFound)
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
				request     = newRequestWithID(t, tt.rawID)
			)

			if tt.setupMock != nil {
				tt.setupMock(graphDBMock)
			}

			handlerSet.GetRelationshipByID(recorder, request)

			assert.Equal(t, tt.wantStatus, recorder.Code)
			if tt.assertBody != nil {
				tt.assertBody(t, recorder.Body.Bytes())
			}
		})
	}
}
