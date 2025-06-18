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

package pgsql

import (
	"slices"
	"strings"

	"github.com/specterops/bloodhound/cypher/models"
)

const (
	WildcardIdentifier Identifier = "*"
	EpochIdentifier    Identifier = "epoch"
)

var reservedIdentifiers = []Identifier{
	EpochIdentifier, WildcardIdentifier,
}

func IsReservedIdentifier(identifier Identifier) bool {
	for _, reservedIdentifier := range reservedIdentifiers {
		if identifier == reservedIdentifier {
			return true
		}
	}

	return false
}

func AsOptionalIdentifier(identifier Identifier) models.Optional[Identifier] {
	return models.ValueOptional(identifier)
}

// IdentifierSet represents a set of identifiers backed by a map[Identifier]struct{} instance.
type IdentifierSet struct {
	identifiers map[Identifier]struct{}
}

func NewIdentifierSet() *IdentifierSet {
	return &IdentifierSet{
		identifiers: map[Identifier]struct{}{},
	}
}

func AllocateIdentifierSet(length int) *IdentifierSet {
	return &IdentifierSet{
		identifiers: make(map[Identifier]struct{}, length),
	}
}

func AsIdentifierSet(identifiers ...Identifier) *IdentifierSet {
	newSet := AllocateIdentifierSet(len(identifiers))

	for _, identifier := range identifiers {
		newSet.Add(identifier)
	}

	return newSet
}

func (s *IdentifierSet) Clear() {
	clear(s.identifiers)
}

func (s *IdentifierSet) Len() int {
	return len(s.identifiers)
}

func (s *IdentifierSet) IsEmpty() bool {
	return len(s.identifiers) == 0
}

func (s *IdentifierSet) Add(identifiers ...Identifier) *IdentifierSet {
	for _, identifier := range identifiers {
		if !s.Contains(identifier) {
			s.identifiers[identifier] = struct{}{}
		}
	}

	return s
}

func (s *IdentifierSet) CheckedAdd(identifier Identifier) bool {
	if s.Contains(identifier) {
		return false
	}

	s.identifiers[identifier] = struct{}{}
	return true
}

func (s *IdentifierSet) Copy() *IdentifierSet {
	copied := AllocateIdentifierSet(len(s.identifiers))
	return copied.MergeSet(s)
}

func (s *IdentifierSet) Remove(others ...Identifier) *IdentifierSet {
	for _, identifierToRemove := range others {
		delete(s.identifiers, identifierToRemove)
	}

	return s
}

func (s *IdentifierSet) RemoveSet(other *IdentifierSet) *IdentifierSet {
	for identifierToRemove := range other.identifiers {
		delete(s.identifiers, identifierToRemove)
	}

	return s
}

func (s *IdentifierSet) MergeSet(other *IdentifierSet) *IdentifierSet {
	for identifierToAdd := range other.identifiers {
		s.identifiers[identifierToAdd] = struct{}{}
	}

	return s
}

func (s *IdentifierSet) Slice() []Identifier {
	copiedIdentifiers := make([]Identifier, 0, len(s.identifiers))

	for identifier := range s.identifiers {
		copiedIdentifiers = append(copiedIdentifiers, identifier)
	}

	// Return the identifiers as a sorted slice
	slices.Sort(copiedIdentifiers)
	return copiedIdentifiers
}

func (s *IdentifierSet) Strings() []string {
	copiedIdentifiers := make([]string, 0, len(s.identifiers))

	for identifier := range s.identifiers {
		copiedIdentifiers = append(copiedIdentifiers, identifier.String())
	}

	slices.Sort(copiedIdentifiers)
	return copiedIdentifiers
}

func (s *IdentifierSet) CombinedKey() Identifier {
	// Join all identifiers in sorted order
	return Identifier(strings.Join(s.Strings(), ""))
}

// Satisfies returns true if the `other`is a subset of `s`
func (s *IdentifierSet) Satisfies(other *IdentifierSet) bool {
	for identifier := range other.identifiers {
		if _, satisfied := s.identifiers[identifier]; !satisfied {
			return false
		}
	}

	return true
}

func (s *IdentifierSet) Matches(other *IdentifierSet) bool {
	return len(s.identifiers) == len(other.identifiers) && s.Satisfies(other)
}

func (s *IdentifierSet) Contains(other Identifier) bool {
	_, contained := s.identifiers[other]
	return contained
}

// SymbolTable is a symbol table that has some generic functions for negotiating unique symbols from identifiers,
// compound identifiers, and other PgSQL AST elements.
type SymbolTable struct {
	table map[string]any
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		table: map[string]any{},
	}
}

func (s *SymbolTable) IsEmpty() bool {
	return len(s.table) == 0
}

func (s *SymbolTable) NotIn(exclusions *SymbolTable) *SymbolTable {
	notIn := NewSymbolTable()

	for symbol, value := range s.table {
		if _, in := exclusions.table[symbol]; !in {
			notIn.Add(value)
		}
	}

	return notIn
}

func (s *SymbolTable) Add(symbol any) {
	switch typedSymbol := symbol.(type) {
	case Identifier:
		s.AddIdentifier(typedSymbol)

	case CompoundIdentifier:
		s.AddCompoundIdentifier(typedSymbol)

	case *SymbolTable:
		for _, nextSymbol := range typedSymbol.table {
			switch typedInnerSymbol := nextSymbol.(type) {
			case Identifier:
				s.AddIdentifier(typedInnerSymbol)

			case CompoundIdentifier:
				s.AddCompoundIdentifier(typedInnerSymbol)
			}
		}
	}
}

func (s *SymbolTable) Contains(symbol any) bool {
	found := false

	switch typedSymbol := symbol.(type) {
	case Identifier:
		_, found = s.table[typedSymbol.String()]

	case CompoundIdentifier:
		_, found = s.table[typedSymbol.String()]

	case *SymbolTable:
		for nextSymbol := range typedSymbol.table {
			_, found = s.table[nextSymbol]

			if !found {
				return false
			}
		}
	}

	return found
}

func (s *SymbolTable) AddTable(symbols *SymbolTable) {
	for key, value := range symbols.table {
		s.table[key] = value
	}
}

func (s *SymbolTable) AddIdentifier(identifier Identifier) {
	s.table[identifier.String()] = identifier
}

func (s *SymbolTable) AddCompoundIdentifier(identifier CompoundIdentifier) {
	s.table[identifier.String()] = identifier
}

func (s *SymbolTable) RootIdentifiers() *IdentifierSet {
	identifiers := NewIdentifierSet()

	for _, identifier := range s.table {
		switch typedIdentifier := identifier.(type) {
		case Identifier:
			identifiers.Add(typedIdentifier)
		case CompoundIdentifier:
			identifiers.Add(typedIdentifier[0])
		}
	}

	return identifiers
}

func (s *SymbolTable) EachIdentifier(each func(next Identifier) bool) {
	for _, untypedIdentifier := range s.table {
		switch typedIdentifier := untypedIdentifier.(type) {
		case Identifier:
			if !each(typedIdentifier) {
				return
			}
		}
	}
}

func (s *SymbolTable) EachCompoundIdentifier(each func(next CompoundIdentifier) bool) {
	for _, untypedIdentifier := range s.table {
		switch typedIdentifier := untypedIdentifier.(type) {
		case CompoundIdentifier:
			if !each(typedIdentifier) {
				return
			}
		}
	}
}
