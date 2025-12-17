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

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
)

func TestOpenGraphSchemaService_UpsertGraphSchemaExtension(t *testing.T) {
	t.Parallel()

	var (
		mockCtrl                      = gomock.NewController(t)
		mockOpenGraphSchemaRepository = mocks.NewMockOpenGraphSchemaRepository(mockCtrl)

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
			IsBuiltin:   true,
		}
		newExtension1 = model.GraphSchemaExtension{
			Name:        "test_extension_2",
			DisplayName: "Test Extension 2",
			Version:     "1.0.0",
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
		setupMocks func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository)
	}
	type args struct {
		ctx         context.Context
		graphSchema model.GraphSchema
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "fail - invalid graph schema",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: model.GraphSchema{},
			},
			wantErr: fmt.Errorf("validation error"),
		},
		{
			name: "fail - error retrieving open graph schema extension",
			fields: fields{
				setupMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
					mockOpenGraphSchemaRepository.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if existingExtension1 exists
						Operator:    model.Equals,
						Value:       existingExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, fmt.Errorf("get extensions - test timeout error"))
				},
			},
			args: args{
				ctx: context.Background(),
				graphSchema: model.GraphSchema{
					GraphSchemaExtension: existingExtension1,
					GraphSchemaNodeKinds: model.GraphSchemaNodeKinds{existingNodeKind1},
				},
			},
			wantErr: fmt.Errorf("get extensions - test timeout error"),
		},

		// New Schema Extension

		{
			name: "fail - error creating new open graph schema extension",
			fields: fields{
				func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
					mockOpenGraphSchemaRepository.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if existingExtension1 exists
						Operator:    model.Equals,
						Value:       newExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

					mockOpenGraphSchemaRepository.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
						newGraphSchema.GraphSchemaExtension.Name, newGraphSchema.GraphSchemaExtension.DisplayName,
						newGraphSchema.GraphSchemaExtension.Version).Return(model.GraphSchemaExtension{}, fmt.Errorf("create extension - test timeout error"))
				},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: newGraphSchema,
			},
			wantErr: fmt.Errorf("create extension - test timeout error"),
		},
		{
			name: "fail - error creating new open graph node kind",
			fields: fields{
				func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
					mockOpenGraphSchemaRepository.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if existingExtension1 exists
						Operator:    model.Equals,
						Value:       newExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

					mockOpenGraphSchemaRepository.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
						newGraphSchema.GraphSchemaExtension.Name, newGraphSchema.GraphSchemaExtension.DisplayName,
						newGraphSchema.GraphSchemaExtension.Version).Return(model.GraphSchemaExtension{
						Serial: model.Serial{
							ID: 1,
							Basic: model.Basic{
								CreatedAt: time.Now(),
								UpdatedAt: time.Now(),
							},
						},
						Name:        newGraphSchema.GraphSchemaExtension.Name,
						DisplayName: newGraphSchema.GraphSchemaExtension.DisplayName,
						Version:     newGraphSchema.GraphSchemaExtension.Version,
						IsBuiltin:   newGraphSchema.GraphSchemaExtension.IsBuiltin,
					}, nil)

					mockOpenGraphSchemaRepository.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(),
						newNodeKind1.Name, int32(1),
						newNodeKind1.DisplayName, newNodeKind1.Description,
						newNodeKind1.IsDisplayKind, newNodeKind1.Icon,
						newNodeKind1.IconColor).Return(model.GraphSchemaNodeKind{}, fmt.Errorf("create node_kind - test timeout error"))
				},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: newGraphSchema,
			},
			wantErr: fmt.Errorf("create node_kind - test timeout error"),
		},
		// {
		// 	name: "fail - error creating new open graph edge kind",
		// 	fields: fields{
		// 		func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
		// 			mockOpenGraphSchemaRepository.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if existingExtension1 exists
		// 				Operator:    model.Equals,
		// 				Value:       newExtension1.Name,
		// 				SetOperator: model.FilterAnd,
		// 			}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

		// 			mockOpenGraphSchemaRepository.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
		// 				newGraphSchema.GraphSchemaExtension.Name, newGraphSchema.GraphSchemaExtension.DisplayName,
		// 				newGraphSchema.GraphSchemaExtension.Version).Return(model.GraphSchemaExtension{
		// 				Serial: model.Serial{
		// 					ID: 1,
		// 					Basic: model.Basic{
		// 						CreatedAt: time.Now(),
		// 						UpdatedAt: time.Now(),
		// 					},
		// 				},
		// 				Name:        newGraphSchema.GraphSchemaExtension.Name,
		// 				DisplayName: newGraphSchema.GraphSchemaExtension.DisplayName,
		// 				Version:     newGraphSchema.GraphSchemaExtension.Version,
		// 				IsBuiltin:   newGraphSchema.GraphSchemaExtension.IsBuiltin,
		// 			}, nil)

		// 			// GenerateMapSynchronizationDiffActions does not maintain preexisting order so
		// 			mockOpenGraphSchemaRepository.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), newNodeKind1.Name,
		// 				int32(1), newNodeKind1.DisplayName, newNodeKind1.Description, newNodeKind1.IsDisplayKind,
		// 				newNodeKind1.Icon, newNodeKind1.IconColor).Return(newNodeKind1, nil)
		// 			mockOpenGraphSchemaRepository.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), newNodeKind2.Name,
		// 				int32(1), newNodeKind2.DisplayName, newNodeKind2.Description, newNodeKind2.IsDisplayKind,
		// 				newNodeKind2.Icon, newNodeKind2.IconColor).Do(func(ctx context.Context, name string, schemaExtensionId int32) {
		// 				compareGraphSchemaNodeKind(t)
		// 			}).Return(newNodeKind2, nil)

		// 			mockOpenGraphSchemaRepository.EXPECT().CreateGraphSchemaEdgeKind(gomock.Any(), newEdgeKind1.Name,
		// 				int32(1), newEdgeKind1.Description, newEdgeKind1.IsTraversable).Return(
		// 				model.GraphSchemaEdgeKind{}, fmt.Errorf("create edge_kind - test timeout error"))
		// 		},
		// 	},
		// 	args: args{
		// 		ctx:         context.Background(),
		// 		graphSchema: newGraphSchema,
		// 	},
		// 	wantErr: fmt.Errorf("create edge_kind - test timeout error"),
		// },

		// Preexisting Schema Extension

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.setupMocks(t, mockOpenGraphSchemaRepository)

			o := &OpenGraphSchemaService{
				openGraphSchemaRepository: mockOpenGraphSchemaRepository,
			}
			err := o.UpsertGraphSchemaExtension(tt.args.ctx, tt.args.graphSchema)
			require.ErrorContains(t, err, tt.wantErr.Error(), "UpsertGraphSchemaExtension() error = %v, wantErr %v", err, tt.wantErr)
		})
	}
}

func Test_validateGraphSchemModel(t *testing.T) {
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
			tt.wantErr(t, validateGraphSchemModel(tt.args.graphSchema), fmt.Sprintf("validateGraphSchemModel(%v)", tt.args.graphSchema))
		})
	}
}
