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
	"net/http"
	"net/url"

	"github.com/specterops/bloodhound/src/api"
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

func (s Client) GetAppConfigs() (appcfg.Parameters, error) {
	if response, err := s.Request(http.MethodGet, "/api/v2/config", nil, nil); err != nil {
		return nil, err
	} else {
		defer response.Body.Close()

		var parameters appcfg.Parameters

		if api.IsErrorResponse(response) {
			return nil, ReadAPIError(response)
		}

		return parameters, api.ReadAPIV2ResponsePayload(&parameters, response)
	}
}

func (s Client) GetAppConfig(parameterKey string) (appcfg.Parameters, error) {
	params := url.Values{"parameter": []string{"eq:" + parameterKey}}

	if response, err := s.Request(http.MethodGet, "/api/v2/config", params, nil); err != nil {
		return nil, err
	} else {
		defer response.Body.Close()

		var parameters appcfg.Parameters

		if api.IsErrorResponse(response) {
			return nil, ReadAPIError(response)
		}

		return parameters, api.ReadAPIV2ResponsePayload(&parameters, response)
	}
}

func (s Client) PutAppConfig(parameter v2.AppConfigUpdateRequest) (appcfg.Parameter, error) {
	var result appcfg.Parameter

	if response, err := s.Request(http.MethodPut, "/api/v2/config", nil, parameter); err != nil {
		return result, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return result, ReadAPIError(response)
		}

		return result, api.ReadAPIV2ResponsePayload(&result, response)
	}
}
