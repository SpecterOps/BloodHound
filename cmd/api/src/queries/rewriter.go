// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package queries

import (
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/cypher/models/walk"
)

type Rewriter struct {
	walk.Visitor[cypher.SyntaxNode]

	HasMutation                 bool
	HasRelationshipTypeShortcut bool
}

func NewRewriter() *Rewriter {
	return &Rewriter{
		Visitor: walk.NewVisitor[cypher.SyntaxNode](),
	}
}

func (s *Rewriter) Enter(node cypher.SyntaxNode) {
	switch typedNode := node.(type) {
	case *cypher.UpdatingClause:
		s.HasMutation = true

	case *cypher.RelationshipPattern:
		for _, kind := range typedNode.Kinds {
			switch kind.String() {
			case "ALL_ATTACK_PATHS":
				s.HasRelationshipTypeShortcut = true
				typedNode.Kinds = append(azure.PathfindingRelationships(), ad.PathfindingRelationships()...)

			case "AZ_ATTACK_PATHS":
				s.HasRelationshipTypeShortcut = true
				typedNode.Kinds = azure.PathfindingRelationships()

			case "AD_ATTACK_PATHS":
				s.HasRelationshipTypeShortcut = true
				typedNode.Kinds = ad.PathfindingRelationships()
			}

			if s.HasRelationshipTypeShortcut {
				break
			}
		}
	}
}
