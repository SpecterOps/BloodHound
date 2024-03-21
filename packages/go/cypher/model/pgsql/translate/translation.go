package translate

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/dawgs/graph"

	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/cypher/model/walk"
)

type Match struct {
	PatternConstraints pgsql.Expression
}

type NodeSelect struct {
	Identifier pgsql.Identifier
	Bound      bool
}

type TraversalStep struct {
	Direction           graph.Direction
	LeftNodeIdentifier  pgsql.Identifier
	LeftNodeBound       bool
	EdgeIdentifier      pgsql.Identifier
	EdgeBound           bool
	RightNodeIdentifier pgsql.Identifier
	RightNodeBound      bool
}

type Pattern struct {
	IsTraversal         bool
	PatternIdentifier   pgsql.Identifier
	PatternBound        bool
	DeclaredIdentifiers map[pgsql.Identifier]struct{}
	TraversalSteps      []*TraversalStep
	NodeSelect          NodeSelect
	LastBoundIdentifier pgsql.Identifier
}

func (s *Pattern) BindNode(identifier pgsql.Identifier) error {
	if s.IsTraversal {
		if numSteps := len(s.TraversalSteps); numSteps == 0 {
			s.TraversalSteps = append(s.TraversalSteps, &TraversalStep{
				LeftNodeIdentifier: identifier,
				LeftNodeBound:      true,
			})
		} else if step := s.TraversalSteps[numSteps-1]; !step.RightNodeBound {
			step.RightNodeIdentifier = identifier
			step.RightNodeBound = true
		} else {
			return fmt.Errorf("unpacked too many nodes for node pattern")
		}
	} else if !s.NodeSelect.Bound {
		s.NodeSelect.Identifier = identifier
		s.NodeSelect.Bound = true
	} else {
		return fmt.Errorf("unpacked too many nodes for node pattern")
	}

	s.LastBoundIdentifier = identifier
	return nil
}

func (s *Pattern) BindEdge(identifier pgsql.Identifier, direction graph.Direction) error {
	if step := s.TraversalSteps[len(s.TraversalSteps)-1]; !step.EdgeBound {
		step.Direction = direction
		step.EdgeIdentifier = identifier
		step.EdgeBound = true
	} else if !step.RightNodeBound {
		return fmt.Errorf("unpacked too many nodes for node pattern")
	} else {
		s.TraversalSteps = append(s.TraversalSteps, &TraversalStep{
			Direction:          direction,
			LeftNodeIdentifier: step.RightNodeIdentifier,
			EdgeIdentifier:     identifier,
			LeftNodeBound:      true,
			EdgeBound:          true,
		})
	}

	s.LastBoundIdentifier = identifier
	return nil
}

type Query struct {
	DeclaredIdentifiers map[pgsql.Identifier]struct{}
}

type Where struct {
	constraints *ConstraintTracker
	translator  *ExpressionTreeTranslator
}

func (s *Where) Consume() pgsql.Expression {
	return s.constraints.ConsumeAll()
}

func (s *Where) SatisfiedConstraints(available pgsql.IdentifierSet) (pgsql.IdentifierSet, pgsql.Expression) {
	return s.constraints.Consume(available)
}

func (s *Where) ConstrainIdentifierSet(identifierSet pgsql.IdentifierSet, constraint pgsql.Expression) {
	s.constraints.Constrain(identifierSet, constraint)
}

func (s *Where) ConstrainIdentifier(identifier pgsql.Identifier, constraint pgsql.Expression) {
	s.ConstrainIdentifierSet(pgsql.AsIdentifierSet(identifier), constraint)
}

func NewWhere() *Where {
	var (
		constraintTracker = NewConstraintTracker()
		translator        = NewExpressionTreeTranslator(constraintTracker)
	)

	return &Where{
		constraints: constraintTracker,
		translator:  translator,
	}
}

type Projection struct {
	Identifier pgsql.Identifier
	Alias      pgsql.Identifier
	AliasBound bool
}

func (s *Projection) SetIdentifier(identifier pgsql.Identifier) {
	s.Identifier = identifier
}

func (s *Projection) SetAlias(alias pgsql.Identifier) {
	s.Alias = alias
	s.AliasBound = true
}

type Projections struct {
	Projections []*Projection
}

func (s *Projections) PushProjection() {
	s.Projections = append(s.Projections, &Projection{})
}

func (s *Projections) CurrentProjection() *Projection {
	return s.Projections[len(s.Projections)-1]
}

type State int

const (
	StateTranslatingStart State = iota
	StateTranslatingPattern
	StateTranslatingMatch
	StateTranslatingWhere
	StateTranslatingProjection
)

type Translator struct {
	walk.HierarchicalVisitor[cypher.SyntaxNode]

	translatedQuery *pgsql.Query
	state           State

	patternTranslation     *Pattern
	patternTranslations    []*Pattern
	whereTranslation       *Where
	matchTranslation       *Match
	projectionTranslations *Projections
	queryTranslation       *Query
	identifierTracker      *IdentifierTracker
	identifierGenerator    *IdentifierGenerator
}

func NewTranslator(query *pgsql.Query) *Translator {
	return &Translator{
		HierarchicalVisitor: walk.NewComposableHierarchicalVisitor[cypher.SyntaxNode](),
		translatedQuery:     query,
		identifierTracker:   NewIdentifierTracker(),
		identifierGenerator: NewIdentifierGenerator(),
	}
}

func (s *Translator) ConstraintTracker() *ConstraintTracker {
	return s.whereTranslation.constraints
}

func cypherVariableSymbol(expression cypher.Expression) (string, bool, error) {
	var variableExpression cypher.Expression

	switch typedExpression := expression.(type) {
	case *cypher.NodePattern:
		variableExpression = typedExpression.Binding

	case *cypher.RelationshipPattern:
		variableExpression = typedExpression.Binding

	case *cypher.PatternPart:
		variableExpression = typedExpression.Binding

	case *cypher.ProjectionItem:
		variableExpression = typedExpression.Binding

	default:
		return "", false, fmt.Errorf("unable to extract variable from expression type: %T", expression)
	}

	if variableExpression == nil {
		return "", false, nil
	}

	switch typedVariableExpression := variableExpression.(type) {
	case *cypher.Variable:
		return typedVariableExpression.Symbol, true, nil

	default:
		return "", false, fmt.Errorf("unknown variable expression type: %T", variableExpression)
	}
}

func OptionalAnd(optional pgsql.Expression, conjoined pgsql.Expression) pgsql.Expression {
	if optional == nil {
		return conjoined
	}

	return &pgsql.BinaryExpression{
		Operator: pgsql.OperatorAnd,
		LOperand: conjoined,
		ROperand: optional,
	}
}

func (s *Translator) Enter(expression cypher.SyntaxNode) {
	switch typedExpression := expression.(type) {
	case *cypher.MultiPartQueryPart:
	case *cypher.SinglePartQuery:
		s.queryTranslation = &Query{
			DeclaredIdentifiers: map[pgsql.Identifier]struct{}{},
		}

	case *cypher.Match:
		s.state = StateTranslatingMatch

		// Start with a fresh match and where clause. Instantiation of the where clause here is necessary since
		// cypher will store identifier constraints in the query pattern which precedes the query where clause.
		s.matchTranslation = &Match{}
		s.whereTranslation = NewWhere()

		// Clear out pattern translations
		s.patternTranslations = s.patternTranslations[:0]

	case *cypher.Where:
		s.state = StateTranslatingWhere

	case *cypher.Variable:
		switch s.state {
		case StateTranslatingWhere:
			if aliasedIdentifier, hasAlias := s.identifierTracker.LookupAlias(typedExpression.Symbol); !hasAlias {
				s.SetErrorf("unable to find aliased identifier %s", typedExpression.Symbol)
			} else {
				// TODO: Invocations like this feel like reaching
				s.whereTranslation.translator.Push(aliasedIdentifier)
			}

		case StateTranslatingProjection:
			if aliasedIdentifier, hasAlias := s.identifierTracker.LookupAlias(typedExpression.Symbol); !hasAlias {
				s.SetErrorf("unable to find aliased identifier %s", typedExpression.Symbol)
			} else {
				s.projectionTranslations.CurrentProjection().SetIdentifier(aliasedIdentifier)
			}
		}

	case *cypher.Literal:
		var literal pgsql.Literal

		if strLiteral, isStr := typedExpression.Value.(string); isStr {
			// Cypher parser wraps string literals with ' characters - unwrap them first
			literal = pgsql.AsLiteral(strLiteral[1 : len(strLiteral)-1])
			literal.Null = typedExpression.Null
		} else {
			literal = pgsql.AsLiteral(typedExpression.Value)
			literal.Null = typedExpression.Null
		}

		switch s.state {
		case StateTranslatingWhere:
			s.whereTranslation.translator.Push(literal)
		}

	case *cypher.PropertyLookup:
		switch s.state {
		case StateTranslatingWhere:
			// Property lookups are binary expressions that use tne -> and ->> operators
			s.whereTranslation.translator.EnterOperator(pgsql.OperatorJSONField)

			if propertyLookupAtom, err := cypher.ExpressionAs[*cypher.Variable](typedExpression.Atom); err != nil {
				s.SetError(err)
			} else if aliasedIdentifier, aliased := s.identifierTracker.LookupAlias(propertyLookupAtom.Symbol); !aliased {
				s.SetErrorf("unable to look up alias for identifier: %s", propertyLookupAtom.Symbol)
			} else {
				// Property lookups must reference the properties column
				s.whereTranslation.translator.Push(pgsql.CompoundIdentifier{aliasedIdentifier, pgsql.ColumnProperties})

				// TODO: Cypher does not support nested property references so the Symbols slice should be a string
				s.whereTranslation.translator.Push(pgsql.AsLiteral(typedExpression.Symbols[0]))
			}
		}

	case *cypher.Projection:
		s.state = StateTranslatingProjection
		s.projectionTranslations = &Projections{}

	case *cypher.ProjectionItem:
		s.projectionTranslations.PushProjection()

		if cypherBinding, hasBinding, err := cypherVariableSymbol(typedExpression); err != nil {
			s.SetError(err)
		} else if hasBinding {
			s.projectionTranslations.CurrentProjection().SetAlias(pgsql.Identifier(cypherBinding))
		}

	case *cypher.PatternPart:
		s.state = StateTranslatingPattern

		s.patternTranslation = &Pattern{
			// We expect this to be a node select if there aren't enough pattern elements for a traversal
			IsTraversal:         len(typedExpression.PatternElements) > 1,
			DeclaredIdentifiers: map[pgsql.Identifier]struct{}{},
		}

		if cypherBinding, hasBinding, err := cypherVariableSymbol(typedExpression); err != nil {
			s.SetError(err)
		} else if hasBinding {
			pgBinding := s.identifierGenerator.PatternBinding()

			// Declare the identifier in the scope of the pattern translation
			s.patternTranslation.DeclaredIdentifiers[pgBinding] = struct{}{}

			// Generate an alias for this binding
			s.identifierTracker.Alias(cypherBinding, pgBinding, pgsql.PathComposite)

			// Record the new binding in the traversal pattern being built
			s.patternTranslation.PatternIdentifier = pgBinding
			s.patternTranslation.PatternBound = true
		}

	case *cypher.NodePattern:
		pgBinding := s.identifierGenerator.NodeBinding()

		// Declare the identifier in the scope of the pattern translation
		s.patternTranslation.DeclaredIdentifiers[pgBinding] = struct{}{}

		// Alias or otherwise track the new binding
		if cypherBinding, hasBinding, err := cypherVariableSymbol(typedExpression); err != nil {
			s.SetError(err)
		} else {
			if hasBinding {
				s.identifierTracker.Alias(cypherBinding, pgBinding, pgsql.NodeComposite)
			} else {
				s.identifierTracker.Track(pgBinding, pgsql.NodeComposite)
			}
		}

		// Apply the binding to the translation
		if err := s.patternTranslation.BindNode(pgBinding); err != nil {
			s.SetError(err)
		}

		// If there's a bound pattern track this identifier as a dependency
		if s.patternTranslation.PatternBound {
			if err := s.identifierTracker.DependsOn(s.patternTranslation.PatternIdentifier, pgBinding); err != nil {
				s.SetError(err)
			}
		}

		// Capture any kind matchers for this node pattern
		if len(typedExpression.Kinds) > 0 {
			s.whereTranslation.ConstrainIdentifier(pgBinding, &pgsql.BinaryExpression{
				Operator: pgsql.OperatorPGArrayOverlap,
				LOperand: pgsql.CompoundIdentifier{pgBinding, pgsql.ColumnKindIDs},
				ROperand: pgsql.AsLiteral(typedExpression.Kinds),
			})
		}

	case *cypher.RelationshipPattern:
		pgBinding := s.identifierGenerator.EdgeBinding()

		// Declare the identifier in the scope of the pattern translation
		s.patternTranslation.DeclaredIdentifiers[pgBinding] = struct{}{}

		// Alias or otherwise track the new binding
		if cypherBinding, hasBinding, err := cypherVariableSymbol(typedExpression); err != nil {
			s.SetError(err)
		} else {
			if hasBinding {
				s.identifierTracker.Alias(cypherBinding, pgBinding, pgsql.EdgeComposite)
			} else {
				s.identifierTracker.Track(pgBinding, pgsql.EdgeComposite)
			}
		}

		// Apply the binding to the translation
		if err := s.patternTranslation.BindEdge(pgBinding, typedExpression.Direction); err != nil {
			s.SetError(err)
		}

		// If there's a bound pattern track this identifier as a dependency
		if s.patternTranslation.PatternBound {
			if err := s.identifierTracker.DependsOn(s.patternTranslation.PatternIdentifier, pgBinding); err != nil {
				s.SetError(err)
			}
		}

		// Capture the kind matchers for this relationship pattern
		if len(typedExpression.Kinds) > 0 {
			s.whereTranslation.ConstrainIdentifier(pgBinding, &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{pgBinding, pgsql.ColumnKindID},
				ROperand: pgsql.FunctionCall{
					Function:   pgsql.FunctionAny,
					Parameters: []pgsql.Expression{pgsql.AsLiteral(typedExpression.Kinds)},
				},
			})
		}

	case *cypher.PartialComparison:
		s.whereTranslation.translator.EnterOperator(pgsql.Operator(typedExpression.Operator))

	case *cypher.PartialArithmeticExpression:
		s.whereTranslation.translator.EnterOperator(pgsql.Operator(typedExpression.Operator))

	case *cypher.Disjunction:
		s.whereTranslation.translator.EnterOperator(pgsql.OperatorOr)

	case *cypher.Conjunction:
		s.whereTranslation.translator.EnterOperator(pgsql.OperatorAnd)
	}
}

func (s *Translator) Exit(expression cypher.SyntaxNode) {
	switch typedExpression := expression.(type) {
	case *cypher.Negation:
		switch s.state {
		case StateTranslatingWhere:
			s.whereTranslation.translator.Push(&pgsql.UnaryExpression{
				Operator: pgsql.OperatorNot,
				Operand:  s.whereTranslation.translator.Pop(),
			})
		}

	case *cypher.PatternPart:
		s.patternTranslations = append(s.patternTranslations, s.patternTranslation)
		s.patternTranslation = nil

		s.state = StateTranslatingMatch

	case *cypher.Where:
		s.state = StateTranslatingMatch

		// Constrain the last operands
		if err := s.whereTranslation.translator.ConstraintRemainingOperands(); err != nil {
			s.SetError(err)
		}

	case *cypher.PropertyLookup:
		switch s.state {
		case StateTranslatingWhere:
			if err := s.whereTranslation.translator.ExitOperator(pgsql.OperatorJSONField); err != nil {
				s.SetError(err)
			}
		}

	case *cypher.PartialComparison:
		switch s.state {
		case StateTranslatingWhere:
			if err := s.whereTranslation.translator.ExitOperator(pgsql.Operator(typedExpression.Operator)); err != nil {
				s.SetError(err)
			}
		}

	case *cypher.PartialArithmeticExpression:
		switch s.state {
		case StateTranslatingWhere:
			if err := s.whereTranslation.translator.ExitOperator(pgsql.Operator(typedExpression.Operator)); err != nil {
				s.SetError(err)
			}
		}

	case *cypher.Disjunction:
		switch s.state {
		case StateTranslatingWhere:
			for idx := 0; idx < typedExpression.Len()-1; idx++ {
				if err := s.whereTranslation.translator.ExitOperator(pgsql.OperatorOr); err != nil {
					s.SetError(err)
				}
			}
		}

	case *cypher.Conjunction:
		switch s.state {
		case StateTranslatingWhere:
			for idx := 0; idx < typedExpression.Len()-1; idx++ {
				if err := s.whereTranslation.translator.ExitOperator(pgsql.OperatorAnd); err != nil {
					s.SetError(err)
				}
			}
		}

	case *cypher.Return:
		s.translateReturn()

	case *cypher.Match:
		s.translateMatch()
	}
}

func (s *Translator) translateMatch() {
	for _, pattern := range s.patternTranslations {
		s.translatePattern(pattern, s.queryTranslation.DeclaredIdentifiers)
	}
}

func (s *Translator) translateReturn() {
	topLevelSelect := pgsql.Select{
		// Combine the remaining constraints for the top-level select
		Where: s.whereTranslation.Consume(),
	}

	for _, projection := range s.projectionTranslations.Projections {
		if trackedIdentifier, isTracked := s.identifierTracker.Lookup(projection.Identifier); !isTracked {
			s.SetErrorf("unable to find identifier %s for projection", projection.Identifier)
		} else {
			if identifierFromClauses, err := trackedIdentifier.BuildFromClauses(); err != nil {
				s.SetError(err)
			} else if identifierProjection, err := trackedIdentifier.BuildProjection(); err != nil {
				s.SetError(err)
			} else {
				topLevelSelect.From = append(topLevelSelect.From, identifierFromClauses...)
				topLevelSelect.Projection = append(topLevelSelect.Projection, identifierProjection)
			}
		}
	}

	s.translatedQuery.Body = topLevelSelect
}

func (s *Translator) translatePattern(pattern *Pattern, declaredIdentifiers map[pgsql.Identifier]struct{}) {
	if pattern.IsTraversal {
		s.translateTraversalPattern(pattern, declaredIdentifiers)
	} else {
		s.translateNodePattern(pattern.NodeSelect.Identifier, declaredIdentifiers)
	}
}

func (s *Translator) translateTraversalPattern(pattern *Pattern, declaredIdentifiers map[pgsql.Identifier]struct{}) {
	for idx, traversalStep := range pattern.TraversalSteps {
		// If this is the first traversal step bind the left pattern node
		if idx == 0 {
			s.translateNodePattern(traversalStep.LeftNodeIdentifier, declaredIdentifiers)
		}

		// Translate the edge and following right node
		s.translateEdgePattern(traversalStep, declaredIdentifiers)

		// If the pattern is bound constrain any projections of the traversal step's identifiers
		if pattern.PatternBound {
			switch traversalStep.Direction {
			case graph.DirectionOutbound:
				// TODO: Move to a projection translation type
				s.whereTranslation.ConstrainIdentifier(pattern.PatternIdentifier, pgsql.NewBinaryExpression(
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnStartID},
					),
					pgsql.OperatorAnd,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnEndID},
					),
				))

			case graph.DirectionInbound:
				s.whereTranslation.ConstrainIdentifier(pattern.PatternIdentifier, pgsql.NewBinaryExpression(
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnEndID},
					),
					pgsql.OperatorAnd,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnStartID},
					),
				))
			}
		}
	}
}

func (s *Translator) translateEdgePattern(traversalStep *TraversalStep, declaredIdentifiers pgsql.IdentifierSet) {
	// Declare the edge identifier
	declaredIdentifiers[traversalStep.EdgeIdentifier] = struct{}{}

	var (
		// Search the edge table
		fromClauses = []pgsql.FromClause{{
			Relation: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: pgsql.AsOptionalIdentifier(traversalStep.EdgeIdentifier),
			},
		}}

		// Find any matched where clause constraints
		requiredIdentifiers, whereClause = s.whereTranslation.SatisfiedConstraints(declaredIdentifiers)
	)

	// Figure out which ID of the edge to join on the left node
	switch traversalStep.Direction {
	case graph.DirectionOutbound:
		whereClause = OptionalAnd(whereClause, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier, pgsql.ColumnID},
		})

	case graph.DirectionInbound:
		whereClause = OptionalAnd(whereClause, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier, pgsql.ColumnID},
		})

	default:
		s.SetErrorf("unsupported direction: %d", traversalStep.Direction)
	}

	// Ensure that the left node is part of the required identifier set
	requiredIdentifiers.Add(traversalStep.LeftNodeIdentifier)

	// Author the required identifiers as from clauses
	for requiredIdentifier := range requiredIdentifiers {
		// Do not emit a from clause for the identifier this pattern belongs to
		if requiredIdentifier == traversalStep.EdgeIdentifier {
			continue
		}

		fromClauses = append(fromClauses, pgsql.FromClause{
			Relation: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{requiredIdentifier},
			},
		})
	}

	// Prepare the next select statement
	s.translatedQuery.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name: traversalStep.EdgeIdentifier,
		},
		Query: pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{pgsql.AsWildcardIdentifier(traversalStep.EdgeIdentifier)},
				From:       fromClauses,
				Where:      whereClause,
			},
		},
	})

	// Declare the end node identifier
	declaredIdentifiers[traversalStep.RightNodeIdentifier] = struct{}{}

	// Search the node table
	fromClauses = []pgsql.FromClause{{
		Relation: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
			Binding: pgsql.AsOptionalIdentifier(traversalStep.RightNodeIdentifier),
		},
	}}

	// Find any matched where clause constraints
	requiredIdentifiers, whereClause = s.whereTranslation.SatisfiedConstraints(declaredIdentifiers)

	// Append our join conditions
	switch traversalStep.Direction {
	case graph.DirectionOutbound:
		whereClause = OptionalAnd(whereClause, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier, pgsql.ColumnID},
		})

	case graph.DirectionInbound:
		whereClause = OptionalAnd(whereClause, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier, pgsql.ColumnID},
		})

	default:
		s.SetErrorf("unsupported direction: %d", traversalStep.Direction)
	}

	// Ensure that the edge is part of the required identifier set
	requiredIdentifiers.Add(traversalStep.EdgeIdentifier)

	// Author the required identifiers as from clauses
	for requiredIdentifier := range requiredIdentifiers {
		// Do not emit a from clause for the identifier this pattern belongs to
		if requiredIdentifier == traversalStep.RightNodeIdentifier {
			continue
		}

		fromClauses = append(fromClauses, pgsql.FromClause{
			Relation: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{requiredIdentifier},
			},
		})
	}

	// Prepare the next select statement
	s.translatedQuery.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name: traversalStep.RightNodeIdentifier,
		},
		Query: pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{pgsql.AsWildcardIdentifier(traversalStep.RightNodeIdentifier)},
				From:       fromClauses,
				Where:      whereClause,
			},
		},
	})
}

func (s *Translator) translateNodePattern(nodeIdentifier pgsql.Identifier, declaredIdentifiers map[pgsql.Identifier]struct{}) {
	// Declare the node identifier
	declaredIdentifiers[nodeIdentifier] = struct{}{}

	var (
		// Search the node table
		fromClauses = []pgsql.FromClause{{
			Relation: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: pgsql.AsOptionalIdentifier(nodeIdentifier),
			},
		}}

		// Find any matched where clause constraints
		requiredIdentifiers, whereClause = s.whereTranslation.SatisfiedConstraints(declaredIdentifiers)
	)

	// Author the required identifiers as from clauses
	for requiredIdentifier := range requiredIdentifiers {
		// Do not emit a from clause for the identifier this pattern belongs to
		if requiredIdentifier == nodeIdentifier {
			continue
		}

		fromClauses = append(fromClauses, pgsql.FromClause{
			Relation: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{requiredIdentifier},
			},
		})
	}

	// Prepare the next select statement
	s.translatedQuery.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name: nodeIdentifier,
		},
		Query: pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{pgsql.AsWildcardIdentifier(nodeIdentifier)},
				From:       fromClauses,
				Where:      whereClause,
			},
		},
	})
}

func Translate(cypherQuery *cypher.RegularQuery) (pgsql.Statement, error) {
	var (
		query = pgsql.Query{
			CommonTableExpressions: &pgsql.With{},
		}

		translator = NewTranslator(&query)
	)

	return query, walk.Cypher(cypherQuery, translator)
}
