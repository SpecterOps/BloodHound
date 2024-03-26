package tools

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	graph_mocks "github.com/specterops/bloodhound/dawgs/graph/mocks"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/src/test/integration/utils"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// func TestPGMigrator(t *testing.T) {
// 	var (
// 		mockCtrl = gomock.NewController(t)
// 		ctx      = context.Background()
// 		schema   = graphschema.DefaultGraphSchema()
// 		graphDB  = graph_mocks.NewMockDatabase(mockCtrl)
// 		dbSwitch = graph.NewDatabaseSwitch(ctx, graphDB)
// 	)

// 	integration.SetupDB(t)

// 	cfg, err := utils.LoadIntegrationTestConfig()
// 	require.Nil(t, err)

// 	migrator := NewPGMigrator(ctx, cfg, schema, dbSwitch)

// 	if err := migrator.startMigration(); err != nil {
// 		log.Errorf("migration failed to start with error: %w", err)
// 	}

// 	for i := 0; i < 5; i++ {
// 		log.Infof("migration state: %v", migrator.state)
// 		time.Sleep(1000 * 1000 * 500)
// 	}

// 	require.Nil(t, true)
// }

func setupTestMigrator(t *testing.T, ctx context.Context) (*PGMigrator, error) {
	var (
		mockCtrl = gomock.NewController(t)
		schema   = graphschema.DefaultGraphSchema()
		graphDB  = graph_mocks.NewMockDatabase(mockCtrl)
		dbSwitch = graph.NewDatabaseSwitch(ctx, graphDB)
	)

	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		return nil, err
	} else {
		return NewPGMigrator(ctx, cfg, schema, dbSwitch), nil
	}
}

func TestSwitchPostgreSQL(t *testing.T) {
	var (
		request  = httptest.NewRequest(http.MethodPut, "/graph-db/switch/pg", nil)
		recorder = httptest.NewRecorder()
		ctx      = request.Context()
	)

	migrator, err := setupTestMigrator(t, ctx)
	require.Nil(t, err)

	err = SetGraphDriver(migrator.serverCtx, migrator.cfg, neo4j.DriverName)
	require.Nil(t, err)

	migrator.SwitchPostgreSQL(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	driver, err := LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)
	require.Equal(t, pg.DriverName, driver)
}

func TestSwitchNeo4j(t *testing.T) {
	var (
		request  = httptest.NewRequest(http.MethodPut, "/graph-db/switch/neo4j", nil)
		recorder = httptest.NewRecorder()
		ctx      = request.Context()
	)

	migrator, err := setupTestMigrator(t, ctx)
	require.Nil(t, err)

	err = SetGraphDriver(migrator.serverCtx, migrator.cfg, pg.DriverName)
	require.Nil(t, err)

	migrator.SwitchNeo4j(recorder, request)

	response := recorder.Result()
	defer response.Body.Close()

	driver, err := LookupGraphDriver(migrator.serverCtx, migrator.cfg)
	require.Nil(t, err)
	require.Equal(t, neo4j.DriverName, driver)
}

// basic steps for runbook:
//
// 1. GET request to /pg-migration/status should return { "state": "idle" }
//
// 2. PUT request to /pg-migration/start starts the migration process
//
// 3. Poll with GET request to /pg-migration/status to see when migration has finished
//     - should return { "state": "migrating" } if process has not completed yet
//     - currently, errors will only surface in the API logs
//
// 4. Once migration has completed, switch db driver to postgres with PUT to /graph-db/switch/pg
//     - Possible to toggle back to neo4j with PUT to /graph-db/switch/neo4j
