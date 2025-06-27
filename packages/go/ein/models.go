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

import "github.com/specterops/dawgs/graph"

// Initialize IngestibleRelationship to ensure the RelProps map can't be nil
func NewIngestibleRelationship(source IngestibleEndpoint, target IngestibleEndpoint, rel IngestibleRel) IngestibleRelationship {
	if rel.RelProps == nil {
		rel.RelProps = make(map[string]any)
	}

	return IngestibleRelationship{
		Source:   source,
		Target:   target,
		RelProps: rel.RelProps,
		RelType:  rel.RelType,
	}
}

// IngestMatchStrategy defines how a node should be matched during ingestion—
// either by its object ID (default) or by its name.
type IngestMatchStrategy string

const (
	MatchByID   IngestMatchStrategy = "id"
	MatchByName IngestMatchStrategy = "name"
)

// IngestibleEndpoint represents a node reference in a relationship to be ingested.
type IngestibleEndpoint struct {
	Value   string              // The actual lookup value (either objectid or name)
	MatchBy IngestMatchStrategy // Strategy used to resolve the node
	Kind    graph.Kind          // Optional kind filter to help disambiguate nodes
}

type IngestibleRel struct {
	RelProps map[string]any
	RelType  graph.Kind
}

// IngestibleRelationship represents a directional relationship between two nodes
// intended for ingestion into the graph database. Both endpoints include resolution
// strategies and optional kind filters.
type IngestibleRelationship struct {
	Source IngestibleEndpoint
	Target IngestibleEndpoint

	RelProps map[string]any
	RelType  graph.Kind
}

func (s IngestibleRelationship) IsValid() bool {
	return s.Target.Value != "" && s.Source.Value != "" && s.RelProps != nil
}

type IngestibleSession struct {
	Source    string
	Target    string
	LogonType int
}

type IngestibleNode struct {
	ObjectID    string
	PropertyMap map[string]any
	Labels      []graph.Kind
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
