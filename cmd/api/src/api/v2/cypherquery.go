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
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/queries"
	"net/http"
)

const (
	errUnauthorizedGraphMutation = errors.Error("unauthorized graph mutation")
)

type CypherQueryPayload struct {
	Query             string `json:"query"`
	IncludeProperties bool   `json:"include_properties,omitempty"`
}

func (s Resources) CypherQuery(response http.ResponseWriter, request *http.Request) {
	var (
		payload       CypherQueryPayload
		preparedQuery queries.PreparedQuery
		graphResponse model.UnifiedGraph
		err           error
	)

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed.", request), response)
		return
	}

	if preparedQuery, err = s.GraphQuery.PrepareCypherQuery(payload.Query); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	if preparedQuery.HasMutation {
		graphResponse, err = s.cypherMutation(request, preparedQuery, payload.IncludeProperties)
	} else {
		graphResponse, err = s.GraphQuery.RawCypherQuery(request.Context(), preparedQuery, payload.IncludeProperties)
	}

	if err != nil {
		if errors.Is(err, errUnauthorizedGraphMutation) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "Permission denied: User may not modify the graph.", request), response)
		} else if util.IsNeoTimeoutError(err) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "transaction timed out, reduce query complexity or try again later", request), response)
		} else {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		}
	} else {
		api.WriteBasicResponse(request.Context(), graphResponse, http.StatusOK, response)
	}
}

func (s Resources) cypherMutation(request *http.Request, preparedQuery queries.PreparedQuery, includeProperties bool) (model.UnifiedGraph, error) {
	var (
		auditLogEntry model.AuditEntry
		graphResponse model.UnifiedGraph
		err           error
	)

	if !s.Authorizer.AllowsPermission(ctx.FromRequest(request).AuthCtx, auth.Permissions().GraphDBMutate) {
		s.Authorizer.AuditLogUnauthorizedAccess(request)
		return model.UnifiedGraph{}, errUnauthorizedGraphMutation
	}

	// All mutation attempts must be audit logged even when failed
	if auditLogEntry, err = model.NewAuditEntry(model.AuditLogActionMutateGraph, model.AuditLogStatusIntent, model.AuditData{"query": preparedQuery.StrippedQuery}); err != nil {
		return model.UnifiedGraph{}, err
	}

	// create an intent audit log
	if err = s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
		return model.UnifiedGraph{}, err
	}

	if graphResponse, err = s.GraphQuery.RawCypherQuery(request.Context(), preparedQuery, includeProperties); err != nil {
		auditLogEntry.Status = model.AuditLogStatusFailure
	} else {
		auditLogEntry.Status = model.AuditLogStatusSuccess
	}

	if err := s.DB.AppendAuditLog(request.Context(), auditLogEntry); err != nil {
		// We want to keep err scoped because having info on the mutation graph response trumps this error
		log.Errorf("failure to create mutation audit log %s", err.Error())
	}

	return graphResponse, err

}
