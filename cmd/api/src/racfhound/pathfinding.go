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

package racfhound

import "github.com/specterops/dawgs/graph"

var pathfindingRelationshipNames = []string{
	"RACFMemberOf",
	"RACFHasSubgroup",
	"RACFSubgroupOf",
	"RACFOwns",
	"RACFHasPrivilege",
	"RACFClassAuth",
	"RACFCanRead",
	"RACFCanUpdate",
	"RACFCanControl",
	"RACFCanAlter",
	"RACFCanWrite",
	"RACFSurrogateFor",
	"RACFCanSubmitAs",
	"RACFStartedAs",
	"RACFAffects",
}

// PathfindingRelationships returns the control relationships Pathfinder may
// traverse. Legacy and namespaced kinds are included while RACFHound migrates
// to its versioned OpenGraph contract.
func PathfindingRelationships() graph.Kinds {
	relationshipKinds := make(graph.Kinds, 0, len(pathfindingRelationshipNames)*2)

	for _, relationshipName := range pathfindingRelationshipNames {
		relationshipKinds = append(
			relationshipKinds,
			graph.StringKind(relationshipName),
			graph.StringKind("racf_"+relationshipName),
		)
	}

	return relationshipKinds
}
