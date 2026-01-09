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
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema"
	schemamocks "github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema/mocks"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestOpenGraphSchemaService_UpsertGraphSchemaExtension(t *testing.T) {
	type mocks struct {
		mockOpenGraphSchema *schemamocks.MockOpenGraphSchemaRepository
	}
	type args struct {
		environments []v2.Environment
	}
	tests := []struct {
		name       string
		setupMocks func(t *testing.T, m *mocks)
		args       args
		expected   error
	}{
		// UpsertSchemaEnvironmentWithPrincipalKinds
		// Validation: Environment Kind
		{
			name: "Error: environment kind name not found in the database",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("environment kind 'Domain' not found"))
			},
			expected: errors.New("failed to upload environments with principal kinds: environment kind 'Domain' not found"),
		},
		{
			name: "Error: failed to retrieve environment kind from database",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error retrieving environment kind 'Domain': error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error retrieving environment kind 'Domain': error"),
		},
		// Validation: Source Kind
		{
			name: "Error: failed to retrieve source kind from database",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error retrieving source kind 'Base': error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error retrieving source kind 'Base': error"),
		},
		{
			name: "Error: source kind name doesn't exist in database, registration fails",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error registering source kind 'Base': error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error registering source kind 'Base': error"),
		},
		{
			name: "Error: source kind name doesn't exist in database, registration succeeds but fetch fails",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error retrieving newly registered source kind 'Base': error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error retrieving newly registered source kind 'Base': error"),
		},
		// Validation: Principal Kind
		{
			name: "Error: principal kind not found in database",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User", "InvalidKind"},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{"User", "InvalidKind"},
				).Return(errors.New("principal kind 'InvalidKind' not found"))
			},
			expected: errors.New("failed to upload environments with principal kinds: principal kind 'InvalidKind' not found"),
		},
		{
			name: "Error: failed to retrieve principal kind from database",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User", "InvalidKind"},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{"User", "InvalidKind"},
				).Return(errors.New("error retrieving principal kind by name 'InvalidKind': error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error retrieving principal kind by name 'InvalidKind': error"),
		},
		// Upsert Schema Environment
		{
			name: "Error: error retrieving schema environment from database",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error upserting schema environment: error retrieving schema environment id 0: error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error upserting schema environment: error retrieving schema environment id 0: error"),
		},
		{
			name: "Error: error deleting schema environment",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error upserting schema environment: error deleting schema environment 5: error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error upserting schema environment: error deleting schema environment 5: error"),
		},
		{
			name: "Error: error creating schema environment after deletion",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error upserting schema environment: error creating schema environment: error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error upserting schema environment: error creating schema environment: error"),
		},
		{
			name: "Error: error creating new schema environment",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{},
				).Return(errors.New("error upserting schema environment: error creating schema environment: error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error upserting schema environment: error creating schema environment: error"),
		},
		// Upsert Principal Kinds
		{
			name: "Error: error getting principal kinds by environment id",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User"},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{"User"},
				).Return(errors.New("error upserting principal kinds: error retrieving existing principal kinds for environment 10: error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error upserting principal kinds: error retrieving existing principal kinds for environment 10: error"),
		},
		{
			name: "Error: openGraphSchemaRepository.UpsertSchemaEnvironmentWithPrincipalKinds error deleting principal kinds",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User"},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{"User"},
				).Return(errors.New("error upserting principal kinds: error deleting principal kind 5 for environment 10: error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error upserting principal kinds: error deleting principal kind 5 for environment 10: error"),
		},
		{
			name: "Error: openGraphSchemaRepository.UpsertSchemaEnvironmentWithPrincipalKinds error creating principal kinds",
			args: args{
				schemaExtensionId: int32(1),
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User"},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{"User"},
				).Return(errors.New("error upserting principal kinds: error creating principal kind 3 for environment 10: error"))
			},
			expected: errors.New("failed to upload environments with principal kinds: error upserting principal kinds: error creating principal kind 3 for environment 10: error"),
		},
		{
			name: "Success: Create new environment with principal kinds",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User", "Computer"},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{"User", "Computer"},
				).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Success: Create environment with source kind registration",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "NewSource",
						PrincipalKinds:  []string{},
					},
				},
			},
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"NewSource",
					[]string{},
				).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Success: Process multiple environments",
			args: args{
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
			setupMocks: func(t *testing.T, m *mocks) {
				t.Helper()
				// First environment
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"Domain",
					"Base",
					[]string{"User"},
				).Return(nil)

				// Second environment
				m.mockOpenGraphSchema.EXPECT().UpsertSchemaEnvironmentWithPrincipalKinds(
					gomock.Any(),
					int32(1),
					"AzureAD",
					"AzureHound",
					[]string{"User", "Group"},
				).Return(nil)
			},
			expected: nil,
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

			service := opengraphschema.NewOpenGraphSchemaService(m.mockOpenGraphSchema)

			err := service.UpsertGraphSchemaExtension(context.Background(), v2.GraphSchemaExtension{
				Environments: tt.args.environments,
			})

			if tt.expected != nil {
				assert.EqualError(t, err, tt.expected.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
