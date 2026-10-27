// Copyright 2026 Specter Ops, Inc.
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

package handlers

//go:generate go tool mockery

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/audit/internal/services"
)

const (
	beforeQueryParameter = "before"
	afterQueryParameter  = "after"
	offsetQueryParameter = "offset"
	defaultLimit         = 1000
)

// Audit defines the service boundary required by the audit handlers.
type Audit interface {
	ListAuditLogs(ctx context.Context, query services.ListQuery) ([]services.Log, int, error)
}

// Handlers contains audit HTTP dependencies.
type Handlers struct {
	audit Audit
}

// NewHandlersContainer initializes the audit handlers.
func NewHandlersContainer(audit Audit) *Handlers {
	return &Handlers{audit: audit}
}

// ListAuditLogs returns a globally filtered, sorted, and paginated view of old and new audit records.
func (s *Handlers) ListAuditLogs(response http.ResponseWriter, request *http.Request) {
	var (
		ctx         = request.Context()
		requestCtx  = bhctx.Get(ctx)
		queryValues = request.URL.Query()
		before      time.Time
		after       time.Time
		err         error
	)

	if before, err = parseTimeQueryParameter(queryValues.Get(beforeQueryParameter), time.Now()); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("invalid %s query parameter: %s", beforeQueryParameter, err), request), response)
		return
	}
	if after, err = parseTimeQueryParameter(queryValues.Get(afterQueryParameter), before.Add(-365*24*time.Hour)); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("invalid %s query parameter: %s", afterQueryParameter, err), request), response)
		return
	}
	if after.After(before) {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("query parameter %q must not be later than %q", afterQueryParameter, beforeQueryParameter), request), response)
		return
	}

	skip := requestCtx.Skip
	if !queryValues.Has("skip") && queryValues.Has(offsetQueryParameter) {
		paging, pagingErr := params.NewPagingParser(params.PagingConfig{DefaultLimit: defaultLimit, SkipKey: offsetQueryParameter}).ParseAndValidate(queryValues)
		if pagingErr != nil {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, pagingErr.Error(), request), response)
			return
		}
		skip = paging.Skip
	}

	sortItems := requestCtx.Sort
	if len(sortItems) == 0 {
		sortItems = params.SortItems{{Field: "created_at", Direction: params.Descending}}
	}

	query := services.ListQuery{
		Before:  before,
		After:   after,
		Skip:    skip,
		Limit:   requestCtx.Limit,
		Filters: requestCtx.Filters,
		Sort:    sortItems,
	}
	logs, count, err := s.audit.ListAuditLogs(ctx, query)
	if err != nil {
		responses.WriteInternalServerError(ctx, err, response)
		return
	}

	responses.WritePaginated(ctx, buildLogListView(logs), query.Limit, query.Skip, count, http.StatusOK, response)
}

func parseTimeQueryParameter(value string, defaultValue time.Time) (time.Time, error) {
	if value == "" {
		return defaultValue, nil
	}

	return time.Parse(time.RFC3339Nano, value)
}
