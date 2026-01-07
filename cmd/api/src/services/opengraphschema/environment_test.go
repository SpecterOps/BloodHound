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
	"testing"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema"
	schemamocks "github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOpenGraphSchemaService_UpsertSchemaEnvironmentWithPrincipalKinds(t *testing.T) {
	type mocks struct {
		mockOpenGraphSchema *schemamocks.MockOpenGraphSchemaRepository
	}
	type args struct {
		schemaExtensionId int32
		environments      []v2.Environment
	}
	tests := []struct {
		name       string
		mocks      mocks
		setupMocks func(t *testing.T, mock *mocks)
		args       args
		expected   error
	}{
		// Validation: Environment Kind
		{
			name: "Error: openGraphSchemaRepository.GetKindByName environment kind name not found in the database",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{}, database.ErrNotFound)
			},
			expected: errors.New("error validating and translating environment: environment kind 'Domain' not found"),
		},
		{
			name: "Error: openGraphSchemaRepository.GetKindByName failed to retrieve environment kind from database",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{}, errors.New("error"))
			},
			expected: errors.New("error validating and translating environment: error retrieving environment kind 'Domain': error"),
		},
		// Validation: Source Kind
		{
			name: "Error: validateAndTranslateSourceKind failed to retrieve source kind from database",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 1}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{}, errors.New("error"))
			},
			expected: errors.New("error validating and translating environment: error retrieving source kind 'Base': error"),
		},
		{
			name: "Error: validateAndTranslateSourceKind source kind name doesn't exist in database, registration fails",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 1}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().RegisterSourceKind(gomock.Any()).Return(func(kind graph.Kind) error {
					return errors.New("error")
				})
			},
			expected: errors.New("error validating and translating environment: error registering source kind 'Base': error"),
		},
		{
			name: "Error: validateAndTranslateSourceKind source kind name doesn't exist in database, registration succeeds but fetch fails",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 1}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().RegisterSourceKind(gomock.Any()).Return(func(kind graph.Kind) error {
					return nil
				})
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{}, errors.New("error"))
			},
			expected: errors.New("error validating and translating environment: error retrieving newly registered source kind 'Base': error"),
		},
		// Validation: Principal Kind
		{
			name: "Error: validateAndTranslatePrincipalKinds principal kind not found in database",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User", "InvalidKind"},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 1}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 2}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "InvalidKind").Return(model.Kind{}, database.ErrNotFound)
			},
			expected: errors.New("error validating and translating environment: principal kind 'InvalidKind' not found"),
		},
		{
			name: "Error: validateAndTranslatePrincipalKinds failed to retrieve principal kind from database",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User", "InvalidKind"},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 1}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 2}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "InvalidKind").Return(model.Kind{}, errors.New("error"))
			},
			expected: errors.New("error validating and translating environment: error retrieving principal kind by name 'InvalidKind': error"),
		},
		// Upsert Schema Environment
		{
			name: "Error: upsertSchemaEnvironment error retrieving schema environment from database",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error retrieving schema environment id 0: error"),
		},
		{
			name: "Error: upsertSchemaEnvironment error deleting schema environment",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 5},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(5)).Return(errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error deleting schema environment 5: error"),
		},
		{
			name: "Error: upsertSchemaEnvironment error creating schema environment after deletion",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 5},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironment(gomock.Any(), int32(5)).Return(nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error creating schema environment: error"),
		},
		{
			name: "Error: upsertSchemaEnvironment error creating new schema environment",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{}, errors.New("error"))
			},
			expected: errors.New("error upserting schema environment: error creating schema environment: error"),
		},
		// Upsert Principal Kinds
		{
			name: "Error: upsertPrincipalKinds error getting principal kinds by environment id",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User"},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Validation and translation
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 3}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, errors.New("error"))
			},
			expected: errors.New("error upserting principal kinds: error retrieving existing principal kinds for environment 10: error"),
		},
		{
			name: "Error: upsertPrincipalKinds error deleting principal kinds",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User"},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Validation and translation
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 3}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return([]model.SchemaEnvironmentPrincipalKind{
					{
						EnvironmentId: int32(10),
						PrincipalKind: int32(5),
					},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().DeleteSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(5)).Return(errors.New("error"))
			},
			expected: errors.New("error upserting principal kinds: error deleting principal kind 5 for environment 10: error"),
		},
		{
			name: "Error: upsertPrincipalKinds error creating principal kinds",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User"},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Validation and translation
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 3}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(3)).Return(model.SchemaEnvironmentPrincipalKind{}, errors.New("error"))
			},
			expected: errors.New("error upserting principal kinds: error creating principal kind 3 for environment 10: error"),
		},
		{
			name: "Success: Create new environment with principal kinds",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User", "Computer"},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Validation and translation
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 4}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Computer").Return(model.Kind{ID: 5}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(4)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(5)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)
			},
			expected: nil,
		},
		{
			name: "Success: Create environment with source kind registration",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "NewSource",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// Validation and translation
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				// Source kind not found, register it
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "NewSource").Return(database.SourceKind{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().RegisterSourceKind(gomock.Any()).Return(func(kind graph.Kind) error {
					return nil
				})
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "NewSource").Return(database.SourceKind{ID: 10}, nil)

				// Environment upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(10)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)

				// Principal kinds upsert
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, database.ErrNotFound)
			},
			expected: nil,
		},
		{
			name: "Success: Process multiple environments",
			args: args{
				schemaExtensionId: int32(3),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User"},
					},
					{
						EnvironmentKind: "AzureAD",
						SourceKind:      "AzureHound",
						PrincipalKinds:  []string{"User", "Group"},
					},
				},
			},
			setupMocks: func(t *testing.T, mocks *mocks) {
				t.Helper()
				// First environment
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Domain").Return(model.Kind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "Base").Return(database.SourceKind{ID: 3}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 4}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(3), int32(3)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 10},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(10)).Return(nil, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(10), int32(4)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)

				// Second environment
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "AzureAD").Return(model.Kind{ID: 5}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSourceKindByName(gomock.Any(), "AzureHound").Return(database.SourceKind{ID: 6}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "User").Return(model.Kind{ID: 4}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetKindByName(gomock.Any(), "Group").Return(model.Kind{ID: 7}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentById(gomock.Any(), int32(0)).Return(model.SchemaEnvironment{}, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironment(gomock.Any(), int32(3), int32(5), int32(6)).Return(model.SchemaEnvironment{
					Serial: model.Serial{ID: 11},
				}, nil)
				mocks.mockOpenGraphSchema.EXPECT().GetSchemaEnvironmentPrincipalKindsByEnvironmentId(gomock.Any(), int32(11)).Return(nil, database.ErrNotFound)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(11), int32(4)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)
				mocks.mockOpenGraphSchema.EXPECT().CreateSchemaEnvironmentPrincipalKind(gomock.Any(), int32(11), int32(7)).Return(model.SchemaEnvironmentPrincipalKind{}, nil)
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

			err := graphService.UpsertSchemaEnvironmentWithPrincipalKinds(context.Background(), tt.args.schemaExtensionId, tt.args.environments)
			if tt.expected != nil {
				assert.EqualError(t, err, tt.expected.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
