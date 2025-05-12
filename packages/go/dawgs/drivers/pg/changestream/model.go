package changestream

import (
	"github.com/cespare/xxhash/v2"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type ChangeType int

const (
	ChangeTypeUpdate ChangeType = 0
	ChangeTypeAdd    ChangeType = 1
	ChangeTypeRemove ChangeType = 2
)

type Change interface {
	Type() ChangeType
}

type change struct {
	changeType ChangeType
}

func (s change) Type() ChangeType {
	return s.changeType
}

type NodeChange struct {
	change

	TargetNodeID string
	Kinds        graph.Kinds
	Properties   *graph.Properties
}

func NewNodeChange(changeType ChangeType, targetNodeID string, kinds graph.Kinds, properties *graph.Properties) *NodeChange {
	return &NodeChange{
		change: change{
			changeType: changeType,
		},
		TargetNodeID: targetNodeID,
		Kinds:        kinds,
		Properties:   properties,
	}
}

func (s NodeChange) CacheKey() string {
	return s.TargetNodeID
}

type EdgeChange struct {
	change

	TargetNodeID  string
	RelatedNodeID string
	Kind          graph.Kind
	Properties    *graph.Properties
}

func NewEdgeChange(changeType ChangeType, targetNodeID, relatedNodeID string, kind graph.Kind, properties *graph.Properties) *EdgeChange {
	return &EdgeChange{
		change: change{
			changeType: changeType,
		},
		TargetNodeID:  targetNodeID,
		RelatedNodeID: relatedNodeID,
		Kind:          kind,
		Properties:    properties,
	}
}
func (s EdgeChange) CacheKey() string {
	return s.TargetNodeID + s.RelatedNodeID + s.Kind.String()
}

func (s EdgeChange) IdentityHash() ([]byte, error) {
	digest := xxhash.New()

	if _, err := digest.Write([]byte(s.CacheKey())); err != nil {
		return nil, err
	}

	return digest.Sum(nil), nil
}

type ChangeLookup struct {
	Type           ChangeType
	PropertiesHash []byte
	Exists         bool
	Changed        bool
}

func (s ChangeLookup) ShouldSubmit() bool {
	return !s.Exists || s.Changed
}
