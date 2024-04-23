package translate

import (
	"fmt"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/model/cypher"
	"github.com/specterops/bloodhound/dawgs/graph"

	"github.com/specterops/bloodhound/cypher/model/pgsql"
	"github.com/specterops/bloodhound/cypher/model/walk"
)

type Match struct {
	PatternConstraints pgsql.Expression
}

type NodeSelect struct {
	Identifier model.Optional[pgsql.Identifier]
}

type Expansion struct {
	Identifier pgsql.Identifier
	MinDepth   model.Optional[int64]
	MaxDepth   model.Optional[int64]
}

type TraversalStep struct {
	Direction           graph.Direction
	Expansion           model.Optional[Expansion]
	LeftNodeIdentifier  model.Optional[pgsql.Identifier]
	EdgeIdentifier      model.Optional[pgsql.Identifier]
	RightNodeIdentifier model.Optional[pgsql.Identifier]
}

type Pattern struct {
	IsTraversal         bool
	PatternIdentifier   model.Optional[pgsql.Identifier]
	DeclaredIdentifiers map[pgsql.Identifier]struct{}
	TraversalSteps      []*TraversalStep
	NodeSelect          NodeSelect
}

func (s *Pattern) ContainsExpansions() bool {
	for _, traversalStep := range s.TraversalSteps {
		if traversalStep.Expansion.Set {
			return true
		}
	}

	return false
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

	// This is our IR
	patternTranslation     *Pattern
	patternTranslations    []*Pattern
	whereTranslation       *Where
	matchTranslation       *Match
	projectionTranslations *Projections
	queryTranslation       *Query

	// Other stuff
	identifierTracker   *IdentifierTracker
	identifierGenerator IdentifierGenerator
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

func (s *Translator) patternToNodeIR(pattern *Pattern, identifier pgsql.Identifier) error {
	identifierOptional := model.ValueOptional(identifier)

	if pattern.IsTraversal {
		if numSteps := len(pattern.TraversalSteps); numSteps == 0 {
			pattern.TraversalSteps = append(pattern.TraversalSteps, &TraversalStep{
				LeftNodeIdentifier: identifierOptional,
			})
		} else if step := pattern.TraversalSteps[numSteps-1]; !step.RightNodeIdentifier.Set {
			// This is part of a continuing pattern element chain. Inspect the previous edge pattern to see if this
			// is the terminal node of an expansion.
			if edgeIdentifier, hasEdgeIdentifier := s.identifierTracker.Lookup(step.EdgeIdentifier.Value); !hasEdgeIdentifier {
				return fmt.Errorf("unable to look up traversal edge by identifier: %s", step.LeftNodeIdentifier.Value)
			} else if edgeIdentifier.DataType == pgsql.ExpansionEdge {
				// This is part of a variable expansion so set the identifier's type accordingly.
				if err := s.identifierTracker.SetType(identifier, pgsql.ExpansionTerminalNode); err != nil {
					return err
				}
			}

			step.RightNodeIdentifier = identifierOptional
		} else {
			return fmt.Errorf("unpacked too many nodes for node pattern")
		}
	} else if !pattern.NodeSelect.Identifier.Set {
		// TODO: Lots of reaching into structs here
		pattern.NodeSelect.Identifier = identifierOptional
	} else {
		return fmt.Errorf("unpacked too many nodes for node pattern")
	}

	return nil
}

func (s *Translator) patternToEdgeIR(pattern *Pattern, identifier pgsql.Identifier, relationshipPattern *cypher.RelationshipPattern) error {
	var (
		traversalStep      = pattern.TraversalSteps[len(pattern.TraversalSteps)-1]
		identifierOptional = model.ValueOptional(identifier)
		expansionOptional  model.Optional[Expansion]
	)

	// Look for any relationship pattern ranges. These indicate some kind of variable expansion of the path pattern.
	if relationshipPattern.Range != nil {
		expansionOptional = model.ValueOptional(Expansion{
			MinDepth: model.PointerOptional(relationshipPattern.Range.StartIndex),
			MaxDepth: model.PointerOptional(relationshipPattern.Range.EndIndex),
		})

		// If there's a pattern range then we modify the data type for the left and right nodes
		// of the pattern to signify that they must be extracted from the path array of the
		// CTE expansion table alias
		if err := s.identifierTracker.SetType(traversalStep.LeftNodeIdentifier.Value, pgsql.ExpansionRootNode); err != nil {
			s.SetError(err)
		}

		if err := s.identifierTracker.SetType(identifier, pgsql.ExpansionEdge); err != nil {
			s.SetError(err)
		}
	}

	if !traversalStep.EdgeIdentifier.Set {
		// If there's no existing edge set for this traversal step, assign this one
		traversalStep.Direction = relationshipPattern.Direction
		traversalStep.EdgeIdentifier = identifierOptional
		traversalStep.Expansion = expansionOptional
	} else if !traversalStep.RightNodeIdentifier.Set {
		return fmt.Errorf("unpacked too many nodes for node pattern")
	} else {
		// If there's an existing edge then append a new traversal step that begins with the previous right
		// node identifier
		pattern.TraversalSteps = append(pattern.TraversalSteps, &TraversalStep{
			Direction:          relationshipPattern.Direction,
			LeftNodeIdentifier: traversalStep.RightNodeIdentifier,
			EdgeIdentifier:     identifierOptional,
			Expansion:          expansionOptional,
		})
	}

	return nil
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
			if identifier, err := s.identifierGenerator.NewIdentifier(pgsql.PathComposite); err != nil {
				s.SetError(err)
			} else {
				// Declare the identifier in the scope of the pattern translation
				s.patternTranslation.DeclaredIdentifiers[identifier] = struct{}{}

				// Generate an alias for this binding
				s.identifierTracker.Alias(cypherBinding, identifier, pgsql.PathComposite)

				// Record the new binding in the traversal pattern being built
				s.patternTranslation.PatternIdentifier = model.ValueOptional(identifier)
			}
		}

	case *cypher.NodePattern:
		if identifier, err := s.identifierGenerator.NewIdentifier(pgsql.NodeComposite); err != nil {
			s.SetError(err)
		} else {
			// Declare the identifier in the scope of the pattern translation
			s.patternTranslation.DeclaredIdentifiers[identifier] = struct{}{}

			// Alias or otherwise track the new binding
			if cypherBinding, hasBinding, err := cypherVariableSymbol(typedExpression); err != nil {
				s.SetError(err)
			} else {
				// Default to using the node type here. The relationship pattern, if any, contains the expansion range
				// information and handling of it must reassign this type if this node is the root or terminal node of
				// an expansion.
				if hasBinding {
					s.identifierTracker.Alias(cypherBinding, identifier, pgsql.NodeComposite)
				} else {
					s.identifierTracker.Track(identifier, pgsql.NodeComposite)
				}
			}

			// Apply the binding to the translation
			if err := s.patternToNodeIR(s.patternTranslation, identifier); err != nil {
				s.SetError(err)
			}

			// If there's a bound pattern track this identifier as a dependency
			if s.patternTranslation.PatternIdentifier.Set {
				if err := s.identifierTracker.DependsOn(s.patternTranslation.PatternIdentifier.Value, identifier); err != nil {
					s.SetError(err)
				}
			}

			// Capture any kind matchers for this node pattern
			if len(typedExpression.Kinds) > 0 {
				s.whereTranslation.ConstrainIdentifier(identifier, &pgsql.BinaryExpression{
					Operator: pgsql.OperatorPGArrayOverlap,
					LOperand: pgsql.CompoundIdentifier{identifier, pgsql.ColumnKindIDs},
					ROperand: pgsql.AsLiteral(typedExpression.Kinds),
				})
			}
		}

	case *cypher.RelationshipPattern:
		if identifier, err := s.identifierGenerator.NewIdentifier(pgsql.EdgeComposite); err != nil {
			s.SetError(err)
		} else {
			// Declare the identifier in the scope of the pattern translation
			s.patternTranslation.DeclaredIdentifiers[identifier] = struct{}{}

			// Alias or otherwise track the new binding
			if cypherBinding, hasBinding, err := cypherVariableSymbol(typedExpression); err != nil {
				s.SetError(err)
			} else {
				dataType := pgsql.EdgeComposite

				if typedExpression.Range != nil {
					dataType = pgsql.ExpansionEdge
				}

				if hasBinding {
					s.identifierTracker.Alias(cypherBinding, identifier, dataType)
				} else {
					s.identifierTracker.Track(identifier, dataType)
				}
			}

			// Apply the binding to the translation
			if err := s.patternToEdgeIR(s.patternTranslation, identifier, typedExpression); err != nil {
				s.SetError(err)
			}

			// If there's a bound pattern track this identifier as a dependency
			if s.patternTranslation.PatternIdentifier.Set {
				if err := s.identifierTracker.DependsOn(s.patternTranslation.PatternIdentifier.Value, identifier); err != nil {
					s.SetError(err)
				}
			}

			// Capture the kind matchers for this relationship pattern
			if len(typedExpression.Kinds) > 0 {
				s.whereTranslation.ConstrainIdentifier(identifier, &pgsql.BinaryExpression{
					Operator: pgsql.OperatorEquals,
					LOperand: pgsql.CompoundIdentifier{identifier, pgsql.ColumnKindID},
					ROperand: pgsql.NewAnyExpression(pgsql.AsLiteral(typedExpression.Kinds)),
				})
			}
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
		s.translateNodePattern(pattern.NodeSelect.Identifier.Value, declaredIdentifiers)
	}
}

func (s *Translator) translateExpansionPattern(pattern *Pattern, declaredIdentifiers map[pgsql.Identifier]struct{}) {
	// All pattern element chains start with a node pattern
	s.translateNodePattern(pattern.TraversalSteps[0].LeftNodeIdentifier.Value, declaredIdentifiers)

	for _, traversalStep := range pattern.TraversalSteps {
		// Translate the edge and following right node
		s.translateEdgePattern(traversalStep, declaredIdentifiers)
	}
}

func (s *Translator) translateTraversalPattern(pattern *Pattern, declaredIdentifiers map[pgsql.Identifier]struct{}) {
	if len(pattern.TraversalSteps) == 0 {
		s.SetError(fmt.Errorf("pattern contained no traversal steps"))
		return
	}

	// TODO: Until we have LOS on what a combined translateTraversalPattern function looks like
	// 		 differentiate between translation of traversals that contain expansions
	if pattern.ContainsExpansions() {
		s.translateExpansionPattern(pattern, declaredIdentifiers)
		return
	}

	for idx, traversalStep := range pattern.TraversalSteps {
		// If this is the first traversal step bind the left pattern node
		if idx == 0 {
			s.translateNodePattern(traversalStep.LeftNodeIdentifier.Value, declaredIdentifiers)
		}

		// Translate the edge and following right node
		s.translateEdgePattern(traversalStep, declaredIdentifiers)

		// If the pattern is bound constrain any projections of the traversal step's identifiers
		if pattern.PatternIdentifier.Set {
			switch traversalStep.Direction {
			case graph.DirectionOutbound:
				// TODO: Move to a projection translation type
				s.whereTranslation.ConstrainIdentifier(pattern.PatternIdentifier.Value, pgsql.NewBinaryExpression(
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier.Value, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnStartID},
					),
					pgsql.OperatorAnd,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier.Value, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnEndID},
					),
				))

			case graph.DirectionInbound:
				s.whereTranslation.ConstrainIdentifier(pattern.PatternIdentifier.Value, pgsql.NewBinaryExpression(
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier.Value, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnEndID},
					),
					pgsql.OperatorAnd,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier.Value, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnStartID},
					),
				))
			}
		}
	}
}

const (
	expansionRootID     pgsql.Identifier = "root_id"
	nextExpansionNodeID pgsql.Identifier = "next_id"
	expansionDepth      pgsql.Identifier = "depth"
	expansionSatisfied  pgsql.Identifier = "satisfied"
	expansionIsCycle    pgsql.Identifier = "is_cycle"
	expansionPath       pgsql.Identifier = "path"
)

var (
	expansionColumns = pgsql.RowShape{
		Columns: []pgsql.Identifier{
			expansionRootID,
			nextExpansionNodeID,
			expansionDepth,
			expansionSatisfied,
			expansionIsCycle,
			expansionPath,
		},
	}
)

func (s *Translator) translateEdgeExpansion(traversalStep *TraversalStep, declaredIdentifiers pgsql.IdentifierSet) {
	// Declare the edge identifier
	declaredIdentifiers[traversalStep.EdgeIdentifier.Value] = struct{}{}

	var (
		wrapperQuery = pgsql.Query{
			CommonTableExpressions: &pgsql.With{
				Recursive: true,
			},
			Body: pgsql.Select{
				Distinct: false,
				Projection: []pgsql.Projection{
					pgsql.AsWildcardIdentifier(traversalStep.EdgeIdentifier.Value),
				},
				From: []pgsql.FromClause{{
					Relation: pgsql.TableReference{
						Name:    pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value},
						Binding: model.EmptyOptional[pgsql.Identifier](),
					},
				}},
			},
		}

		expansionEdgeIdentifier, _ = s.identifierGenerator.NewIdentifier(pgsql.EdgeComposite)
		rctePrimerFromClauses      = []pgsql.FromClause{{
			Relation: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: pgsql.AsOptionalIdentifier(expansionEdgeIdentifier),
			},
		}}

		rcteUnion = pgsql.SetOperation{
			Operator: pgsql.OperatorUnion,
			All:      true,
		}

		// Find any matched where clause constraints
		requiredIdentifiers, edgeConstraints = s.whereTranslation.SatisfiedConstraints(declaredIdentifiers)
	)

	// Figure out which ID of the edge to join on the left node
	switch traversalStep.Direction {
	case graph.DirectionOutbound:
		edgeConstraints = OptionalAnd(edgeConstraints, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier.Value, pgsql.ColumnID},
		})

	case graph.DirectionInbound:
		edgeConstraints = OptionalAnd(edgeConstraints, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier.Value, pgsql.ColumnID},
		})

	default:
		s.SetErrorf("unsupported direction: %d", traversalStep.Direction)
	}

	// Rewrite identifiers in the where clause to close over the interstitial edges
	if err := RewriteExpressionIdentifiers(edgeConstraints, traversalStep.EdgeIdentifier.Value, expansionEdgeIdentifier); err != nil {
		s.SetError(err)
	}

	// Ensure that the left node is part of the required identifier set
	requiredIdentifiers.Add(traversalStep.LeftNodeIdentifier.Value)

	// Author the required identifiers as from clauses
	for requiredIdentifier := range requiredIdentifiers {
		// Do not emit a from clause for the identifier this pattern belongs to
		if requiredIdentifier == traversalStep.EdgeIdentifier.Value {
			continue
		}

		rctePrimerFromClauses = append(rctePrimerFromClauses, pgsql.FromClause{
			Relation: pgsql.TableReference{
				Name: pgsql.CompoundIdentifier{requiredIdentifier},
			},
		})
	}

	// Primer union
	rcteUnion.LOperand = pgsql.Select{
		Projection: []pgsql.Projection{
			pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnStartID},
			pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnEndID},
			pgsql.AsLiteral(1),
			pgsql.AsLiteral(false),
			pgsql.NewBinaryExpression(
				pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnStartID},
				pgsql.OperatorEquals,
				pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnEndID},
			),
			pgsql.ArrayLiteral{
				Values: []pgsql.Expression{
					pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnID},
				},
			},
		},
		From:  rctePrimerFromClauses,
		Where: edgeConstraints,
	}

	edgeConstraints = pgsql.NewBinaryExpression(
		pgsql.UnaryExpression{
			Operator: pgsql.OperatorNot,
			Operand:  pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionIsCycle},
		},
		pgsql.OperatorAnd,
		pgsql.UnaryExpression{
			Operator: pgsql.OperatorNot,
			Operand:  pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionSatisfied},
		})

	// Continuation
	rctePrimerFromClauses = []pgsql.FromClause{{
		// Bind the ongoing expansion first
		Relation: pgsql.TableReference{
			Name: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value},
		},
	}, {
		// Bind the edge table for the next expansion lookup
		Relation: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
			Binding: pgsql.AsOptionalIdentifier(expansionEdgeIdentifier),
		},

		// Join on the node blindly for now
		Joins: []pgsql.Join{{
			Table: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
				Binding: pgsql.AsOptionalIdentifier(traversalStep.RightNodeIdentifier.Value),
			},
			JoinOperator: pgsql.JoinOperator{
				JoinType: pgsql.JoinTypeInner,
				Constraint: pgsql.NewBinaryExpression(
					pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier.Value, pgsql.ColumnID},
					pgsql.OperatorEquals,
					pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnEndID},
				),
			},
		}},
	}}

	// Find any matched constraints for the right node identifier as these are written in the projection of the
	// recursive union
	_, terminalNodeConstraints := s.whereTranslation.SatisfiedConstraints(pgsql.AsIdentifierSet(traversalStep.RightNodeIdentifier.Value))

	if rOperandTerminalCriteria, err := pgsql.ExpressionAs[pgsql.Projection](terminalNodeConstraints); err == nil {
		if rOperandTerminalCriteria != nil {
			rcteUnion.ROperand = pgsql.Select{
				Projection: []pgsql.Projection{
					pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionRootID},
					pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnEndID},
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionDepth},
						pgsql.OperatorAdd,
						pgsql.AsLiteral(1),
					),
					rOperandTerminalCriteria,
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionPath}),
					),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionPath},
						pgsql.OperatorConcatenate,
						pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnID},
					),
				},
				From:  rctePrimerFromClauses,
				Where: edgeConstraints,
			}
		} else {
			rcteUnion.ROperand = pgsql.Select{
				Projection: []pgsql.Projection{
					pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionRootID},
					pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnEndID},
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionDepth},
						pgsql.OperatorAdd,
						pgsql.AsLiteral(1),
					),
					pgsql.AsLiteral(false),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnID},
						pgsql.OperatorEquals,
						pgsql.NewAnyExpression(pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionPath}),
					),
					pgsql.NewBinaryExpression(
						pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, expansionPath},
						pgsql.OperatorConcatenate,
						pgsql.CompoundIdentifier{expansionEdgeIdentifier, pgsql.ColumnID},
					),
				},
				From:  rctePrimerFromClauses,
				Where: edgeConstraints,
			}
		}
	} else if err != nil {
		s.SetError(err)
	} else {

	}

	//
	wrapperQuery.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name:  traversalStep.EdgeIdentifier.Value,
			Shape: model.ValueOptional(expansionColumns),
		},
		Query: pgsql.Query{
			Body: rcteUnion,
		},
	})

	s.translatedQuery.AddCTE(pgsql.CommonTableExpression{
		Alias: pgsql.TableAlias{
			Name: traversalStep.EdgeIdentifier.Value,
		},
		Query: wrapperQuery,
	})
}

func (s *Translator) translateEdgePattern(traversalStep *TraversalStep, declaredIdentifiers pgsql.IdentifierSet) {
	// Declare the edge identifier
	declaredIdentifiers[traversalStep.EdgeIdentifier.Value] = struct{}{}

	if traversalStep.Expansion.Set {
		s.translateEdgeExpansion(traversalStep, declaredIdentifiers)
		return
	}

	var (
		// Search the edge table
		fromClauses = []pgsql.FromClause{{
			Relation: pgsql.TableReference{
				Name:    pgsql.CompoundIdentifier{pgsql.TableEdge},
				Binding: pgsql.AsOptionalIdentifier(traversalStep.EdgeIdentifier.Value),
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
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier.Value, pgsql.ColumnID},
		})

	case graph.DirectionInbound:
		whereClause = OptionalAnd(whereClause, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.LeftNodeIdentifier.Value, pgsql.ColumnID},
		})

	default:
		s.SetErrorf("unsupported direction: %d", traversalStep.Direction)
	}

	// Ensure that the left node is part of the required identifier set
	requiredIdentifiers.Add(traversalStep.LeftNodeIdentifier.Value)

	// Author the required identifiers as from clauses
	for requiredIdentifier := range requiredIdentifiers {
		// Do not emit a from clause for the identifier this pattern belongs to
		if requiredIdentifier == traversalStep.EdgeIdentifier.Value {
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
			Name: traversalStep.EdgeIdentifier.Value,
		},
		Query: pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{pgsql.AsWildcardIdentifier(traversalStep.EdgeIdentifier.Value)},
				From:       fromClauses,
				Where:      whereClause,
			},
		},
	})

	// Declare the end node identifier
	declaredIdentifiers[traversalStep.RightNodeIdentifier.Value] = struct{}{}

	// Search the node table
	fromClauses = []pgsql.FromClause{{
		Relation: pgsql.TableReference{
			Name:    pgsql.CompoundIdentifier{pgsql.TableNode},
			Binding: pgsql.AsOptionalIdentifier(traversalStep.RightNodeIdentifier.Value),
		},
	}}

	// Find any matched where clause constraints
	requiredIdentifiers, whereClause = s.whereTranslation.SatisfiedConstraints(declaredIdentifiers)

	// Append our join conditions
	switch traversalStep.Direction {
	case graph.DirectionOutbound:
		whereClause = OptionalAnd(whereClause, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnEndID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier.Value, pgsql.ColumnID},
		})

	case graph.DirectionInbound:
		whereClause = OptionalAnd(whereClause, &pgsql.BinaryExpression{
			Operator: pgsql.OperatorEquals,
			ROperand: pgsql.CompoundIdentifier{traversalStep.EdgeIdentifier.Value, pgsql.ColumnStartID},
			LOperand: pgsql.CompoundIdentifier{traversalStep.RightNodeIdentifier.Value, pgsql.ColumnID},
		})

	default:
		s.SetErrorf("unsupported direction: %d", traversalStep.Direction)
	}

	// Ensure that the edge is part of the required identifier set
	requiredIdentifiers.Add(traversalStep.EdgeIdentifier.Value)

	// Author the required identifiers as from clauses
	for requiredIdentifier := range requiredIdentifiers {
		// Do not emit a from clause for the identifier this pattern belongs to
		if requiredIdentifier == traversalStep.RightNodeIdentifier.Value {
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
			Name: traversalStep.RightNodeIdentifier.Value,
		},
		Query: pgsql.Query{
			Body: pgsql.Select{
				Projection: []pgsql.Projection{pgsql.AsWildcardIdentifier(traversalStep.RightNodeIdentifier.Value)},
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

	if err := walk.Cypher(cypherQuery, translator); err != nil {
		return query, err
	}

	return query, nil
}
