package apiclient

import (
	"fmt"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"net/http"
)

func (s Client) GetFeatureFlags() ([]appcfg.FeatureFlag, error) {
	var featureFlags []appcfg.FeatureFlag

	if response, err := s.Request(http.MethodGet, "/api/v2/features", nil, nil); err != nil {
		return nil, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return nil, ReadAPIError(response)
		}

		return featureFlags, api.ReadAPIV2ResponsePayload(&featureFlags, response)
	}
}

func (s Client) GetFeatureFlag(key string) (appcfg.FeatureFlag, error) {
	if flags, err := s.GetFeatureFlags(); err != nil {
		return appcfg.FeatureFlag{}, err
	} else {
		for _, flag := range flags {
			if flag.Key == key {
				return flag, nil
			}
		}
	}

	return appcfg.FeatureFlag{}, fmt.Errorf("flag with key %s not found", key)
}

func (s Client) ToggleFeatureFlag(key string) error {
	var result appcfg.Parameter

	if flag, err := s.GetFeatureFlag(key); err != nil {
		return err
	} else if response, err := s.Request(http.MethodPut, fmt.Sprintf("/api/v2/features/%d/toggle", flag.ID), nil, nil); err != nil {
		return err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return ReadAPIError(response)
		}

		return api.ReadAPIV2ResponsePayload(&result, response)
	}
}
