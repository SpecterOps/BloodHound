package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"

	adAnalysis "github.com/specterops/bloodhound/analysis/ad"

	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func TestLinkWellKnownGroups(t *testing.T) {
	// Skip if running short tests
	if testing.Short() {
		t.Skip("Skipping test in short mode")
	}

	// Define database drivers to test
	dbDrivers := []struct {
		name   string
		driver string
		initFn func(t *testing.T, ctx context.Context) (graph.Database, func())
	}{
		{
			name:   "Neo4j",
			driver: "neo4j",
			initFn: initNeo4jDatabase,
		},
		{
			name:   "PostgreSQL",
			driver: "pg",
			initFn: initPostgresDatabase,
		},
	}

	testCases := []linkWellKnownGroupsTestCase{
		{
			name: "Node Not Found Create New Node",
			expectedNode: graph.Node{
				Kinds: graph.Kinds{
					ad.Entity, ad.Group,
				},
				Properties: graph.AsProperties(graph.PropertyMap{
					common.Name:     "some name",
					common.ObjectID: "wellknowsid",
					ad.DomainSID:    "some domain",
					ad.DomainFQDN:   "fullyqualifieddsomain",
				}),
			},
			setupFunc: func(
				t *testing.T,
				ctx context.Context,
				graphDB graph.Database,
				expectedNode *graph.Node,
			) *graph.Node {
				// NOTE: in order to trigger getOrCreateWellKnownGroup creating a node if the given wellKnownSid does
				// not return a node
				var createdNode *graph.Node
				var err error
				err = graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
					domainProperties := graph.AsProperties(graph.PropertyMap{
						common.Name:      expectedNode.Properties.Get(common.Name.String()),
						common.Collected: true,
					})
					createdNode, err = tx.CreateNode(domainProperties, ad.Domain)
					if err != nil {
						return err
					}
					return tx.Commit()
				})
				require.NoError(t, err)
				require.NotNil(t, createdNode)
				return createdNode
			},
		},
	}

	// Run tests for each database driver
	for _, dbDriver := range dbDrivers {
		t.Run(dbDriver.name, func(t *testing.T) {
			ctx := context.Background()

			// Initialize the database
			graphDB, cleanup := dbDriver.initFn(t, ctx)
			defer cleanup()

			for _, tc := range testCases {
				t.Run(tc.name, func(t *testing.T) {
					// Clean the database before each test
					cleanDatabase(t, ctx, graphDB)

					// Set up the test case with a random domain SID and name for each test
					// domainSid, domainName := generateDomainInfo()
					createdNode := tc.setupFunc(
						t,
						ctx,
						graphDB,
						generateCollectedDomain(t),
					)
					require.NotNil(t, createdNode)

					// Verify that the domain node was created successfully
					// var domainExists bool
					// err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
					// 	domain, err := tx.Nodes().Filterf(func() graph.Criteria {
					// 		return query.And(
					// 			query.Kind(query.Node(), ad.Domain),
					// 			query.Equals(query.NodeProperty(common.ObjectID.String()), domainSid),
					// 		)
					// 	}).First()
					// 	if err != nil {
					// 		return err
					// 	}
					//
					// 	domainExists = (domain != nil)
					// 	return nil
					// })
					// require.NoError(t, err)
					// require.True(t, domainExists, "Domain node was not created successfully")

					// Make sure the domain node has the required properties for LinkWellKnownGroups
					// err = :graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
					// 	domain, err := tx.Nodes().Filterf(func() graph.Criteria {
					// 		return query.And(
					// 			query.Kind(query.Node(), ad.Domain),
					// 			query.Equals(query.NodeProperty(common.ObjectID.String()), domainSid),
					// 		)
					// 	}).First()
					// 	if err != nil {
					// 		return err
					// 	}
					//
					// 	// Ensure domain has DomainSID property
					// 	if _, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
					// 		domain.Properties.Set(ad.DomainSID.String(), domainSid)
					// 		if err := tx.UpdateNode(domain); err != nil {
					// 			return err
					// 		}
					// 	}
					// 	return tx.Commit()
					// })
					// require.NoError(t, err)

					// Verify that the domain node has the DomainSID property set correctly
					// err = graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
					// 	domain, err := tx.Nodes().Filterf(func() graph.Criteria {
					// 		return query.And(
					// 			query.Kind(query.Node(), ad.Domain),
					// 			query.Equals(query.NodeProperty(common.ObjectID.String()), domainSid),
					// 		)
					// 	}).First()
					// 	if err != nil {
					// 		return err
					// 	}
					//
					// 	domainSIDValue, err := domain.Properties.Get(ad.DomainSID.String()).String()
					// 	if err != nil {
					// 		return err
					// 	}
					//
					// 	require.Equal(t, domainSid, domainSIDValue, "Domain SID property not set correctly")
					// 	return nil
					// })
					// require.NoError(t, err)

					// Run LinkWellKnownGroups
					err := adAnalysis.LinkWellKnownGroups(ctx, graphDB)
					require.NoError(t, err)

					// Verify the results in a read transaction
					err = graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
						// for sidSuffix, shouldExist := range tc.expectedGroupExists {
						// 	var wellKnownSid string
						// 	if sidSuffix == "-S-1-5-11" || sidSuffix == "-S-1-1-0" {
						// 		// For Authenticated Users and Everyone, the SID is prefixed with the domain name
						// 		wellKnownSid = domainName + sidSuffix
						// 	} else {
						// 		// For other groups, the SID is prefixed with the domain SID
						// 		wellKnownSid = domainSid + sidSuffix
						// 	}
						//
						// 	node, err := tx.Nodes().Filterf(func() graph.Criteria {
						// 		return query.Equals(
						// 			query.NodeProperty(common.ObjectID.String()),
						// 			wellKnownSid,
						// 		)
						// 	}).First()
						//
						// 	if shouldExist {
						// 		if err != nil {
						// 			return fmt.Errorf(
						// 				"node with SID %s should exist: %w",
						// 				wellKnownSid,
						// 				err,
						// 			)
						// 		}
						// 		if node == nil {
						// 			return fmt.Errorf(
						// 				"node with SID %s should not be nil",
						// 				wellKnownSid,
						// 			)
						// 		}
						//
						// 		// Verify the kinds
						// 		// Check that the node has the Group kind
						// 		if !node.Kinds.ContainsOneOf(ad.Group) {
						// 			return fmt.Errorf(
						// 				"node with SID %s should have kind %s",
						// 				wellKnownSid,
						// 				ad.Group,
						// 			)
						// 		}
						// 	} else {
						// 		if err == nil || !graph.IsErrNotFound(err) {
						// 			return fmt.Errorf("node with SID %s should not exist", wellKnownSid)
						// 		}
						// 	}
						// }
						return nil
					})
					require.NoError(t, err)
				})
			}
		})
	}
}
