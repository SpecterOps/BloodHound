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
	EpochIdentifier,
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
