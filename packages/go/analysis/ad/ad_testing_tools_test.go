package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"

	"github.com/rs/xid"

	faker "github.com/go-faker/faker/v4"
)

func generateCollectedDomain() *graph.Node {
	return &graph.Node{
		Kinds: graph.Kinds{
			ad.Domain,
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			common.Collected: true,
			common.Name:      faker.DomainName(),
			ad.DomainSID:     xid.New().String(),
			ad.DomainFQDN:    faker.DomainName(),
		}),
	}
}

type linkWellKnownGroupsTestCase struct {
	name         string
	expectedNode graph.Node
	setupFunc    func(
		t *testing.T,
		ctx context.Context,
		graphDB graph.Database,
		expectedNode *graph.Node,
	) *graph.Node
}
