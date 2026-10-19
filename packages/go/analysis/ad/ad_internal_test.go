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

	adschema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestIsManagedServiceAccount(t *testing.T) {
	testCases := []struct {
		name                  string
		properties            *graph.Properties
		expected              bool
		expectedPropertyError bool
	}{
		{
			name:       "gMSA",
			properties: graph.NewProperties().Set(adschema.GMSA.String(), true),
			expected:   true,
		},
		{
			name:       "sMSA",
			properties: graph.NewProperties().Set(adschema.MSA.String(), true),
			expected:   true,
		},
		{
			name: "both properties false",
			properties: graph.NewProperties().Set(adschema.GMSA.String(), false).
				Set(adschema.MSA.String(), false),
		},
		{
			name:       "properties absent",
			properties: graph.NewProperties(),
		},
		{
			name:                  "invalid gMSA property",
			properties:            graph.NewProperties().Set(adschema.GMSA.String(), "true"),
			expectedPropertyError: true,
		},
		{
			name:                  "invalid sMSA property",
			properties:            graph.NewProperties().Set(adschema.MSA.String(), "true"),
			expectedPropertyError: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			node := graph.NewNode(1, testCase.properties, adschema.User)

			actual, err := isManagedServiceAccount(node)
			if testCase.expectedPropertyError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, testCase.expected, actual)
		})
	}
}
