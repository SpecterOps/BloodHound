package ein_test

import (
	"fmt"
	"github.com/bloodhoundad/azurehound/v2/models"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
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
	assert.Len(t, expectedRels, 1)
	expectedRel := expectedRels[0]
	assert.Equal(t, expectedRel.Source, strings.ToUpper(fmt.Sprintf("%s@%s", testData.RoleDefinitionId, testData.TenantId)))
	assert.Equal(t, expectedRel.RelType, azure.AZRoleEligible)
	assert.Equal(t, expectedRel.Target, strings.ToUpper(testData.PrincipalId))
}
