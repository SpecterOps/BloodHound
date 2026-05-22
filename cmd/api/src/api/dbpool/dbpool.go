package dbpool

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/dawgs/drivers/pg"
)

const (
	poolInitConnectionTimeout = time.Second * 10
)

func getConnectionString(cfg config.DatabaseConfiguration) string {
	if cfg.EnableRDSIAMAuth {
		// Only enable the BeforeConnect handler if RDS IAM Auth is enabled
		cfg.Endpoint = cfg.LookupEndpoint()
	}
	return cfg.PostgreSQLConnectionString()
}

func NewDawgsPool(cfg config.DatabaseConfiguration) (*pgxpool.Pool, error) {
	return pg.NewPool(getConnectionString(cfg))
}

func NewAppPool(cfg config.DatabaseConfiguration) (*pgxpool.Pool, error) {
	return newPool(getConnectionString(cfg))
}

func newPool(connString string) (*pgxpool.Pool, error) {
	poolCtx, done := context.WithTimeout(context.Background(), poolInitConnectionTimeout)
	defer done()

	poolCfg, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, err
	}

	// TODO: Min and Max connections for the pool should be configurable
	poolCfg.MinConns = 5
	poolCfg.MaxConns = 50

	return pgxpool.NewWithConfig(poolCtx, poolCfg)
}
