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

func mapValue(kindMapper KindMapper) func(rawValue, target any) (bool, error) {
	return func(rawValue, target any) (bool, error) {
		switch typedTarget := target.(type) {
		case *graph.Relationship:
			if compositeMap, typeOK := rawValue.(map[string]any); !typeOK {
				return false, fmt.Errorf("unexpected edge composite backing type: %T", rawValue)
			} else {
				edge := edgeComposite{}

				if edge.TryMap(compositeMap) {
					if err := edge.ToRelationship(kindMapper, typedTarget); err != nil {
						return false, err
					}
				} else {
					return false, nil
				}
			}

		case *graph.Node:
			if compositeMap, typeOK := rawValue.(map[string]any); !typeOK {
				return false, fmt.Errorf("unexpected node composite backing type: %T", rawValue)
			} else {
				node := nodeComposite{}

				if node.TryMap(compositeMap) {
					if err := node.ToNode(kindMapper, typedTarget); err != nil {
						return false, err
					}
				} else {
					return false, nil
				}
			}

		case *graph.Path:
			if compositeMap, typeOK := rawValue.(map[string]any); !typeOK {
				return false, fmt.Errorf("unexpected node composite backing type: %T", rawValue)
			} else {
				path := pathComposite{}

				if path.TryMap(compositeMap) {
					if err := path.ToPath(kindMapper, typedTarget); err != nil {
						return false, err
					}
				} else {
					return false, nil
				}
			}

		default:
			return false, nil
		}

		return true, nil
	}
}

func NewValueMapper(values []any, kindMapper KindMapper) graph.ValueMapper {
	return graph.NewValueMapper(values, mapValue(kindMapper))
}
