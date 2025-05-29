package client

import (
	"errors"
	"github.com/specterops/bloodhound/src/api"
	"net/http"
)

func ReadAPIError(response *http.Response) error {
	if response.StatusCode == http.StatusNotFound {
		return errors.New("API returned a 404 error")
	}

	var apiError api.ErrorWrapper

	if err := api.ReadAPIV2ErrorResponsePayload(&apiError, response); err != nil {
		return err
	}

	return apiError
}
