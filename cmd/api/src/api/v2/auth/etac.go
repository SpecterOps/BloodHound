package auth

import (
	"github.com/specterops/bloodhound/cmd/api/src/api"
	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// handleETACRequest will modify the user passed in to assign an etac list or grant all environment access
// and will return an error on bad requests
func handleETACRequest(etacRequest v2.UpdateUserETACListRequest, roles model.Roles, user *model.User, ) error {
	// Administrators and Power Users may not have an ETAC list applied to them
	if roles.Has(model.Role{Name: auth.RoleAdministrator}) || roles.Has(model.Role{Name: auth.RolePowerUser}) {
		return errors.New(api.ErrorResponseETACInvalidRoles)
	}

	// The user may not request all environments and have an ETAC list applied to them
	if len(etacRequest.Environments) != 0 && etacRequest.AllEnvironments {
		return errors.New(api.ErrorResponseETACBadRequest)
	}

	user.AllEnvironments = etacRequest.AllEnvironments

	if etacRequest.AllEnvironments {
		user.EnvironmentAccessControl = make([]model.EnvironmentAccess, 0)
	} else {
		environments := make([]model.EnvironmentAccess, 0, len(etacRequest.Environments))
		for _, environment := range etacRequest.Environments {
			environments = append(environments, model.EnvironmentAccess{
				UserID:      user.ID.String(),
				Environment: environment,
			})
		}
		user.EnvironmentAccessControl = environments
	}

	return nil
}
