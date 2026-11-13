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

package ad_test

import (
	"testing"

	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func kindsContain(kinds []graph.Kind, target graph.Kind) bool {
	for _, k := range kinds {
		if k.Is(target) {
			return true
		}
	}

	return false
}

// TestDMSASchemaRegistration locks the dMSA schema contract so a future `cue` regeneration can't silently drop the
// node kind or any edge from the lists that make them ingestable, post-processable, and pathfindable.
func TestDMSASchemaRegistration(t *testing.T) {
	require.True(t, kindsContain(adSchema.NodeKinds(), adSchema.DMSA), "DMSA must be a registered node kind")

	for _, edge := range []graph.Kind{
		adSchema.DMSAPrecededBy,
		adSchema.CreateDMSA,
		adSchema.BadSuccessor,
		adSchema.OuroborosSelfEnroll,
		adSchema.WriteSupersededManagedAccountLink,
		adSchema.WriteSupersededServiceAccountState,
	} {
		require.Truef(t, kindsContain(adSchema.Relationships(), edge), "%s must be a registered relationship kind", edge)
	}

	for _, edge := range []graph.Kind{adSchema.DMSAPrecededBy, adSchema.CreateDMSA, adSchema.BadSuccessor, adSchema.OuroborosSelfEnroll} {
		require.Truef(t, kindsContain(adSchema.PathfindingRelationships(), edge), "%s must be pathfinding-traversable", edge)
	}

	// CreateDMSA and the attribute write permissions are ingest-time ACL edges routed by ein.ParseACEData;
	// BadSuccessor and OuroborosSelfEnroll are ACL edges produced by post-processing.
	for _, acl := range []graph.Kind{
		adSchema.CreateDMSA,
		adSchema.BadSuccessor,
		adSchema.OuroborosSelfEnroll,
		adSchema.WriteSupersededManagedAccountLink,
		adSchema.WriteSupersededServiceAccountState,
	} {
		require.Truef(t, kindsContain(adSchema.ACLRelationships(), acl), "%s must be an ACL relationship", acl)
		require.Truef(t, adSchema.IsACLKind(acl), "%s must be recognized by IsACLKind", acl)
	}

	// DMSAPrecededBy is a structural ingest edge (from msDS-ManagedAccountPrecededByLink), not an ACL edge.
	require.False(t, kindsContain(adSchema.ACLRelationships(), adSchema.DMSAPrecededBy), "DMSAPrecededBy must not be an ACL relationship")
	require.False(t, adSchema.IsACLKind(adSchema.DMSAPrecededBy), "DMSAPrecededBy must not be an ACL kind")
}
