// Copyright 2024 Specter Ops, Inc.
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

package auth

import (
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/database"
)

// DeleteSSOProvider deletes a sso_provider with the matching id
func (s ManagementResource) DeleteSSOProvider(response http.ResponseWriter, request *http.Request) {
	var (
		rawSSOProviderID = mux.Vars(request)[api.URIPathVariableSSOProviderID]
	)

	// Convert the incoming string url param to an int
	if ssoProviderID, err := strconv.Atoi(rawSSOProviderID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if err = s.db.DeleteSSOProvider(request.Context(), ssoProviderID); errors.Is(err, database.ErrNotFound) {
		// Handle error if requested record could not be found
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, err.Error(), request), response)
	} else if err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}
