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

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
)

const (
	queryParamSchemaEnvironmentKindID = "schema_environment_kind_id"
	queryParamSchemaExtensionID       = "schema_extension_id"
	filterCreatedAt                   = "created_at"
	queryParamStart                   = "start"
	queryParamEnd                     = "end"
	errorInvalidRFC3339               = "invalid RFC-3339 datetime format: %v"
)

type AggregationView struct {
	RunID             string    `json:"run_id"`
	ExtensionID       int32     `json:"extension_id"`
	EnvironmentKindID int32     `json:"environment_kind_id"`
	MetricType        string    `json:"metric_type"`
	MetricName        string    `json:"metric_name"`
	MetricValue       float64   `json:"metric_value"`
	KindID            *int32    `json:"kind_id"`
	ID                int32     `json:"id"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	// The legacy Serial model exposes sql.NullTime even though deleted rows are excluded.
	DeletedAt sql.NullTime `json:"deleted_at"`
}

// AggregationsView owns the complete time-windowed, paginated response envelope.
type AggregationsView struct {
	Start *time.Time        `json:"start,omitempty"`
	End   *time.Time        `json:"end,omitempty"`
	Count int               `json:"count"`
	Limit int               `json:"limit"`
	Skip  int               `json:"skip"`
	Data  []AggregationView `json:"data"`
}

func BuildAggregationsView(aggregations []services.Aggregation, start, end time.Time, count, skip, limit int) AggregationsView {
	var view = AggregationsView{Count: count, Limit: limit, Skip: skip}
	if !start.IsZero() {
		view.Start = &start
	}
	if !end.IsZero() {
		view.End = &end
	}
	for _, aggregation := range aggregations {
		view.Data = append(view.Data, AggregationView{
			ID: aggregation.ID, RunID: aggregation.RunID, ExtensionID: aggregation.SchemaExtensionID,
			EnvironmentKindID: aggregation.SchemaEnvironmentKindID, MetricType: aggregation.MetricType,
			MetricName: aggregation.MetricName, MetricValue: aggregation.MetricValue, KindID: aggregation.KindID,
			CreatedAt: aggregation.CreatedAt, UpdatedAt: aggregation.UpdatedAt,
		})
	}
	return view
}

func (s AggregationsView) IsSortable(field string) bool {
	return field == "created_at" || field == "updated_at"
}
func (s AggregationsView) JSONView() ([]byte, error) { return json.Marshal(s) }

func (s *Handlers) ListAggregations(response http.ResponseWriter, request *http.Request) {
	var (
		ctx                        = request.Context()
		queryParams                = request.URL.Query()
		bhCtx                      = bhctx.Get(ctx)
		defaultEnd, defaultStart   = params.DefaultTimeRange()
		queryFilters               = make(params.Filters)
		start, end                 time.Time
		rawSchemaEnvironmentKindID string
		schemaEnvironmentKindID    int64
		err                        error
	)
	if rawSchemaEnvironmentKindID = queryParams.Get(queryParamSchemaEnvironmentKindID); rawSchemaEnvironmentKindID == "" {
		responses.WriteError(ctx, http.StatusBadRequest, "missing required query parameter: schema_environment_kind_id", response)
		return
	}
	if schemaEnvironmentKindID, err = strconv.ParseInt(rawSchemaEnvironmentKindID, 10, 32); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "query parameter must be an integer: schema_environment_kind_id", response)
		return
	}

	queryFilters[queryParamSchemaEnvironmentKindID] = []params.Filter{{Field: queryParamSchemaEnvironmentKindID, Operator: params.Equals, Value: strconv.FormatInt(schemaEnvironmentKindID, 10)}}

	if start, err = params.ParseTimeQueryParameter(queryParams, queryParamStart, defaultStart); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, fmt.Sprintf(errorInvalidRFC3339, queryParams[queryParamStart]), response)
		return
	}

	if end, err = params.ParseTimeQueryParameter(queryParams, queryParamEnd, defaultEnd); err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, fmt.Sprintf(errorInvalidRFC3339, queryParams[queryParamEnd]), response)
		return
	}

	if rawExtensionID := queryParams.Get(queryParamSchemaExtensionID); rawExtensionID != "" {
		var extensionID int64
		if extensionID, err = strconv.ParseInt(rawExtensionID, 10, 32); err != nil {
			responses.WriteError(ctx, http.StatusBadRequest, "query parameter must be an integer: schema_extension_id", response)
			return
		}
		queryFilters[queryParamSchemaExtensionID] = []params.Filter{{Field: queryParamSchemaExtensionID, Operator: params.Equals, Value: strconv.FormatInt(extensionID, 10)}}
	}

	queryFilters[filterCreatedAt] = []params.Filter{
		{Field: filterCreatedAt, Operator: params.GreaterThanOrEquals, Value: start.Format(time.RFC3339Nano)},
		{Field: filterCreatedAt, Operator: params.LessThanOrEquals, Value: end.Format(time.RFC3339Nano)},
	}
	aggregations, count, err := s.service.ListAggregations(ctx, queryFilters, bhCtx.Sort, bhCtx.Skip, bhCtx.Limit)
	if err != nil {
		handleAggregationError(ctx, err, response)
		return
	}
	responses.WriteJSON(ctx, BuildAggregationsView(aggregations, start, end, count, bhCtx.Skip, bhCtx.Limit), http.StatusOK, response)
}

func handleAggregationError(ctx context.Context, err error, response http.ResponseWriter) {
	switch {
	case errors.Is(err, services.ErrExtensionNotFound):
		responses.WriteError(ctx, http.StatusNotFound, "resource not found", response)
	case errors.Is(err, services.ErrInvalidQuery):
		responses.WriteError(ctx, http.StatusBadRequest, "invalid aggregation query", response)
	default:
		responses.WriteInternalServerError(ctx, err, response)
	}
}
