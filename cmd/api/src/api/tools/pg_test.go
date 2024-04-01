package tools

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/dawgs/util/size"
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
		migrator = setupTestMigrator(t, ctx, graphDB)
	)

	// lookup creates the database_switch table if needed
	driver, err := LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)

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
		migrator = setupTestMigrator(t, ctx, graphDB)
	)

	driver, err := LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)

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

func TestCancelMigration(t *testing.T) {
	var (
		mockCtrl        = gomock.NewController(t)
		graphDB         = graph_mocks.NewMockDatabase(mockCtrl)
		request         = httptest.NewRequest(http.MethodPut, "/pg-migrate/cancel", nil)
		recorder        = httptest.NewRecorder()
		ctx, cancelFunc = context.WithCancel(request.Context())
		migrator        = setupTestMigrator(t, ctx, graphDB)
	)

	// This seems kinda hacky
	migrator.migrationCancelFunc = cancelFunc
	migrator.advanceState(stateMigrating, stateIdle)

	migrator.MigrationCancel(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	require.Equal(t, http.StatusAccepted, response.StatusCode)

	require.Nil(t, true)
}

func TestPGMigrator(t *testing.T) {
	testContext := integration.NewGraphTestContext(t, graphschema.DefaultGraphSchema())
	integration.SetupDB(t)

	testContext.DatabaseTestWithSetup(func(harness *integration.HarnessDetails) error {
		harness.DBMigrateHarness.Setup(testContext)
		return nil
	}, func(harness integration.HarnessDetails, neo4jDB graph.Database) {
		var (
			request  = httptest.NewRequest(http.MethodPut, "/pg-migrate/start", nil)
			recorder = httptest.NewRecorder()
			migrator = setupTestMigrator(t, testContext.Context(), neo4jDB)
		)

		// Start migration process
		migrator.MigrationStart(recorder, request)
		response := recorder.Result()
		defer response.Body.Close()

		require.Equal(t, http.StatusAccepted, response.StatusCode)

		// Poll migration status handler until we see an "idle" status
		for {
			if migratorState := checkMigrationStatus(t, migrator); migratorState == stateMigrating {
				log.Infof("Migration in progress, waiting...")
				time.Sleep(1000 * 1000 * 100) // 1/10th of a second
			} else if migratorState == stateIdle {
				break
			} else {
				t.Fatalf("Encountered invalid migration status: %s", migratorState)
			}
		}

		// WIP: validate nodes/edges/types in pg

		pgDB, err := dawgs.Open(migrator.serverCtx, pg.DriverName, dawgs.Config{
			TraversalMemoryLimit: size.Gibibyte,
			DriverCfg:            migrator.cfg.Database.PostgreSQLConnectionString(),
		})
		require.Nil(t, err)

		pgDB.ReadTransaction(migrator.serverCtx, func(tx graph.Transaction) error {
			return tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for next := range cursor.Chan() {
					log.Infof("confirming node: %+v", next)
					return nil
				}

				return cursor.Error()
			})

		})
		require.Nil(t, true)
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
