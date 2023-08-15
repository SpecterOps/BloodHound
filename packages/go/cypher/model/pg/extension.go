package pg

import "github.com/specterops/bloodhound/cypher/model"

func Copy[T any](value T) T {
	return model.Copy(value, func(value T) (T, bool) {
		var valueCopy T

		switch typedValue := any(value).(type) {
		case *AnnotatedVariable:
			valueCopy = any(typedValue.copy()).(T)

		case *AnnotatedKindMatcher:
			valueCopy = any(typedValue.copy()).(T)

		default:
			return valueCopy, false
		}

		return valueCopy, true
	})
}

func CollectPGSQLTypes(nextCursor *model.WalkCursor, expression model.Expression) bool {
	switch typedExpression := expression.(type) {
	case *PropertiesReference:
		model.Collect(nextCursor, typedExpression.Reference)

	case *AnnotatedPropertyLookup:
		model.CollectExpression(nextCursor, typedExpression.Atom)

	case *AnnotatedKindMatcher:
		model.CollectExpression(nextCursor, typedExpression.Reference)

	case *Entity:
		model.Collect(nextCursor, typedExpression.Binding)

	case *Subquery:
		model.CollectSlice(nextCursor, typedExpression.PatternElements)
		model.CollectExpression(nextCursor, typedExpression.Filter)

	case *PropertyMutation:
		model.Collect(nextCursor, typedExpression.Reference)
		model.Collect(nextCursor, typedExpression.Removals)
		model.Collect(nextCursor, typedExpression.Additions)

	case *Delete:
		model.Collect(nextCursor, typedExpression.Binding)

	case *KindMutation:
		model.Collect(nextCursor, typedExpression.Variable)
		model.Collect(nextCursor, typedExpression.Removals)
		model.Collect(nextCursor, typedExpression.Additions)

	case *NodeKindsReference:
		model.CollectExpression(nextCursor, typedExpression.Variable)

	case *EdgeKindReference:
		model.CollectExpression(nextCursor, typedExpression.Variable)

	case *AnnotatedLiteral, *AnnotatedVariable, *AnnotatedParameter:
		// Valid types but no descent

	default:
		return false
	}

	return true
}
