package identity

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/identity/internal/appdb"
	"github.com/specterops/bloodhound/server/identity/internal/handlers"
	"github.com/specterops/bloodhound/server/identity/internal/routes"
	"github.com/specterops/bloodhound/server/identity/internal/services"
)

// Register builds the analysis store -> service -> handler chain and attaches
// the analysis routes to the provided router. It is called from the modules
// registry and receives only the infrastructure it directly needs.
func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	routes.Register(routerInst, handlerSet)
}
