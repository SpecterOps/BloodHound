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

package azure_test

import (
	"slices"
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

func TestEntraDSRelationshipTraversability(t *testing.T) {
	for _, relationship := range []struct {
		kind        graph.Kind
		traversable bool
		control     bool
	}{
		{kind: azure.EntraDSContributor, traversable: false, control: false},
		{kind: azure.ManageEntraDS, traversable: true, control: true},
		{kind: azure.SyncedToEntraDSUser, traversable: true, control: false},
		{kind: azure.SyncedToEntraDSGroup, traversable: false, control: false},
		{kind: azure.AddEntraDSGroupMember, traversable: true, control: false},
		{kind: azure.EntraDSFor, traversable: false, control: false},
		{kind: azure.ManageEntraDSSync, traversable: true, control: false},
		{kind: azure.ManageEntraDSSyncFilter, traversable: true, control: false},
	} {
		if !slices.Contains(azure.Relationships(), relationship.kind) {
			t.Errorf("%s must remain a recognized relationship kind", relationship.kind)
		}

		if actual := slices.Contains(azure.ControlRelationships(), relationship.kind); actual != relationship.control {
			t.Errorf("%s control relationship status: got %t, want %t", relationship.kind, actual, relationship.control)
		}

		if actual := slices.Contains(azure.PathfindingRelationships(), relationship.kind); actual != relationship.traversable {
			t.Errorf("%s pathfinding relationship status: got %t, want %t", relationship.kind, actual, relationship.traversable)
		}
	}
}
