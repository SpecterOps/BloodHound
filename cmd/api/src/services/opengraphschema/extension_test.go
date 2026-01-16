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
		{
			name: "Error: openGraphSchemaRepository.UpsertGraphSchemaExtension error",
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
				expectedEnvs := []database.EnvironmentInput{
					{
						EnvironmentKindName: "Domain",
						SourceKindName:      "Base",
						PrincipalKinds:      []string{"User"},
					},
				}
				m.mockOpenGraphSchema.EXPECT().UpsertGraphSchemaExtension(
					gomock.Any(),
					int32(1),
					expectedEnvs,
				).Return(errors.New("error"))
			},
			expected: errors.New("error upserting graph extension: error"),
		},
		{
			name: "Success: single environment",
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
				expectedEnvs := []database.EnvironmentInput{
					{
						EnvironmentKindName: "Domain",
						SourceKindName:      "Base",
						PrincipalKinds:      []string{"User", "Computer"},
					},
				}
				m.mockOpenGraphSchema.EXPECT().UpsertGraphSchemaExtension(
					gomock.Any(),
					int32(1),
					expectedEnvs,
				).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Success: multiple environments",
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
				expectedEnvs := []database.EnvironmentInput{
					{
						EnvironmentKindName: "Domain",
						SourceKindName:      "Base",
						PrincipalKinds:      []string{"User"},
					},
					{
						EnvironmentKindName: "AzureAD",
						SourceKindName:      "AzureHound",
						PrincipalKinds:      []string{"User", "Group"},
					},
				}
				m.mockOpenGraphSchema.EXPECT().UpsertGraphSchemaExtension(
					gomock.Any(),
					int32(1),
					expectedEnvs,
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
