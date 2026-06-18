package appcfg

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/appcfg/internal/appdb"
	"github.com/specterops/bloodhound/server/appcfg/internal/handlers"
	"github.com/specterops/bloodhound/server/appcfg/internal/routes"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
)

func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlers(svc)
	)

	routes.Register(routerInst, handlerSet)
}
