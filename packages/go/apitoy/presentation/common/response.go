package common

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/apitoy/app"
	"github.com/specterops/bloodhound/src/api"
)

func WriteErrorResponse(ctx context.Context, request *http.Request, response http.ResponseWriter, err error) {
	status, msg := handleSensitiveError(err)
	api.WriteErrorResponse(ctx, api.BuildErrorResponse(status, msg.Error(), request), response)
}

func handleSensitiveError(err error) (int, error) {
	if errors.Is(err, app.ErrNotFound) {
		return http.StatusNotFound, fmt.Errorf(api.ErrorResponseDetailsResourceNotFound)
	} else if errors.Is(err, app.ErrFileValidation) || errors.Is(err, app.ErrInvalidJSON) {
		return http.StatusBadRequest, err
	} else if errors.Is(err, app.ErrGeneralApplicationFailure) {
		return http.StatusInternalServerError, err
	} else if errors.Is(err, app.ErrGenericDatabaseFailure) {
		log.Errorf("Database error occurred: %v", err)
		return http.StatusInternalServerError, fmt.Errorf(api.ErrorResponseDetailsInternalServerError)
	} else if err != nil {
		log.Errorf("Unknown error type occurred: %v", err)
		return http.StatusInternalServerError, fmt.Errorf(api.ErrorResponseDetailsInternalServerError)
	} else {
		log.Errorf("Nil error passed to handleSensitiveError, please fix")
		return http.StatusInternalServerError, fmt.Errorf(api.ErrorResponseDetailsInternalServerError)
	}
}
