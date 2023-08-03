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
	v2 "github.com/specterops/bloodhound/src/api/v2"
	"github.com/specterops/bloodhound/src/model"
)

func (s Client) CypherSearch(request v2.CypherSearch) (model.UnifiedGraph, error) {
	var graphResponse model.UnifiedGraph

	if response, err := s.Request(http.MethodPost, "api/v2/graphs/cypher", nil, request); err != nil {
		return graphResponse, err
	} else {
		defer response.Body.Close()

		if api.IsErrorResponse(response) {
			return graphResponse, ReadAPIError(response)
		}

		return graphResponse, api.ReadAPIV2ResponsePayload(&graphResponse, response)
	}
}
