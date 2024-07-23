package v2

import (
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model"

	"github.com/stretchr/testify/assert"
)

func Test_SetNodeProperties(t *testing.T) {
	tests := []struct {
		name     string
		nodes    graph.NodeSet
		expected model.DomainSelectors
	}{
		{
			name: "basic case",
			nodes: graph.NodeSet{
				1: &graph.Node{
					ID: 1,
					Properties: &graph.Properties{
						Map: map[string]any{
							common.ObjectID.String():  "1",
							common.Name.String():      "Node1",
							common.Collected.String(): false},
					},
				},
			},
			expected: model.DomainSelectors{
				{
					Type:      "active-directory",
					Name:      "Node1",
					ObjectID:  "1",
					Collected: false,
				},
			},
		},
		{
			name: "azure tenant",
			nodes: graph.NodeSet{
				1: &graph.Node{
					Properties: &graph.Properties{
						Map: map[string]any{
							common.ObjectID.String():  "2",
							common.Name.String():      "Node2",
							common.Collected.String(): true,
						},
					},
					Kinds: graph.Kinds{azure.Tenant},
				},
			},
			expected: model.DomainSelectors{
				{
					Type:      "azure",
					Name:      "Node2",
					ObjectID:  "2",
					Collected: true,
				},
			},
		},
		{
			name: "missing properties",
			nodes: graph.NodeSet{
				1: &graph.Node{
					Properties: &graph.Properties{
						Map: map[string]any{},
					},
				},
			},
			expected: model.DomainSelectors{
				{
					Type:      "active-directory",
					Name:      "NO NAME",
					ObjectID:  "NO OBJECT ID",
					Collected: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := setNodeProperties(tt.nodes)
			assert.Equal(t, tt.expected, got)
		})
	}
}
