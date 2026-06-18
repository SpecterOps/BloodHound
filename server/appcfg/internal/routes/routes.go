package routes

import (
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/appcfg/internal/handlers"
)

func Register(routerInst *router.Router, handlers *handlers.Handlers) {
	routerInst.GET("/api/v2/datapipe/status", handlers.GetDatapipeStatus).RequireAuth()
}
