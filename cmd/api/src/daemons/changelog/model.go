// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package changelog

import (
	"fmt"
	"time"

	"github.com/cespare/xxhash/v2"
	"github.com/specterops/dawgs/graph"
)

type Change interface {
	IdentityKey() uint64   // identity hash
	Hash() (uint64, error) // content hash
	Apply(batch graph.Batch) error
}

var (
	propLastSeen = "lastseen"
	propObjectID = "objectid"
	// ignoredPropertiesKeys defines a set of node/edge properties that are
	// excluded from content hashing. These fields are typically volatile
	// (timestamps, collection metadata, environment-specific IDs) and are
	// not meaningful indicators of a substantive graph change.
	ignoredPropertiesKeys = map[string]struct{}{
		propLastSeen:    {},
		propObjectID:    {},
		"lastcollected": {},
		"isinherited":   {},
		"domainsid":     {},
		"isacl":         {},
		"tenantid":      {},
	}
)

type NodeChange struct {
	NodeID     string
	Properties *graph.Properties
	Kinds      graph.Kinds
}

func NewNodeChange(nodeID string, kinds graph.Kinds, properties *graph.Properties) *NodeChange {
	return &NodeChange{
		NodeID:     nodeID,
		Kinds:      kinds,
		Properties: properties,
	}
}

func (s NodeChange) IdentityKey() uint64 {
	hash := xxhash.Sum64String(s.NodeID)
	return hash
}

func (s NodeChange) Hash() (uint64, error) {
	props := s.Properties
	if props == nil {
		props = graph.NewProperties()
	}

	h := xxhash.New()

	if err := props.HashInto(h, ignoredPropertiesKeys); err != nil {
		return 0, fmt.Errorf("node properties hash error: %w", err)
	} else if err := s.Kinds.HashInto(h); err != nil {
		return 0, fmt.Errorf("node kinds hash error: %w", err)
	} else {
		return h.Sum64(), nil
	}
}

func (s NodeChange) Apply(batch graph.Batch) error {
	lastSeen, err := getTimeProp(s.Properties, propLastSeen)
	if err != nil {
		return fmt.Errorf("node: get %s: %w", propLastSeen, err)
	}
	objectID, err := getStringProp(s.Properties, propObjectID)
	if err != nil {
		return fmt.Errorf("node: get %s: %w", propObjectID, err)
	}

	update := graph.NodeUpdate{
		Node: graph.PrepareNode(
			graph.NewProperties().SetAll(map[string]any{
				propLastSeen: lastSeen,
				propObjectID: objectID,
			}),
		),
		IdentityProperties: []string{propObjectID},
	}

	if err := batch.UpdateNodeBy(update); err != nil {
		return fmt.Errorf("node: update objectid=%s: %w", objectID, err)
	}
	return nil
}

type EdgeChange struct {
	SourceNodeID string
	TargetNodeID string
	Kind         graph.Kind
	Properties   *graph.Properties
}

func NewEdgeChange(sourceNodeID, targetNodeID string, kind graph.Kind, properties *graph.Properties) *EdgeChange {
	return &EdgeChange{
		SourceNodeID: sourceNodeID,
		TargetNodeID: targetNodeID,
		Kind:         kind,
		Properties:   properties,
	}
}

func (s EdgeChange) IdentityKey() uint64 {
	identity := s.SourceNodeID + "|" + s.TargetNodeID + "|" + s.Kind.String()
	hash := xxhash.Sum64String(identity)
	return hash
}

func (s EdgeChange) Hash() (uint64, error) {
	props := s.Properties
	if props == nil {
		props = graph.NewProperties()
	}

	h := xxhash.New()

	if err := props.HashInto(h, ignoredPropertiesKeys); err != nil {
		return 0, fmt.Errorf("edge properties hash error: %w", err)
	} else {
		return h.Sum64(), nil
	}
}

func (s EdgeChange) Apply(batch graph.Batch) error {
	lastSeen, err := getTimeProp(s.Properties, propLastSeen)
	if err != nil {
		return fmt.Errorf("edge: get %q: %w", propLastSeen, err)
	}

	update := graph.RelationshipUpdate{
		Start:                   graph.PrepareNode(graph.NewProperties().SetAll(map[string]any{propObjectID: s.SourceNodeID, propLastSeen: lastSeen})),
		StartIdentityProperties: []string{propObjectID},
		End:                     graph.PrepareNode(graph.NewProperties().SetAll(map[string]any{propObjectID: s.TargetNodeID, propLastSeen: lastSeen})),
		EndIdentityProperties:   []string{propObjectID},
		Relationship:            graph.PrepareRelationship(graph.NewProperties().Set(propLastSeen, lastSeen), s.Kind),
	}

	if err := batch.UpdateRelationshipBy(update); err != nil {
		return fmt.Errorf("edge: update %v (%s->%s): %w", s.Kind, s.SourceNodeID, s.TargetNodeID, err)
	}
	return nil
}

func getTimeProp(props *graph.Properties, key string) (time.Time, error) {
	if props == nil {
		return time.Time{}, fmt.Errorf("missing properties for %q", key)
	}
	return props.Get(key).Time()
}

func getStringProp(props *graph.Properties, key string) (string, error) {
	if props == nil {
		return "", fmt.Errorf("missing properties for %q", key)
	}
	return props.Get(key).String()
}
