//go:build integration

package graphify_test

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/lab/generic"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/services/graphify"
	"github.com/specterops/bloodhound/src/services/upload"
	"github.com/stretchr/testify/require"
)

func TestVersion6Ingest(t *testing.T) {
	// TODO: Create setup function (should return a valid DB and GraphDB handle at minimum)
	t.Parallel()
	var (
		ctx = context.Background()
		// TODO: Read in configuration from integration testing file, no hard coded DB connections
		conf = pgtestdb.Config{
			DriverName:                "pgx",
			User:                      "bloodhound",
			Password:                  "bloodhoundcommunityedition",
			Host:                      "localhost",
			Port:                      "65432",
			Options:                   "sslmode=disable",
			ForceTerminateConnections: true,
		}
		connConf = pgtestdb.Custom(t, conf, pgtestdb.NoopMigrator{})

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
	)

	//#region Setup for dbs
	pool, err := pg.NewPool(connConf.URL())
	require.NoError(t, err)

	gormDB, err := database.OpenDatabase(connConf.URL())
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, auth.NewIdentityResolver())

	// TODO: make sure to migrate the DB so we have a good blank slate

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.NoError(t, err)

	err = graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema())
	require.NoError(t, err)

	ingestSchema, err := upload.LoadIngestSchema()
	require.NoError(t, err)
	//#endregion

	service := graphify.NewGraphifyService(ctx, db, graphDB, config.Configuration{}, ingestSchema)

	//#region hacks until refactor for embedfs
	cmd := exec.Command("rm", "-r", "fixtures/tmp")
	err = cmd.Run()
	require.NoError(t, err, "%s", cmd.Err)

	cmd = exec.Command("cp", "-r", "fixtures/v6ingest/", "fixtures/tmp/")
	err = cmd.Run()
	require.NoError(t, err, "%s", cmd.Err)
	//#endregion

	for _, file := range files {
		total, failed, err := service.ProcessIngestFile(ctx, model.IngestTask{FileName: file, FileType: model.FileTypeJson}, time.Now())
		require.NoError(t, err)
		require.Zero(t, failed)
		require.Equal(t, 1, total)
	}

	// TODO: use an embedded filesystem instead of bodging dirfs here
	expected, err := generic.LoadGraphFromFile(os.DirFS("fixtures"), "v6expected.json")
	require.NoError(t, err)
	// TODO: decide if AssertDatabaseGraph should get a DB handle or transaction (should we worry about wrapping transactions
	// at this level or let the function do it)
	generic.AssertDatabaseGraph(t, ctx, graphDB, &expected)
}
