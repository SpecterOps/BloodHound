package dbpool

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/dawgs/drivers/pg"
)

const (
	poolInitConnectionTimeout = time.Second * 10
)

func NewDawgsPool(cfg config.DatabaseConfiguration) (*pgxpool.Pool, error) {
	return newPool(cfg, true)
}

func NewAppPool(cfg config.DatabaseConfiguration) (*pgxpool.Pool, error) {
	return newPool(cfg, false)
}

func newPool(cfg config.DatabaseConfiguration, enableGraphHooks bool) (*pgxpool.Pool, error) {
	poolCtx, done := context.WithTimeout(context.Background(), poolInitConnectionTimeout)
	defer done()

	poolCfg, err := pgxpool.ParseConfig(cfg.PostgreSQLConnectionString())
	if err != nil {
		return nil, err
	}

	// TODO: Min and Max connections for the pool should be configurable
	poolCfg.MinConns = 5
	poolCfg.MaxConns = 50

	if enableGraphHooks {
		// Bind functions to the AfterConnect and AfterRelease hooks to ensure that composite type registration occurs.
		// Without composite type registration, the pgx connection type will not be able to marshal PG OIDs to their
		// respective Golang structs.
		poolCfg.AfterConnect = pg.AfterPooledConnectionEstablished
		poolCfg.AfterRelease = pg.AfterPooledConnectionRelease
	}

	if cfg.EnableRDSIAMAuth {
		// Only enable the BeforeConnect handler if RDS IAM Auth is enabled
		cfg.Endpoint = cfg.LookupEndpoint()
		poolCfg.BeforeConnect = func(ctx context.Context, connCfg *pgx.ConnConfig) error {
			if newPoolCfg, err := pgxpool.ParseConfig(cfg.PostgreSQLConnectionString()); err != nil {
				return err
			} else {
				connCfg.Password = newPoolCfg.ConnConfig.Password
			}

			return nil
		}
	}

	pool, err := pgxpool.NewWithConfig(poolCtx, poolCfg)
	if err != nil {
		return nil, err
	}

	return pool, nil
}
