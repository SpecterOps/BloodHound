// Copyright 2026 Specter Ops, Inc.
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

package services

import (
	"context"
	"errors"
)

// Kind is the domain representation of a relationship kind, pairing the kind name
// recorded on the graph relationship with the integer identifier assigned to it in
// the schema_relationship_kinds table.
type Kind struct {
	ID   int32
	Name string
}

// Relationship is the domain representation of a graph relationship together with
// its resolved kind, endpoint node ids, and properties.
type Relationship struct {
	ID           int64
	SourceNodeID int64
	TargetNodeID int64
	Kind         Kind
	Properties   map[string]any
}

// ErrRelationshipNotFound indicates that no graph relationship exists for the requested id.
var ErrRelationshipNotFound = errors.New("relationship not found")

// ErrKindNotFound indicates that the relationship's kind has no entry in the
// schema_relationship_kinds table.
var ErrKindNotFound = errors.New("relationship kind not found")

// GetRelationship returns the relationship identified by the graph-assigned id with its
// kind resolved to the integer identifier from the kind table. ErrRelationshipNotFound is
// returned when the relationship does not exist; ErrKindNotFound when the relationship's
// kind has no entry in the kind table.
func (s *Service) GetRelationship(ctx context.Context, id int64) (Relationship, error) {
	var (
		relationship Relationship
		kind         Kind
		err          error
	)

	relationship, err = s.db.GetRelationship(ctx, id)
	if err != nil {
		return Relationship{}, err
	}

	kind, err = s.db.GetKindByName(ctx, relationship.Kind.Name)
	if err != nil {
		return Relationship{}, err
	}

	relationship.Kind = kind
	return relationship, nil
}
