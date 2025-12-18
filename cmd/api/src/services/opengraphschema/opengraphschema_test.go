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
	"strconv"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
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

		mockOpenGraphSchemaExtensionRepository = mocks.NewMockOpenGraphSchemaExtensionRepository(mockCtrl)
		mockOpenGraphSchemaNodeKindRepository  = mocks.NewMockOpenGraphSchemaNodeKindRepository(mockCtrl)
		mockOpenGraphSchemaEdgeKindRepository  = mocks.NewMockOpenGraphSchemaEdgeKindRepository(mockCtrl)
		mockOpenGraphSchemaPropertyRepository  = mocks.NewMockOpenGraphSchemaPropertyRepository(mockCtrl)
		mockGraphDBKindsRepository             = mocks.NewMockGraphDBKindRepository(mockCtrl)

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
		setupGraphSchemaExtensionRepositoryMocks func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository)
		setupGraphSchemaNodeKindRepositoryMocks  func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository)
		setupGraphSchemaEdgeKindRepositoryMocks  func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository)
		setupGraphSchemaPropertyRepositoryMocks  func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository)
		setupGraphDBKindsRepositoryMock          func(t *testing.T, mock *mocks.MockGraphDBKindRepository)
	}
	type args struct {
		ctx         context.Context
		graphSchema model.GraphSchema
	}

	type mapSyncActionsWant struct {
		nodeKindsToCreate  map[string]model.GraphSchemaNodeKind
		edgeKindsToCreate  map[string]model.GraphSchemaEdgeKind
		propertiesToCreate map[string]model.GraphSchemaProperty
		nodeKindsToUpdate  map[string]model.GraphSchemaNodeKind
		edgeKindsToUpdate  map[string]model.GraphSchemaEdgeKind
		propertiesToUpdate map[string]model.GraphSchemaProperty
		nodeKindsToDelete  map[string]model.GraphSchemaNodeKind
		edgeKindsToDelete  map[string]model.GraphSchemaEdgeKind
		propertiesToDelete map[string]model.GraphSchemaProperty
	}

	// GenerateMapSynchronizationDiffActions does not preserve ordering so the following CRUD mocks
	// perform a Do function which removes the provided kind/property from the item-function map.
	// The last Do function for each CRUD Operation will check to see if their respective wantAction's
	// map length is 0 to ensure all actions are accounted for.

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr error
	}{
		{
			name: "fail - invalid graph schema",
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {},
				setupGraphSchemaNodeKindRepositoryMocks:  func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {},
				setupGraphSchemaEdgeKindRepositoryMocks:  func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {},
				setupGraphSchemaPropertyRepositoryMocks:  func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
				setupGraphDBKindsRepositoryMock:          func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
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
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {
					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if existingExtension1 exists
						Operator:    model.Equals,
						Value:       existingExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, fmt.Errorf("get extensions - test timeout error"))
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
				setupGraphDBKindsRepositoryMock:         func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
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
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {
					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if existingExtension1 exists
						Operator:    model.Equals,
						Value:       newExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

					mock.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
						newGraphSchema.GraphSchemaExtension.Name, newGraphSchema.GraphSchemaExtension.DisplayName,
						newGraphSchema.GraphSchemaExtension.Version).Return(model.GraphSchemaExtension{}, fmt.Errorf("create extension - test timeout error"))
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
				setupGraphDBKindsRepositoryMock:         func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
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
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {

					// GenerateMapSynchronizationDiffActions does not preserve ordering so the following CRUD mocks
					// perform a Do function which removes the provided kind/property from the item-function map.
					// The last Do function for each CRUD Operation will check to see if their respective wantAction's
					// map length is 0 to ensure all actions are accounted for.

					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if existingExtension1 exists
						Operator:    model.Equals,
						Value:       newExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

					mock.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
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
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {

					nodeKindsToCreate := map[string]model.GraphSchemaNodeKind{
						newNodeKind1.Name: {
							Name:              newNodeKind1.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind1.DisplayName,
							Description:       newNodeKind1.Description,
							IsDisplayKind:     newNodeKind1.IsDisplayKind,
							Icon:              newNodeKind1.Icon,
							IconColor:         newNodeKind1.IconColor,
						},
						newNodeKind2.Name: {
							Name:              newNodeKind2.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind2.DisplayName,
							Description:       newNodeKind2.Description,
							IsDisplayKind:     newNodeKind2.IsDisplayKind,
							Icon:              newNodeKind2.Icon,
							IconColor:         newNodeKind2.IconColor,
						},
					}
					mock.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), gomock.AnyOf(newNodeKind1.Name, newNodeKind2.Name),
						int32(1), gomock.AnyOf(newNodeKind1.DisplayName, newNodeKind2.DisplayName), gomock.AnyOf(newNodeKind1.Description, newNodeKind2.Description),
						gomock.AnyOf(newNodeKind1.IsDisplayKind, newNodeKind2.IsDisplayKind),
						gomock.AnyOf(newNodeKind1.Icon, newNodeKind2.Icon), gomock.AnyOf(newNodeKind1.IconColor, newNodeKind2.IconColor)).Do(func(ctx context.Context, name string, schemaExtensionId int32, displayName string, description string, isDisplayKind bool, icon string, iconColor string) {
						if want, ok := nodeKindsToCreate[name]; !ok {
							require.Fail(t, "create - node kind is not expected: %s", name)
						} else {
							compareGraphSchemaNodeKind(t, model.GraphSchemaNodeKind{
								Name:              name,
								SchemaExtensionId: schemaExtensionId,
								DisplayName:       displayName,
								Description:       description,
								IsDisplayKind:     isDisplayKind,
								Icon:              icon,
								IconColor:         iconColor,
							}, want)
							delete(nodeKindsToCreate, name)
							require.Lenf(t, nodeKindsToCreate, len(nodeKindsToCreate),
								"unexpected number of node kinds not created: %v, want: %d", nodeKindsToCreate, len(nodeKindsToCreate))
						}
					}).Return(model.GraphSchemaNodeKind{}, fmt.Errorf("create node_kind - test timeout error"))

				},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
				setupGraphDBKindsRepositoryMock:         func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: newGraphSchema,
			},
			wantErr: fmt.Errorf("create node_kind - test timeout error"),
		},
		{
			name: "fail - error creating new open graph edge kind",
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {

					// GenerateMapSynchronizationDiffActions does not preserve ordering so the following CRUD mocks
					// perform a Do function which removes the provided kind/property from the item-function map.
					// The last Do function for each CRUD Operation will check to see if their respective wantAction's
					// map length is 0 to ensure all actions are accounted for.

					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{
						Operator:    model.Equals,
						Value:       newExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

					mock.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
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
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {

					nodeKindsToCreate := map[string]model.GraphSchemaNodeKind{
						newNodeKind1.Name: {
							Name:              newNodeKind1.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind1.DisplayName,
							Description:       newNodeKind1.Description,
							IsDisplayKind:     newNodeKind1.IsDisplayKind,
							Icon:              newNodeKind1.Icon,
							IconColor:         newNodeKind1.IconColor,
						},
						newNodeKind2.Name: {
							Name:              newNodeKind2.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind2.DisplayName,
							Description:       newNodeKind2.Description,
							IsDisplayKind:     newNodeKind2.IsDisplayKind,
							Icon:              newNodeKind2.Icon,
							IconColor:         newNodeKind2.IconColor,
						},
					}

					mock.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), gomock.AnyOf(newNodeKind1.Name, newNodeKind2.Name),
						int32(1), gomock.AnyOf(newNodeKind1.DisplayName, newNodeKind2.DisplayName), gomock.AnyOf(newNodeKind1.Description, newNodeKind2.Description),
						gomock.AnyOf(newNodeKind1.IsDisplayKind, newNodeKind2.IsDisplayKind),
						gomock.AnyOf(newNodeKind1.Icon, newNodeKind2.Icon), gomock.AnyOf(newNodeKind1.IconColor,
							newNodeKind2.IconColor),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, displayName string, description string, isDisplayKind bool, icon string, iconColor string) {
						if want, ok := nodeKindsToCreate[name]; !ok {
							require.Fail(t, "create - node kind is not expected: %s", name)
						} else {
							compareGraphSchemaNodeKind(t, model.GraphSchemaNodeKind{
								Name:              name,
								SchemaExtensionId: schemaExtensionId,
								DisplayName:       displayName,
								Description:       description,
								IsDisplayKind:     isDisplayKind,
								Icon:              icon,
								IconColor:         iconColor,
							}, want)
							delete(nodeKindsToCreate, name)
						}
					}).Return(newNodeKind1, nil)

					mock.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), gomock.AnyOf(newNodeKind1.Name, newNodeKind2.Name),
						int32(1), gomock.AnyOf(newNodeKind1.DisplayName, newNodeKind2.DisplayName),
						gomock.AnyOf(newNodeKind1.Description, newNodeKind2.Description),
						gomock.AnyOf(newNodeKind1.IsDisplayKind, newNodeKind2.IsDisplayKind),
						gomock.AnyOf(newNodeKind1.Icon, newNodeKind2.Icon),
						gomock.AnyOf(newNodeKind1.IconColor, newNodeKind2.IconColor),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, displayName string, description string, isDisplayKind bool, icon string, iconColor string) {
						if want, ok := nodeKindsToCreate[name]; !ok {
							require.Fail(t, "create - node kind is not expected: %s", name)
						} else {
							compareGraphSchemaNodeKind(t, model.GraphSchemaNodeKind{
								Name:              name,
								SchemaExtensionId: schemaExtensionId,
								DisplayName:       displayName,
								Description:       description,
								IsDisplayKind:     isDisplayKind,
								Icon:              icon,
								IconColor:         iconColor,
							}, want)
							delete(nodeKindsToCreate, name)
							require.Lenf(t, nodeKindsToCreate, 0, "create node_kind - not all nodes were created: %v", nodeKindsToCreate)
						}
					}).Return(newNodeKind2, nil)
				},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {

					edgeKindsToCreate := map[string]model.GraphSchemaEdgeKind{
						newEdgeKind1.Name: {
							SchemaExtensionId: 1,
							Name:              newEdgeKind1.Name,
							Description:       newEdgeKind1.Description,
							IsTraversable:     newEdgeKind1.IsTraversable,
						},
						newEdgeKind2.Name: {
							SchemaExtensionId: 1,
							Name:              newEdgeKind2.Name,
							Description:       newEdgeKind2.Description,
							IsTraversable:     newEdgeKind2.IsTraversable,
						},
					}

					mock.EXPECT().CreateGraphSchemaEdgeKind(gomock.Any(), gomock.AnyOf(newEdgeKind1.Name, newEdgeKind2.Name), int32(1),
						gomock.AnyOf(newEdgeKind1.Description, newNodeKind2.Description), gomock.AnyOf(newEdgeKind1.IsTraversable, newEdgeKind2.IsTraversable),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) {
						if want, ok := edgeKindsToCreate[name]; !ok {
							require.Fail(t, "unexpected create edge kind: %s", name)
						} else {
							compareGraphSchemaEdgeKind(t, model.GraphSchemaEdgeKind{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								IsTraversable:     isTraversable,
							}, want)
							delete(edgeKindsToCreate, name)
							require.Lenf(t, edgeKindsToCreate, 1,
								"unexpected number of edge kinds not created: %v, want: %d", edgeKindsToCreate, 1)
						}
					}).Return(
						model.GraphSchemaEdgeKind{}, fmt.Errorf("create edge_kind - test timeout error"))
				},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
				setupGraphDBKindsRepositoryMock:         func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: newGraphSchema,
			},
			wantErr: fmt.Errorf("create edge_kind - test timeout error"),
		},
		{
			name: "fail - error creating new open graph schema property",
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {

					// GenerateMapSynchronizationDiffActions does not preserve ordering so the following CRUD mocks
					// perform a Do function which removes the provided kind/property from the item-function map.
					// The last Do function for each CRUD Operation will check to see if their respective wantAction's
					// map length is 0 to ensure all actions are accounted for.

					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{
						Operator:    model.Equals,
						Value:       newExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

					mock.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
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
				},

				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {

					nodeKindsToCreate := map[string]model.GraphSchemaNodeKind{
						newNodeKind1.Name: {
							Name:              newNodeKind1.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind1.DisplayName,
							Description:       newNodeKind1.Description,
							IsDisplayKind:     newNodeKind1.IsDisplayKind,
							Icon:              newNodeKind1.Icon,
							IconColor:         newNodeKind1.IconColor,
						},
						newNodeKind2.Name: {
							Name:              newNodeKind2.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind2.DisplayName,
							Description:       newNodeKind2.Description,
							IsDisplayKind:     newNodeKind2.IsDisplayKind,
							Icon:              newNodeKind2.Icon,
							IconColor:         newNodeKind2.IconColor,
						},
					}

					mock.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), gomock.AnyOf(newNodeKind1.Name, newNodeKind2.Name),
						int32(1), gomock.AnyOf(newNodeKind1.DisplayName, newNodeKind2.DisplayName), gomock.AnyOf(newNodeKind1.Description, newNodeKind2.Description),
						gomock.AnyOf(newNodeKind1.IsDisplayKind, newNodeKind2.IsDisplayKind),
						gomock.AnyOf(newNodeKind1.Icon, newNodeKind2.Icon), gomock.AnyOf(newNodeKind1.IconColor,
							newNodeKind2.IconColor),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, displayName string, description string, isDisplayKind bool, icon string, iconColor string) {
						if want, ok := nodeKindsToCreate[name]; !ok {
							require.Fail(t, "create - node kind is not expected: %s", name)
						} else {
							compareGraphSchemaNodeKind(t, model.GraphSchemaNodeKind{
								Name:              name,
								SchemaExtensionId: schemaExtensionId,
								DisplayName:       displayName,
								Description:       description,
								IsDisplayKind:     isDisplayKind,
								Icon:              icon,
								IconColor:         iconColor,
							}, want)
							delete(nodeKindsToCreate, name)
						}
					}).Return(newNodeKind1, nil)

					mock.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), gomock.AnyOf(newNodeKind1.Name, newNodeKind2.Name),
						int32(1), gomock.AnyOf(newNodeKind1.DisplayName, newNodeKind2.DisplayName),
						gomock.AnyOf(newNodeKind1.Description, newNodeKind2.Description),
						gomock.AnyOf(newNodeKind1.IsDisplayKind, newNodeKind2.IsDisplayKind),
						gomock.AnyOf(newNodeKind1.Icon, newNodeKind2.Icon),
						gomock.AnyOf(newNodeKind1.IconColor, newNodeKind2.IconColor),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, displayName string, description string, isDisplayKind bool, icon string, iconColor string) {
						if want, ok := nodeKindsToCreate[name]; !ok {
							require.Fail(t, "create - node kind is not expected: %s", name)
						} else {
							compareGraphSchemaNodeKind(t, model.GraphSchemaNodeKind{
								Name:              name,
								SchemaExtensionId: schemaExtensionId,
								DisplayName:       displayName,
								Description:       description,
								IsDisplayKind:     isDisplayKind,
								Icon:              icon,
								IconColor:         iconColor,
							}, want)
							delete(nodeKindsToCreate, name)
							require.Lenf(t, nodeKindsToCreate, 0, "create node_kind - not all nodes were created: %v", nodeKindsToCreate)
						}
					}).Return(newNodeKind2, nil)
				},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {

					edgeKindsToCreate := map[string]model.GraphSchemaEdgeKind{
						newEdgeKind1.Name: {
							SchemaExtensionId: 1,
							Name:              newEdgeKind1.Name,
							Description:       newEdgeKind1.Description,
							IsTraversable:     newEdgeKind1.IsTraversable,
						},
						newEdgeKind2.Name: {
							SchemaExtensionId: 1,
							Name:              newEdgeKind2.Name,
							Description:       newEdgeKind2.Description,
							IsTraversable:     newEdgeKind2.IsTraversable,
						},
					}

					mock.EXPECT().CreateGraphSchemaEdgeKind(gomock.Any(), gomock.AnyOf(newEdgeKind1.Name, newEdgeKind2.Name), int32(1),
						gomock.AnyOf(newEdgeKind1.Description, newNodeKind2.Description), gomock.AnyOf(newEdgeKind1.IsTraversable, newEdgeKind2.IsTraversable),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) {
						if want, ok := edgeKindsToCreate[name]; !ok {
							require.Fail(t, "unexpected create edge kind: %s", name)
						} else {
							compareGraphSchemaEdgeKind(t, model.GraphSchemaEdgeKind{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								IsTraversable:     isTraversable,
							}, want)
							delete(edgeKindsToCreate, name)
						}
					}).Return(newEdgeKind1, nil)
					mock.EXPECT().CreateGraphSchemaEdgeKind(gomock.Any(), gomock.AnyOf(newEdgeKind1.Name, newEdgeKind2.Name), int32(1),
						gomock.AnyOf(newEdgeKind1.Description, newNodeKind2.Description), gomock.AnyOf(newEdgeKind1.IsTraversable, newEdgeKind2.IsTraversable),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) {
						if want, ok := edgeKindsToCreate[name]; !ok {
							require.Fail(t, "unexpected create edge kind: %s", name)
						} else {
							compareGraphSchemaEdgeKind(t, model.GraphSchemaEdgeKind{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								IsTraversable:     isTraversable,
							}, want)
							delete(edgeKindsToCreate, name)
							require.Lenf(t, edgeKindsToCreate, 0,
								"unexpected number of edge kinds not created: %v, want: %d", edgeKindsToCreate, 0)
						}
					}).Return(newEdgeKind2, nil)
				},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {

					propertiesToCreate := map[string]model.GraphSchemaProperty{
						newProperty1.Name: {
							SchemaExtensionId: 1,
							Name:              newProperty1.Name,
							DisplayName:       newProperty1.DisplayName,
							DataType:          newProperty1.DataType,
							Description:       newProperty1.Description,
						},
						newProperty2.Name: {
							SchemaExtensionId: 1,
							Name:              newProperty2.Name,
							DisplayName:       newProperty2.DisplayName,
							DataType:          newProperty2.DataType,
							Description:       newProperty2.Description,
						},
					}
					mock.EXPECT().CreateGraphSchemaProperty(gomock.Any(), int32(1), gomock.AnyOf(newProperty1.Name, newProperty2.Name),
						gomock.AnyOf(newProperty1.DisplayName, newProperty2.DisplayName), gomock.AnyOf(newProperty1.DataType, newProperty2.DataType),
						gomock.AnyOf(newProperty1.Description, newProperty2.Description),
					).Do(func(ctx context.Context, schemaExtensionId int32, name string, DisplayName, dataType, description string) {
						if want, ok := propertiesToCreate[name]; !ok {
							require.Fail(t, "unexpected create property: %s", name)
						} else {
							compareGraphSchemaProperty(t, model.GraphSchemaProperty{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								DataType:          dataType,
								DisplayName:       DisplayName,
							}, want)
							delete(propertiesToCreate, name)
							require.Lenf(t, propertiesToCreate, 1,
								"unexpected number of properties not created: %v, want: %d", propertiesToCreate, 1)
						}
					}).Return(model.GraphSchemaProperty{}, fmt.Errorf("create property - test timeout error"))
				},
				setupGraphDBKindsRepositoryMock: func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: newGraphSchema,
			},
			wantErr: fmt.Errorf("create property - test timeout error"),
		},
		// TODO: CREATE FAIL REFRESH KINDS TEST
		{
			name: "success - create new graph schema",
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {

					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{
						Operator:    model.Equals,
						Value:       newExtension1.Name,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, database.ErrNotFound)

					mock.EXPECT().CreateGraphSchemaExtension(gomock.Any(),
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
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {

					nodeKindsToCreate := map[string]model.GraphSchemaNodeKind{
						newNodeKind1.Name: {
							Name:              newNodeKind1.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind1.DisplayName,
							Description:       newNodeKind1.Description,
							IsDisplayKind:     newNodeKind1.IsDisplayKind,
							Icon:              newNodeKind1.Icon,
							IconColor:         newNodeKind1.IconColor,
						},
						newNodeKind2.Name: {
							Name:              newNodeKind2.Name,
							SchemaExtensionId: 1,
							DisplayName:       newNodeKind2.DisplayName,
							Description:       newNodeKind2.Description,
							IsDisplayKind:     newNodeKind2.IsDisplayKind,
							Icon:              newNodeKind2.Icon,
							IconColor:         newNodeKind2.IconColor,
						},
					}

					mock.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), gomock.AnyOf(newNodeKind1.Name, newNodeKind2.Name),
						int32(1), gomock.AnyOf(newNodeKind1.DisplayName, newNodeKind2.DisplayName), gomock.AnyOf(newNodeKind1.Description, newNodeKind2.Description),
						gomock.AnyOf(newNodeKind1.IsDisplayKind, newNodeKind2.IsDisplayKind),
						gomock.AnyOf(newNodeKind1.Icon, newNodeKind2.Icon), gomock.AnyOf(newNodeKind1.IconColor,
							newNodeKind2.IconColor),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, displayName string, description string, isDisplayKind bool, icon string, iconColor string) {
						if want, ok := nodeKindsToCreate[name]; !ok {
							require.Fail(t, "create - node kind is not expected: %s", name)
						} else {
							compareGraphSchemaNodeKind(t, model.GraphSchemaNodeKind{
								Name:              name,
								SchemaExtensionId: schemaExtensionId,
								DisplayName:       displayName,
								Description:       description,
								IsDisplayKind:     isDisplayKind,
								Icon:              icon,
								IconColor:         iconColor,
							}, want)
							delete(nodeKindsToCreate, name)
						}
					}).Return(newNodeKind1, nil)

					mock.EXPECT().CreateGraphSchemaNodeKind(gomock.Any(), gomock.AnyOf(newNodeKind1.Name, newNodeKind2.Name),
						int32(1), gomock.AnyOf(newNodeKind1.DisplayName, newNodeKind2.DisplayName),
						gomock.AnyOf(newNodeKind1.Description, newNodeKind2.Description),
						gomock.AnyOf(newNodeKind1.IsDisplayKind, newNodeKind2.IsDisplayKind),
						gomock.AnyOf(newNodeKind1.Icon, newNodeKind2.Icon),
						gomock.AnyOf(newNodeKind1.IconColor, newNodeKind2.IconColor),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, displayName string, description string, isDisplayKind bool, icon string, iconColor string) {
						if want, ok := nodeKindsToCreate[name]; !ok {
							require.Fail(t, "create - node kind is not expected: %s", name)
						} else {
							compareGraphSchemaNodeKind(t, model.GraphSchemaNodeKind{
								Name:              name,
								SchemaExtensionId: schemaExtensionId,
								DisplayName:       displayName,
								Description:       description,
								IsDisplayKind:     isDisplayKind,
								Icon:              icon,
								IconColor:         iconColor,
							}, want)
							delete(nodeKindsToCreate, name)
							require.Lenf(t, nodeKindsToCreate, 0, "create node_kind - not all nodes were created: %v", nodeKindsToCreate)
						}
					}).Return(newNodeKind2, nil)
				},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {

					edgeKindsToCreate := map[string]model.GraphSchemaEdgeKind{
						newEdgeKind1.Name: {
							SchemaExtensionId: 1,
							Name:              newEdgeKind1.Name,
							Description:       newEdgeKind1.Description,
							IsTraversable:     newEdgeKind1.IsTraversable,
						},
						newEdgeKind2.Name: {
							SchemaExtensionId: 1,
							Name:              newEdgeKind2.Name,
							Description:       newEdgeKind2.Description,
							IsTraversable:     newEdgeKind2.IsTraversable,
						},
					}

					mock.EXPECT().CreateGraphSchemaEdgeKind(gomock.Any(), gomock.AnyOf(newEdgeKind1.Name, newEdgeKind2.Name), int32(1),
						gomock.AnyOf(newEdgeKind1.Description, newNodeKind2.Description), gomock.AnyOf(newEdgeKind1.IsTraversable, newEdgeKind2.IsTraversable),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) {
						if want, ok := edgeKindsToCreate[name]; !ok {
							require.Fail(t, "unexpected create edge kind: %s", name)
						} else {
							compareGraphSchemaEdgeKind(t, model.GraphSchemaEdgeKind{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								IsTraversable:     isTraversable,
							}, want)
							delete(edgeKindsToCreate, name)
						}
					}).Return(newEdgeKind1, nil)
					mock.EXPECT().CreateGraphSchemaEdgeKind(gomock.Any(), gomock.AnyOf(newEdgeKind1.Name, newEdgeKind2.Name), int32(1),
						gomock.AnyOf(newEdgeKind1.Description, newNodeKind2.Description), gomock.AnyOf(newEdgeKind1.IsTraversable, newEdgeKind2.IsTraversable),
					).Do(func(ctx context.Context, name string, schemaExtensionId int32, description string, isTraversable bool) {
						if want, ok := edgeKindsToCreate[name]; !ok {
							require.Fail(t, "unexpected create edge kind: %s", name)
						} else {
							compareGraphSchemaEdgeKind(t, model.GraphSchemaEdgeKind{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								IsTraversable:     isTraversable,
							}, want)
							delete(edgeKindsToCreate, name)
							require.Lenf(t, edgeKindsToCreate, 0,
								"unexpected number of edge kinds not created: %v, want: %d", edgeKindsToCreate, 0)
						}
					}).Return(newEdgeKind2, nil)
				},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {

					propertiesToCreate := map[string]model.GraphSchemaProperty{
						newProperty1.Name: {
							SchemaExtensionId: 1,
							Name:              newProperty1.Name,
							DisplayName:       newProperty1.DisplayName,
							DataType:          newProperty1.DataType,
							Description:       newProperty1.Description,
						},
						newProperty2.Name: {
							SchemaExtensionId: 1,
							Name:              newProperty2.Name,
							DisplayName:       newProperty2.DisplayName,
							DataType:          newProperty2.DataType,
							Description:       newProperty2.Description,
						},
					}

					mock.EXPECT().CreateGraphSchemaProperty(gomock.Any(), int32(1), gomock.AnyOf(newProperty1.Name, newProperty2.Name),
						gomock.AnyOf(newProperty1.DisplayName, newProperty2.DisplayName), gomock.AnyOf(newProperty1.DataType, newProperty2.DataType),
						gomock.AnyOf(newProperty1.Description, newProperty2.Description),
					).Do(func(ctx context.Context, schemaExtensionId int32, name string, DisplayName, dataType, description string) {
						if want, ok := propertiesToCreate[name]; !ok {
							require.Fail(t, "unexpected create property: %s", name)
						} else {
							compareGraphSchemaProperty(t, model.GraphSchemaProperty{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								DataType:          dataType,
								DisplayName:       DisplayName,
							}, want)
							delete(propertiesToCreate, name)
						}
					}).Return(newProperty1, nil)

					mock.EXPECT().CreateGraphSchemaProperty(gomock.Any(), int32(1), gomock.AnyOf(newProperty1.Name, newProperty2.Name),
						gomock.AnyOf(newProperty1.DisplayName, newProperty2.DisplayName), gomock.AnyOf(newProperty1.DataType, newProperty2.DataType),
						gomock.AnyOf(newProperty1.Description, newProperty2.Description),
					).Do(func(ctx context.Context, schemaExtensionId int32, name string, DisplayName, dataType, description string) {
						if want, ok := propertiesToCreate[name]; !ok {
							require.Fail(t, "unexpected create property: %s", name)
						} else {
							compareGraphSchemaProperty(t, model.GraphSchemaProperty{
								SchemaExtensionId: schemaExtensionId,
								Name:              name,
								Description:       description,
								DataType:          dataType,
								DisplayName:       DisplayName,
							}, want)
							delete(propertiesToCreate, name)
							require.Lenf(t, propertiesToCreate, 0,
								"unexpected number of properties not created: %v, want: %d", propertiesToCreate, 0)
						}
					}).Return(newProperty2, nil)
				},
				setupGraphDBKindsRepositoryMock: func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {
					mock.EXPECT().RefreshKinds(gomock.Any()).Return(nil)
				},
			},
			args: args{
				ctx:         context.Background(),
				graphSchema: newGraphSchema,
			},
			wantErr: nil,
		},

		// Preexisting Schema Extension

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.setupGraphSchemaExtensionRepositoryMocks(t, mockOpenGraphSchemaExtensionRepository)
			tt.fields.setupGraphSchemaNodeKindRepositoryMocks(t, mockOpenGraphSchemaNodeKindRepository)
			tt.fields.setupGraphSchemaEdgeKindRepositoryMocks(t, mockOpenGraphSchemaEdgeKindRepository)
			tt.fields.setupGraphSchemaPropertyRepositoryMocks(t, mockOpenGraphSchemaPropertyRepository)
			tt.fields.setupGraphDBKindsRepositoryMock(t, mockGraphDBKindsRepository)

			o := &OpenGraphSchemaService{
				openGraphSchemaExtensionRepository: mockOpenGraphSchemaExtensionRepository,
				openGraphSchemaNodeRepository:      mockOpenGraphSchemaNodeKindRepository,
				openGraphSchemaEdgeRepository:      mockOpenGraphSchemaEdgeKindRepository,
				openGraphSchemaPropertyRepository:  mockOpenGraphSchemaPropertyRepository,
				graphDBKindRepository:              mockGraphDBKindsRepository,
			}
			err := o.UpsertGraphSchemaExtension(tt.args.ctx, tt.args.graphSchema)
			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error(), "UpsertGraphSchemaExtension() error = %v, wantErr %v", err, tt.wantErr)
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

func TestOpenGraphSchemaService_getGraphSchemaByExtensionName(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)

		mockOpenGraphSchemaExtensionRepository = mocks.NewMockOpenGraphSchemaExtensionRepository(mockCtrl)
		mockOpenGraphSchemaNodeKindRepository  = mocks.NewMockOpenGraphSchemaNodeKindRepository(mockCtrl)
		mockOpenGraphSchemaEdgeKindRepository  = mocks.NewMockOpenGraphSchemaEdgeKindRepository(mockCtrl)
		mockOpenGraphSchemaPropertyRepository  = mocks.NewMockOpenGraphSchemaPropertyRepository(mockCtrl)

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
	)

	type fields struct {
		setupGraphSchemaExtensionRepositoryMocks func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository)
		setupGraphSchemaNodeKindRepositoryMocks  func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository)
		setupGraphSchemaEdgeKindRepositoryMocks  func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository)
		setupGraphSchemaPropertyRepositoryMocks  func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository)
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
			name: "fail - GetGraphSchemaExtensions error",
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {
					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if extension exists
						Operator:    model.Equals,
						Value:       testExtensionName,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, fmt.Errorf("GetGraphSchemaExtensions error"))
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {

				},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
			},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: fmt.Errorf("GetGraphSchemaExtensions error"),
		},
		{
			name: "fail - GetGraphSchemaNodeKinds error",
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {
					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if extension exists
						Operator:    model.Equals,
						Value:       testExtensionName,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{testExtension}, 1, nil)
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {
					mock.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(),
						model.Filters{"schema_extension_id": []model.Filter{{
							Operator:    model.Equals,
							Value:       strconv.FormatInt(int64(testExtension.ID), 10),
							SetOperator: model.FilterAnd,
						}}}, model.Sort{}, 0, 0).Return(model.GraphSchemaNodeKinds{}, 0, fmt.Errorf("GetGraphSchemaNodeKinds error"))
				},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
			},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: fmt.Errorf("GetGraphSchemaNodeKinds error"),
		},
		{
			name: "fail - GetGraphSchemaEdgeKinds error",
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {
					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if extension exists
						Operator:    model.Equals,
						Value:       testExtensionName,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{testExtension}, 1, nil)
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {
					mock.EXPECT().GetGraphSchemaNodeKinds(gomock.Any(),
						model.Filters{"schema_extension_id": []model.Filter{{
							Operator:    model.Equals,
							Value:       strconv.FormatInt(int64(testExtension.ID), 10),
							SetOperator: model.FilterAnd,
						}}}, model.Sort{}, 0, 0).Return(model.GraphSchemaNodeKinds{testNodeKind1}, 1, nil)
				},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {
					mock.EXPECT().GetGraphSchemaEdgeKinds(gomock.Any(),
						model.Filters{"schema_extension_id": []model.Filter{{
							Operator:    model.Equals,
							Value:       strconv.FormatInt(int64(testExtension.ID), 10),
							SetOperator: model.FilterAnd,
						}}}, model.Sort{}, 0, 0).Return(model.GraphSchemaEdgeKinds{}, 0, fmt.Errorf("GetGraphSchemaEdgeKinds error"))
				},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
			},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: fmt.Errorf("GetGraphSchemaEdgeKinds error"),
		},
		{
			name: "success - no GetGraphSchemaExtensions results", // Will result in new graph schema extension
			fields: fields{
				setupGraphSchemaExtensionRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaExtensionRepository) {
					mock.EXPECT().GetGraphSchemaExtensions(gomock.Any(), model.Filters{"name": []model.Filter{{ // check to see if extension exists
						Operator:    model.Equals,
						Value:       testExtensionName,
						SetOperator: model.FilterAnd,
					}}}, model.Sort{}, 0, 1).Return(model.GraphSchemaExtensions{}, 0, nil)
				},
				setupGraphSchemaNodeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaNodeKindRepository) {},
				setupGraphSchemaEdgeKindRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaEdgeKindRepository) {},
				setupGraphSchemaPropertyRepositoryMocks: func(t *testing.T, mock *mocks.MockOpenGraphSchemaPropertyRepository) {},
			},
			args: args{
				ctx:           context.Background(),
				extensionName: testExtensionName,
			},
			want:    model.GraphSchema{},
			wantErr: database.ErrNotFound,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.fields.setupGraphSchemaExtensionRepositoryMocks(t, mockOpenGraphSchemaExtensionRepository)
			tt.fields.setupGraphSchemaNodeKindRepositoryMocks(t, mockOpenGraphSchemaNodeKindRepository)
			tt.fields.setupGraphSchemaEdgeKindRepositoryMocks(t, mockOpenGraphSchemaEdgeKindRepository)
			tt.fields.setupGraphSchemaPropertyRepositoryMocks(t, mockOpenGraphSchemaPropertyRepository)

			o := &OpenGraphSchemaService{
				openGraphSchemaExtensionRepository: mockOpenGraphSchemaExtensionRepository,
				openGraphSchemaNodeRepository:      mockOpenGraphSchemaNodeKindRepository,
				openGraphSchemaEdgeRepository:      mockOpenGraphSchemaEdgeKindRepository,
				openGraphSchemaPropertyRepository:  mockOpenGraphSchemaPropertyRepository,
			}
			got, err := o.getGraphSchemaByExtensionName(tt.args.ctx, tt.args.extensionName)
			if tt.wantErr != nil {
				require.EqualErrorf(t, err, tt.wantErr.Error(), "getGraphSchemaByExtensionName() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				require.NoError(t, err)
				compareGraphSchema(t, tt.want, got)
			}
		})
	}
}

func compareGraphSchemaNodeKind(t *testing.T, got, want model.GraphSchemaNodeKind) {
	t.Helper()
	require.Equalf(t, got.ID, want.ID, "GraphSchemaNodeKinds - ID mismatch - got: %v, mapSyncActionsWant: %v", got.ID, want.ID)
	require.Equalf(t, want.Name, got.Name, "GraphSchemaNodeKind - name mismatch - got: %v, mapSyncActionsWant: %v", got.Name, want.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaNodeKind - extension_id mismatch - got: %d, mapSynActionsWant: %d", got.SchemaExtensionId, want.SchemaExtensionId)
	require.Equalf(t, want.DisplayName, got.DisplayName, "GraphSchemaNodeKind - display_name mismatch - got: %s, mapSyncActionsWant: %s", got.DisplayName, want.DisplayName)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaNodeKind - description mismatch - got: %s, mapSyncActionsWant: %s", got.Description, want.Description)
	require.Equalf(t, want.IsDisplayKind, got.IsDisplayKind, "GraphSchemaNodeKind - is_display_kind mismatch - got: %s, mapSyncActionsWant: %s", got.IsDisplayKind, want.IsDisplayKind)
	require.Equalf(t, want.Icon, got.Icon, "GraphSchemaNodeKind - icon mismatch - got: %s, mapSyncActionsWant: %s", got.Icon, want.Icon)
	require.Equalf(t, want.IconColor, got.IconColor, "GraphSchemaNodeKind - icon_color mismatch - got: %s, mapSyncActionsWant: %s", got.IconColor, want.IconColor)
	require.Equalf(t, want.CreatedAt, got.CreatedAt, "GraphSchemaNodeKind - created_at mismatch - got: %s, mapSyncActionsWant: %s", got.CreatedAt.String(), want.CreatedAt.String())
	require.Equalf(t, want.UpdatedAt, got.UpdatedAt, "GraphSchemaNodeKind - updated_at mismatch - got: %s, mapSyncActionsWant: %s", got.UpdatedAt.String(), want.UpdatedAt.String())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaNodeKind - deleted_at is not null")
}

func compareGraphSchemaProperty(t *testing.T, got, want model.GraphSchemaProperty) {
	t.Helper()
	require.Equalf(t, got.ID, want.ID, "GraphSchemaProperty - ID mismatch - got: %v, mapSyncActionsWant: %v", got.ID, want.ID)
	require.Equalf(t, want.Name, got.Name, "GraphSchemaProperty - name mismatch - got: %v, want: %v", got.Name, want.Name)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaProperty - schema_extension_id mismatch - got: %v, want: %v", got.SchemaExtensionId, want.SchemaExtensionId)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaProperty - description mismatch - got: %v, want: %v", got.Description, want.Description)
	require.Equalf(t, want.DisplayName, got.DisplayName, "GraphSchemaProperty - display_name mismatch - got: %v, want: %v", got.DisplayName, want.DisplayName)
	require.Equalf(t, want.DataType, got.DataType, "GraphSchemaProperty - data_type mismatch - got: %v, want: %v", got.DataType, want.DataType)
	require.Equalf(t, want.CreatedAt, got.CreatedAt, "GraphSchemaProperty - created_at mismatch - got: %s, mapSyncActionsWant: %s", got.CreatedAt.String(), want.CreatedAt.String())
	require.Equalf(t, want.UpdatedAt, got.UpdatedAt, "GraphSchemaProperty - updated_at mismatch - got: %s, mapSyncActionsWant: %s", got.UpdatedAt.String(), want.UpdatedAt.String())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaProperty - deleted_at is not null")

}

func compareGraphSchemaEdgeKind(t *testing.T, got, want model.GraphSchemaEdgeKind) {
	t.Helper()
	require.Equalf(t, got.ID, want.ID, "GraphSchemaEdgeKind - ID mismatch - got: %v, mapSyncActionsWant: %v", got.ID, want.ID)
	require.Equalf(t, want.Name, got.Name, "GraphSchemaEdgeKind - name mismatch - got %v, want %v", got.Name, want.Name)
	require.Equalf(t, want.Description, got.Description, "GraphSchemaEdgeKind - description mismatch - got %v, want %v", got.Description, want.Description)
	require.Equalf(t, want.IsTraversable, got.IsTraversable, "GraphSchemaEdgeKind - IsTraversable mismatch - got %t, want %t", got.IsTraversable, want.IsTraversable)
	require.Equalf(t, want.SchemaExtensionId, got.SchemaExtensionId, "GraphSchemaEdgeKind - SchemaExtensionId mismatch - got %d, want %d", got.SchemaExtensionId, want.SchemaExtensionId)
	require.Equalf(t, want.CreatedAt, got.CreatedAt, "GraphSchemaEdgeKind - created_at mismatch - got: %s, mapSyncActionsWant: %s", got.CreatedAt.String(), want.CreatedAt.String())
	require.Equalf(t, want.UpdatedAt, got.UpdatedAt, "GraphSchemaEdgeKind - updated_at mismatch - got: %s, mapSyncActionsWant: %s", got.UpdatedAt.String(), want.UpdatedAt.String())
	require.Equalf(t, false, got.DeletedAt.Valid, "GraphSchemaEdgeKind - deleted_at is not null")
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

// compareGraphSchemaNodeKinds - compares the returned list of model.GraphSchemaNodeKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaNodeKinds(t *testing.T, got, want model.GraphSchemaNodeKinds) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaNodeKinds")
	for i, schemaNodeKind := range got {
		compareGraphSchemaNodeKind(t, schemaNodeKind, want[i])
	}
}

// compareGraphSchemaEdgeKinds - compares the returned list of model.GraphSchemaEdgeKinds with the expected results.
// Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaEdgeKinds(t *testing.T, got, want model.GraphSchemaEdgeKinds) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaEdgeKinds")
	for i, schemaEdgeKind := range got {
		compareGraphSchemaEdgeKind(t, schemaEdgeKind, want[i])
	}
}

// compareGraphSchemaProperties - compares the returned list of model.GraphSchemaProperties with the expected results.
// // Since this is used to compare filtered and paginated results ORDER MATTERS for the expected result.
func compareGraphSchemaProperties(t *testing.T, got, want model.GraphSchemaProperties) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaProperties")
	for i, schemaProperty := range got {
		compareGraphSchemaProperty(t, schemaProperty, want[i])
	}
}

func compareGraphSchema(t *testing.T, got, want model.GraphSchema) {
	t.Helper()
	compareGraphSchemaExtension(t, got.GraphSchemaExtension, want.GraphSchemaExtension)
	compareGraphSchemaNodeKinds(t, got.GraphSchemaNodeKinds, want.GraphSchemaNodeKinds)
	compareGraphSchemaEdgeKinds(t, got.GraphSchemaEdgeKinds, want.GraphSchemaEdgeKinds)
	compareGraphSchemaProperties(t, got.GraphSchemaProperties, want.GraphSchemaProperties)
}
