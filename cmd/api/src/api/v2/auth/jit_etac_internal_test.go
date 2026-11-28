// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package auth

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/crewjam/saml"
	authModel "github.com/specterops/bloodhound/cmd/api/src/auth"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// TestJITUserUpsertETAC verifies SAML and OIDC persist environment access with role changes while preserving explicit access on unchanged users.
func TestJITUserUpsertETAC(t *testing.T) {
	var (
		roles = model.Roles{
			{Name: authModel.RoleUser, Serial: model.Serial{ID: 1}},
			{Name: authModel.RoleAdministrator, Serial: model.Serial{ID: 2}},
			{Name: authModel.RolePowerUser, Serial: model.Serial{ID: 3}},
			{Name: authModel.RoleAuditor, Serial: model.Serial{ID: 4}},
			{Name: authModel.RoleReadOnly, Serial: model.Serial{ID: 5}},
			{Name: authModel.RoleUploadOnly, Serial: model.Serial{ID: 6}},
		}
		updateFailure = errors.New("user persistence failed")
		testCases     = []struct {
			name                    string
			currentRole             int
			incomingRole            int
			etacEnabled             bool
			roleProvision           bool
			allEnvironments         bool
			environments            []model.EnvironmentTargetedAccessControl
			expectUpdate            bool
			expectedAllEnvironments bool
			updateError             error
		}{
			{
				name: "promote to administrator", incomingRole: 1,
				etacEnabled: true, roleProvision: true, expectUpdate: true, expectedAllEnvironments: true,
			},
			{
				name: "promote to power user", incomingRole: 2,
				etacEnabled: true, roleProvision: true, expectUpdate: true, expectedAllEnvironments: true,
			},
			{
				name: "promote to auditor", incomingRole: 3,
				etacEnabled: true, roleProvision: true, expectUpdate: true, expectedAllEnvironments: true,
			},
			{
				name: "demote to user", currentRole: 1,
				etacEnabled: true, roleProvision: true, allEnvironments: true, expectUpdate: true,
			},
			{
				name: "demote to read only", currentRole: 1, incomingRole: 4,
				etacEnabled: true, roleProvision: true, allEnvironments: true, expectUpdate: true,
			},
			{
				name: "demote to upload only", currentRole: 1, incomingRole: 5,
				etacEnabled: true, roleProvision: true, allEnvironments: true, expectUpdate: true,
			},
			{
				name: "disabled ETAC permits all environments on role change", currentRole: 1,
				roleProvision: true, expectUpdate: true, expectedAllEnvironments: true,
			},
			{
				name:        "unchanged role preserves manual all environment grant",
				etacEnabled: true, roleProvision: true, allEnvironments: true,
			},
			{
				name:        "unchanged role preserves assigned environment list",
				etacEnabled: true, roleProvision: true,
				environments: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "assigned-domain"}},
			},
			{
				name: "disabled role provisioning preserves existing access", currentRole: 4, incomingRole: 1,
				etacEnabled:  true,
				environments: []model.EnvironmentTargetedAccessControl{{EnvironmentID: "assigned-domain"}},
			},
			{
				name: "update failure is returned", incomingRole: 1,
				etacEnabled: true, roleProvision: true, expectUpdate: true, expectedAllEnvironments: true,
				updateError: updateFailure,
			},
		}
	)

	for _, protocol := range []string{"SAML", "OIDC"} {
		t.Run(protocol, func(t *testing.T) {
			for _, testCase := range testCases {
				t.Run(testCase.name, func(t *testing.T) {
					var (
						controller = gomock.NewController(t)
						database   = dbMocks.NewMockDatabase(controller)
						user       = model.User{
							PrincipalName:                    "user@example.test",
							Roles:                            model.Roles{roles[testCase.currentRole]},
							AllEnvironments:                  testCase.allEnvironments,
							EnvironmentTargetedAccessControl: testCase.environments,
						}
						provider = model.SSOProvider{
							SAMLProvider: &model.SAMLProvider{},
							Config: model.SSOProviderConfig{
								AutoProvision: model.SSOProviderAutoProvisionConfig{
									Enabled:       true,
									RoleProvision: testCase.roleProvision,
									DefaultRoleId: roles[0].ID,
								},
							},
						}
						roleClaim = "bh-" + strings.ReplaceAll(strings.ToLower(roles[testCase.incomingRole].Name), " ", "-")
						dogTags   = dogtags.NewTestService(dogtags.TestOverrides{
							Bools: map[dogtags.BoolDogTag]bool{dogtags.ETAC_ENABLED: testCase.etacEnabled},
						})
						err error
					)

					database.EXPECT().GetAllRoles(gomock.Any(), "", model.SQLFilter{}).Return(roles, nil)
					database.EXPECT().LookupUser(gomock.Any(), user.PrincipalName).Return(user, nil)
					if testCase.expectUpdate {
						database.EXPECT().UpdateUser(gomock.Any(), gomock.Any()).DoAndReturn(
							func(_ context.Context, updatedUser model.User) error {
								require.Equal(t, model.Roles{roles[testCase.incomingRole]}, updatedUser.Roles)
								require.Equal(t, testCase.expectedAllEnvironments, updatedUser.AllEnvironments)
								require.Equal(t, user.PrincipalName, updatedUser.PrincipalName)
								return testCase.updateError
							})
					}

					switch protocol {
					case "SAML":
						assertion := &saml.Assertion{
							AttributeStatements: []saml.AttributeStatement{{
								Attributes: []saml.Attribute{{
									Name:   model.MicrosoftClaimsRole,
									Values: []saml.AttributeValue{{Value: roleClaim}},
								}},
							}},
						}
						err = jitSAMLUserUpsert(context.Background(), provider, user.PrincipalName, assertion, database, dogTags)
					case "OIDC":
						claims := oidcClaims{Email: user.PrincipalName, Roles: []string{roleClaim}}
						err = jitOIDCUserUpsert(context.Background(), provider, claims, database, dogTags)
					}

					if testCase.updateError != nil {
						require.ErrorContains(t, err, "update user: "+testCase.updateError.Error())
					} else {
						require.NoError(t, err)
					}
				})
			}
		})
	}
}
