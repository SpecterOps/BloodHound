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

package model

import (
	"github.com/specterops/bloodhound/dawgs/graph"
)

type IndexChangeSet struct {
	NodeIndexesToRemove     []string
	EdgeIndexesToRemove     []string
	NodeConstraintsToRemove []string
	EdgeConstraintsToRemove []string
	NodeIndexesToAdd        map[string]graph.Index
	EdgeIndexesToAdd        map[string]graph.Index
	NodeConstraintsToAdd    map[string]graph.Constraint
	EdgeConstraintsToAdd    map[string]graph.Constraint
}

func NewIndexChangeSet() IndexChangeSet {
	return IndexChangeSet{
		NodeIndexesToAdd:     map[string]graph.Index{},
		NodeConstraintsToAdd: map[string]graph.Constraint{},
		EdgeIndexesToAdd:     map[string]graph.Index{},
		EdgeConstraintsToAdd: map[string]graph.Constraint{},
	}
}

type GraphPartition struct {
	Name        string
	Indexes     map[string]graph.Index
	Constraints map[string]graph.Constraint
}

func NewGraphPartition(name string) GraphPartition {
	return GraphPartition{
		Name:        name,
		Indexes:     map[string]graph.Index{},
		Constraints: map[string]graph.Constraint{},
	}
}

func NewGraphPartitionFromSchema(name string, indexes []graph.Index, constraints []graph.Constraint) GraphPartition {
	graphPartition := GraphPartition{
		Name:        name,
		Indexes:     make(map[string]graph.Index, len(indexes)),
		Constraints: make(map[string]graph.Constraint, len(constraints)),
	}

	for _, index := range indexes {
		graphPartition.Indexes[IndexName(name, index)] = index
	}

	for _, constraint := range constraints {
		graphPartition.Constraints[ConstraintName(name, constraint)] = constraint
	}

	return graphPartition
}

type GraphPartitions struct {
	Node GraphPartition
	Edge GraphPartition
}

type Graph struct {
	ID         int32
	Name       string
	Partitions GraphPartitions
}
