//go:build integration

package graphify_test

import (
	"context"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/lab/generic"
	"github.com/specterops/dawgs"
	"github.com/specterops/dawgs/drivers/pg"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

type IntegrationTestSuite struct {
	Context         context.Context
	GraphifyService graphify.GraphifyService
	GraphDB         graph.Database
	BHDatabase      *database.BloodhoundDB
	WorkDir         string
}

// setupIntegrationTestSuite initializes and returns a test suite containing
// all necessary dependencies for integration tests, including a connected
// graph database instance and a configured graph service.
func setupIntegrationTestSuite(t *testing.T, fixturesPath string) IntegrationTestSuite {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, getPostgresConfig(t), pgtestdb.NoopMigrator{})
		workDir  = t.TempDir()
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

	err = os.CopyFS(workDir, os.DirFS(fixturesPath))
	require.NoError(t, err)

	err = os.Mkdir(path.Join(workDir, "tmp"), 0755)
	require.NoError(t, err)

	cfg := config.Configuration{
		WorkDir: workDir,
	}

	return IntegrationTestSuite{
		Context:         ctx,
		GraphifyService: graphify.NewGraphifyService(ctx, db, graphDB, cfg, ingestSchema),
		GraphDB:         graphDB,
		BHDatabase:      db,
		WorkDir:         workDir,
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

func TestVersion6IngestJSON(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", t.Name(), "v6ingest")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "computers.json"),
			path.Join(testSuite.WorkDir, "containers.json"),
			path.Join(testSuite.WorkDir, "domains.json"),
			path.Join(testSuite.WorkDir, "gpos.json"),
			path.Join(testSuite.WorkDir, "groups.json"),
			path.Join(testSuite.WorkDir, "ous.json"),
			path.Join(testSuite.WorkDir, "sessions.json"),
			path.Join(testSuite.WorkDir, "users.json"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", t.Name())), "v6expected.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}

func TestVersion6IngestZIP(t *testing.T) {
	t.Parallel()
	var (
		ctx = context.Background()

		fixturesPath = path.Join("fixtures", t.Name(), "v6ingest")

		testSuite = setupIntegrationTestSuite(t, fixturesPath)

		files = []string{
			path.Join(testSuite.WorkDir, "archive.zip"),
		}
	)

	defer teardownIntegrationTestSuite(t, &testSuite)

	for _, file := range files {
		total, failed, err := testSuite.GraphifyService.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeZip}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 8, total)
	}

	expected, err := generic.LoadGraphFromFile(os.DirFS(path.Join("fixtures", t.Name())), "v6expected.json")
	require.NoError(t, err)
	generic.AssertDatabaseGraph(t, ctx, testSuite.GraphDB, &expected)
}
