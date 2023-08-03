// Copyright 2023 Specter Ops, Inc.
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

package v2

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
)

func Test_parseRelationshipKindsParam(t *testing.T) {
	validKinds := graph.Kinds(ad.Relationships()).Concatenate(azure.Relationships())

	// Default case
	kinds, operator, err := parseRelationshipKindsParam(validKinds, "")

	require.Nil(t, err)
	require.Equal(t, "in", operator)
	require.Equal(t, len(validKinds), len(kinds))

	// Valid parameter definition
	kinds, operator, err = parseRelationshipKindsParam(validKinds, "in:Contains,GenericAll,AZResetPassword")

	require.Nil(t, err)
	require.Equal(t, "in", operator)
	require.Equal(t, 3, len(kinds))
	require.True(t, kinds.ContainsOneOf(ad.Contains))
	require.True(t, kinds.ContainsOneOf(ad.GenericAll))
	require.True(t, kinds.ContainsOneOf(azure.ResetPassword))

	// Expect an error if we can't find a matching kind
	_, _, err = parseRelationshipKindsParam(validKinds, "in:Contains,GenericAll,NOTAKIND")
	require.NotNil(t, err)

	// Expect an error if the operator is broken
	_, _, err = parseRelationshipKindsParam(validKinds, "LOLNO:Contains,GenericAll")
	require.NotNil(t, err)
}
