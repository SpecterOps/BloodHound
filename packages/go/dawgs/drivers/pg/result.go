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
	"github.com/jackc/pgx/v5"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type queryResult struct {
	rows       pgx.Rows
	kindMapper KindMapper
}

func (s *queryResult) Next() bool {
	return s.rows.Next()
}

func (s *queryResult) Values() (graph.ValueMapper, error) {
	if values, err := s.rows.Values(); err != nil {
		return nil, err
	} else {
		return NewValueMapper(values, s.kindMapper), nil
	}
}

func (s *queryResult) Scan(targets ...any) error {
	var translatedTargets []any

	for _, target := range targets {
		switch target.(type) {
		case *graph.Relationship:
			translatedTargets = append(translatedTargets, &edgeComposite{})

		case *graph.Node:
			translatedTargets = append(translatedTargets, &nodeComposite{})

		case *graph.Kind:
			translatedTargets = append(translatedTargets, new(int16))

		case *graph.Kinds:
			translatedTargets = append(translatedTargets, &[]int16{})

		default:
			translatedTargets = append(translatedTargets, target)
		}
	}

	if err := s.rows.Scan(translatedTargets...); err != nil {
		return err
	}

	for idx, translatedTarget := range translatedTargets {
		switch typedTarget := translatedTarget.(type) {
		case *edgeComposite:
			if err := typedTarget.ToRelationship(s.kindMapper, targets[idx].(*graph.Relationship)); err != nil {
				return err
			}

		case *nodeComposite:
			if err := typedTarget.ToNode(s.kindMapper, targets[idx].(*graph.Node)); err != nil {
				return err
			}

		case *int16:
			if kindPtr, isKindType := targets[idx].(*graph.Kind); isKindType {
				if kind, hasKind := s.kindMapper.MapKindID(*typedTarget); !hasKind {
					return fmt.Errorf("unable to map kind ID %d", *typedTarget)
				} else {
					*kindPtr = kind
				}
			}

		case *[]int16:
			if kindsPtr, isKindsType := targets[idx].(*graph.Kinds); isKindsType {
				if kinds, missingKindIDs := s.kindMapper.MapKindIDs(*typedTarget...); len(missingKindIDs) > 0 {
					return fmt.Errorf("unable to map kind IDs %+v", missingKindIDs)
				} else {
					*kindsPtr = kinds
				}
			}
		}
	}

	return nil
}

func (s *queryResult) Error() error {
	return s.rows.Err()
}

func (s *queryResult) Close() {
	s.rows.Close()
}
