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

	"github.com/specterops/bloodhound/src/api"
)

func (s Client) Version() (map[string]any, error) {
	version := make(map[string]any)
	if response, err := s.Request(http.MethodGet, "api/version", nil, nil); err != nil {
		return version, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return version, ReadAPIError(response)
		}

		return version, api.ReadAPIV2ResponsePayload(&version, response)
	}
}
