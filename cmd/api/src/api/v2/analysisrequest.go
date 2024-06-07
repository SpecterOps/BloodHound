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

package v2

import (
	"database/sql"
	"net/http"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
)

func (s Resources) GetAnalysisRequest(response http.ResponseWriter, request *http.Request) {
	if analRequest, err := s.DB.GetAnalysisRequest(request.Context()); err != nil && !errors.Is(err, sql.ErrNoRows) {
		api.HandleDatabaseError(request, response, err)
	} else if errors.Is(err, sql.ErrNoRows) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, api.ErrorResponseDetailsResourceNotFound, request), response)
	} else {
		api.WriteBasicResponse(request.Context(), analRequest, http.StatusOK, response)
	}
}

func (s Resources) RequestAnalysis(response http.ResponseWriter, request *http.Request) {
	defer log.Measure(log.LevelDebug, "Requesting analysis")()

	var userId string
	if user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
		log.Warnf("encountered request analysis for unknown user, this shouldn't happen")
		userId = "unknown-user"
	} else {
		userId = user.ID.String()
	}

	if err := s.DB.RequestAnalysis(request.Context(), userId); err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	response.WriteHeader(http.StatusAccepted)
}
