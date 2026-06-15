// Package featureflags is a self-contained feature-flag library. It owns the
// feature-flag domain (the FeatureFlag type, the Database port and the Service
// in service.go), the PostgreSQL adapter (Store in sql.go) and the Register
// entry point that wires them together so callers obtain a ready-to-use service
// without reaching into the storage layer.
package featureflags

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/featureflags/internal/appdb"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

const (
	FeatureOpenHoundSupport = services.FeatureOpenHoundSupport
	FeatureAlerts           = services.FeatureAlerts
)

type Service interface {
	IsEnabled(ctx context.Context, key string) (bool, error)
}

// Register wires the feature-flag service to its PostgreSQL store and returns
// the constructed service for use by BHE feature slices.
func Register(pool *pgxpool.Pool) Service {
	return services.NewService(appdb.NewStore(pool))
}
