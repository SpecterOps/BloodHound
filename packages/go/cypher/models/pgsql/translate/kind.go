// Copyright 2025 Specter Ops, Inc.
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

package translate

import (
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/cypher/models/pgsql/pgd"
)

func newPGKindIDMatcher(scope *Scope, treeTranslator *ExpressionTreeTranslator, binding *BoundIdentifier, kindIDs []int16) error {
	kindIDsLiteral := pgsql.NewLiteral(kindIDs, pgsql.Int2Array)

	switch binding.DataType {
	case pgsql.NodeComposite, pgsql.ExpansionRootNode, pgsql.ExpansionTerminalNode:
		treeTranslator.PushOperand(pgd.Column(binding.Identifier, pgsql.ColumnKindIDs))
		treeTranslator.PushOperand(kindIDsLiteral)

		return treeTranslator.CompleteBinaryExpression(scope, pgsql.OperatorPGArrayOverlap)

	case pgsql.EdgeComposite, pgsql.ExpansionEdge:
		treeTranslator.PushOperand(pgsql.CompoundIdentifier{binding.Identifier, pgsql.ColumnKindID})
		treeTranslator.PushOperand(pgsql.NewAnyExpressionHinted(kindIDsLiteral))

		return treeTranslator.CompleteBinaryExpression(scope, pgsql.OperatorEquals)
	}

	return fmt.Errorf("unexpected kind matcher reference data type: %s", binding.DataType)
}

func (s *Translator) translateKindMatcher(kindMatcher *cypher.KindMatcher) error {
	if operand, err := s.treeTranslator.PopOperand(); err != nil {
		return errors.New("expected kind matcher to have one valid operand")
	} else if identifier, isIdentifier := operand.(pgsql.Identifier); !isIdentifier {
		return fmt.Errorf("expected variable for kind matcher reference but found type: %T", operand)
	} else if binding, resolved := s.scope.Lookup(identifier); !resolved {
		return fmt.Errorf("unable to find identifier %s", identifier)
	} else if kindIDs, err := s.kindMapper.MapKinds(kindMatcher.Kinds); err != nil {
		return fmt.Errorf("failed to translate kinds: %w", err)
	} else {
		return newPGKindIDMatcher(s.scope, s.treeTranslator, binding, kindIDs)
	}
}
