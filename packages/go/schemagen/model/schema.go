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

type StringEnum struct {
	Symbol         string
	Schema         string
	Name           string
	Representation string
}

func (s StringEnum) GetRepresentation() string {
	if s.Representation == "" {
		return s.Symbol
	}

	return s.Representation
}

func (s StringEnum) GetName() string {
	if s.Name == "" {
		return s.Symbol
	}

	return s.Name
}

type Graph struct {
	Properties        []StringEnum
	NodeKinds         []StringEnum
	RelationshipKinds []StringEnum
}

type Azure struct {
	Properties                       []StringEnum
	NodeKinds                        []StringEnum
	RelationshipKinds                []StringEnum
	AppRoleTransitRelationshipKinds  []StringEnum
	AbusableAppRoleRelationshipKinds []StringEnum
	ControlRelationshipKinds         []StringEnum
	ExecutionPrivilegeKinds          []StringEnum
	PathfindingRelationships         []StringEnum
}

type ActiveDirectory struct {
	Properties                   []StringEnum
	NodeKinds                    []StringEnum
	RelationshipKinds            []StringEnum
	ACLRelationships             []StringEnum
	PathfindingRelationships     []StringEnum
	EdgeCompositionRelationships []StringEnum
}
