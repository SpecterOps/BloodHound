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
package opengraphschema

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestDiffMapsToSyncActions(t *testing.T) {
	var (
		kind1 = model.GraphSchemaNodeKind{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:              "Kind1",
			SchemaExtensionId: 1,
			DisplayName:       "Kind 1",
			Description:       "Test Kind 1",
			IsDisplayKind:     false,
			Icon:              "icon",
			IconColor:         "blue",
		}
		kind2 = model.GraphSchemaNodeKind{
			Serial: model.Serial{
				ID: 2,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:              "Kind2",
			SchemaExtensionId: 1,
			DisplayName:       "Kind 2",
			Description:       "Test Kind 2",
			IsDisplayKind:     false,
			Icon:              "icon",
			IconColor:         "blue",
		}
		updatedKind1 = model.GraphSchemaNodeKind{
			Serial: model.Serial{
				ID: kind1.ID,
				Basic: model.Basic{
					CreatedAt: kind1.CreatedAt,
					UpdatedAt: kind1.UpdatedAt,
				},
			},
			Name:              "Kind1",
			SchemaExtensionId: kind1.SchemaExtensionId,
			DisplayName:       "Updated Kind 1",
			Description:       "Test Updated Kind 1",
			IsDisplayKind:     false,
			Icon:              "icon",
			IconColor:         "blue",
		}
		kind3 = model.GraphSchemaNodeKind{
			Name:              "Kind3",
			SchemaExtensionId: 0,
			DisplayName:       "Kind 3",
			Description:       "Test Kind 3",
			IsDisplayKind:     false,
			Icon:              "icon",
			IconColor:         "blue",
		}
	)

	type args[K comparable, V any] struct {
		dst      map[K]V
		src      map[K]V
		onUpsert func(*V, *V)
	}
	type testCase[K comparable, V any] struct {
		name string
		args args[K, V]
		want MapDiffActions[V]
	}
	tests := []testCase[string, model.GraphSchemaNodeKind]{
		{
			name: "convertGraphSchemaNodeKinds",
			args: args[string, model.GraphSchemaNodeKind]{
				dst: map[string]model.GraphSchemaNodeKind{
					kind1.Name: kind1,
					kind2.Name: kind2,
				},
				src: map[string]model.GraphSchemaNodeKind{
					updatedKind1.Name: updatedKind1,
					kind3.Name:        kind3,
				},
				onUpsert: convertGraphSchemaNodeKinds,
			},
			want: MapDiffActions[model.GraphSchemaNodeKind]{
				ItemsToDelete: model.GraphSchemaNodeKinds{kind2},
				ItemsToUpsert: model.GraphSchemaNodeKinds{updatedKind1, kind3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateMapSynchronizationDiffActions(tt.args.dst, tt.args.src, tt.args.onUpsert)
			compareGraphSchemaNodeKinds(t, got.ItemsToUpsert, tt.want.ItemsToUpsert)
			compareGraphSchemaNodeKinds(t, got.ItemsToDelete, tt.want.ItemsToDelete)
		})
	}
}

// compareGraphSchemaNodeKinds - compares the returned list of model.GraphSchemaNodeKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaNodeKinds(t *testing.T, got, want model.GraphSchemaNodeKinds) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaNodeKinds")
	for i, schemaNodeKind := range got {
		compareGraphSchemaNodeKind(t, schemaNodeKind, want[i])
	}
}

func compareGraphSchemaNodeKind(t *testing.T, got, want model.GraphSchemaNodeKind) {
	t.Helper()
	require.Equalf(t, got.ID, want.ID, "GraphSchemaNodeKinds - ID mismatch - got: %v, want: %v", got.ID, want.ID)
	require.Equalf(t, want.Name, got.Name, "GraphSchemaNodeKind - name mismatch - got: %v, want: %v", got.Name, want.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaNodeKind(%v) - extension_id mismatch", got.SchemaExtensionId)
	require.Equalf(t, want.DisplayName, got.DisplayName, "GraphSchemaNodeKind(%v) - display_name mismatch", got.DisplayName)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaNodeKind(%v) - description mismatch", got.Description)
	require.Equalf(t, want.IsDisplayKind, got.IsDisplayKind, "GraphSchemaNodeKind(%v) - is_display_kind mismatch", got.IsDisplayKind)
	require.Equalf(t, want.Icon, got.Icon, "GraphSchemaNodeKind(%v) - icon mismatch", got.Icon)
	require.Equalf(t, want.IconColor, got.IconColor, "GraphSchemaNodeKind(%v) - icon_color mismatch", got.IconColor)
	require.Equalf(t, want.CreatedAt, got.CreatedAt, "GraphSchemaNodeKind(%v) - created_at mismatch", got.CreatedAt.IsZero())
	require.Equalf(t, want.UpdatedAt, got.UpdatedAt, "GraphSchemaNodeKind(%v) - updated_at mismatch", got.UpdatedAt.IsZero())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaNodeKind(%v) - deleted_at is not null", got.DeletedAt.Valid)
}
