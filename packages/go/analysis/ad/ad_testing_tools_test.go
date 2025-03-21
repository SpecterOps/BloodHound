package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"

	"github.com/specterops/bloodhound/analysis/ad/internal"
	"github.com/specterops/bloodhound/analysis/ad/wellknown"

	"github.com/rs/xid"
	"github.com/stretchr/testify/require"

	faker "github.com/go-faker/faker/v4"
)

func fetchNode(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	queryCriterias ...graph.Criteria,
) (
	*graph.Node,
	error,
) {
	t.Helper()

	var fetchedNode *graph.Node
	err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		filteredNode, err := tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(queryCriterias...)
		}).First()
		if err != nil {
			return err
		}
		fetchedNode = filteredNode

		return nil
	})
	return fetchedNode, err
}

func assertNodeExists(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	expectedNode *graph.Node,
	queryCriterias ...graph.Criteria,
) {
	t.Helper()

	domainSID, err := expectedNode.Properties.Get(ad.DomainSID.String()).String()
	require.NoError(t, err)
	domainFQDN, err := expectedNode.Properties.Get(ad.DomainFQDN.String()).String()
	require.NoError(t, err)
	nodeName, err := expectedNode.Properties.Get(common.Name.String()).String()
	require.NoError(t, err)

	defaultQueryCriterias := []graph.Criteria{
		query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
		query.Equals(query.NodeProperty(ad.DomainFQDN.String()), domainFQDN),
		query.Equals(query.NodeProperty(common.Name.String()), nodeName),
	}

	queryCriterias = append(queryCriterias, defaultQueryCriterias)
	fetchedNode, err := fetchNode(t, ctx, graphDB, queryCriterias...)
	require.NoError(t, err)
	require.NotNil(t, fetchedNode)
}

func createNode(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	nodeToCreate *graph.Node,
) *graph.Node {
	t.Helper()
	require.NotNil(t, graphDB)
	require.NotNil(t, nodeToCreate)

	if ctx == nil {
		ctx = t.Context()
	}

	var createdNode *graph.Node
	var err error
	err = graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		createdNode, err = tx.CreateNode(nodeToCreate.Properties, nodeToCreate.Kinds...)
		if err != nil {
			return err
		}
		return tx.Commit()
	})
	require.NoError(t, err)
	require.NotNil(t, createdNode)

	assertNodeExists(t, ctx, graphDB, createdNode)
	return createdNode
}

func createCollectedDomainNode(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
) *graph.Node {
	return createNode(
		t,
		ctx,
		graphDB,
		generateCollectedDomain(),
	)
}

func createWellKnownGroup(
	t *testing.T,
	ctx context.Context,
	graphDB graph.Database,
	domainNode *graph.Node,
	sidSuffix wellknown.SIDSuffix,
	nodeNamePrefix wellknown.NodeNamePrefix,
) *graph.Node {
	return createNode(
		t,
		ctx,
		graphDB,
		generateWellKnownGroup(
			t,
			domainNode,
			sidSuffix,
			nodeNamePrefix,
		),
	)
}

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

func generateWellKnownGroup(
	t *testing.T,
	domainNode *graph.Node,
	sidSuffix wellknown.SIDSuffix,
	nodeNamePrefix wellknown.NodeNamePrefix,
) *graph.Node {
	require.NotNil(t, domainNode)
	require.NotNil(t, sidSuffix)
	require.NotNil(t, nodeNamePrefix)

	domainSID, domainName, err := internal.ReadDomainIDandNameAsString(domainNode)
	require.NoError(t, err)

	var wellKnownSID string
	if sidSuffix.String() == wellknown.DomainUsersSIDSuffix.String() {
		wellKnownSID = sidSuffix.PrependPrefix(domainSID)
	} else {
		wellKnownSID = sidSuffix.PrependPrefix(domainName)
	}

	return &graph.Node{
		Kinds: graph.Kinds{
			ad.Entity,
			ad.Group,
		},
		Properties: graph.AsProperties(graph.PropertyMap{
			ad.DomainFQDN:    domainName,
			ad.DomainSID:     domainSID,
			common.Collected: true,
			common.Name:      nodeNamePrefix.AppendSuffix(domainName),
			common.ObjectID:  wellKnownSID,
		}),
	}
}
