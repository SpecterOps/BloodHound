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
	"github.com/specterops/bloodhound/cypher/model"
)

func Variable(name string) *model.Variable {
	return &model.Variable{
		Symbol: name,
	}
}

func Identity(entity model.Expression) *model.FunctionInvocation {
	return &model.FunctionInvocation{
		Name:      "id",
		Arguments: []model.Expression{entity},
	}
}

const (
	PathSymbol      = "p"
	NodeSymbol      = "n"
	EdgeSymbol      = "r"
	EdgeStartSymbol = "s"
	EdgeEndSymbol   = "e"
)

func Node() *model.Variable {
	return Variable(NodeSymbol)
}

func NodeID() *model.FunctionInvocation {
	return Identity(Node())
}

func Relationship() *model.Variable {
	return Variable(EdgeSymbol)
}

func RelationshipID() *model.FunctionInvocation {
	return Identity(Relationship())
}

func Start() *model.Variable {
	return Variable(EdgeStartSymbol)
}

func StartID() *model.FunctionInvocation {
	return Identity(Start())
}

func End() *model.Variable {
	return Variable(EdgeEndSymbol)
}

func EndID() *model.FunctionInvocation {
	return Identity(End())
}

func Path() *model.Variable {
	return Variable(PathSymbol)
}
