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
		findings     []v2.Finding
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
				findings: []v2.Finding{
					{
						Name:             "Finding",
						DisplayName:      "DisplayName",
						RelationshipKind: "Domain",
						EnvironmentKind:  "Domain",
						SourceKind:       "Base",
						Remediation: v2.Remediation{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
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
				expectedFindings := []database.FindingInput{
					{
						Name:                 "Finding",
						DisplayName:          "DisplayName",
						RelationshipKindName: "Domain",
						EnvironmentKindName:  "Domain",
						SourceKindName:       "Base",
						RemediationInput: database.RemediationInput{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
					},
				}
				m.mockOpenGraphSchema.EXPECT().UpsertGraphSchemaExtension(
					gomock.Any(),
					int32(1),
					expectedEnvs,
					expectedFindings,
				).Return(errors.New("error"))
			},
			expected: errors.New("error upserting graph extension: error"),
		},
		{
			name: "Success: single environment with single finding",
			args: args{
				environments: []v2.Environment{
					{
						EnvironmentKind: "Domain",
						SourceKind:      "Base",
						PrincipalKinds:  []string{"User", "Computer"},
					},
				},
				findings: []v2.Finding{
					{
						Name:             "Finding",
						DisplayName:      "DisplayName",
						RelationshipKind: "Domain",
						EnvironmentKind:  "Domain",
						SourceKind:       "Base",
						Remediation: v2.Remediation{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
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
				expectedFindings := []database.FindingInput{
					{
						Name:                 "Finding",
						DisplayName:          "DisplayName",
						RelationshipKindName: "Domain",
						EnvironmentKindName:  "Domain",
						SourceKindName:       "Base",
						RemediationInput: database.RemediationInput{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
					},
				}
				m.mockOpenGraphSchema.EXPECT().UpsertGraphSchemaExtension(
					gomock.Any(),
					int32(1),
					expectedEnvs,
					expectedFindings,
				).Return(nil)
			},
			expected: nil,
		},
		{
			name: "Success: multiple environments with multiple findings",
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
				findings: []v2.Finding{
					{
						Name:             "Finding1",
						DisplayName:      "DisplayName1",
						RelationshipKind: "Domain",
						EnvironmentKind:  "Domain",
						SourceKind:       "Base",
						Remediation: v2.Remediation{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
					},
					{
						Name:             "Finding2",
						DisplayName:      "DisplayName2",
						RelationshipKind: "Domain",
						EnvironmentKind:  "Domain",
						SourceKind:       "Base",
						Remediation: v2.Remediation{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
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
				expectedFindings := []database.FindingInput{
					{
						Name:                 "Finding1",
						DisplayName:          "DisplayName1",
						RelationshipKindName: "Domain",
						EnvironmentKindName:  "Domain",
						SourceKindName:       "Base",
						RemediationInput: database.RemediationInput{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
					},
					{
						Name:                 "Finding2",
						DisplayName:          "DisplayName2",
						RelationshipKindName: "Domain",
						EnvironmentKindName:  "Domain",
						SourceKindName:       "Base",
						RemediationInput: database.RemediationInput{
							ShortDescription: "Short Description",
							LongDescription:  "Long Description",
							ShortRemediation: "Short Remediation",
							LongRemediation:  "Long Remediation",
						},
					},
				}
				m.mockOpenGraphSchema.EXPECT().UpsertGraphSchemaExtension(
					gomock.Any(),
					int32(1),
					expectedEnvs,
					expectedFindings,
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
				Findings:     tt.args.findings,
			})

			if tt.expected != nil {
				assert.EqualError(t, err, tt.expected.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
