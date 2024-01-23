// Copyright 2023 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package apiclient

import (
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model/appcfg"
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
