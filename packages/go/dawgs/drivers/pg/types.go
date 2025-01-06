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
	"context"
	"fmt"

	"github.com/specterops/bloodhound/dawgs/graph"
)

type edgeComposite struct {
	ID         int64
	StartID    int64
	EndID      int64
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
	} else {
		switch typedSrc := src.(type) {
		case int8:
			switch typedDst := any(dst).(type) {
			case *int8:
				*typedDst = typedSrc
			case *int16:
				*typedDst = int16(typedSrc)
			case *int32:
				*typedDst = int32(typedSrc)
			case *int64:
				*typedDst = int64(typedSrc)
			case *int:
				*typedDst = int(typedSrc)
			default:
				return fmt.Errorf("unable to cast and assign value type: %T", src)
			}

		case int16:
			switch typedDst := any(dst).(type) {
			case *int16:
				*typedDst = typedSrc
			case *int32:
				*typedDst = int32(typedSrc)
			case *int64:
				*typedDst = int64(typedSrc)
			case *int:
				*typedDst = int(typedSrc)
			default:
				return fmt.Errorf("unable to cast and assign value type: %T", src)
			}

		case int32:
			switch typedDst := any(dst).(type) {
			case *int32:
				*typedDst = typedSrc
			case *int64:
				*typedDst = int64(typedSrc)
			case *int:
				*typedDst = int(typedSrc)
			default:
				return fmt.Errorf("unable to cast and assign value type: %T", src)
			}

		case int64:
			switch typedDst := any(dst).(type) {
			case *int64:
				*typedDst = typedSrc
			case *int:
				*typedDst = int(typedSrc)
			default:
				return fmt.Errorf("unable to cast and assign value type: %T", src)
			}

		case int:
			switch typedDst := any(dst).(type) {
			case *int64:
				*typedDst = int64(typedSrc)
			case *int:
				*typedDst = typedSrc
			default:
				return fmt.Errorf("unable to cast and assign value type: %T", src)
			}

		case T:
			*dst = typedSrc

		default:
			return fmt.Errorf("unable to cast and assign value type: %T", src)
		}
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

func (s *edgeComposite) ToRelationship(ctx context.Context, kindMapper KindMapper, relationship *graph.Relationship) error {
	if kinds, err := kindMapper.MapKindIDs(ctx, s.KindID); err != nil {
		return err
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
	ID         int64
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

func (s *nodeComposite) ToNode(ctx context.Context, kindMapper KindMapper, node *graph.Node) error {
	if kinds, err := kindMapper.MapKindIDs(ctx, s.KindIDs...); err != nil {
		return err
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
	if rawNodes, hasNodes := compositeMap["nodes"]; hasNodes {
		if typedRawNodes, typeOK := rawNodes.([]any); !typeOK {
			return fmt.Errorf("")
		} else {
			for _, rawNode := range typedRawNodes {
				switch typedNode := rawNode.(type) {
				case map[string]any:
					var node nodeComposite

					if err := node.FromMap(typedNode); err != nil {
						return err
					}

					s.Nodes = append(s.Nodes, node)

				default:
					return fmt.Errorf("unexpected type for raw node: %T", rawNode)
				}
			}
		}
	}

	if rawEdges, hasEdges := compositeMap["edges"]; hasEdges {
		if typedRawEdges, typeOK := rawEdges.([]any); !typeOK {
			return fmt.Errorf("")
		} else {
			for _, rawEdge := range typedRawEdges {
				switch typedNode := rawEdge.(type) {
				case map[string]any:
					var edge edgeComposite

					if err := edge.FromMap(typedNode); err != nil {
						return err
					}

					s.Edges = append(s.Edges, edge)

				default:
					return fmt.Errorf("unexpected type for raw edge: %T", rawEdge)
				}
			}
		}
	}

	return nil
}

func (s *pathComposite) ToPath(ctx context.Context, kindMapper KindMapper, path *graph.Path) error {
	path.Nodes = make([]*graph.Node, len(s.Nodes))

	for idx, pgNode := range s.Nodes {
		dawgsNode := &graph.Node{}

		if err := pgNode.ToNode(ctx, kindMapper, dawgsNode); err != nil {
			return err
		}

		path.Nodes[idx] = dawgsNode
	}

	path.Edges = make([]*graph.Relationship, len(s.Edges))

	for idx, pgEdge := range s.Edges {
		dawgsRelationship := &graph.Relationship{}

		if err := pgEdge.ToRelationship(ctx, kindMapper, dawgsRelationship); err != nil {
			return err
		}

		path.Edges[idx] = dawgsRelationship
	}

	return nil
}
