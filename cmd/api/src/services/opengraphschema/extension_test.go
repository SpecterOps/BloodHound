// Copyright 2026 Specter Ops, Inc.
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
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema"
	schemamocks "github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestOpenGraphSchemaService_UpsertGraphSchemaExtension -
func TestOpenGraphSchemaService_UpsertGraphSchemaExtension(t *testing.T) {
	t.Parallel()

	type fields struct {
		setupOpenGraphSchemaRepositoryMock func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaRepository)
		setupGraphDBKindsRepositoryMock    func(t *testing.T, mock *schemamocks.MockGraphDBKindRepository)
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
				setupOpenGraphSchemaRepositoryMock: func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaRepository) {},
				setupGraphDBKindsRepositoryMock:    func(t *testing.T, mock *schemamocks.MockGraphDBKindRepository) {},
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
				func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaRepository) {
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
				func(t *testing.T, mock *schemamocks.MockGraphDBKindRepository) {},
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
		{
			name: "fail - duplicate namespace", // duplicate namespaces are not caught during validation and will be returned as an error from UpsertOpenGraphExtension
			fields: fields{
				func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaRepository) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name:      "Test extension",
							Version:   "1.0.0",
							Namespace: "DEFAULT",
						},
						NodeKindsInput: model.NodesInput{{
							Name: "DEFAULT_node kind 1",
						}},
					}).Return(false, fmt.Errorf("%w: DEFAULT", model.ErrDuplicateGraphSchemaExtensionNamespace))
				},
				func(t *testing.T, mock *schemamocks.MockGraphDBKindRepository) {},
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
			wantErr:     fmt.Errorf("%w: %v", model.ErrGraphExtensionValidation, fmt.Errorf("%w: %s", model.ErrDuplicateGraphSchemaExtensionNamespace, "DEFAULT")),
			wantUpdated: false,
		},
		{
			name: "fail - graph kinds refresh error",
			fields: fields{
				func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaRepository) {
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
				func(t *testing.T, mock *schemamocks.MockGraphDBKindRepository) {
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
			wantErr:     model.ErrGraphDBRefreshKinds,
			wantUpdated: false,
		},
		{
			name: "success - inserted",
			fields: fields{
				func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaRepository) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name:      "Test extension",
							Version:   "1.0.0",
							Namespace: "DEFAULT",
						},
						NodeKindsInput: model.NodesInput{
							{
								Name: "DEFAULT_node kind 1",
							},
							{
								Name: "DEFAULT_Domain",
							},
							{
								Name: "DEFAULT_User",
							},
							{
								Name: "DEFAULT_Group",
							},
							{
								Name: "DEFAULT_AzureAD",
							},
						},
						RelationshipKindsInput: model.RelationshipsInput{{
							Name: "DEFAULT_Relationship_Kind_1",
						}},
						EnvironmentsInput: []model.EnvironmentInput{
							{
								EnvironmentKindName: "DEFAULT_Domain",
								SourceKindName:      "Base",
								PrincipalKinds:      []string{"DEFAULT_User"},
							},
							{
								EnvironmentKindName: "DEFAULT_AzureAD",
								SourceKindName:      "AzureHound",
								PrincipalKinds:      []string{"DEFAULT_User", "DEFAULT_Group"},
							},
						},
						RelationshipFindingsInput: model.RelationshipFindingsInput{
							{
								Name:                 "DEFAULT_Finding_1",
								DisplayName:          "Finding 1",
								SourceKindName:       "Base",
								RelationshipKindName: "DEFAULT_Relationship_Kind_1",
								EnvironmentKindName:  "DEFAULT_Domain",
								RemediationInput:     model.RemediationInput{},
							},
						},
					}).Return(false, nil)
				},
				func(t *testing.T, mock *schemamocks.MockGraphDBKindRepository) {
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
					NodeKindsInput: model.NodesInput{
						{
							Name: "DEFAULT_node kind 1",
						},
						{
							Name: "DEFAULT_Domain",
						},
						{
							Name: "DEFAULT_User",
						},
						{
							Name: "DEFAULT_Group",
						},
						{
							Name: "DEFAULT_AzureAD",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{{
						Name: "DEFAULT_Relationship_Kind_1",
					}},
					EnvironmentsInput: []model.EnvironmentInput{
						{
							EnvironmentKindName: "DEFAULT_Domain",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"DEFAULT_User"},
						},
						{
							EnvironmentKindName: "DEFAULT_AzureAD",
							SourceKindName:      "AzureHound",
							PrincipalKinds:      []string{"DEFAULT_User", "DEFAULT_Group"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                 "DEFAULT_Finding_1",
							DisplayName:          "Finding 1",
							SourceKindName:       "Base",
							RelationshipKindName: "DEFAULT_Relationship_Kind_1",
							EnvironmentKindName:  "DEFAULT_Domain",
							RemediationInput:     model.RemediationInput{},
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
				func(t *testing.T, mock *schemamocks.MockOpenGraphSchemaRepository) {
					mock.EXPECT().UpsertOpenGraphExtension(gomock.Any(), model.GraphExtensionInput{
						ExtensionInput: model.ExtensionInput{
							Name:      "Test extension",
							Version:   "1.0.0",
							Namespace: "DEFAULT",
						},
						NodeKindsInput: model.NodesInput{
							{
								Name: "DEFAULT_node kind 1",
							},
							{
								Name: "DEFAULT_Domain",
							},
							{
								Name: "DEFAULT_User",
							},
							{
								Name: "DEFAULT_Group",
							},
							{
								Name: "DEFAULT_AzureAD",
							},
						},
						RelationshipKindsInput: model.RelationshipsInput{{
							Name: "DEFAULT_Relationship_Kind_1",
						}},
						EnvironmentsInput: []model.EnvironmentInput{
							{
								EnvironmentKindName: "DEFAULT_Domain",
								SourceKindName:      "Base",
								PrincipalKinds:      []string{"DEFAULT_User"},
							},
							{
								EnvironmentKindName: "DEFAULT_AzureAD",
								SourceKindName:      "AzureHound",
								PrincipalKinds:      []string{"DEFAULT_User", "DEFAULT_Group"},
							},
						},
						RelationshipFindingsInput: model.RelationshipFindingsInput{
							{
								Name:                 "DEFAULT_Finding_1",
								DisplayName:          "Finding 1",
								SourceKindName:       "Base",
								RelationshipKindName: "DEFAULT_Relationship_Kind_1",
								EnvironmentKindName:  "DEFAULT_Domain",
								RemediationInput:     model.RemediationInput{},
							},
						},
					}).Return(true, nil)
				},
				func(t *testing.T, mock *schemamocks.MockGraphDBKindRepository) {
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
					NodeKindsInput: model.NodesInput{
						{
							Name: "DEFAULT_node kind 1",
						},
						{
							Name: "DEFAULT_Domain",
						},
						{
							Name: "DEFAULT_User",
						},
						{
							Name: "DEFAULT_Group",
						},
						{
							Name: "DEFAULT_AzureAD",
						},
					},
					RelationshipKindsInput: model.RelationshipsInput{{
						Name: "DEFAULT_Relationship_Kind_1",
					}},
					EnvironmentsInput: []model.EnvironmentInput{
						{
							EnvironmentKindName: "DEFAULT_Domain",
							SourceKindName:      "Base",
							PrincipalKinds:      []string{"DEFAULT_User"},
						},
						{
							EnvironmentKindName: "DEFAULT_AzureAD",
							SourceKindName:      "AzureHound",
							PrincipalKinds:      []string{"DEFAULT_User", "DEFAULT_Group"},
						},
					},
					RelationshipFindingsInput: model.RelationshipFindingsInput{
						{
							Name:                 "DEFAULT_Finding_1",
							DisplayName:          "Finding 1",
							SourceKindName:       "Base",
							RelationshipKindName: "DEFAULT_Relationship_Kind_1",
							EnvironmentKindName:  "DEFAULT_Domain",
							RemediationInput:     model.RemediationInput{},
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
			t.Parallel()

			var (
				mockCtrl = gomock.NewController(t)

				mockOpenGraphSchemaRepository = schemamocks.NewMockOpenGraphSchemaRepository(mockCtrl)
				mockGraphDBKindsRepository    = schemamocks.NewMockGraphDBKindRepository(mockCtrl)
			)

			defer mockCtrl.Finish()

			tt.fields.setupOpenGraphSchemaRepositoryMock(t, mockOpenGraphSchemaRepository)
			tt.fields.setupGraphDBKindsRepositoryMock(t, mockGraphDBKindsRepository)

			o := opengraphschema.NewOpenGraphSchemaService(mockOpenGraphSchemaRepository, mockGraphDBKindsRepository)
			updated, err := o.UpsertOpenGraphExtension(tt.args.ctx, tt.args.graphExtension)
			if tt.wantErr != nil {
				require.ErrorContains(t, err, tt.wantErr.Error(), "UpsertOpenGraphExtension() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if tt.wantUpdated != updated {
				require.Fail(t, "expected graph schema to be updated")
			}
		})
	}
}

func TestOpenGraphSchemaService_ListExtensions(t *testing.T) {
	t.Parallel()

	type mocks struct {
		mockOpenGraphSchema *schemamocks.MockOpenGraphSchemaRepository
	}
	type expected struct {
		extensions model.GraphSchemaExtensions
		err        error
	}
	tests := []struct {
		name       string
		setupMocks func(t *testing.T, m *mocks)
		expected   expected
	}{
		{
			name: "Error: openGraphSchemaRepository.GetGraphSchemaExtensions error",
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().GetGraphSchemaExtensions(
					gomock.Any(),
					model.Filters{},
					model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}},
					0, 0).Return(model.GraphSchemaExtensions{}, 0, errors.New("error"))
			},
			expected: expected{
				extensions: model.GraphSchemaExtensions{},
				err:        errors.New("error retrieving graph extensions: error"),
			},
		},
		{
			name: "Success: single extension",
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().GetGraphSchemaExtensions(
					gomock.Any(),
					model.Filters{},
					model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}},
					0, 0).Return(model.GraphSchemaExtensions{
					{
						Serial: model.Serial{
							ID: int32(1),
						},
						Name:        "Name 1",
						DisplayName: "Display Name 1",
						Version:     "v1.0.0",
						IsBuiltin:   false,
					},
				}, 1, nil,
				)
			},
			expected: expected{
				extensions: model.GraphSchemaExtensions{
					{
						Serial: model.Serial{
							ID: int32(1),
						},
						Name:        "Name 1",
						DisplayName: "Display Name 1",
						Version:     "v1.0.0",
						IsBuiltin:   false,
					},
				},
				err: nil,
			},
		},
		{
			name: "Success: multiple extensions",
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().GetGraphSchemaExtensions(
					gomock.Any(),
					model.Filters{},
					model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}},
					0, 0).Return(model.GraphSchemaExtensions{
					{
						Serial: model.Serial{
							ID: int32(1),
						},
						Name:        "Name 1",
						DisplayName: "Display Name 1",
						Version:     "v1.0.0",
						IsBuiltin:   false,
					},
					{
						Serial: model.Serial{
							ID: int32(2),
						},
						Name:        "Name 2",
						DisplayName: "Display Name 2",
						Version:     "v2.0.0",
						IsBuiltin:   true,
					},
				}, 2, nil,
				)
			},
			expected: expected{
				extensions: model.GraphSchemaExtensions{
					{
						Serial: model.Serial{
							ID: int32(1),
						},
						Name:        "Name 1",
						DisplayName: "Display Name 1",
						Version:     "v1.0.0",
					},
					{
						Serial: model.Serial{
							ID: int32(2),
						},
						Name:        "Name 2",
						DisplayName: "Display Name 2",
						Version:     "v2.0.0",
						IsBuiltin:   true,
					},
				},
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)

			m := &mocks{
				mockOpenGraphSchema: schemamocks.NewMockOpenGraphSchemaRepository(ctrl),
			}

			tt.setupMocks(t, m)

			service := opengraphschema.NewOpenGraphSchemaService(m.mockOpenGraphSchema, nil)

			res, err := service.ListExtensions(context.Background())

			if tt.expected.err != nil {
				assert.EqualError(t, err, tt.expected.err.Error())
				assert.Equal(t, tt.expected.extensions, res)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected.extensions, res)
			}
		})
	}
}

func TestOpenGraphSchemaService_DeleteExtension(t *testing.T) {
	t.Parallel()

	type mocks struct {
		mockOpenGraphSchema *schemamocks.MockOpenGraphSchemaRepository
		mockGraphDB         *schemamocks.MockGraphDBKindRepository
	}
	type args struct {
		extensionID int32
	}
	type expected struct {
		err error
	}
	tests := []struct {
		name       string
		args       args
		setupMocks func(t *testing.T, m *mocks)
		expected   expected
	}{
		{
			name: "Error: failed to delete graph schema extension",
			args: args{
				extensionID: int32(1),
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().DeleteGraphSchemaExtension(
					gomock.Any(), int32(1)).Return(errors.New("error"))
			},
			expected: expected{
				err: errors.New("error deleting graph extension: error"),
			},
		},
		{
			name: "Error: failed to refresh kinds after deleting extension",
			args: args{
				extensionID: int32(1),
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().DeleteGraphSchemaExtension(
					gomock.Any(), int32(1)).Return(nil)
				m.mockGraphDB.EXPECT().RefreshKinds(gomock.Any()).Return(errors.New("error"))
			},
			expected: expected{
				err: errors.New("error refreshing graph db kinds: error"),
			},
		},
		{
			name: "Success",
			args: args{
				extensionID: int32(1),
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().DeleteGraphSchemaExtension(
					gomock.Any(), int32(1)).Return(nil)
				m.mockGraphDB.EXPECT().RefreshKinds(gomock.Any()).Return(nil)
			},
			expected: expected{
				err: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctrl := gomock.NewController(t)

			m := &mocks{
				mockOpenGraphSchema: schemamocks.NewMockOpenGraphSchemaRepository(ctrl),
				mockGraphDB:         schemamocks.NewMockGraphDBKindRepository(ctrl),
			}

			tt.setupMocks(t, m)

			service := opengraphschema.NewOpenGraphSchemaService(m.mockOpenGraphSchema, m.mockGraphDB)

			err := service.DeleteExtension(context.Background(), tt.args.extensionID)
			if tt.expected.err != nil {
				assert.EqualError(t, err, tt.expected.err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
