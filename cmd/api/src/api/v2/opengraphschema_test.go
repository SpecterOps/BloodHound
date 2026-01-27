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

package v2_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	uuid2 "github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestResources_OpenGraphSchemaIngest(t *testing.T) {

	var (
		mockCtrl             = gomock.NewController(t)
		mockOpenGraphService = mocks.NewMockOpenGraphSchemaService(mockCtrl)
		userId, err          = uuid2.NewV4()

		graphExtension = v2.GraphExtensionPayload{
			GraphSchemaExtension: v2.GraphSchemaExtensionPayload{
				Name:        "Test_Extension",
				DisplayName: "Test Extension",
				Version:     "1.0.0",
			},
			GraphSchemaProperties: []v2.GraphSchemaPropertiesPayload{
				{
					Name:        "Property_1",
					DisplayName: "Property 1",
					DataType:    "string",
					Description: "a property",
				},
			},
			GraphSchemaEdgeKinds: []v2.GraphSchemaEdgeKindsPayload{
				{
					Name:          "GraphSchemaEdgeKind_1",
					Description:   "GraphSchemaRelationshipKind 1",
					IsTraversable: true,
				},
			},
			GraphSchemaNodeKinds: []v2.GraphSchemaNodeKindsPayload{
				{
					Name:          "GraphSchemaNodeKind_1",
					DisplayName:   "GraphSchemaNodeKind 1",
					Description:   "a graph schema node",
					IsDisplayKind: true,
					Icon:          "User",
					IconColor:     "blue",
				},
			},
			GraphEnvironments: []v2.EnvironmentPayload{
				{
					EnvironmentKind: "EnvironmentInput",
					SourceKind:      "Source_Kind_1",
					PrincipalKinds:  []string{"User"},
				},
			},
			GraphFinding: []v2.FindingsPayload{
				{
					Name:             "Finding_1",
					DisplayName:      "Finding 1",
					SourceKind:       "Source_Kind_1",
					RelationshipKind: "GraphSchemaEdgeKind_1",
					EnvironmentKind:  "EnvironmentInput",
					Remediation: v2.RemediationPayload{
						ShortDescription: "remediation for Finding_1",
						LongDescription:  "a remediation for Finding 1",
						ShortRemediation: "do x",
						LongRemediation:  "do x but better",
					},
				},
			},
		}
		serviceGraphExtension = model.GraphExtensionInput{
			ExtensionInput: model.ExtensionInput{
				Name:        "Test_Extension",
				DisplayName: "Test Extension",
				Version:     "1.0.0",
			},
			PropertiesInput: model.PropertiesInput{
				{
					Name:        "Property_1",
					DisplayName: "Property 1",
					DataType:    "string",
					Description: "a property",
				},
			},
			RelationshipKindsInput: model.RelationshipsInput{
				{
					Name:          "GraphSchemaEdgeKind_1",
					Description:   "GraphSchemaRelationshipKind 1",
					IsTraversable: true,
				},
			},
			NodeKindsInput: model.NodesInput{
				{
					Name:          "GraphSchemaNodeKind_1",
					DisplayName:   "GraphSchemaNodeKind 1",
					Description:   "a graph schema node",
					IsDisplayKind: true,
					Icon:          "User",
					IconColor:     "blue",
				},
			},
			EnvironmentsInput: []model.EnvironmentInput{
				{
					EnvironmentKind: "EnvironmentInput",
					SourceKind:      "Source_Kind_1",
					PrincipalKinds:  []string{"User"},
				},
			},
			FindingsInput: []model.FindingInput{
				{
					Name:             "Finding_1",
					DisplayName:      "Finding 1",
					SourceKind:       "Source_Kind_1",
					RelationshipKind: "GraphSchemaEdgeKind_1",
					EnvironmentKind:  "EnvironmentInput",
					Remediation: model.RemediationInput{
						ShortDescription: "remediation for Finding_1",
						LongDescription:  "a remediation for Finding 1",
						ShortRemediation: "do x",
						LongRemediation:  "do x but better",
					},
				},
			},
		}
	)
	defer mockCtrl.Finish()
	require.NoError(t, err)

	type fields struct {
		setupOpenGraphServiceMock func(t *testing.T, repository *mocks.MockOpenGraphSchemaService)
	}
	type args struct {
		buildRequest func() *http.Request
	}
	type want struct {
		responseCode int
		err          error
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   want
	}{
		{
			name: "fail - no user",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *mocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequest(http.MethodPut, "/api/v2/extensions", nil)
					require.NoError(t, err)
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: No associated user found"),
			},
		},
		{
			name: "fail - user is not admin",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *mocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithOwnerId(userId), http.MethodPut, "/api/v2/extensions", nil)
					require.NoError(t, err)
					return req
				},
			},
			want: want{
				responseCode: http.StatusForbidden,
				err:          fmt.Errorf("Code: 403 - errors: user does not have sufficient permissions to create or update an extension"),
			},
		},
		{
			name: "fail - open graph extension payload cannot be empty",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *mocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut, "/api/v2/extensions", nil)
					require.NoError(t, err)
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: open graph extension payload cannot be empty"),
			},
		},
		{
			name: "fail - invalid content type",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *mocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					var body []byte
					body, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut, "/api/v2/extensions", bytes.NewReader(body))
					require.NoError(t, err)
					req.Header.Set("content-type", "invalid")
					return req
				},
			},
			want: want{
				responseCode: http.StatusUnsupportedMediaType,
				err:          fmt.Errorf("Code: 415 - errors: invalid content-type: [invalid]; Content type must be application/json"),
			},
		},
		{
			name: "fail - unable to decode graph schema",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, repository *mocks.MockOpenGraphSchemaService) {},
			},
			args: args{
				func() *http.Request {
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader([]byte("awfawf")))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: unable to decode graph extension payload: invalid character 'a' looking for beginning of value"),
			},
		},
		{
			name: "fail - UpsertOpenGraphExtension - generic error",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *mocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, fmt.Errorf("generic error"))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusInternalServerError,
				err:          fmt.Errorf("Code: 500 - errors: an internal error has occurred that is preventing the service from servicing this request"),
			},
		},
		{
			name: "fail - UpsertOpenGraphExtension - validation error",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *mocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, fmt.Errorf("%w: some_error", model.ErrGraphExtensionValidation))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: %w: some_error", model.ErrGraphExtensionValidation),
			},
		},
		{
			name: "fail - UpsertOpenGraphExtension - cannot modify a built-in graph extension",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *mocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false,
						fmt.Errorf("Error upserting graph extension: %w", model.ErrGraphExtensionBuiltIn))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusBadRequest,
				err:          fmt.Errorf("Code: 400 - errors: Error upserting graph extension: %w", model.ErrGraphExtensionBuiltIn),
			},
		},
		{
			name: "fail - unable to refresh graph db kinds",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *mocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, fmt.Errorf("%w: graph_db error", model.ErrGraphDBRefreshKinds))
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusInternalServerError,
				err:          fmt.Errorf("Code: 500 - errors: an internal error has occurred that is preventing the service from servicing this request"),
			},
		},
		{
			name: "success - updated graph extension",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *mocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(true, nil)
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusOK,
			},
		},
		{
			name: "success - inserted new graph extension",
			fields: fields{
				setupOpenGraphServiceMock: func(t *testing.T, mock *mocks.MockOpenGraphSchemaService) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), serviceGraphExtension).Return(false, nil)
				},
			},
			args: args{
				func() *http.Request {
					var jsonPayload []byte
					jsonPayload, err = json.Marshal(graphExtension)
					require.NoError(t, err)
					req, err := http.NewRequestWithContext(createContextWithAdminOwnerId(userId), http.MethodPut,
						"/api/v2/extensions", bytes.NewReader(jsonPayload))
					require.NoError(t, err)
					req.Header.Set("content-type", mediatypes.ApplicationJson.String())
					return req
				},
			},
			want: want{
				responseCode: http.StatusCreated,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				response = httptest.NewRecorder()
				request  = tt.args.buildRequest()
			)
			tt.fields.setupOpenGraphServiceMock(t, mockOpenGraphService)

			s := v2.Resources{
				OpenGraphSchemaService: mockOpenGraphService,
			}

			s.OpenGraphSchemaIngest(response, request)
			require.Equal(t, tt.want.responseCode, response.Code) // If success, only a 200 or 201 header response is made
			if tt.want.responseCode != http.StatusOK && tt.want.responseCode != http.StatusCreated {
				// Failure responses
				var errWrapper api.ErrorWrapper
				err := json.Unmarshal(response.Body.Bytes(), &errWrapper)
				require.NoError(t, err)
				require.EqualErrorf(t, errWrapper, tt.want.err.Error(), "Unexpected error: %v", errWrapper.Error())
			}
		})
	}
}

func Test_convertGraphExtensionPayloadToGraphExtension(t *testing.T) {
	type args struct {
		payload v2.GraphExtensionPayload
	}
	tests := []struct {
		name string
		args args
		want model.GraphExtensionInput
	}{
		{
			name: "success",
			args: args{
				payload: v2.GraphExtensionPayload{
					GraphSchemaExtension: v2.GraphSchemaExtensionPayload{
						Name:        "Test_Extension",
						DisplayName: "Test Extension",
						Version:     "1.0.0",
					},
					GraphSchemaProperties: []v2.GraphSchemaPropertiesPayload{
						{
							Name:        "Property_1",
							DisplayName: "Property 1",
							DataType:    "string",
							Description: "a property",
						},
					},
					GraphSchemaEdgeKinds: []v2.GraphSchemaEdgeKindsPayload{
						{
							Name:          "GraphSchemaEdgeKind_1",
							Description:   "GraphSchemaRelationshipKind 1",
							IsTraversable: true,
						},
					},
					GraphSchemaNodeKinds: []v2.GraphSchemaNodeKindsPayload{
						{
							Name:          "GraphSchemaNodeKind_1",
							DisplayName:   "GraphSchemaNodeKind 1",
							Description:   "a graph schema node",
							IsDisplayKind: true,
							Icon:          "User",
							IconColor:     "blue",
						},
					},
					GraphEnvironments: []v2.EnvironmentPayload{
						{
							EnvironmentKind: "EnvironmentInput",
							SourceKind:      "Source_Kind_1",
							PrincipalKinds:  []string{"User"},
						},
					},
					GraphFinding: []v2.FindingsPayload{
						{
							Name:             "Finding_1",
							DisplayName:      "Finding 1",
							SourceKind:       "Source_Kind_1",
							RelationshipKind: "GraphSchemaEdgeKind_1",
							EnvironmentKind:  "EnvironmentInput",
							Remediation: v2.RemediationPayload{
								ShortDescription: "remediation for Finding_1",
								LongDescription:  "a remediation for Finding 1",
								ShortRemediation: "do x",
								LongRemediation:  "do x but better",
							},
						},
					},
				},
			},
			want: model.GraphExtensionInput{
				ExtensionInput: model.ExtensionInput{
					Name:        "Test_Extension",
					DisplayName: "Test Extension",
					Version:     "1.0.0",
				},
				PropertiesInput: model.PropertiesInput{
					{
						Name:        "Property_1",
						DisplayName: "Property 1",
						DataType:    "string",
						Description: "a property",
					},
				},
				RelationshipKindsInput: model.RelationshipsInput{
					{
						Name:          "GraphSchemaEdgeKind_1",
						Description:   "GraphSchemaRelationshipKind 1",
						IsTraversable: true,
					},
				},
				NodeKindsInput: model.NodesInput{
					{
						Name:          "GraphSchemaNodeKind_1",
						DisplayName:   "GraphSchemaNodeKind 1",
						Description:   "a graph schema node",
						IsDisplayKind: true,
						Icon:          "User",
						IconColor:     "blue",
					},
				},
				EnvironmentsInput: []model.EnvironmentInput{
					{
						EnvironmentKind: "EnvironmentInput",
						SourceKind:      "Source_Kind_1",
						PrincipalKinds:  []string{"User"},
					},
				},
				FindingsInput: []model.FindingInput{
					{
						Name:             "Finding_1",
						DisplayName:      "Finding 1",
						SourceKind:       "Source_Kind_1",
						RelationshipKind: "GraphSchemaEdgeKind_1",
						EnvironmentKind:  "EnvironmentInput",
						Remediation: model.RemediationInput{
							ShortDescription: "remediation for Finding_1",
							LongDescription:  "a remediation for Finding 1",
							ShortRemediation: "do x",
							LongRemediation:  "do x but better",
						},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, v2.ConvertGraphExtensionPayloadToGraphExtension(tt.args.payload), "ConvertGraphExtensionPayloadToGraphExtension(%v)", tt.args.payload)
		})
	}
}
