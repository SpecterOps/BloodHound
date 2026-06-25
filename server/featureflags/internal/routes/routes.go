package routes

import (
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/server/featureflags/internal/handlers"
)

func Register(routerInst *router.Router, handlers *handlers.Handlers) {
	var permissions = auth.Permissions()

	routerInst.GET("/api/v2/features", handlers.GetAllFlags).RequirePermissions(permissions.AppReadApplicationConfiguration)
	routerInst.PUT("/api/v2/features/{feature_id}/toggle", handlers.ToggleFlag).RequirePermissions(permissions.AppWriteApplicationConfiguration)
}
