package translate

import (
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
	if variable, isVariable := kindMatcher.Reference.(*cypher.Variable); !isVariable {
		return fmt.Errorf("expected variable for kind matcher reference but found type: %T", kindMatcher.Reference)
	} else if binding, resolved := s.scope.LookupString(variable.Symbol); !resolved {
		return fmt.Errorf("unable to find identifier %s", variable.Symbol)
	} else if kindIDs, err := s.kindMapper.MapKinds(s.ctx, kindMatcher.Kinds); err != nil {
		return fmt.Errorf("failed to translate kinds: %w", err)
	} else {
		return newPGKindIDMatcher(s.scope, s.treeTranslator, binding, kindIDs)
	}
}
