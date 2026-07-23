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
	"cmp"
	"context"
	"errors"
	"slices"
)

// Relationship is the domain representation of a graph relationship together with
// its resolved kind, endpoint node ids, and properties.
type Relationship struct {
	ID           int64
	SourceNodeID int64
	TargetNodeID int64
	Kind         Kind
	Properties   map[string]any
	KindInfos    []KindInfo
}

// ErrRelationshipNotFound indicates that no graph relationship exists for the requested id.
var ErrRelationshipNotFound = errors.New("relationship not found")

// ErrKindNotFound indicates that the relationship's kind has no entry in the
// schema_relationship_kinds table.
var ErrKindNotFound = errors.New("relationship kind not found")

// GetRelationship returns the relationship identified by the graph-assigned id with its
// kind resolved to the integer identifier from the kind table. ErrRelationshipNotFound is
// returned when the relationship does not exist. If the kind has no entry in the
// schema_relationship_kinds table, the relationship is returned with Kind.Name populated
// and Kind.ID set to nil (best-effort resolution).
func (s *Service) GetRelationship(ctx context.Context, id int64, includeKindInfo bool) (Relationship, error) {
	var (
		relationship Relationship
		kind         Kind
		err          error
	)

	if relationship, err = s.db.GetRelationship(ctx, id); err != nil {
		return Relationship{}, err
	} else if kind, err = s.db.GetKindByName(ctx, relationship.Kind.Name); errors.Is(err, ErrKindNotFound) {
		// Kind exists in the graph but not in schema_relationship_kinds; return with nil ID
		return relationship, nil
	} else if err != nil {
		return Relationship{}, err
	} else {
		relationship.Kind = kind
	}

	if includeKindInfo && relationship.Kind.ID != nil {
		kindInfos, err := s.db.GetKindInfos(ctx, relationship.Kind.Name)
		if err != nil {
			return Relationship{}, err
		}

		var allKindInfos []KindInfo
		for _, kindInfo := range kindInfos {
			if kindInfo.RelationshipKindID == nil {
				continue
			}

			allKindInfos = append(allKindInfos, kindInfo)
		}

		// Sort all kind infos by Position -> Title -> Kind ID
		slices.SortFunc(allKindInfos, func(left, right KindInfo) int {
			if result := cmp.Compare(left.Position, right.Position); result != 0 {
				return result
			} else if result := cmp.Compare(left.Title, right.Title); result != 0 {
				return result
			} else {
				return cmp.Compare(*left.RelationshipKindID, *right.RelationshipKindID)
			}
		})

		relationship.KindInfos = allKindInfos
	}

	return relationship, nil
}
