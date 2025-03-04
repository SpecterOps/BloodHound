package ad_test

import (
	"context"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/stretchr/testify/require"

	adAnalysis "github.com/specterops/bloodhound/analysis/ad"
	"github.com/specterops/bloodhound/analysis/ad/wellknown"
)

type linkWellKnownGroupsTestCase struct {
	name      string
	setupFunc func(
		t *testing.T,
		ctx context.Context,
		graphDB graph.Database,
	) *graph.Node
	assertionFunc func(
		t *testing.T,
		ctx context.Context,
		graphDB graph.Database,
		expectedNode *graph.Node,
	)
}

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

	wellKnownGroups := []struct {
		sidSuffix      wellknown.SIDSuffix
		nodeNamePrefix wellknown.NodeNamePrefix
	}{
		{
			sidSuffix:      wellknown.DomainUsersSIDSuffix,
			nodeNamePrefix: wellknown.DomainUsersNodeNamePrefix,
		},
	}

	testCases := []linkWellKnownGroupsTestCase{
		{
			name: "Verifies that linking all well-known groups succeeds when they already exist.",
			setupFunc: func(
				t *testing.T,
				ctx context.Context,
				graphDB graph.Database,
			) *graph.Node {
				// NOTE: Testing the scenario requires created the wellknown groups ahead of time asserting their
				// execution when asserting the scenario prior to asserting the expected outcome for
				// LinkWellKnownGroups
				createdCollectedDomainNode := createCollectedDomainNode(t, ctx, graphDB)
				for i := range wellKnownGroups {
					createdNode := createWellKnownGroup(
						t,
						ctx,
						graphDB,
						createdCollectedDomainNode,
						wellKnownGroups[i].sidSuffix,
						wellKnownGroups[i].nodeNamePrefix,
					)
					assertNodeExists(
						t,
						ctx,
						graphDB,
						createdNode,
						query.StringEndsWith(common.ObjectID.String(), wellKnownGroups[i].sidSuffix.String()),
						query.StringStartsWith(common.Name.String(), wellKnownGroups[i].nodeNamePrefix.String()),
					)
				}
				return createdCollectedDomainNode
			},
			assertionFunc: func(
				t *testing.T,
				ctx context.Context,
				graphDB graph.Database,
				expectedNode *graph.Node,
			) {
				// assert that the relationships exist
				// for i := range wellKnownGroups {
				// 	wellKnownNode, err := fetchNode(
				// 		t,
				// 		ctx,
				// 		graphDB,
				// 		expectedNode,
				// 		query.StringEndsWith(common.ObjectID.String(), wellKnownGroups[i].sidSuffix.String()),
				// 		query.StringStartsWith(common.Name.String(), wellKnownGroups[i].nodeNamePrefix.String()),
				// 	)
				// 	require.NoError(t, err)
				// }
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
					)
					require.NotNil(t, createdNode)

					// Run LinkWellKnownGroups
					err := adAnalysis.LinkWellKnownGroups(ctx, graphDB)
					require.NoError(t, err)

					tc.assertionFunc(t, ctx, graphDB, createdNode)
					require.NoError(t, err)
				})
			}
		})
	}
}
