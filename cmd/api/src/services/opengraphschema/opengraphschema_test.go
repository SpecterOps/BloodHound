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
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestOpenGraphSchemaService_UpsertGraphSchemaExtension -
//
// Mocks:
//
// GenerateMapSynchronizationDiffActions does not preserve ordering so the following CRUD mocks
// perform a Do function which removes the provided kind/property from the item-function map.
// The last Do function for each CRUD Operation will check to see if their respective wantAction's
// map length is 0 to ensure all actions are accounted for.
func TestOpenGraphSchemaService_UpsertGraphSchemaExtension(t *testing.T) {
	t.Parallel()

	var (
		mockCtrl = gomock.NewController(t)

		mockOpenGraphSchemaRepository = mocks.NewMockOpenGraphSchemaRepository(mockCtrl)
		mockGraphDBKindsRepository    = mocks.NewMockGraphDBKindRepository(mockCtrl)

		existingExtension1 = model.GraphSchemaExtension{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:        "test_extension_1",
			DisplayName: "Test Extension 1",
			Version:     "1.0.0",
			IsBuiltin:   false,
		}

		builtInExtension = model.GraphSchemaExtension{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:        "test_extension_1",
			DisplayName: "Test Extension 1",
			Version:     "1.0.0",
			IsBuiltin:   true,
		}

		newExtension1 = model.GraphSchemaExtension{
			Name:        "test_extension_2",
			DisplayName: "Test Extension 2",
			Version:     "1.0.0",
			IsBuiltin:   false,
		}

		updatedExtension = model.GraphSchemaExtension{
			Name:        "test_extension_1",
			DisplayName: "Test Extension 1",
			Version:     "2.0.0",
			IsBuiltin:   false,
		}

		existingNodeKind1 = model.GraphSchemaNodeKind{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:              "node_kind_1",
			SchemaExtensionId: existingExtension1.ID,
			DisplayName:       "Node Kind 1",
			Description:       "a test node kind",
			IsDisplayKind:     true,
			Icon:              "desktop",
			IconColor:         "blue",
		}
		existingNodeKind2 = model.GraphSchemaNodeKind{
			Serial: model.Serial{
				ID: 2,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			Name:              "node_kind_2",
			SchemaExtensionId: existingExtension1.ID,
			DisplayName:       "Node Kind 2",
			Description:       "a test node kind",
			IsDisplayKind:     true,
			Icon:              "user",
			IconColor:         "red",
		}
		newNodeKind1 = model.GraphSchemaNodeKind{
			Name:          "new_node_kind_1",
			DisplayName:   "New Node Kind 1",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "desktop",
			IconColor:     "blue",
		}
		newNodeKind2 = model.GraphSchemaNodeKind{
			Name:          "new_node_kind_2",
			DisplayName:   "New Node Kind 2",
			Description:   "a test node kind",
			IsDisplayKind: true,
			Icon:          "user",
			IconColor:     "green",
		}

		existingEdgeKind1 = model.GraphSchemaEdgeKind{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			SchemaExtensionId: existingExtension1.ID,
			Name:              "edge_kind_1",
			Description:       "a test edge kind",
			IsTraversable:     true,
		}
		existingEdgeKind2 = model.GraphSchemaEdgeKind{
			Serial: model.Serial{
				ID: 2,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			SchemaExtensionId: existingExtension1.ID,
			Name:              "edge_kind_2",
			Description:       "a test edge kind",
			IsTraversable:     true,
		}
		newEdgeKind1 = model.GraphSchemaEdgeKind{
			Name:          "new_edge_kind_1",
			Description:   "a test edge kind",
			IsTraversable: true,
		}
		newEdgeKind2 = model.GraphSchemaEdgeKind{
			Name:          "new_edge_kind_2",
			Description:   "a test edge kind",
			IsTraversable: true,
		}
		existingProperty1 = model.GraphSchemaProperty{
			Serial: model.Serial{
				ID: 1,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			SchemaExtensionId: existingExtension1.ID,
			Name:              "property_1",
			DisplayName:       "Property 1",
			DataType:          "string",
			Description:       "a test property",
		}
		existingProperty2 = model.GraphSchemaProperty{
			Serial: model.Serial{
				ID: 2,
				Basic: model.Basic{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
			SchemaExtensionId: existingExtension1.ID,
			Name:              "property_2",
			DisplayName:       "Property 2",
			DataType:          "integer",
			Description:       "a test property",
		}
		newProperty1 = model.GraphSchemaProperty{
			Name:        "property_1",
			DisplayName: "Property 1",
			DataType:    "integer",
			Description: "a test property",
		}
		newProperty2 = model.GraphSchemaProperty{
			Name:        "property_2",
			DisplayName: "Property 2",
			DataType:    "string",
			Description: "a test property",
		}

		newGraphSchema = model.GraphSchema{
			GraphSchemaExtension:  newExtension1,
			GraphSchemaNodeKinds:  model.GraphSchemaNodeKinds{newNodeKind1, newNodeKind2},
			GraphSchemaEdgeKinds:  model.GraphSchemaEdgeKinds{newEdgeKind1, newEdgeKind2},
			GraphSchemaProperties: model.GraphSchemaProperties{newProperty1, newProperty2},
		}
		_ = model.GraphSchema{
			GraphSchemaExtension:  existingExtension1,
			GraphSchemaNodeKinds:  model.GraphSchemaNodeKinds{existingNodeKind1, existingNodeKind2},
			GraphSchemaEdgeKinds:  model.GraphSchemaEdgeKinds{existingEdgeKind1, existingEdgeKind2},
			GraphSchemaProperties: model.GraphSchemaProperties{existingProperty1, existingProperty2},
		}
	)

	defer mockCtrl.Finish()

	type fields struct {
		setupOpenGraphSchemaRepositoryMock func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository)
		setupGraphDBKindsRepositoryMock    func(t *testing.T, mock *mocks.MockGraphDBKindRepository)
	}
	type args struct {
		ctx         context.Context
		graphSchema model.GraphSchema
	}

	tests := []struct {
		name        string
		fields      fields
		args        args
		wantErr     error
		wantUpdated bool
	}{
		{
			name: "fail - invalid graph schema",
			fields: fields{
				setupOpenGraphSchemaRepositoryMock: func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {},
				setupGraphDBKindsRepositoryMock:    func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: model.GraphSchema{},
			},
			wantErr:     fmt.Errorf("validation error"),
			wantUpdated: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.setupOpenGraphSchemaRepositoryMock(t, mockOpenGraphSchemaRepository)
			tt.fields.setupGraphDBKindsRepositoryMock(t, mockGraphDBKindsRepository)

			o := &OpenGraphSchemaService{
				openGraphSchemaRepository: mockOpenGraphSchemaRepository,
				graphDBKindRepository:     mockGraphDBKindsRepository,
			}
			updated, err := o.UpsertGraphSchemaExtension(tt.args.ctx, tt.args.graphSchema)
			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error(), "UpsertGraphSchemaExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantUpdated != updated {
				require.Fail(t, "expected graph schema to be updated")
			}
		})
	}
}

func Test_validateGraphSchemaModel(t *testing.T) {
	type args struct {
		graphSchema model.GraphSchema
	}
	tests := []struct {
		name    string
		args    args
		wantErr require.ErrorAssertionFunc
	}{
		{
			name: "fail - empty extension name",
			args: args{
				graphSchema: model.GraphSchema{},
			},
			wantErr: require.Error,
		},
		{
			name: "fail - empty graph schema nodes",
			args: args{
				graphSchema: model.GraphSchema{
					GraphSchemaExtension: model.GraphSchemaExtension{
						Name: "Test extension",
					},
				},
			},
			wantErr: require.Error,
		},
		{
			name: "success - valid model.GraphSchemaExtension",
			args: args{
				graphSchema: model.GraphSchema{
					GraphSchemaExtension: model.GraphSchemaExtension{
						Name: "Test extension",
					},
					GraphSchemaNodeKinds: model.GraphSchemaNodeKinds{{
						Name:              "node kind 1",
						SchemaExtensionId: 1,
					}},
				},
			},
			wantErr: require.NoError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, validateGraphSchemaModel(tt.args.graphSchema), fmt.Sprintf("validateGraphSchemaModel(%v)", tt.args.graphSchema))
		})
	}
}
