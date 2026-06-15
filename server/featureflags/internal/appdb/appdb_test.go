package appdb_test

import (
	"context"
	"errors"
	"testing"

	"github.com/pashagolub/pgxmock/v4"
	"github.com/specterops/bloodhound/server/featureflags/internal/appdb"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// expectedSelectSQL is compared via pgxmock.QueryMatcherEqual, which
// whitespace-normalises both sides; column order, table name, the WHERE
// predicate and the parameter shape are load-bearing.
const expectedSelectSQL = `SELECT id, created_at, updated_at, key, name, description, enabled, user_updatable FROM feature_flags WHERE key = $1 LIMIT $2`

func newTestStore(t *testing.T) (*appdb.Store, pgxmock.PgxPoolIface) {
	t.Helper()
	pool, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	require.NoError(t, err)
	t.Cleanup(pool.Close)
	return appdb.NewStore(pool), pool
}

func flagColumns() []string {
	return []string{"id", "created_at", "updated_at", "key", "name", "description", "enabled", "user_updatable"}
}

func TestStore_GetFlagByKey(t *testing.T) {
	var (
		ctx   = context.Background()
		dbErr = errors.New("connection refused")
	)

	tests := []struct {
		name         string
		expectations func(pool pgxmock.PgxPoolIface)
		wantFlag     services.FeatureFlag
		wantErr      error
	}{
		{
			name: "returns the feature flag on success",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectSQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnRows(
					pool.NewRows(flagColumns()).AddRow(
						int32(7), nil, nil, services.FeatureOpenHoundSupport, "OpenHound Support", "desc", true, false,
					),
				)
			},
			wantFlag: services.FeatureFlag{ID: 7, Key: services.FeatureOpenHoundSupport, Name: "OpenHound Support", Description: "desc", Enabled: true},
		},
		{
			name: "maps zero rows to ErrNotFound",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectSQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnRows(
					pool.NewRows(flagColumns()),
				)
			},
			wantErr: services.ErrNotFound,
		},
		{
			name: "propagates other database errors",
			expectations: func(pool pgxmock.PgxPoolIface) {
				pool.ExpectQuery(expectedSelectSQL).WithArgs(services.FeatureOpenHoundSupport, 1).WillReturnError(dbErr)
			},
			wantErr: dbErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store, pool := newTestStore(t)
			tt.expectations(pool)

			flag, err := store.GetFlagByKey(ctx, services.FeatureOpenHoundSupport)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.wantFlag, flag)
			}
			require.NoError(t, pool.ExpectationsWereMet())
		})
	}
}
