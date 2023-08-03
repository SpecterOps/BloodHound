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
	"net/url"

	"github.com/specterops/bloodhound/src/api"
)

func (s Client) GetUserSessionsList(objectID string) ([]string, error) {
	var sessionList []string

	params := url.Values{
		"type": {
			"list",
		},
	}

	if response, err := s.Request(http.MethodGet, fmt.Sprintf("/api/v2/entities/users/%s/sessions", objectID), params, nil); err != nil {
		return nil, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return nil, ReadAPIError(response)
		}

		return sessionList, api.ReadAPIV2ResponsePayload(&sessionList, response)
	}
}

func (s Client) GetUserSessionsGraph(objectID string) (map[string]any, error) {
	var sessionGraph map[string]any

	params := url.Values{
		"type": {
			"graph",
		},
	}

	if response, err := s.Request(http.MethodGet, fmt.Sprintf("/api/v2/entities/users/%s/sessions", objectID), params, nil); err != nil {
		return nil, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return nil, ReadAPIError(response)
		}

		return sessionGraph, api.ReadAPIV2ResponsePayload(&sessionGraph, response)
	}
}
