// Copyright 2023 Specter Ops, Inc.
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

package pg

import (
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type edgeComposite struct {
	ID         int32
	StartID    int32
	EndID      int32
	KindID     int16
	Properties map[string]any
}

func castSlice[T any](raw any) ([]T, error) {
	if rawSlice, typeOK := raw.([]any); !typeOK {
		return nil, fmt.Errorf("expected raw type []any but received %T", raw)
	} else {
		sliceCopy := make([]T, len(rawSlice))

		for idx, rawValue := range rawSlice {
			if typedValue, typeOK := rawValue.(T); !typeOK {
				var empty T
				return nil, fmt.Errorf("expected type %T but received %T", empty, rawValue)
			} else {
				sliceCopy[idx] = typedValue
			}
		}

		return sliceCopy, nil
	}
}

func castMapValueAsSliceOf[T any](compositeMap map[string]any, key string) ([]T, error) {
	if src, hasKey := compositeMap[key]; !hasKey {
		return nil, fmt.Errorf("composite map does not contain expected key %s", key)
	} else {
		return castSlice[T](src)
	}
}

func castAndAssignMapValue[T any](compositeMap map[string]any, key string, dst *T) error {
	if src, hasKey := compositeMap[key]; !hasKey {
		return fmt.Errorf("composite map does not contain expected key %s", key)
	} else if typed, typeOK := src.(T); !typeOK {
		var empty T
		return fmt.Errorf("expected type %T but received %T", empty, src)
	} else {
		*dst = typed
	}

	return nil
}

func (s *edgeComposite) TryMap(compositeMap map[string]any) bool {
	return s.FromMap(compositeMap) == nil
}

func (s *edgeComposite) FromMap(compositeMap map[string]any) error {
	if err := castAndAssignMapValue(compositeMap, "id", &s.ID); err != nil {
		return err
	}

	if err := castAndAssignMapValue(compositeMap, "start_id", &s.StartID); err != nil {
		return err
	}

	if err := castAndAssignMapValue(compositeMap, "end_id", &s.EndID); err != nil {
		return err
	}

	if err := castAndAssignMapValue(compositeMap, "kind_id", &s.KindID); err != nil {
		return err
	}

	if err := castAndAssignMapValue(compositeMap, "properties", &s.Properties); err != nil {
		return err
	}

	return nil
}

func (s *edgeComposite) ToRelationship(kindMapper KindMapper, relationship *graph.Relationship) error {
	if kinds, missingIDs := kindMapper.MapKindIDs(s.KindID); len(missingIDs) > 0 {
		return fmt.Errorf("edge references the following unknown kind IDs: %v", missingIDs)
	} else {
		relationship.Kind = kinds[0]
	}

	relationship.ID = graph.ID(s.ID)
	relationship.StartID = graph.ID(s.StartID)
	relationship.EndID = graph.ID(s.EndID)
	relationship.Properties = graph.AsProperties(s.Properties)

	return nil
}

type nodeComposite struct {
	ID         int32
	KindIDs    []int16
	Properties map[string]any
}

func (s *nodeComposite) TryMap(compositeMap map[string]any) bool {
	return s.FromMap(compositeMap) == nil
}

func (s *nodeComposite) FromMap(compositeMap map[string]any) error {
	if err := castAndAssignMapValue(compositeMap, "id", &s.ID); err != nil {
		return err
	}

	if kindIDs, err := castMapValueAsSliceOf[int16](compositeMap, "kind_ids"); err != nil {
		return err
	} else {
		s.KindIDs = kindIDs
	}

	if err := castAndAssignMapValue(compositeMap, "properties", &s.Properties); err != nil {
		return err
	}

	return nil
}

func (s *nodeComposite) ToNode(kindMapper KindMapper, node *graph.Node) error {
	if kinds, missingIDs := kindMapper.MapKindIDs(s.KindIDs...); len(missingIDs) > 0 {
		return fmt.Errorf("node references the following unknown kind IDs: %v", missingIDs)
	} else {
		node.Kinds = kinds
	}

	node.ID = graph.ID(s.ID)
	node.Properties = graph.AsProperties(s.Properties)

	return nil
}

type pathComposite struct {
	Nodes []nodeComposite
	Edges []edgeComposite
}

func (s *pathComposite) TryMap(compositeMap map[string]any) bool {
	return s.FromMap(compositeMap) == nil
}

func (s *pathComposite) FromMap(compositeMap map[string]any) error {
	return nil
}

func (s *pathComposite) ToPath(kindMapper KindMapper, path *graph.Path) error {
	path.Nodes = make([]*graph.Node, len(s.Nodes))

	for idx, pgNode := range s.Nodes {
		dawgsNode := &graph.Node{}

		if err := pgNode.ToNode(kindMapper, dawgsNode); err != nil {
			return err
		}

		path.Nodes[idx] = dawgsNode
	}

	path.Edges = make([]*graph.Relationship, len(s.Edges))

	for idx, pgEdge := range s.Edges {
		dawgsRelationship := &graph.Relationship{}

		if err := pgEdge.ToRelationship(kindMapper, dawgsRelationship); err != nil {
			return err
		}

		path.Edges[idx] = dawgsRelationship
	}

	return nil
}
