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

package pgutil

import (
	"context"
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
)

var nextKindID = int16(1)

type InMemoryKindMapper struct {
	KindToID map[graph.Kind]int16
	IDToKind map[int16]graph.Kind
}

func NewInMemoryKindMapper() *InMemoryKindMapper {
	return &InMemoryKindMapper{
		KindToID: map[graph.Kind]int16{},
		IDToKind: map[int16]graph.Kind{},
	}
}

func (s *InMemoryKindMapper) MapKindID(ctx context.Context, kindID int16) (graph.Kind, error) {
	if kind, hasKind := s.IDToKind[kindID]; hasKind {
		return kind, nil
	}

	return nil, fmt.Errorf("kind not found for id %d", kindID)
}

func (s *InMemoryKindMapper) MapKindIDs(ctx context.Context, kindIDs ...int16) (graph.Kinds, error) {
	kinds := make(graph.Kinds, len(kindIDs))

	for idx, kindID := range kindIDs {
		if kind, err := s.MapKindID(ctx, kindID); err != nil {
			return nil, err
		} else {
			kinds[idx] = kind
		}
	}

	return kinds, nil
}

func (s *InMemoryKindMapper) MapKind(ctx context.Context, kind graph.Kind) (int16, error) {
	if id, hasID := s.KindToID[kind]; hasID {
		return id, nil
	}

	return 0, fmt.Errorf("no id found for kind %s", kind)
}

func (s *InMemoryKindMapper) mapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
	var (
		ids     = make([]int16, 0, len(kinds))
		missing = make(graph.Kinds, 0, len(kinds))
	)

	for _, kind := range kinds {
		if id, found := s.KindToID[kind]; !found {
			missing = append(missing, kind)
		} else {
			ids = append(ids, id)
		}
	}

	return ids, missing
}
func (s *InMemoryKindMapper) MapKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error) {
	if ids, missing := s.mapKinds(kinds); len(missing) > 0 {
		return nil, fmt.Errorf("missing kinds: %v", missing)
	} else {
		return ids, nil
	}
}

func (s *InMemoryKindMapper) AssertKinds(ctx context.Context, kinds graph.Kinds) ([]int16, error) {
	ids, missing := s.mapKinds(kinds)

	for _, kind := range missing {
		ids = append(ids, s.Put(kind))
	}

	return ids, nil
}

func (s *InMemoryKindMapper) Put(kind graph.Kind) int16 {
	kindID := nextKindID
	nextKindID += 1

	s.KindToID[kind] = kindID
	s.IDToKind[kindID] = kind

	return kindID
}
