// Copyright 2025 Specter Ops, Inc.
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
package opengraphschema_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema"
	schemamocks "github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOpenGraphSchemaService_UpsertSchemaEnvironment(t *testing.T) {
	type mocks struct {
		mockOpenGraphSchema *schemamocks.MockOpenGraphSchemaRepository
	}
	type args struct {
		graphSchema model.SchemaEnvironment
	}
	tests := []struct {
		name       string
		mocks      mocks
		setupMocks func(t *testing.T, mock *mocks)
		args       args
		expected   error
	}{
		{
			name: "Error: Validation - Failed to retrieve source kinds",
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{}, errors.New("error"))
			},
			expected: errors.New("error validating schema environment: error retrieving source kinds: error"),
		},
		{
			name: "Error: Validation - Environment Kind doesn't exist in Kinds table",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(1),
					SourceKindId:      int32(1),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
			},
			expected: errors.New("error validating schema environment: invalid environment kind id 1"),
		},
		{
			name: "Error: Validation - Source Kind doesn't exist in Kinds table",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(1),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
			},
			expected: errors.New("error validating schema environment: invalid source kind id 1"),
		},
		{
			name: "Error: Validation - Source Kind doesn't exist in Kinds table",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(1),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(1),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
			},
			expected: errors.New("error validating schema environment: invalid source kind id 1"),
		},
		{
			name: "Error: GetSchemaEnvironmentById err",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error retrieving schema environment id 0: error"),
		},
		{
			name: "Error: GetSchemaEnvironmentById success but error occurs when deleting schema env",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(0)).Return(errors.New("error"))
			},
			expected: errors.New("error deleting schema environment 0: error"),
		},
		{
			name: "Error: schema environment id found in database, env deleted successfully, but fails to recreate",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(0)).Return(nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error creating schema environment 0: error"),
		},
		{
			name: "Error: GetSchemaEnvironmentById returns database.ErrNotFound but error occurs when creating schema env",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error creating schema environment: error"),
		},
		{
			name: "Success: environment id exists, so delete and re-create",
			args: args{
				graphSchema: model.SchemaEnvironment{
					SchemaExtensionId: int32(3),
					EnvironmentKindId: int32(3),
					SourceKindId:      int32(3),
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKinds(gomock.Any()).Return([]database.SourceKind{
					{
						ID:   3,
						Name: graph.StringKind("kind"),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(0)).Return(nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{}, nil)
			},
			expected: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			mocks := &mocks{
				mockOpenGraphSchema: schemamocks.NewMockOpenGraphSchemaRepository(ctrl),
			}

			tt.setupMocks(t, mocks)

			graphService := opengraphschema.NewOpenGraphSchemaService(mocks.mockOpenGraphSchema)

			err := graphService.UpsertSchemaEnvironment(context.Background(), tt.args.graphSchema)
			if tt.expected != nil {
				assert.EqualError(t, tt.expected, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
