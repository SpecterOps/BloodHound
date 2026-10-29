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
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/packages/go/responses"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
)

type DataQuality interface {
	ListAggregations(ctx context.Context, query services.AggregationQuery) ([]services.Aggregation, int, error)
}

type Handlers struct{ service DataQuality }

func NewHandlersContainer(service DataQuality) *Handlers { return &Handlers{service: service} }

func (s *Handlers) ListAggregations(response http.ResponseWriter, request *http.Request) {
	var (
		ctx               = request.Context()
		queryParameters   = request.URL.Query()
		bloodhoundContext = bhctx.Get(ctx)
		end               = time.Now()
		query             = services.AggregationQuery{
			Start: end.Add(-30 * 24 * time.Hour), End: end,
			Sort: bloodhoundContext.Sort, Skip: bloodhoundContext.Skip, Limit: bloodhoundContext.Limit,
		}
	)
	rawKindID := queryParameters.Get("schema_environment_kind_id")
	if rawKindID == "" {
		responses.WriteError(ctx, http.StatusBadRequest, "missing required query parameter: schema_environment_kind_id", response)
		return
	}
	kindID, err := strconv.ParseInt(rawKindID, 10, 32)
	if err != nil {
		responses.WriteError(ctx, http.StatusBadRequest, "query parameter must be an integer: schema_environment_kind_id", response)
		return
	}
	query.EnvironmentKindID = int32(kindID)
	for _, parameter := range []struct {
		name   string
		target *time.Time
	}{{"start", &query.Start}, {"end", &query.End}} {
		if raw := queryParameters.Get(parameter.name); raw != "" {
			parsed, err := time.Parse(time.RFC3339Nano, raw)
			if err != nil {
				responses.WriteError(ctx, http.StatusBadRequest, fmt.Sprintf("invalid RFC-3339 datetime format: %s", queryParameters[parameter.name]), response)
				return
			}
			*parameter.target = parsed
		}
	}
	if raw := queryParameters.Get("schema_extension_id"); raw != "" {
		extensionID, err := strconv.ParseInt(raw, 10, 32)
		if err != nil {
			responses.WriteError(ctx, http.StatusBadRequest, "query parameter must be an integer: schema_extension_id", response)
			return
		}
		query.ExtensionID = new(int32(extensionID))
	}
	aggregations, count, err := s.service.ListAggregations(ctx, query)
	if err != nil {
		handleAggregationError(ctx, err, response)
		return
	}
	responses.WriteJSON(ctx, BuildAggregationsView(aggregations, query, count), http.StatusOK, response)
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
