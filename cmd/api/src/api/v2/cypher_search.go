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

package v2

import (
	"net/http"

	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/queries"
)

type CypherSearch struct {
	Query             string `json:"query"`
	IncludeProperties bool   `json:"include_properties,omitempty"`
}

func (s Resources) CypherSearch(response http.ResponseWriter, request *http.Request) {
	var payload CypherSearch

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed.", request), response,
		)
	} else if graphResponse, err := s.GraphQuery.RawCypherSearch(request.Context(), payload.Query, payload.IncludeProperties); err != nil {
		if queries.IsQueryError(err) {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response,
			)
		} else if util.IsNeoTimeoutError(err) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "transaction timed out, reduce query complexity or try again later", request), response)
		} else {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response,
			)
		}
	} else {
		api.WriteBasicResponse(request.Context(), graphResponse, http.StatusOK, response)
	}

}
