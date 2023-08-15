package model_test

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestComboNodeElement_FromNode(t *testing.T) {
	element := model.ComboNodeElement{}

	// Creating a node with all the properties holding invalid value types.
	// Then we will correct the data type one by one and verify that
	// the error evolves accordingly, until we have no error returned.
	node := &graph.Node{
		Kinds: graph.Kinds{ad.Domain},
		Properties: &graph.Properties{
			Map: map[string]any{
				common.Description.String():   123,
				ad.DistinguishedName.String(): 123,
				ad.Domain.String():            123,
				ad.DomainSID.String():         123,
				ad.HighValue.String():         123,
				common.LastSeen.String():      123,
				common.ObjectID.String():      123,
				common.SystemTags.String():    123,
				"category":                    123,
				"level":                       "foo",
			},
		},
	}

	require.Contains(t, element.FromNode(node).Error(), "Description")

	node.Properties.Set(common.Description.String(), "description")
	require.Contains(t, element.FromNode(node).Error(), "DistinguishedName")

	node.Properties.Set(ad.DistinguishedName.String(), "distinguishedName")
	require.Contains(t, element.FromNode(node).Error(), "Domain")

	node.Properties.Set(ad.Domain.String(), "domain")
	require.Contains(t, element.FromNode(node).Error(), "DomainSID")

	node.Properties.Set(ad.DomainSID.String(), "domainsid")
	require.Contains(t, element.FromNode(node).Error(), "HighValue")

	node.Properties.Set(ad.HighValue.String(), true)
	require.Contains(t, element.FromNode(node).Error(), "LastSeen")

	node.Properties.Set(common.LastSeen.String(), time.Now())
	require.Contains(t, element.FromNode(node).Error(), "ObjectID")

	node.Properties.Set(common.ObjectID.String(), "objectid")
	require.Contains(t, element.FromNode(node).Error(), "SystemTags")

	node.Properties.Set(common.SystemTags.String(), "systemtags")
	require.Contains(t, element.FromNode(node).Error(), "Category")

	node.Properties.Set("category", "category")
	require.Contains(t, element.FromNode(node).Error(), "Level")

	node.Properties.Set("level", 1)
	require.Nil(t, element.FromNode(node))
	require.Equal(t, ad.Domain.String(), element.Kind)
}
