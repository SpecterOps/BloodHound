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

package ein

import "github.com/specterops/bloodhound/dawgs/graph"

type IngestibleRelationship struct {
	Source     string
	SourceType graph.Kind
	TargetType graph.Kind
	Target     string
	RelProps   map[string]any
	RelType    graph.Kind
}

func (s IngestibleRelationship) IsValid() bool {
	return s.Target != "" && s.Source != ""
}

type IngestibleSession struct {
	Source    string
	Target    string
	LogonType int
}

type IngestibleNode struct {
	ObjectID    string
	PropertyMap map[string]any
	Label       graph.Kind
}

func (s IngestibleNode) IsValid() bool {
	return s.ObjectID != ""
}

type ParsedLocalGroupData struct {
	Relationships []IngestibleRelationship
	Nodes         []IngestibleNode
}

type ParsedDomainTrustData struct {
	TrustRelationships []IngestibleRelationship
	ExtraNodeProps     []IngestibleNode
}

type ParsedGroupMembershipData struct {
	RegularMembers           []IngestibleRelationship
	DistinguishedNameMembers []IngestibleRelationship
}
