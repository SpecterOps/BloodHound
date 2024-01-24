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

package graph

type IndexType int

const (
	UnsupportedIndex IndexType = 0
	BTreeIndex       IndexType = 1
	TextSearchIndex  IndexType = 2
)

func (s IndexType) String() string {
	switch s {
	case BTreeIndex:
		return "btree"

	case TextSearchIndex:
		return "fts"

	default:
		return "invalid"
	}
}

type Index struct {
	Name  string
	Field string
	Type  IndexType
}

type Constraint Index

type Graph struct {
	Name            string
	Nodes           Kinds
	Edges           Kinds
	NodeConstraints []Constraint
	EdgeConstraints []Constraint
	NodeIndexes     []Index
	EdgeIndexes     []Index
}

type Schema struct {
	Graphs       []Graph
	DefaultGraph Graph
}
