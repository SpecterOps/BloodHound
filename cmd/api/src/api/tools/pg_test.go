package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestSwitchPostgreSQL(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		graphDB  = graph_mocks.NewMockDatabase(mockCtrl)
		request  = httptest.NewRequest(http.MethodPut, "/graph-db/switch/pg", nil)
		recorder = httptest.NewRecorder()
		ctx      = request.Context()
	)

	migrator := setupTestMigrator(t, ctx, graphDB)

	// lookup sets the driver from config and creates the database_switch table if needed
	driver, err := LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)

	// Set starting value to neo4j
	if driver != neo4j.DriverName {
		err = SetGraphDriver(migrator.serverCtx, migrator.cfg, neo4j.DriverName)
		require.Nil(t, err)
	}

	migrator.SwitchPostgreSQL(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	driver, err = LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)
	require.Equal(t, pg.DriverName, driver)
}

func TestSwitchNeo4j(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		graphDB  = graph_mocks.NewMockDatabase(mockCtrl)
		request  = httptest.NewRequest(http.MethodPut, "/graph-db/switch/neo4j", nil)
		recorder = httptest.NewRecorder()
		ctx      = request.Context()
	)

	migrator := setupTestMigrator(t, ctx, graphDB)

	// lookup sets the driver from config and creates the database_switch table if needed
	driver, err := LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)

	// Set starting value to pg
	if driver != pg.DriverName {
		err = SetGraphDriver(migrator.serverCtx, migrator.cfg, pg.DriverName)
		require.Nil(t, err)
	}

	migrator.SwitchNeo4j(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	driver, err = LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)
	require.Equal(t, neo4j.DriverName, driver)
}

func TestPGMigrator(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	integration.SetupDB(t)

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.DBMigrateHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, graphDB graph.Database) {
		var (
			request  = httptest.NewRequest(http.MethodPut, "/pg-migrate/start", nil)
			recorder = httptest.NewRecorder()
			migrator = setupTestMigrator(t, testContext.Context(), graphDB)
		)

		// Start migration process
		migrator.MigrationStart(recorder, request)
		response := recorder.Result()
		defer response.Body.Close()

		require.Equal(t, http.StatusAccepted, response.StatusCode)

		// Poll migration status handler until we see an "idle" status
		for {
			if migratorState := checkMigrationStatus(t, migrator); migratorState == stateMigrating {
				log.Infof("Migration in progress, waiting 1 second...")
				time.Sleep(1000 * 1000 * 100)
			} else if migratorState == stateIdle {
				break
			} else {
				t.Fatalf("Encountered invalid migration status: %s", migratorState)
			}
		}

		// TODO: validate nodes/edges/types in pg
	})
}

func checkMigrationStatus(t *testing.T, migrator *PGMigrator) MigratorState {
	var (
		request  = httptest.NewRequest(http.MethodGet, "/pg-migrate/status", nil)
		recorder = httptest.NewRecorder()
		body     struct{ State MigratorState }
	)

	migrator.MigrationStatus(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	require.Equal(t, http.StatusOK, response.StatusCode)

	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatal("failed to decode json")
	}
	require.NotEmpty(t, body.State)

	return body.State
}

func setupTestMigrator(t *testing.T, ctx context.Context, graphDB graph.Database) *PGMigrator {
	var (
		schema   = graphschema.DefaultGraphSchema()
		dbSwitch = graph.NewDatabaseSwitch(ctx, graphDB)
	)

	cfg, err := utils.LoadIntegrationTestConfig()
	require.Nil(t, err)

	return NewPGMigrator(ctx, cfg, schema, dbSwitch)
}
