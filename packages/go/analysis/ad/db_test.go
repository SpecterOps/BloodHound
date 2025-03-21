package ad_test

import (
	"context"
	"os"
	"testing"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/services"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/graphschema/common"
)

// Get database configuration from environment variables or use defaults
func getDatabaseConfig(t *testing.T, driver string) config.Configuration {
	neo4jURI := os.Getenv("BLOODHOUND_TESTING_NEO4J_URI")
	if neo4jURI == "" {
		neo4jURI = "neo4j://neo4j:bloodhoundcommunityedition@localhost:7687/" // Default Neo4j URI
	}

	postgresURI := os.Getenv("BLOODHOUND_TESTING_POSTGRES_URI")
	if postgresURI == "" {
		postgresURI = "postgres://bloodhound:bloodhoundcommunityedition@localhost:5432/bloodhound?sslmode=disable" // Default PostgreSQL URI
	}

	cfg := config.Configuration{
		GraphDriver: driver,
		Neo4J: config.DatabaseConfiguration{
			Connection: neo4jURI,
		},
		Database: config.DatabaseConfiguration{
			Connection: postgresURI,
		},
	}

	return cfg
}

// Initialize database using the Initializer pattern
func initDatabase(t *testing.T, ctx context.Context, driver string) (graph.Database, func()) {
	t.Helper()
	// Create configuration with database settings
	cfg := getDatabaseConfig(t, driver)
	require.NotEmpty(t, cfg)

	// Create an initializer with the database connector
	initializer := bootstrap.Initializer[*database.BloodhoundDB, *graph.DatabaseSwitch]{
		Configuration: cfg,
		DBConnector:   services.ConnectDatabases,
	}
	require.NotNil(t, initializer)

	// Connect to the database using the initializer's DBConnector
	connections, err := initializer.DBConnector(ctx, cfg)
	require.NotNil(t, connections)
	require.NoError(t, err, "Failed to connect to database")

	// Initialize the schema
	schema := graph.Schema{
		DefaultGraph: graph.Graph{
			Name: "bloodhound",
			NodeConstraints: []graph.Constraint{
				{
					Name:  "object_id_constraint",
					Field: common.ObjectID.String(),
					Type:  graph.BTreeIndex,
				},
			},
		},
	}

	// Assert schema on the graph database
	err = connections.Graph.AssertSchema(ctx, schema)
	require.NoError(t, err, "Failed to assert schema")

	// Return the graph database and a cleanup function
	cleanup := func() {
		t.Logf("closing database connections")
		t.Logf("closing GraphDB")
		if err := connections.Graph.Close(ctx); err != nil {
			t.Logf("Error closing graph database: %v", err)
		}
		t.Logf("closing RDMS")
		connections.RDMS.Close(ctx)
	}

	return connections.Graph, cleanup
}

// Initialize Neo4j database
func initNeo4jDatabase(t *testing.T, ctx context.Context) (graph.Database, func()) {
	return initDatabase(t, ctx, "neo4j")
}

// Initialize PostgreSQL database
func initPostgresDatabase(t *testing.T, ctx context.Context) (graph.Database, func()) {
	return initDatabase(t, ctx, "pg")
}

// Helper function to clean the database
func cleanDatabase(t *testing.T, ctx context.Context, graphDB graph.Database) {
	t.Helper()
	err := graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
		// Delete all nodes and relationships
		result := tx.Raw("MATCH (n) DETACH DELETE n", nil)
		defer result.Close()
		if err := result.Error(); err != nil {
			return err
		}
		return tx.Commit()
	})
	require.NoError(t, err, "Failed to clean the database")
}
