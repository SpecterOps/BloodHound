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

package test

import (
	"github.com/specterops/bloodhound/dawgs/graph"
)

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

func (s *InMemoryKindMapper) MapKinds(kinds graph.Kinds) ([]int16, graph.Kinds) {
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

func (s *InMemoryKindMapper) Put(kind graph.Kind, id int16) {
	s.KindToID[kind] = id
	s.IDToKind[id] = kind
}
