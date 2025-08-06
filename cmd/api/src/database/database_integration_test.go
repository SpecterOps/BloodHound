//go:build integration
// +build integration

package database_test

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context         context.Context
	GraphifyService graphify.GraphifyService
	BHDatabase      *database.BloodhoundDB
	WorkDir         string
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests, including a connected
// graph database instance and a configured graph service.
func setupIntegrationTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
		workDir  = t.TempDir()
	)

	// #region Setup for dbs
	pool, err := pg.NewPool(connConf.URL())
	require.NoError(t, err)

	gormDB, err := database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.NoError(t, err)

	err = migrations.NewGraphMigrator(graphDB).Migrate(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	err = db.Migrate(ctx)
	require.NoError(t, err)

	err = graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	ingestSchema, err := upload.LoadIngestSchema()
	require.NoError(t, err)

	// #endregion

	cfg := config.Configuration{
		WorkDir: workDir,
	}

	return IntegrationTestSuite{
		Context:         ctx,
		GraphifyService: graphify.NewGraphifyService(ctx, db, graphDB, cfg, ingestSchema),
		BHDatabase:      db,
		WorkDir:         workDir,
	}
}

// getPostgresConfig reads key/value pairs from the default integration
// config file and creates a pgtestdb configuration object.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	config, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(config.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

func teardownIntegrationTestSuite(t *testing.T, suite *IntegrationTestSuite) {
	t.Helper()

	if suite.BHDatabase != nil {
		suite.BHDatabase.Close(suite.Context)
	}
}
