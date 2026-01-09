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
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestDatabase_getGraphSchemaByExtensionName(t *testing.T) {
	var (
		testExtensionName = "test_extension"
		testExtension     = model.GraphSchemaExtension{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:        testExtensionName,
			Version:     "1.0.0",
			IsBuiltin:   true,
			DisplayName: "Test Extension",
		}
		testNodeKind1 = model.GraphSchemaNodeKind{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:              "Test_Node_Kind_1",
			SchemaExtensionId: testExtension.ID,
			DisplayName:       "Test Node Kind 1",
			Description:       "a test node kind",
			IsDisplayKind:     true,
			Icon:              "user",
			IconColor:         "blue",
		}
		testEdgeKind1 = model.GraphSchemaEdgeKind{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			SchemaExtensionId: testExtension.ID,
			Name:              "Test_Edge_Kind_1",
			Description:       "Test Edge Kind 1",
			IsTraversable:     true,
		}
		testProperty1 = model.GraphSchemaProperty{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			SchemaExtensionId: testExtension.ID,
			Name:              "Test_Property_1",
			DisplayName:       "Test Property 1",
			DataType:          "string",
			Description:       "Test Property 1",
		}
	)

	type fields struct {
	}
	type args struct {
		ctx           context.Context
		extensionName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    model.GraphSchema
		wantErr error
	}{
		{
			name:   "fail - GetGraphSchemaExtensions error",
			fields: fields{},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: fmt.Errorf("GetGraphSchemaExtensions error"),
		},
		{
			name:   "fail - GetGraphSchemaNodeKinds error",
			fields: fields{},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: fmt.Errorf("GetGraphSchemaNodeKinds error"),
		},
		{
			name:   "fail - GetGraphSchemaEdgeKinds error",
			fields: fields{},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: fmt.Errorf("GetGraphSchemaEdgeKinds error"),
		},
		{
			name:   "fail - GetGraphSchemaEdgeKinds error",
			fields: fields{},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: fmt.Errorf("GetGraphSchemaProperties error"),
		},
		{
			name:   "success - existing but empty GraphSchemaExtension", // schema extension exists but there are no nodes, edges or properties linked to the extension
			fields: fields{},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want: model.GraphSchema{
				GraphSchemaExtension: testExtension,
			},
			wantErr: nil,
		},
		{
			name:   "success - GraphSchemaExtension",
			fields: fields{},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want: model.GraphSchema{
				GraphSchemaExtension:  testExtension,
				GraphSchemaNodeKinds:  model.GraphSchemaNodeKinds{testNodeKind1},
				GraphSchemaProperties: model.GraphSchemaProperties{testProperty1},
				GraphSchemaEdgeKinds:  model.GraphSchemaEdgeKinds{testEdgeKind1},
			},
			wantErr: nil,
		},
		{
			name:   "success - no GetGraphSchemaExtensions results", // Will result in new graph schema extension
			fields: fields{},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			s := &database.BloodhoundDB{}
			got, err := s.getGraphSchemaByExtensionName(tt.args.ctx, tt.args.extensionName)
			if tt.wantErr != nil {
				require.EqualErrorf(t, err, tt.wantErr.Error(), "getGraphSchemaByExtensionName() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoError(t, err)
				compareGraphSchema(t, got, tt.want)
			}
		})
	}
}

func compareGraphSchemaExtension(t *testing.T, got, want model.GraphSchemaExtension) {
	t.Helper()
	require.Equalf(t, got.ID, want.ID, "GraphSchemaExtension - ID mismatch - got: %v, want: %v", got.ID, want.ID)
	require.Equalf(t, got.Name, want.Name, "GraphSchemaExtension - name mismatch - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, got.DisplayName, want.DisplayName, "GraphSchemaExtension - display_name mismatch - got %v, want %v", got.DisplayName, want.DisplayName)
	require.Equalf(t, got.Version, want.Version, "GraphSchemaExtension - version mismatch - got %v, want %v", got.Version, want.Version)
	require.Equalf(t, got.IsBuiltin, want.IsBuiltin, "GraphSchemaExtension - is_built mismatch - got %t, want %t", got.IsBuiltin, want.IsBuiltin)
	require.Equalf(t, want.CreatedAt, got.CreatedAt, "GraphSchemaExtension - created_at mismatch - got: %s, mapSyncActionsWant: %s", got.CreatedAt.String(), want.CreatedAt.String())
	require.Equalf(t, want.UpdatedAt, got.UpdatedAt, "GraphSchemaExtension - updated_at mismatch - got: %s, mapSyncActionsWant: %s", got.UpdatedAt.String(), want.UpdatedAt.String())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaExtension - deleted_at is not null")
}

func compareGraphSchema(t *testing.T, got, want model.GraphSchema) {
	t.Helper()
	compareGraphSchemaExtension(t, got.GraphSchemaExtension, want.GraphSchemaExtension)
	compareGraphSchemaNodeKinds(t, got.GraphSchemaNodeKinds, want.GraphSchemaNodeKinds)
	compareGraphSchemaEdgeKinds(t, got.GraphSchemaEdgeKinds, want.GraphSchemaEdgeKinds)
	compareGraphSchemaProperties(t, got.GraphSchemaProperties, want.GraphSchemaProperties)
}
