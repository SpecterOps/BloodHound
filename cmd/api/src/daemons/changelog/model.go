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
	propLastSeen          = "lastseen"
	propObjectID          = "objectid"
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

	if propertiesHash, err := props.Hash(ignoredPropertiesKeys); err != nil {
		return 0, fmt.Errorf("node properties hash error: %w", err)
	} else if kindsHash, err := s.Kinds.Hash(); err != nil {
		return 0, fmt.Errorf("node kinds hash error: %w", err)
	} else {
		combined := append(propertiesHash, kindsHash...)
		return xxhash.Sum64(combined), nil
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

	if dataHash, err := props.Hash(ignoredPropertiesKeys); err != nil {
		return 0, fmt.Errorf("edge properties hash error: %w", err)
	} else {
		return xxhash.Sum64(dataHash), nil
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
	return props.Get(key).Time()
}

func getStringProp(props *graph.Properties, key string) (string, error) {
	return props.Get(key).String()
}
