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

	"github.com/specterops/bloodhound/cypher/models"
	"github.com/specterops/bloodhound/cypher/models/pgsql"
)

// IdentifierGenerator is a map that creates a unique identifier for each call with a given
// data type. This ensures that renamed identifiers in queries do not conflict with each other.
type IdentifierGenerator map[pgsql.DataType]int

func (s IdentifierGenerator) NewIdentifier(dataType pgsql.DataType) (pgsql.Identifier, error) {
	var prefixStr string

	switch dataType {
	case pgsql.ExpansionPattern:
		prefixStr = "ex"
	case pgsql.ExpansionPath:
		prefixStr = "ep"
	case pgsql.PathComposite:
		prefixStr = "pc"
	case pgsql.NodeComposite:
		prefixStr = "n"
	case pgsql.EdgeComposite:
		prefixStr = "e"
	case pgsql.Scope:
		prefixStr = "s"
	case pgsql.ParameterIdentifier:
		prefixStr = "pi"
	default:
		// Make this data type the unknown generic
		dataType = pgsql.UnknownDataType
		prefixStr = "i"
	}

	var (
		nextID    = s[dataType]
		nextIDStr = strconv.Itoa(nextID)
	)

	// Increment the ID
	s[dataType] = nextID + 1

	return pgsql.Identifier(prefixStr + nextIDStr), nil
}

func NewIdentifierGenerator() IdentifierGenerator {
	return IdentifierGenerator{}
}

func previousValidFrame(query *Query, partFrame *Frame) (*Frame, bool) {
	if partFrame.Previous == nil {
		return nil, false
	}

	if currentQueryPart := query.CurrentPart(); currentQueryPart.Frame != nil && partFrame.Previous.Binding.Identifier == currentQueryPart.Frame.Binding.Identifier {
		// If the part's previous frame matches the query part's frame identifier then it's possible that
		// this current part is a multipart query part. In this case there still may be a valid frame
		// to source references from
		return currentQueryPart.Frame.Previous, currentQueryPart.Frame.Previous != nil
	}

	return partFrame.Previous, true
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

func (s *Frame) Stash(identifier pgsql.Identifier) {
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

func (s *BoundIdentifier) MaterializedBy(frame *Frame) {
	s.LastProjection = frame
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

func (s *BoundIdentifier) Dematerialize() {
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
