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

//go:build integration

package database_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestDatabase_GetGraphSchemaByExtensionName(t *testing.T) {
	t.Parallel()

	testSuite := setupIntegrationTestSuite(t)

	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		err               error
		testExtensionName = "test_extension"
		testExtension     = model.GraphSchemaExtension{
			Name:        testExtensionName,
			Version:     "1.0.0",
			IsBuiltin:   false,
			DisplayName: "Test Extension",
		}
		nodeKind1 = model.GraphSchemaNodeKind{
			Name:              "Test_Node_Kind_1",
			SchemaExtensionId: testExtension.ID,
			DisplayName:       "Test Node Kind 1",
			Description:       "a test node kind",
			IsDisplayKind:     true,
			Icon:              "user",
			IconColor:         "blue",
		}
		edgeKind1 = model.GraphSchemaEdgeKind{
			SchemaExtensionId: testExtension.ID,
			Name:              "Test_Edge_Kind_1",
			Description:       "Test Edge Kind 1",
			IsTraversable:     true,
		}
		property1 = model.GraphSchemaProperty{
			SchemaExtensionId: testExtension.ID,
			Name:              "Test_Property_1",
			DisplayName:       "Test Property 1",
			DataType:          "string",
			Description:       "Test Property 1",
		}
	)

	type fields struct {
		setup    func(t *testing.T) model.GraphSchema
		teardown func(t *testing.T)
	}
	type args struct {
		ctx           context.Context
		extensionName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{

		{
			name: "success - existing but empty GraphSchemaExtension", // schema extension exists but there are no nodes, edges or properties linked to the extension
			fields: fields{
				setup: func(t *testing.T) model.GraphSchema {
					testExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version)
					require.NoError(t, err)
					return model.GraphSchema{
						GraphSchemaExtension: testExtension,
					}
				},
				teardown: func(t *testing.T) {
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, testExtension.ID)
					require.NoError(t, err)
				},
			},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtension.Name,
			},
		},

		{
			name: "success - GraphSchemaExtension",
			fields: fields{
				setup: func(t *testing.T) model.GraphSchema {
					testExtension, err = testSuite.BHDatabase.CreateGraphSchemaExtension(testSuite.Context, testExtension.Name, testExtension.DisplayName,
						testExtension.Version)
					require.NoError(t, err)
					nodeKind1, err = testSuite.BHDatabase.CreateGraphSchemaNodeKind(testSuite.Context, nodeKind1.Name,
						testExtension.ID, nodeKind1.DisplayName, nodeKind1.Description, nodeKind1.IsDisplayKind,
						nodeKind1.Icon, nodeKind1.IconColor)
					require.NoError(t, err)
					edgeKind1, err = testSuite.BHDatabase.CreateGraphSchemaEdgeKind(testSuite.Context, edgeKind1.Name,
						testExtension.ID, edgeKind1.Description, edgeKind1.IsTraversable)
					require.NoError(t, err)
					property1, err = testSuite.BHDatabase.CreateGraphSchemaProperty(testSuite.Context, testExtension.ID,
						property1.Name, property1.DisplayName, property1.DataType, property1.Description)
					require.NoError(t, err)

					return model.GraphSchema{
						GraphSchemaExtension:  testExtension,
						GraphSchemaNodeKinds:  model.GraphSchemaNodeKinds{nodeKind1},
						GraphSchemaEdgeKinds:  model.GraphSchemaEdgeKinds{edgeKind1},
						GraphSchemaProperties: model.GraphSchemaProperties{property1},
					}
				},
				teardown: func(t *testing.T) {
					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, testExtension.ID)
					require.NoError(t, err)
				},
			},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
		},
		{
			name: "success - no GetGraphSchemaExtensions results", // Will result in new graph schema extension
			fields: fields{
				setup: func(t *testing.T) model.GraphSchema {
					return model.GraphSchema{}
				},
				teardown: func(t *testing.T) {},
			},
			args: args{
				ctx:           context.Background(),
				extensionName: "non_existing_extension",
			},
			wantErr: database.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.fields.setup(t)
			if got, err := testSuite.BHDatabase.GetGraphSchemaByExtensionName(tt.args.ctx, tt.args.extensionName); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
				return
			} else {
				require.NoError(t, err)
				compareGraphSchema(t, got, want)
				tt.fields.teardown(t)
			}
		})
	}
}

func TestBloodhoundDB_UpsertGraphSchemaExtension(t *testing.T) {
	t.Parallel()
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	var (
		err            error
		got            bool
		gotGraphSchema model.GraphSchema

		testExtensionName = "Upsert_New_Test_Extension"
		testExtension     = model.GraphSchemaExtension{
			Name:        testExtensionName,
			Version:     "1.0.0",
			DisplayName: "Test Extension",
		}
		nodeKind1 = model.GraphSchemaNodeKind{
			Name:              "Upsert_New_Test_Node_Kind_1",
			SchemaExtensionId: testExtension.ID,
			DisplayName:       "Test Node Kind 1",
			Description:       "a test node kind",
			IsDisplayKind:     true,
			Icon:              "user",
			IconColor:         "blue",
		}
		edgeKind1 = model.GraphSchemaEdgeKind{
			SchemaExtensionId: testExtension.ID,
			Name:              "Upsert_New_Test_Edge_Kind_1",
			Description:       "Test Edge Kind 1",
			IsTraversable:     true,
		}
		property1 = model.GraphSchemaProperty{
			SchemaExtensionId: testExtension.ID,
			Name:              "Upsert_New_Test_Property_1",
			DisplayName:       "Test Property 1",
			DataType:          "string",
			Description:       "Test Property 1",
		}
		testGraphSchema = model.GraphSchema{
			GraphSchemaExtension:  testExtension,
			GraphSchemaNodeKinds:  model.GraphSchemaNodeKinds{nodeKind1},
			GraphSchemaEdgeKinds:  model.GraphSchemaEdgeKinds{edgeKind1},
			GraphSchemaProperties: model.GraphSchemaProperties{property1},
		}
	)

	type fields struct {
		setup    func(t *testing.T)
		teardown func(t *testing.T)
	}

	type args struct {
		ctx         context.Context
		graphSchema model.GraphSchema
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "success - create new OpenGraph extension",
			fields: fields{
				setup: func(t *testing.T) {},
				teardown: func(t *testing.T) {
					gotGraphSchema, err = testSuite.BHDatabase.GetGraphSchemaByExtensionName(testSuite.Context, testExtensionName)
					require.NoError(t, err)
					compareGraphSchema(t, gotGraphSchema, testGraphSchema)

					err = testSuite.BHDatabase.DeleteGraphSchemaExtension(testSuite.Context, gotGraphSchema.GraphSchemaExtension.ID)
					require.NoError(t, err)

					_, err = testSuite.BHDatabase.GetGraphSchemaByExtensionName(testSuite.Context, testExtension.Name)
					require.Equal(t, database.ErrNotFound, err)
				},
			},
			args: args{
				ctx:         testSuite.Context,
				graphSchema: testGraphSchema,
			},
			wantErr: nil,
			want:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.setup(t)

			if got, err = testSuite.BHDatabase.UpsertOpenGraphExtension(tt.args.ctx, tt.args.graphSchema); tt.wantErr != nil {
				require.EqualError(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
				require.Equalf(t, tt.want, got, "UpsertOpenGraphExtension(%v, %v)", tt.args.ctx, tt.args.graphSchema)
			}
			tt.fields.teardown(t)
		})
	}
}

func compareGraphSchemaExtension(t *testing.T, got, want model.GraphSchemaExtension) {
	t.Helper()
	require.Greaterf(t, got.ID, int32(0), "GraphSchemaExtension - ID mismatch - got: %v", got.ID)
	require.Equalf(t, want.Name, got.Name, "GraphSchemaExtension - name mismatch - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.DisplayName, got.DisplayName, "GraphSchemaExtension - display_name mismatch - got %v, want %v", got.DisplayName, want.DisplayName)
	require.Equalf(t, want.Version, got.Version, "GraphSchemaExtension - version mismatch - got %v, want %v", got.Version, want.Version)
	require.Equalf(t, want.IsBuiltin, got.IsBuiltin, "GraphSchemaExtension - is_built mismatch - got %t, want %t", got.IsBuiltin, want.IsBuiltin)
	require.Equalf(t, false, got.CreatedAt.IsZero(), "GraphSchemaExtension - created_at mismatch - got: %s", got.CreatedAt.String())
	require.Equalf(t, false, got.UpdatedAt.IsZero(), "GraphSchemaExtension - updated_at mismatch - got: %s", got.UpdatedAt.String())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaExtension - deleted_at is not null")
}

func compareGraphSchema(t *testing.T, got, want model.GraphSchema) {
	t.Helper()
	compareGraphSchemaExtension(t, got.GraphSchemaExtension, want.GraphSchemaExtension)
	compareGraphSchemaNodeKinds(t, got.GraphSchemaNodeKinds, want.GraphSchemaNodeKinds)
	compareGraphSchemaEdgeKinds(t, got.GraphSchemaEdgeKinds, want.GraphSchemaEdgeKinds)
	compareGraphSchemaProperties(t, got.GraphSchemaProperties, want.GraphSchemaProperties)
}
