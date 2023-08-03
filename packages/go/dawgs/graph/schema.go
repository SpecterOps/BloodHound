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

import (
	"fmt"
	"strings"
)

type IndexType int

const (
	UnsupportedIndex    IndexType = 0
	BTreeIndex          IndexType = 1
	FullTextSearchIndex IndexType = 2
)

func (s IndexType) String() string {
	switch s {
	case BTreeIndex:
		return "BTreeIndex"
	case FullTextSearchIndex:
		return "FullTextSearchIndex"
	case UnsupportedIndex:
		fallthrough
	default:
		return "UnsupportedIndex"
	}
}

type ConstraintSchema struct {
	Name      string
	IndexType IndexType
}

func (s ConstraintSchema) Equals(other ConstraintSchema) bool {
	return s.Name == other.Name && s.IndexType == other.IndexType
}

type IndexSchema struct {
	Name      string
	IndexType IndexType
}

func (s IndexSchema) Equals(other IndexSchema) bool {
	return s.Name == other.Name && s.IndexType == other.IndexType
}

type KindSchema struct {
	Kind                Kind
	PropertyIndices     map[string]IndexSchema
	PropertyConstraints map[string]ConstraintSchema
}

func (s *KindSchema) Name() string {
	return s.Kind.String()
}

func (s *KindSchema) Constraint(property, name string, indexType IndexType) {
	s.PropertyConstraints[property] = ConstraintSchema{
		Name:      name,
		IndexType: indexType,
	}
}

func (s *KindSchema) ConstrainProperty(property string, indexType IndexType) {
	s.Constraint(property, fmt.Sprintf("%s_%s_constraint", strings.ToLower(s.Name()), strings.ToLower(property)), indexType)
}

func (s *KindSchema) Index(property, name string, indexType IndexType) {
	s.PropertyIndices[property] = IndexSchema{
		Name:      name,
		IndexType: indexType,
	}
}

func (s *KindSchema) IndexProperty(property string, indexType IndexType) {
	s.Index(property, fmt.Sprintf("%s_%s_index", strings.ToLower(s.Name()), strings.ToLower(property)), indexType)
}

type KindSchemaContinuation struct {
	kinds []*KindSchema
}

func (s KindSchemaContinuation) Constrain(name string, indexType IndexType) KindSchemaContinuation {
	for _, label := range s.kinds {
		label.ConstrainProperty(name, indexType)
	}

	return s
}

func (s KindSchemaContinuation) Index(name string, indexType IndexType) KindSchemaContinuation {
	for _, label := range s.kinds {
		label.IndexProperty(name, indexType)
	}

	return s
}

type Schema struct {
	Kinds map[Kind]*KindSchema
}

func NewSchema() *Schema {
	return &Schema{
		Kinds: make(map[Kind]*KindSchema),
	}
}

func (s *Schema) ForKinds(kinds ...Kind) KindSchemaContinuation {
	var selectedKinds []*KindSchema

	for _, kind := range kinds {
		if kind, found := s.Kinds[kind]; found {
			selectedKinds = append(selectedKinds, kind)
		}
	}

	return KindSchemaContinuation{
		kinds: selectedKinds,
	}
}

func (s *Schema) Kind(kind Kind) *KindSchema {
	return s.Kinds[kind]
}

func (s *Schema) EnsureKind(kind Kind) *KindSchema {
	if label, found := s.Kinds[kind]; found {
		return label
	} else {
		newLabel := &KindSchema{
			Kind:                kind,
			PropertyIndices:     make(map[string]IndexSchema),
			PropertyConstraints: make(map[string]ConstraintSchema),
		}

		s.Kinds[kind] = newLabel
		return newLabel
	}
}

func (s *Schema) DefineKinds(kinds ...Kind) {
	for _, kind := range kinds {
		s.Kinds[kind] = &KindSchema{
			Kind:                kind,
			PropertyIndices:     make(map[string]IndexSchema),
			PropertyConstraints: make(map[string]ConstraintSchema),
		}
	}
}

func (s *Schema) ConstrainProperty(name string, indexType IndexType) {
	for _, kindSchema := range s.Kinds {
		kindSchema.PropertyConstraints[name] = ConstraintSchema{
			Name:      fmt.Sprintf("%s_%s_constraint", strings.ToLower(kindSchema.Name()), strings.ToLower(name)),
			IndexType: indexType,
		}
	}
}

func (s *Schema) IndexProperty(name string, indexType IndexType) {
	for _, labelSchema := range s.Kinds {
		labelSchema.PropertyIndices[name] = IndexSchema{
			Name:      fmt.Sprintf("%s_%s_index", strings.ToLower(labelSchema.Name()), strings.ToLower(name)),
			IndexType: indexType,
		}
	}
}

func (s *Schema) String() string {
	output := strings.Builder{}

	for _, kindSchema := range s.Kinds {
		output.WriteString("Label: ")
		output.WriteString(kindSchema.Name())
		output.WriteRune('\n')

		for propertyName, constraint := range kindSchema.PropertyConstraints {
			output.WriteString("\t")
			output.WriteString(propertyName)
			output.WriteString(" ")
			output.WriteString(constraint.Name)
			output.WriteString("[")
			output.WriteString(constraint.IndexType.String())
			output.WriteString("]\n")
		}

		for propertyName, index := range kindSchema.PropertyIndices {
			output.WriteString("\t")
			output.WriteString(propertyName)
			output.WriteString(" ")
			output.WriteString(index.Name)
			output.WriteString("[")
			output.WriteString(index.IndexType.String())
			output.WriteString("]\n")
		}

		output.WriteRune('\n')
	}

	return output.String()
}
