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

package query

import (
	"strconv"

	"github.com/specterops/bloodhound/cypher/models/walk"

	"github.com/specterops/bloodhound/cypher/models/cypher"
)

type ParameterRewriter struct {
	walk.Visitor[cypher.SyntaxNode]

	Parameters     map[string]any
	parameterIndex int
}

func NewParameterRewriter() *ParameterRewriter {
	return &ParameterRewriter{
		Visitor:        walk.NewVisitor[cypher.SyntaxNode](),
		Parameters:     map[string]any{},
		parameterIndex: 0,
	}
}

func (s *ParameterRewriter) Enter(node cypher.SyntaxNode) {
	switch typedNode := node.(type) {
	case *cypher.Parameter:
		var (
			nextParameterIndex    = s.parameterIndex
			nextParameterIndexStr = "p" + strconv.Itoa(nextParameterIndex)
		)

		// Increment the parameter index first
		s.parameterIndex++

		// Record the parameter in our map and then bind the symbol in the model
		s.Parameters[nextParameterIndexStr] = typedNode.Value
		typedNode.Symbol = nextParameterIndexStr
	}
}
