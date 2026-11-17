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

func TestIsCertTemplateValidForESC16(t *testing.T) {
	var (
		testCases = []struct {
			name                  string
			schemaVersion         any
			authorizedSignatures  any
			authenticationEnabled any
			requiresApproval      any
			expected              bool
			expectError           bool
		}{
			{name: "schema one without signatures", schemaVersion: float64(1), authenticationEnabled: true, requiresApproval: false, expected: true},
			{name: "schema one ignores signatures", schemaVersion: float64(1), authorizedSignatures: float64(2), authenticationEnabled: true, requiresApproval: false, expected: true},
			{name: "schema one still requires authentication", schemaVersion: float64(1), authenticationEnabled: false, requiresApproval: false},
			{name: "schema one requires authentication property", schemaVersion: float64(1), requiresApproval: false, expectError: true},
			{name: "schema one still rejects manager approval", schemaVersion: float64(1), authenticationEnabled: true, requiresApproval: true},
			{name: "schema one requires manager approval property", schemaVersion: float64(1), authenticationEnabled: true, expectError: true},
			{name: "schema two with zero signatures", schemaVersion: float64(2), authorizedSignatures: float64(0), authenticationEnabled: true, requiresApproval: false, expected: true},
			{name: "schema three with zero signatures", schemaVersion: float64(3), authorizedSignatures: float64(0), authenticationEnabled: true, requiresApproval: false, expected: true},
			{name: "schema two with positive signatures", schemaVersion: float64(2), authorizedSignatures: float64(1), authenticationEnabled: true, requiresApproval: false},
			{name: "schema two with negative signatures", schemaVersion: float64(2), authorizedSignatures: float64(-1), authenticationEnabled: true, requiresApproval: false},
			{name: "schema two without signatures", schemaVersion: float64(2), authenticationEnabled: true, requiresApproval: false, expectError: true},
			{name: "missing schema version", authorizedSignatures: float64(0), authenticationEnabled: true, requiresApproval: false, expectError: true},
			{name: "zero schema version", schemaVersion: float64(0), authorizedSignatures: float64(0), authenticationEnabled: true, requiresApproval: false},
			{name: "negative schema version", schemaVersion: float64(-1), authorizedSignatures: float64(0), authenticationEnabled: true, requiresApproval: false},
		}
	)

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				properties = graph.NewProperties()
			)
			if testCase.schemaVersion != nil {
				properties.Set(adSchema.SchemaVersion.String(), testCase.schemaVersion)
			}
			if testCase.authorizedSignatures != nil {
				properties.Set(adSchema.AuthorizedSignatures.String(), testCase.authorizedSignatures)
			}
			if testCase.authenticationEnabled != nil {
				properties.Set(adSchema.AuthenticationEnabled.String(), testCase.authenticationEnabled)
			}
			if testCase.requiresApproval != nil {
				properties.Set(adSchema.RequiresManagerApproval.String(), testCase.requiresApproval)
			}

			valid, err := isCertTemplateValidForESC16(graph.NewNode(1, properties, adSchema.CertTemplate))
			if testCase.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.expected, valid)
		})
	}
}
