// Copyright 2024 Specter Ops, Inc.
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
	"fmt"

	"github.com/specterops/bloodhound/cypher/models"
	cypher "github.com/specterops/bloodhound/cypher/models/cypher"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
	"github.com/specterops/bloodhound/dawgs/graph"
)

const (
	expansionRootID    pgsql.Identifier = "root_id"
	expansionNextID    pgsql.Identifier = "next_id"
	expansionDepth     pgsql.Identifier = "depth"
	expansionSatisfied pgsql.Identifier = "satisfied"
	expansionIsCycle   pgsql.Identifier = "is_cycle"
	expansionPath      pgsql.Identifier = "path"
)

func expansionColumns() pgsql.RecordShape {
	return pgsql.RecordShape{
		Columns: []pgsql.Identifier{
			expansionRootID,
			expansionNextID,
			expansionDepth,
			expansionSatisfied,
			expansionIsCycle,
			expansionPath,
		},
	}
}

type Match struct {
	Pattern *Pattern
}

type NodeSelect struct {
	Frame      *Frame
	Binding    *BoundIdentifier
	Select     pgsql.Select
	Constraint *Constraint
}

type Expansion struct {
	Frame       *Frame
	PathBinding *BoundIdentifier
	MinDepth    models.Optional[int64]
	MaxDepth    models.Optional[int64]

	PrimerProjection  []pgsql.SelectItem
	PrimerConstraints pgsql.Expression

	RecursiveProjection  []pgsql.SelectItem
	RecursiveConstraints pgsql.Expression

	LeftNodeJoinCondition    pgsql.Expression
	ExpansionEdgeConstraints pgsql.Expression
	ExpansionNodeConstraints pgsql.Expression
	TerminalNodeConstraints  pgsql.Expression

	Projection  []pgsql.SelectItem
	Constraints pgsql.Expression
}

type PatternSegment struct {
	Frame                  *Frame
	Direction              graph.Direction
	Expansion              models.Optional[Expansion]
	LeftNode               *BoundIdentifier
	LeftNodeBound          bool
	LeftNodeConstraints    pgsql.Expression
	LeftNodeJoinCondition  pgsql.Expression
	Edge                   *BoundIdentifier
	EdgeConstraints        *Constraint
	EdgeJoinCondition      pgsql.Expression
	RightNode              *BoundIdentifier
	RightNodeBound         bool
	RightNodeConstraints   pgsql.Expression
	RightNodeJoinCondition pgsql.Expression
	Definitions            []*BoundIdentifier
	Projection             []pgsql.SelectItem
}

func (s *PatternSegment) FlipNodes() {
	oldLeftNode := s.LeftNode
	s.LeftNode = s.RightNode
	s.RightNode = oldLeftNode

	switch s.Direction {
	case graph.DirectionOutbound:
		s.Direction = graph.DirectionInbound
	case graph.DirectionInbound:
		s.Direction = graph.DirectionOutbound
	}
}

// TerminalNode will find the terminal node of this pattern segment based on the segment's direction
func (s *PatternSegment) TerminalNode() (*BoundIdentifier, error) {
	switch s.Direction {
	case graph.DirectionInbound:
		return s.LeftNode, nil
	case graph.DirectionOutbound:
		return s.RightNode, nil
	default:
		return nil, fmt.Errorf("unsupported direction: %v", s.Direction)
	}
}

// TerminalNode will find the root node of this pattern segment based on the segment's direction
func (s *PatternSegment) RootNode() (*BoundIdentifier, error) {
	switch s.Direction {
	case graph.DirectionInbound:
		return s.RightNode, nil
	case graph.DirectionOutbound:
		return s.LeftNode, nil
	default:
		return nil, fmt.Errorf("unsupported direction: %v", s.Direction)
	}
}

type PatternPart struct {
	IsTraversal      bool
	ShortestPath     bool
	AllShortestPaths bool
	PatternBinding   models.Optional[*BoundIdentifier]
	TraversalSteps   []*PatternSegment
	NodeSelect       NodeSelect
	Constraints      *ConstraintTracker
}

func (s *PatternPart) LastStep() *PatternSegment {
	return s.TraversalSteps[len(s.TraversalSteps)-1]
}

func (s *PatternPart) ContainsExpansions() bool {
	for _, traversalStep := range s.TraversalSteps {
		if traversalStep.Expansion.Set {
			return true
		}
	}

	return false
}

type Pattern struct {
	Parts []*PatternPart
}

func (s *Pattern) Reset() {
	s.Parts = s.Parts[:0]
}

func (s *Pattern) NewPart() *PatternPart {
	newPatternPart := &PatternPart{
		Constraints: NewConstraintTracker(),
	}

	s.Parts = append(s.Parts, newPatternPart)
	return newPatternPart
}

func (s *Pattern) CurrentPart() *PatternPart {
	return s.Parts[len(s.Parts)-1]
}

type Query struct {
	Parts []*QueryPart
	Scope *Scope
}

func (s *Query) HasParts() bool {
	return len(s.Parts) > 0
}

func (s *Query) CurrentPart() *QueryPart {
	return s.Parts[len(s.Parts)-1]
}

func (s *Query) PreparePart(numReadingClauses, numUpdatingClauses int, allocateFrame bool) error {
	newPart := &QueryPart{
		Model: &pgsql.Query{
			CommonTableExpressions: &pgsql.With{},
		},

		numReadingClauses:  numReadingClauses,
		numUpdatingClauses: numUpdatingClauses,
		mutations:          NewMutations(),
		properties:         map[string]pgsql.Expression{},
	}

	if allocateFrame {
		if frame, err := s.Scope.PushFrame(); err != nil {
			return err
		} else {
			newPart.Frame = frame
		}
	}

	s.Parts = append(s.Parts, newPart)
	return nil
}

type QueryPart struct {
	Model   *pgsql.Query
	Frame   *Frame
	Updates []*Mutations
	OrderBy []pgsql.OrderBy
	Skip    models.Optional[pgsql.Expression]
	Limit   models.Optional[pgsql.Expression]

	numReadingClauses  int
	numUpdatingClauses int

	// The fields below are meant to be used to build each component as the source AST is walked. There's some
	// repetition of some of the exported fields above which is intentional and may be a good refactor target
	// in the future
	patternPredicates []*pgsql.Future[*Pattern]
	properties        map[string]pgsql.Expression
	currentPattern    *Pattern
	stashedPattern    *Pattern
	projections       *Projections
	mutations         *Mutations
	fromClauses       []pgsql.FromClause
}

func (s *QueryPart) AddFromClause(clause pgsql.FromClause) {
	s.fromClauses = append(s.fromClauses, clause)
}

func (s *QueryPart) ConsumeFromClauses() []pgsql.FromClause {
	fromClauses := s.fromClauses
	s.fromClauses = nil

	return fromClauses
}

func (s *QueryPart) RestoreStashedPattern() {
	s.currentPattern = s.stashedPattern
	s.stashedPattern = nil
}

func (s *QueryPart) StashCurrentPattern() {
	s.stashedPattern = s.ConsumeCurrentPattern()
}

func (s *QueryPart) AddPatternPredicateFuture(predicateFuture *pgsql.Future[*Pattern]) {
	s.patternPredicates = append(s.patternPredicates, predicateFuture)
}

func (s *QueryPart) ConsumeCurrentPattern() *Pattern {
	currentPattern := s.currentPattern
	s.currentPattern = &Pattern{}

	return currentPattern
}

func (s *QueryPart) HasProjections() bool {
	return s.projections != nil && len(s.projections.Items) > 0
}

func (s *QueryPart) PrepareProjections(distinct bool) {
	s.projections = &Projections{
		Distinct: distinct,
	}
}

func (s *QueryPart) PrepareMutations() {
	if s.mutations == nil {
		s.mutations = NewMutations()
	}
}

func (s *QueryPart) HasMutations() bool {
	return s.mutations != nil && s.mutations.Updates.Len() > 0
}

func (s *QueryPart) HasDeletions() bool {
	return s.mutations != nil && s.mutations.Deletions.Len() > 0
}

func (s *QueryPart) PrepareProjection() {
	s.projections.Items = append(s.projections.Items, &Projection{})
}

func (s *QueryPart) CurrentProjection() *Projection {
	return s.projections.Current()
}

func (s *QueryPart) HasProperties() bool {
	return len(s.properties) > 0
}

func (s *QueryPart) AddProperty(key string, expression pgsql.Expression) {
	s.properties[key] = expression
}

func (s *QueryPart) ConsumeProperties() map[string]pgsql.Expression {
	properties := s.properties
	s.properties = map[string]pgsql.Expression{}

	return properties
}

func (s *QueryPart) CurrentOrderBy() *pgsql.OrderBy {
	return &s.OrderBy[len(s.OrderBy)-1]
}

type Projection struct {
	SelectItem pgsql.SelectItem
	Alias      models.Optional[pgsql.Identifier]
}

func (s *Projection) SetIdentifier(identifier pgsql.Identifier) {
	s.SelectItem = identifier
}

func (s *Projection) SetAlias(alias pgsql.Identifier) {
	s.Alias = models.ValueOptional(alias)
}

type Removal struct {
	Field string
}

type LabelAssignment struct {
	Kinds pgsql.Expression
}

type PropertyAssignment struct {
	Field           string
	Operator        pgsql.Operator
	ValueExpression pgsql.Expression
}

type Update struct {
	Frame               *Frame
	JoinConstraint      pgsql.Expression
	Projection          []pgsql.SelectItem
	TargetBinding       *BoundIdentifier
	UpdateBinding       *BoundIdentifier
	Removals            *IndexedSlice[string, Removal]
	PropertyAssignments *IndexedSlice[string, PropertyAssignment]
	KindRemovals        graph.Kinds
	KindAssignments     graph.Kinds
}

type Delete struct {
	Frame         *Frame
	TargetBinding *BoundIdentifier
	UpdateBinding *BoundIdentifier
}

type Mutations struct {
	Deletions *IndexedSlice[pgsql.Identifier, *Delete]
	Updates   *IndexedSlice[pgsql.Identifier, *Update]
}

func NewMutations() *Mutations {
	return &Mutations{
		Deletions: NewIndexedSlice[pgsql.Identifier, *Delete](),
		Updates:   NewIndexedSlice[pgsql.Identifier, *Update](),
	}
}

func (s *Mutations) AddDeletion(scope *Scope, targetIdentifier pgsql.Identifier, frame *Frame) (*Delete, error) {
	if targetBinding, bound := scope.Lookup(targetIdentifier); !bound {
		return nil, fmt.Errorf("invalid identifier: %s", targetIdentifier)
	} else if updateBinding, err := scope.DefineNew(targetBinding.DataType); err != nil {
		return nil, err
	} else {
		deletion := &Delete{
			TargetBinding: targetBinding,
			UpdateBinding: updateBinding,
			Frame:         frame,
		}

		s.Deletions.Put(targetIdentifier, deletion)
		return deletion, nil
	}
}

func (s *Mutations) newIdentifierAssignment(scope *Scope, targetBinding *BoundIdentifier) (*Update, error) {
	if updateBinding, err := scope.DefineNew(targetBinding.DataType); err != nil {
		return nil, err
	} else {
		// Create a unique scope binding for this mutation since there may be assignments that also
		// target the same identifier later in the query
		newUpdates := &Update{
			TargetBinding:       targetBinding,
			UpdateBinding:       updateBinding,
			PropertyAssignments: NewIndexedSlice[string, PropertyAssignment](),
			Removals:            NewIndexedSlice[string, Removal](),
		}

		s.Updates.Put(targetBinding.Identifier, newUpdates)
		return newUpdates, nil
	}
}

func (s *Mutations) getIdentifierMutation(scope *Scope, targetIdentifier pgsql.Identifier) (*Update, error) {
	if targetBinding, bound := scope.Lookup(targetIdentifier); !bound {
		return nil, fmt.Errorf("invalid identifier: %s", targetIdentifier)
	} else if existingAssignments, hasExisting := s.Updates.Get(targetIdentifier); hasExisting {
		return existingAssignments, nil
	} else {
		return s.newIdentifierAssignment(scope, targetBinding)
	}
}

func (s *Mutations) AddPropertyRemoval(scope *Scope, propertyLookup PropertyLookup) error {
	if mutation, err := s.getIdentifierMutation(scope, propertyLookup.Reference.Root()); err != nil {
		return err
	} else {
		mutation.Removals.Put(propertyLookup.Field, Removal{
			Field: propertyLookup.Field,
		})
	}

	return nil
}

func (s *Mutations) AddPropertyAssignment(scope *Scope, propertyLookup PropertyLookup, operator pgsql.Operator, assignmentValueExpression pgsql.Expression) error {
	if mutation, err := s.getIdentifierMutation(scope, propertyLookup.Reference.Root()); err != nil {
		return err
	} else if err := RewriteFrameBindings(scope, assignmentValueExpression); err != nil {
		return err
	} else {
		mutation.PropertyAssignments.Put(propertyLookup.Field, PropertyAssignment{
			Field:           propertyLookup.Field,
			Operator:        operator,
			ValueExpression: assignmentValueExpression,
		})
	}

	return nil
}

func (s *Mutations) AddKindAssignment(scope *Scope, targetIdentifier pgsql.Identifier, kinds graph.Kinds) error {
	if mutation, err := s.getIdentifierMutation(scope, targetIdentifier); err != nil {
		return err
	} else {
		mutation.KindAssignments = append(mutation.KindAssignments, kinds...)
	}

	return nil
}

func (s *Mutations) AddKindRemoval(scope *Scope, targetIdentifier pgsql.Identifier, kinds graph.Kinds) error {
	if mutation, err := s.getIdentifierMutation(scope, targetIdentifier); err != nil {
		return err
	} else {
		mutation.KindRemovals = append(mutation.KindRemovals, kinds...)
	}

	return nil
}

type Projections struct {
	Distinct bool
	Frame    *Frame
	Items    []*Projection
	GroupBy  []pgsql.SelectItem
}

func (s *Projections) Add(projection *Projection) {
	s.Items = append(s.Items, projection)
}

func (s *Projections) Current() *Projection {
	return s.Items[len(s.Items)-1]
}

func extractIdentifierFromCypherExpression(expression cypher.Expression) (pgsql.Identifier, bool, error) {
	if expression == nil {
		return "", false, nil
	}

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

	case *cypher.Variable:
		variableExpression = typedExpression

	default:
		return "", false, fmt.Errorf("unable to extract variable from expression type: %T", expression)
	}

	if variableExpression == nil {
		return "", false, nil
	}

	switch typedVariableExpression := variableExpression.(type) {
	case *cypher.Variable:
		return pgsql.Identifier(typedVariableExpression.Symbol), true, nil

	default:
		return "", false, fmt.Errorf("unknown variable expression type: %T", variableExpression)
	}
}
