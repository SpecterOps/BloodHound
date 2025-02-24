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
	"strconv"

	"github.com/specterops/bloodhound/cypher/models/walk"

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

type IdentifierGenerator map[pgsql.DataType]int

func (s IdentifierGenerator) NewIdentifier(dataType pgsql.DataType) (pgsql.Identifier, error) {
	var (
		nextID    = s[dataType]
		nextIDStr = strconv.Itoa(nextID)
	)

	// Increment the ID
	s[dataType] = nextID + 1

	switch dataType {
	case pgsql.ExpansionPattern:
		return pgsql.Identifier("ex" + nextIDStr), nil
	case pgsql.ExpansionPath:
		return pgsql.Identifier("ep" + nextIDStr), nil
	case pgsql.PathComposite:
		return pgsql.Identifier("pc" + nextIDStr), nil
	case pgsql.NodeComposite:
		return pgsql.Identifier("n" + nextIDStr), nil
	case pgsql.EdgeComposite:
		return pgsql.Identifier("e" + nextIDStr), nil
	case pgsql.Scope:
		return pgsql.Identifier("s" + nextIDStr), nil
	case pgsql.ParameterIdentifier:
		return pgsql.Identifier("pi" + nextIDStr), nil
	default:
		return pgsql.Identifier("i" + nextIDStr), nil
	}
}

func NewIdentifierGenerator() IdentifierGenerator {
	return IdentifierGenerator{}
}

type Constraint struct {
	Dependencies *pgsql.IdentifierSet
	Expression   pgsql.Expression
}

func (s *Constraint) Merge(other *Constraint) error {
	if other.Dependencies != nil && other.Expression != nil {
		newExpression := pgsql.OptionalAnd(s.Expression, other.Expression)

		switch typedNewExpression := newExpression.(type) {
		case *pgsql.UnaryExpression:
			if err := applyUnaryExpressionTypeHints(typedNewExpression); err != nil {
				return err
			}

		case *pgsql.BinaryExpression:
			if err := applyBinaryExpressionTypeHints(typedNewExpression); err != nil {
				return err
			}
		}

		s.Dependencies.MergeSet(other.Dependencies)
		s.Expression = newExpression
	}

	return nil
}

// ConstraintTracker is a tool for associating constraints (e.g. binary or unary expressions
// that constrain a set of identifiers) with the identifier set they constrain.
//
// This is useful for rewriting a where-clause so that conjoined components can be isolated:
//
// Where Clause:
//
// where a.name = 'a' and b.name = 'b' and c.name = 'c' and a.num_a > 1 and a.ef = b.ef + c.ef
//
// Isolated Constraints:
//
//	"a":           a.name = 'a' and a.num_a > 1
//	"b":           b.name = 'b'
//	"c":           c.name = 'c'
//	"a", "b", "c": a.ef = b.ef + c.ef
type ConstraintTracker struct {
	Constraints []*Constraint
}

func NewConstraintTracker() *ConstraintTracker {
	return &ConstraintTracker{}
}

func (s *ConstraintTracker) HasConstraints(scope *pgsql.IdentifierSet) (bool, error) {
	for idx := 0; idx < len(s.Constraints); idx++ {
		nextConstraint := s.Constraints[idx]

		if syntaxNodeSatisfied, err := isSyntaxNodeSatisfied(nextConstraint.Expression); err != nil {
			return false, err
		} else if syntaxNodeSatisfied && scope.Satisfies(nextConstraint.Dependencies) {
			return true, nil
		}
	}

	return false, nil
}

func (s *ConstraintTracker) ConsumeAll() (*Constraint, error) {
	var (
		constraintExpressions = make([]pgsql.Expression, len(s.Constraints))
		matchedDependencies   = pgsql.NewIdentifierSet()
	)

	for idx, constraint := range s.Constraints {
		constraintExpressions[idx] = constraint.Expression
		matchedDependencies.MergeSet(constraint.Dependencies)
	}

	// Clear the internal constraint slice
	s.Constraints = s.Constraints[:0]

	if conjoined, err := ConjoinExpressions(constraintExpressions); err != nil {
		return nil, err
	} else {
		return &Constraint{
			Dependencies: matchedDependencies,
			Expression:   conjoined,
		}, nil
	}
}

func isSyntaxNodeSatisfied(syntaxNode pgsql.SyntaxNode) (bool, error) {
	var (
		satisfied = true
		err       = walk.WalkPgSQL(syntaxNode, walk.NewSimpleVisitor[pgsql.SyntaxNode](
			func(node pgsql.SyntaxNode, errorHandler walk.CancelableErrorHandler) {
				switch typedNode := node.(type) {
				case pgsql.SyntaxNodeFuture:
					satisfied = typedNode.Satisfied()

					if !satisfied {
						errorHandler.SetDone()
					}
				}
			},
		))
	)

	return satisfied, err
}

/*
ConsumeSet takes a given scope (a set of identifiers considered in-scope) and locates all constraints that can
be satisfied by the scope's identifiers.

```

	scope := pgsql.IdentifierSet{
		"a": struct{}{},
		"b": struct{}{},
	}

	tracker := ConstraintTracker{
		Constraints: []*Constraint{{
			Dependencies: pgsql.IdentifierSet{
				"a": struct{}{},
			},
			Expression: &pgsql.BinaryExpression{
				Operator: pgsql.OperatorEquals,
				LOperand: pgsql.CompoundIdentifier{"a", "name"},
				ROperand: pgsql.Literal{
					Value: "a",
				},
			},
		}},
	}

	satisfiedScope, expression := tracker.ConsumeSet(scope)

```
*/
func (s *ConstraintTracker) ConsumeSet(scope *pgsql.IdentifierSet) (*Constraint, error) {
	var (
		matchedDependencies   = pgsql.NewIdentifierSet()
		constraintExpressions []pgsql.Expression
	)

	for idx := 0; idx < len(s.Constraints); {
		nextConstraint := s.Constraints[idx]

		// If this is a syntax node that has not been realized do not allow the constraint it represents
		// to be consumed even if the dependencies are satisfied
		if syntaxNodeSatisfied, err := isSyntaxNodeSatisfied(nextConstraint.Expression); err != nil {
			return nil, err
		} else if !syntaxNodeSatisfied || !scope.Satisfies(nextConstraint.Dependencies) {
			// This constraint isn't satisfied, move to the next one
			idx += 1
		} else {
			// Remove this constraint
			s.Constraints = append(s.Constraints[:idx], s.Constraints[idx+1:]...)

			// Append the constraint as a conjoined expression
			constraintExpressions = append(constraintExpressions, nextConstraint.Expression)

			// Track which identifiers were satisfied
			matchedDependencies.MergeSet(nextConstraint.Dependencies)
		}
	}

	if conjoined, err := ConjoinExpressions(constraintExpressions); err != nil {
		return nil, err
	} else {
		return &Constraint{
			Dependencies: matchedDependencies,
			Expression:   conjoined,
		}, nil
	}
}

func (s *ConstraintTracker) Constrain(dependencies *pgsql.IdentifierSet, constraintExpression pgsql.Expression) error {
	for _, constraint := range s.Constraints {
		if constraint.Dependencies.Matches(dependencies) {
			joinedExpression := pgsql.NewBinaryExpression(
				constraintExpression,
				pgsql.OperatorAnd,
				constraint.Expression,
			)

			if err := applyBinaryExpressionTypeHints(joinedExpression); err != nil {
				return err
			}

			constraint.Expression = joinedExpression
			return nil
		}
	}

	s.Constraints = append(s.Constraints, &Constraint{
		Dependencies: dependencies,
		Expression:   constraintExpression,
	})

	return nil
}

// Frame represents a snapshot of all identifiers defined and visible in a given scope
type Frame struct {
	id              int
	Previous        *Frame
	Binding         *BoundIdentifier
	Visible         *pgsql.IdentifierSet
	stashedVisible  *pgsql.IdentifierSet
	Exported        *pgsql.IdentifierSet
	stashedExported *pgsql.IdentifierSet
}

func (s *Frame) RestoreStashed() {
	s.Visible.MergeSet(s.stashedVisible)
	s.Exported.MergeSet(s.stashedExported)
}

func (s *Frame) Known() *pgsql.IdentifierSet {
	return s.Visible.Copy().MergeSet(s.Exported)
}

func (s *Frame) Unexport(identifer pgsql.Identifier) {
	s.Exported.Remove(identifer)
}

func (s *Frame) Export(identifier pgsql.Identifier) {
	s.Exported.Add(identifier)
}

func (s *Frame) Veil(identifier pgsql.Identifier) {
	if s.Exported.Contains(identifier) {
		s.stashedExported.Add(identifier)
		s.Exported.Remove(identifier)
	}

	if s.Visible.Contains(identifier) {
		s.stashedVisible.Add(identifier)
		s.Visible.Remove(identifier)
	}
}

func (s *Frame) Reveal(identifier pgsql.Identifier) {
	s.Visible.Add(identifier)
}

// Scope contains all identifier definitions and their temporal resolutions in a []*Frame field.
//
// Frames may be pushed onto the stack, advancing the scope of the query to the next component. Frames
// may be popped from the stack, rewinding visibility to an earlier temporal state. This is useful
// when navigating subqueries and nested expressions that require their own descendent scope lifecycle.
//
// Each frame is associated with an identifier that represents the query AST element that contains
// all visible projections. This is required when disambiguating references that otherwise belong to
// a frame.
type Scope struct {
	nextFrameID int
	stack       []*Frame
	generator   IdentifierGenerator
	aliases     map[pgsql.Identifier]pgsql.Identifier
	definitions map[pgsql.Identifier]*BoundIdentifier
}

func NewScope() *Scope {
	return &Scope{
		nextFrameID: 0,
		generator:   NewIdentifierGenerator(),
		aliases:     map[pgsql.Identifier]pgsql.Identifier{},
		definitions: map[pgsql.Identifier]*BoundIdentifier{},
	}
}

func (s *Scope) PruneDefinitions(protectedIdentifiers *pgsql.IdentifierSet) error {
	var (
		prunedAliases     = make(map[pgsql.Identifier]pgsql.Identifier, len(s.aliases))
		prunedDefinitions = make(map[pgsql.Identifier]*BoundIdentifier, len(s.definitions))
	)

	for _, protectedIdentifier := range protectedIdentifiers.Slice() {
		if definition, hasDefinition := s.definitions[protectedIdentifier]; !hasDefinition {
			return fmt.Errorf("unable to find definition for protected identifier: %s", protectedIdentifier)
		} else {
			prunedDefinitions[protectedIdentifier] = definition
		}

		for alias, identifier := range s.aliases {
			if identifier == protectedIdentifier {
				prunedAliases[alias] = protectedIdentifier
				break
			}
		}
	}

	s.definitions = prunedDefinitions
	s.aliases = prunedAliases

	// Prune scope to only what's being exported by the with statement
	currentFrame := s.CurrentFrame()

	currentFrame.Visible = protectedIdentifiers.Copy()
	currentFrame.Exported = protectedIdentifiers.Copy()

	return nil
}

func (s *Scope) Snapshot() *Scope {
	stackCopy := make([]*Frame, len(s.stack))
	copy(stackCopy, s.stack)

	aliasesCopy := make(map[pgsql.Identifier]pgsql.Identifier)
	for k, v := range s.aliases {
		aliasesCopy[k] = v
	}

	definitionsCopy := make(map[pgsql.Identifier]*BoundIdentifier)
	for k, v := range s.definitions {
		definitionsCopy[k] = v.Copy()
	}

	return &Scope{
		nextFrameID: s.nextFrameID,
		stack:       stackCopy,
		generator:   s.generator,
		aliases:     aliasesCopy,
		definitions: definitionsCopy,
	}
}

func (s *Scope) FrameAt(depth int) *Frame {
	if len(s.stack) <= depth {
		return nil
	}

	return s.stack[len(s.stack)-depth-1]
}

func (s *Scope) PreviousFrame() *Frame {
	return s.FrameAt(1)
}

func (s *Scope) CurrentFrame() *Frame {
	return s.FrameAt(0)
}

func (s *Scope) ReferenceFrame() *Frame {
	if previousFrame := s.PreviousFrame(); previousFrame != nil {
		return previousFrame
	}

	return s.CurrentFrame()
}

func (s *Scope) PopFrame() error {
	if len(s.stack) <= 0 {
		return fmt.Errorf("no frame to pop")
	}

	s.stack = s.stack[:len(s.stack)-1]
	return nil
}

func (s *Scope) UnwindToFrame(frame *Frame) error {
	found := false

	for idx := len(s.stack) - 1; idx >= 0; idx-- {
		if found = s.stack[idx].id == frame.id; found {
			s.stack = s.stack[:idx+1]
			break
		}
	}

	if !found {
		return fmt.Errorf("unable to pop frame with ID %d", frame.id)
	}

	return nil
}

func (s *Scope) PushFrame() (*Frame, error) {
	newFrame := &Frame{
		id:              s.nextFrameID,
		Visible:         pgsql.NewIdentifierSet(),
		stashedVisible:  pgsql.NewIdentifierSet(),
		Exported:        pgsql.NewIdentifierSet(),
		stashedExported: pgsql.NewIdentifierSet(),
	}

	s.nextFrameID += 1

	if nextScopeBinding, err := s.DefineNew(pgsql.Scope); err != nil {
		return nil, err
	} else {
		newFrame.Binding = nextScopeBinding
	}

	if currentFrame := s.CurrentFrame(); currentFrame != nil {
		if len(s.stack) > 0 {
			newFrame.Previous = s.stack[len(s.stack)-1]
		}

		newFrame.Visible = currentFrame.Exported.Copy()
		newFrame.Exported = currentFrame.Exported.Copy()
	} else {
		newFrame.Visible = pgsql.NewIdentifierSet()
	}

	s.stack = append(s.stack, newFrame)
	return newFrame, nil
}

func (s *Scope) CurrentFrameBinding() *BoundIdentifier {
	if currentFrame := s.CurrentFrame(); currentFrame != nil {
		return currentFrame.Binding
	}

	return nil
}

func (s *Scope) IsMaterialized(identifier pgsql.Identifier) bool {
	if binding, isBound := s.definitions[identifier]; isBound {
		return binding.LastProjection != nil
	}

	return false
}

func (s *Scope) Visible() *pgsql.IdentifierSet {
	return s.CurrentFrame().Visible.Copy()
}

func (s *Scope) Lookup(identifier pgsql.Identifier) (*BoundIdentifier, bool) {
	binding, hasBinding := s.definitions[identifier]
	return binding, hasBinding
}

func (s *Scope) LookupBindings(identifiers ...pgsql.Identifier) ([]*BoundIdentifier, error) {
	bindings := make([]*BoundIdentifier, len(identifiers))

	for idx, identifier := range identifiers {
		if binding, bound := s.definitions[identifier]; !bound {
			return nil, fmt.Errorf("missing bound identifier: %s", identifier)
		} else {
			bindings[idx] = binding
		}
	}

	return bindings, nil
}

func (s *Scope) Alias(alias pgsql.Identifier, binding *BoundIdentifier) {
	binding.Alias = models.ValueOptional(alias)
	s.aliases[alias] = binding.Identifier
}

func (s *Scope) Declare(identifier pgsql.Identifier) {
	s.CurrentFrame().Visible.Add(identifier)
}

func (s *Scope) DefineNew(dataType pgsql.DataType) (*BoundIdentifier, error) {
	if newIdentifier, err := s.generator.NewIdentifier(dataType); err != nil {
		return nil, err
	} else {
		return s.Define(newIdentifier, dataType), nil
	}
}

func (s *Scope) AliasedLookup(identifier pgsql.Identifier) (*BoundIdentifier, bool) {
	if alias, aliased := s.aliases[identifier]; aliased {
		return s.Lookup(alias)
	}

	return nil, false
}

func (s *Scope) LookupString(identifierString string) (*BoundIdentifier, bool) {
	return s.AliasedLookup(pgsql.Identifier(identifierString))
}

func (s *Scope) Define(identifier pgsql.Identifier, dataType pgsql.DataType) *BoundIdentifier {
	boundIdentifier := &BoundIdentifier{
		Identifier: identifier,
		DataType:   dataType,
	}

	s.definitions[identifier] = boundIdentifier
	return boundIdentifier
}

// BoundIdentifier is a declared query identifier bound to the current scope frame.
//
// Bound identifiers have two states:
//   - Defined - the translation code is aware of this identifier and its type
//   - Visible - the identifier has been projected into the query's scope and can be referenced
//
// Bound identifiers may also be aliased if the source query contains an alias for the identifier. In the
// openCypher query `match (n) return n as e` the projection for `n` is aliased as `e`. The translations
// will eagerly bind anonymous identifiers for traversal steps and rebind existing identifiers and their
// aliases to prevent naming collisions.
type BoundIdentifier struct {
	Identifier     pgsql.Identifier
	Alias          models.Optional[pgsql.Identifier]
	Parameter      models.Optional[*pgsql.Parameter]
	LastProjection *Frame
	Dependencies   []*BoundIdentifier
	DataType       pgsql.DataType
}

func (s *BoundIdentifier) Copy() *BoundIdentifier {
	dependenciesCopy := make([]*BoundIdentifier, len(s.Dependencies))
	copy(dependenciesCopy, s.Dependencies)

	return &BoundIdentifier{
		Identifier:     s.Identifier,
		Alias:          s.Alias,
		Parameter:      s.Parameter,
		LastProjection: s.LastProjection,
		Dependencies:   dependenciesCopy,
		DataType:       s.DataType,
	}
}

func (s *BoundIdentifier) DetachFromFrame() {
	s.LastProjection = nil
	s.Dependencies = nil
}

func (s *BoundIdentifier) Aliased() pgsql.Identifier {
	if s.Alias.Set {
		return s.Alias.Value
	}

	return s.Identifier
}

func (s *BoundIdentifier) DependOn(other *BoundIdentifier) {
	s.Dependencies = append(s.Dependencies, other)
}

func (s *BoundIdentifier) Link(other *BoundIdentifier) {
	s.DependOn(other)
	other.DependOn(s)
}

func (s *BoundIdentifier) FirstDependencyByType(dataType pgsql.DataType) (*BoundIdentifier, bool) {
	for _, dependency := range s.Dependencies {
		if dependency.DataType == dataType {
			return dependency, true
		}
	}

	return nil, false
}
