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

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestOpenGraphSchemaService_UpsertGraphSchemaExtension -
func TestOpenGraphSchemaService_UpsertGraphSchemaExtension(t *testing.T) {
	t.Parallel()

	var (
		mockCtrl = gomock.NewController(t)

		mockOpenGraphSchemaRepository = mocks.NewMockOpenGraphSchemaRepository(mockCtrl)
		mockGraphDBKindsRepository    = mocks.NewMockGraphDBKindRepository(mockCtrl)
	)

	defer mockCtrl.Finish()

	type fields struct {
		setupOpenGraphSchemaRepositoryMock func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository)
		setupGraphDBKindsRepositoryMock    func(t *testing.T, mock *mocks.MockGraphDBKindRepository)
	}
	type args struct {
		ctx            context.Context
		graphExtension model.GraphExtensionInput
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
				ctx:            context.Background(),
				graphExtension: model.GraphExtensionInput{},
			},
			wantErr:     model.ErrGraphExtensionValidation,
			wantUpdated: false,
		},
		{
			name: "fail - UpsertOpenGraphExtension error",
			fields: fields{
				func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name:      "Test extension",
							Version:   "1.0.0",
							Namespace: "DEFAULT",
						},
						NodeKindsInput: model.NodesInput{{
							Name: "DEFAULT_node kind 1",
						}},
					}).Return(false, fmt.Errorf("test error"))
				},
				func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {},
			},
			args: args{
				ctx: context.Background(),
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "DEFAULT",
					},
					NodeKindsInput: model.NodesInput{{
						Name: "DEFAULT_node kind 1",
					}},
				},
			},
			wantErr:     fmt.Errorf("test error"),
			wantUpdated: false,
		},
		{ // Open Graph TODO: Want error if kinds refresh fails?
			name: "success - fail refresh (does not return an error)",
			fields: fields{
				func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name:      "Test extension",
							Version:   "1.0.0",
							Namespace: "DEFAULT",
						},
						NodeKindsInput: model.NodesInput{{
							Name: "DEFAULT_node kind 1",
						}},
					}).Return(false, nil)
				},
				func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {
					mock.EXPECT().RefreshKinds(gomock.Any()).Return(fmt.Errorf("test error"))
				},
			},
			args: args{
				ctx: context.Background(),
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "DEFAULT",
					},
					NodeKindsInput: model.NodesInput{{
						Name: "DEFAULT_node kind 1",
					}},
				},
			},
			wantErr:     nil,
			wantUpdated: false,
		},
		{
			name: "success - inserted",
			fields: fields{
				func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name:      "Test extension",
							Version:   "1.0.0",
							Namespace: "DEFAULT",
						},
						NodeKindsInput: model.NodesInput{{
							Name: "DEFAULT_node kind 1",
						}},
						EnvironmentsInput: []model.EnvironmentInput{
							{
								EnvironmentKind: "DEFAULT_Domain",
								SourceKind:      "Base",
								PrincipalKinds:  []string{"DEFAULT_User"},
							},
							{
								EnvironmentKind: "DEFAULT_AzureAD",
								SourceKind:      "AzureHound",
								PrincipalKinds:  []string{"DEFAULT_User", "DEFAULT_Group"},
							},
						},
					}).Return(false, nil)
				},
				func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {
					mock.EXPECT().RefreshKinds(gomock.Any()).Return(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "DEFAULT",
					},
					NodeKindsInput: model.NodesInput{{
						Name: "DEFAULT_node kind 1",
					}},
					EnvironmentsInput: []model.EnvironmentInput{
						{
							EnvironmentKind: "DEFAULT_Domain",
							SourceKind:      "Base",
							PrincipalKinds:  []string{"DEFAULT_User"},
						},
						{
							EnvironmentKind: "DEFAULT_AzureAD",
							SourceKind:      "AzureHound",
							PrincipalKinds:  []string{"DEFAULT_User", "DEFAULT_Group"},
						},
					},
				},
			},
			wantErr:     nil,
			wantUpdated: false,
		},
		{
			name: "success - updated",
			fields: fields{
				func(t *testing.T, mock *mocks.MockOpenGraphSchemaRepository) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name:      "Test extension",
							Version:   "1.0.0",
							Namespace: "DEFAULT",
						},
						NodeKindsInput: model.NodesInput{{
							Name: "DEFAULT_node kind 1",
						}},
						EnvironmentsInput: []model.EnvironmentInput{
							{
								EnvironmentKind: "DEFAULT_Domain",
								SourceKind:      "Base",
								PrincipalKinds:  []string{"DEFAULT_User"},
							},
						},
					}).Return(true, nil)
				},
				func(t *testing.T, mock *mocks.MockGraphDBKindRepository) {
					mock.EXPECT().RefreshKinds(gomock.Any()).Return(nil)
				},
			},
			args: args{
				ctx: context.Background(),
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "DEFAULT",
					},
					NodeKindsInput: model.NodesInput{{
						Name: "DEFAULT_node kind 1",
					}},
					EnvironmentsInput: []model.EnvironmentInput{
						{
							EnvironmentKind: "DEFAULT_Domain",
							SourceKind:      "Base",
							PrincipalKinds:  []string{"DEFAULT_User"},
						},
					},
				},
			},
			wantErr:     nil,
			wantUpdated: true,
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
			updated, err := o.UpsertOpenGraphExtension(tt.args.ctx, tt.args.graphExtension)
			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error(), "UpsertOpenGraphExtension() error = %v, wantErr %v", err, tt.wantErr)
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
		graphExtension model.GraphExtensionInput
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "fail - empty extension name",
			args: args{
				graphExtension: model.GraphExtensionInput{},
			},
			wantErr: fmt.Errorf("graph schema extension name is required"),
		},
		{
			name: "fail - empty extension version",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name: "Test extension",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension version is required"),
		},
		{
			name: "fail - empty extension namespace",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:    "Test extension",
						Version: "1.0.0",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema extension namespace is required"),
		},
		{
			name: "fail - empty graph schema nodes",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
				},
			},
			wantErr: fmt.Errorf("graph schema node kinds are required"),
		},
		{
			name: "fail - duplicate kinds - two node kinds",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph kinds: AD_node kind 1"),
		},
		{
			name: "fail - node kind missing namespace",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema kind node kind 1 is missing extension namespace"),
		},
		{
			name: "fail - duplicate kinds - two edge kinds",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "AD_edge kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph kinds: AD_edge kind 1"),
		},
		{
			name: "fail - edge kind missing namespace",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "edge kind 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("graph schema edge kind edge kind 1 is missing extension namespace"),
		},
		{
			name: "fail - duplicate properties",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_node kind 1",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "AD_edge kind 2",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 1",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph properties: property 1"),
		},
		{
			name: "fail - duplicate kinds - same edge and node kind",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{
						{
							Name: "AD_a_duplicate_graph_kind",
						},
						{
							Name: "AD_node kind 2",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{
						{
							Name: "AD_edge kind 1",
						},
						{
							Name: "AD_a_duplicate_graph_kind",
						},
					},
					PropertiesInput: model.PropertiesInput{
						{
							Name: "property 1",
						},
						{
							Name: "property 2",
						},
					},
				},
			},
			wantErr: fmt.Errorf("duplicate graph kinds: %s", "AD_a_duplicate_graph_kind"),
		},
		{
			name: "success - valid model.ExtensionInput",
			args: args{
				graphExtension: model.GraphExtensionInput{
					ExtensionInput: model.ExtensionInput{
						Name:      "Test extension",
						Version:   "1.0.0",
						Namespace: "AD",
					},
					NodeKindsInput: model.NodesInput{{
						Name: "AD_node kind 1",
					}},
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := validateGraphExtension(tt.args.graphExtension); tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}
