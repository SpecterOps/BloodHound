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

package bloodhoundgraph

import (
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type ManifestedRelationship struct {
	Start        *graph.Node
	End          *graph.Node
	Relationship *graph.Relationship
}

type ManifestedRelationshipSet []ManifestedRelationship

func ManifestedRelationshipSetToBloodHoundGraph(relationships ManifestedRelationshipSet) map[string]any {
	result := make(map[string]any)

	for _, rel := range relationships {
		result[fmt.Sprintf("rel_%d", rel.Relationship.ID)] = RelationshipToBloodHoundGraph(rel.Relationship)
		result[rel.Start.ID.String()] = NodeToBloodHoundGraph(rel.Start)
		result[rel.End.ID.String()] = NodeToBloodHoundGraph(rel.End)
	}

	return result
}
