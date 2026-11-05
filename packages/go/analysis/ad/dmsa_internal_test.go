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

func TestDMSADomainHasUnpatchedKDC(t *testing.T) {
	testCases := []struct {
		name     string
		builds   []windowsBuildVersion
		expected bool
	}{
		{name: "no collected builds", expected: false},
		{name: "earlier server release", builds: []windowsBuildVersion{{Build: 20348, Revision: 9999}}, expected: false},
		{name: "unpatched server 2025", builds: []windowsBuildVersion{{Build: 26100, Revision: 4945}}, expected: true},
		{name: "first patched server 2025", builds: []windowsBuildVersion{{Build: 26100, Revision: 4946}}, expected: false},
		{name: "later server 2025 update", builds: []windowsBuildVersion{{Build: 26100, Revision: 5000}}, expected: false},
		{
			name: "mixed server 2025 patch levels",
			builds: []windowsBuildVersion{
				{Build: 26100, Revision: 6000},
				{Build: 26100, Revision: 4000},
			},
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, domainHasUnpatchedDMSAKDC(testCase.builds))
		})
	}
}

func TestDMSAVictimRights(t *testing.T) {
	testCases := []struct {
		name     string
		rights   graph.Kinds
		expected bool
	}{
		{name: "generic all", rights: graph.Kinds{adSchema.GenericAll}, expected: true},
		{name: "generic write", rights: graph.Kinds{adSchema.GenericWrite}, expected: true},
		{
			name: "both required attributes",
			rights: graph.Kinds{
				adSchema.WriteSupersededManagedAccountLink,
				adSchema.WriteSupersededServiceAccountState,
			},
			expected: true,
		},
		{name: "managed account link only", rights: graph.Kinds{adSchema.WriteSupersededManagedAccountLink}, expected: false},
		{name: "service account state only", rights: graph.Kinds{adSchema.WriteSupersededServiceAccountState}, expected: false},
		{name: "unrelated right", rights: graph.Kinds{adSchema.WriteSPN}, expected: false},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, qualifiesBadSuccessorVictim(testCase.rights))
		})
	}
}
