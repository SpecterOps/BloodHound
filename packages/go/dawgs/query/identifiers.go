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
	"github.com/specterops/bloodhound/cypher/models/cypher"
)

func Variable(name string) *cypher.Variable {
	return &cypher.Variable{
		Symbol: name,
	}
}

func Identity(entity cypher.Expression) *cypher.FunctionInvocation {
	return &cypher.FunctionInvocation{
		Name:      "id",
		Arguments: []cypher.Expression{entity},
	}
}

const (
	PathSymbol      = "p"
	NodeSymbol      = "n"
	EdgeSymbol      = "r"
	EdgeStartSymbol = "s"
	EdgeEndSymbol   = "e"
)

func Node() *cypher.Variable {
	return Variable(NodeSymbol)
}

func NodeID() *cypher.FunctionInvocation {
	return Identity(Node())
}

func Relationship() *cypher.Variable {
	return Variable(EdgeSymbol)
}

func RelationshipID() *cypher.FunctionInvocation {
	return Identity(Relationship())
}

func Start() *cypher.Variable {
	return Variable(EdgeStartSymbol)
}

func StartID() *cypher.FunctionInvocation {
	return Identity(Start())
}

func End() *cypher.Variable {
	return Variable(EdgeEndSymbol)
}

func EndID() *cypher.FunctionInvocation {
	return Identity(End())
}

func Path() *cypher.Variable {
	return Variable(PathSymbol)
}
