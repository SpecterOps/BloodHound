//go:build integration

package graphify_test

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/lab/generic"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/graphify"
	"github.com/specterops/bloodhound/src/services/upload"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context         context.Context
	GraphifyService graphify.GraphifyService
	GraphDB         graph.Database
	BHDatabase      *database.BloodhoundDB
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests, including a connected
// graph database instance and a configured graph service.
func setupIntegrationTestSuite(t *testing.T) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	//#region Setup for dbs
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

	//#endregion

	return IntegrationTestSuite{
		Context:         ctx,
		GraphifyService: graphify.NewGraphifyService(ctx, db, graphDB, config.Configuration{}, ingestSchema),
		GraphDB:         graphDB,
		BHDatabase:      db,
	}
}

func teardownIntegrationTestSuite(t *testing.T, suite *IntegrationTestSuite) {
	t.Helper()

	suite.GraphDB.Close(suite.Context)
	suite.BHDatabase.Close(suite.Context)
}

// getPostgresConfig reads key/value pairs from the default integration
// config file and creates a pgtestdb configuration object.
func getPostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	config, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	entries := strings.Split(config.Database.Connection, " ")

	environmentMap := make(map[string]string)
	for _, entry := range entries {
		parts := strings.Split(entry, "=")
		environmentMap[parts[0]] = parts[1]
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

func TestVersion6Ingest(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		files = []string{
			"fixtures/tmp/computers.json",
			"fixtures/tmp/containers.json",
			"fixtures/tmp/domains.json",
			"fixtures/tmp/gpos.json",
			"fixtures/tmp/groups.json",
			"fixtures/tmp/ous.json",
			"fixtures/tmp/sessions.json",
			"fixtures/tmp/users.json",
		}

		testSuite = setupIntegrationTestSuite(t)
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	//#region hacks until refactor for embedfs
	cmd := exec.Command("rm", "-r", "fixtures/tmp")
	err := cmd.Run()
	require.NoError(t, err, "%s", cmd.Err)

	cmd = exec.Command("cp", "-r", "fixtures/v6ingest/", "fixtures/tmp/")
	err = cmd.Run()
	require.NoError(t, err, "%s", cmd.Err)
	//#endregion

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	// TODO: use an embedded filesystem instead of bodging dirfs here
	expected, err := generic.LoadGraphFromFile(os.DirFS("fixtures"), "v6expected.json")
	require.NoError(t, err)
	// TODO: decide if AssertDatabaseGraph should get a DB handle or transaction (should we worry about wrapping transactions
	// at this level or let the function do it)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}
