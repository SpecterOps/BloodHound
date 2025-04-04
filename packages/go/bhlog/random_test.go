package bhlog_test

import (
	"context"
	"testing"

	_ "github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/graphschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRandom(t *testing.T) {
	var ctx = context.Background()

	conf := pgtestdb.Config{
		DriverName: "pgx",
		User:       "bhe",
		Password:   "weneedbetterpasswords",
		Host:       "localhost",
		Port:       "55432",
		Options:    "sslmode=disable",
	}

	connConf := pgtestdb.Custom(t, conf, pgtestdb.NoopMigrator{})

	pool, err := pg.NewPool(connConf.URL())
	require.Nil(t, err)

	graphDB, err := dawgs.Open(ctx, pg.DriverName, dawgs.Config{
		GraphQueryMemoryLimit: 1024 * 1024 * 1024 * 2,
		ConnectionString:      connConf.URL(),
		Pool:                  pool,
	})
	require.Nil(t, err)

	err = graphDB.AssertSchema(ctx, graphschema.DefaultGraphSchema())
	assert.Nil(t, err)
}
