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

package ad

import (
	"testing"

	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestIsStartCertTemplateValidESC3(t *testing.T) {
	testCases := []struct {
		name                 string
		schemaVersion        any
		authorizedSignatures any
		requiresApproval     any
		expected             bool
	}{
		{
			name:             "schema version one ignores missing authorized signatures",
			schemaVersion:    float64(1),
			requiresApproval: false,
			expected:         true,
		},
		{
			name:                 "schema version one ignores authorized signatures",
			schemaVersion:        float64(1),
			authorizedSignatures: float64(2),
			requiresApproval:     false,
			expected:             true,
		},
		{
			name:                 "schema version two requires zero authorized signatures",
			schemaVersion:        float64(2),
			authorizedSignatures: float64(0),
			requiresApproval:     false,
			expected:             true,
		},
		{
			name:                 "schema version two rejects positive authorized signatures",
			schemaVersion:        float64(2),
			authorizedSignatures: float64(1),
			requiresApproval:     false,
		},
		{
			name:                 "schema version two rejects negative authorized signatures",
			schemaVersion:        float64(2),
			authorizedSignatures: float64(-1),
			requiresApproval:     false,
		},
		{
			name:                 "missing schema version is invalid",
			authorizedSignatures: float64(0),
			requiresApproval:     false,
		},
		{
			name:                 "malformed schema version is invalid",
			schemaVersion:        "not-a-version",
			authorizedSignatures: float64(0),
			requiresApproval:     false,
		},
		{
			name:                 "zero schema version is invalid",
			schemaVersion:        float64(0),
			authorizedSignatures: float64(0),
			requiresApproval:     false,
		},
		{
			name:                 "negative schema version is invalid",
			schemaVersion:        float64(-1),
			authorizedSignatures: float64(0),
			requiresApproval:     false,
		},
		{
			name:                 "manager approval is invalid",
			schemaVersion:        float64(1),
			authorizedSignatures: float64(0),
			requiresApproval:     true,
		},
		{
			name:                 "missing manager approval is invalid",
			schemaVersion:        float64(1),
			authorizedSignatures: float64(0),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			properties := graph.NewProperties()
			if testCase.schemaVersion != nil {
				properties.Set(adSchema.SchemaVersion.String(), testCase.schemaVersion)
			}
			if testCase.authorizedSignatures != nil {
				properties.Set(adSchema.AuthorizedSignatures.String(), testCase.authorizedSignatures)
			}
			if testCase.requiresApproval != nil {
				properties.Set(adSchema.RequiresManagerApproval.String(), testCase.requiresApproval)
			}

			certTemplate := graph.NewNode(1, properties, adSchema.CertTemplate)
			require.Equal(t, testCase.expected, isStartCertTemplateValidESC3(certTemplate))
		})
	}
}

func TestEnterpriseCAHasEnrollmentAgentRestrictions(t *testing.T) {
	testCases := []struct {
		name            string
		collected       any
		hasRestrictions any
		expected        bool
	}{
		{name: "properties absent"},
		{name: "not collected", collected: false, hasRestrictions: true},
		{name: "collected and unrestricted", collected: true, hasRestrictions: false},
		{name: "collected with missing restriction property", collected: true},
		{name: "collected with malformed restriction property", collected: true, hasRestrictions: "true"},
		{name: "collected with restrictions", collected: true, hasRestrictions: true, expected: true},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			properties := graph.NewProperties()
			if testCase.collected != nil {
				properties.Set(adSchema.EnrollmentAgentRestrictionsCollected.String(), testCase.collected)
			}
			if testCase.hasRestrictions != nil {
				properties.Set(adSchema.HasEnrollmentAgentRestrictions.String(), testCase.hasRestrictions)
			}

			enterpriseCA := graph.NewNode(1, properties, adSchema.EnterpriseCA)
			require.Equal(t, testCase.expected, enterpriseCAHasEnrollmentAgentRestrictions(enterpriseCA))
		})
	}
}
