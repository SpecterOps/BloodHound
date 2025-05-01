// Copyright 2025 Specter Ops, Inc.
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

package ein_test

import (
	"fmt"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/bloodhoundad/azurehound/v2/models"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/azure"
)

func TestConvertAzureRoleEligibilityScheduleInstanceToRel(t *testing.T) {
	testData := models.RoleEligibilityScheduleInstance{
		Id:               "lAPpYvVpN0KRkAEhdxReELKn6QMIlSROgkgWZy9fE3c-1-e",
		RoleDefinitionId: "62e90394-69f5-4237-9190-012177145e10",
		PrincipalId:      "03e9a7b2-9508-4e24-8248-16672f5f1377",
		DirectoryScopeId: "/",
		StartDateTime:    "2024-01-04T01:22:36.867Z",
		TenantId:         "6c12b0b0-b2cc-4a73-8252-0b94bfca2145",
	}

	expectedRels := ein.ConvertAzureRoleEligibilityScheduleInstanceToRel(testData)
	require.Len(t, expectedRels, 1)
	expectedRel := expectedRels[0]
	require.Equal(t, expectedRel.Source, strings.ToUpper(fmt.Sprintf("%s@%s", testData.RoleDefinitionId, testData.TenantId)))
	require.Equal(t, expectedRel.RelType, azure.AZRoleEligible)
	require.Equal(t, expectedRel.Target, strings.ToUpper(testData.PrincipalId))
}
