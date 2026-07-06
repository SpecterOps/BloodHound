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
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

const (
	ErrNoTenantId                string = "no tenant id specified in url"
	ErrNoPlatformId              string = "no platform id specified in url"
	ErrInvalidPlatformId         string = "invalid platform id specified in url: %v"
	ErrNoEnvironmentId           string = "environmentid is required"
	ErrEnvironmentIdDoesNotExist string = "environmentid does not exist"
	ErrUnknownUser               string = "unknown user"
	ErrNoAccess                  string = "user does not have permission to access this environment"
)

func (s Resources) GetDatabaseCompleteness(response http.ResponseWriter, request *http.Request) {
	defer measure.ContextMeasureWithThreshold(request.Context(), slog.LevelDebug, "Get Current Database Completeness")()

	result := make(map[string]float64)

	if err := s.Graph.ReadTransaction(request.Context(), func(tx graph.Transaction) error {
		if userSessionCompleteness, err := ad.FetchUserSessionCompleteness(tx); err != nil {
			return err
		} else {
			result["LocalGroupCompleteness"] = userSessionCompleteness
		}

		if localGroupCompleteness, err := ad.FetchLocalGroupCompleteness(tx); err != nil {
			return err
		} else {
			result["SessionCompleteness"] = localGroupCompleteness
		}

		return nil
	}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Error getting quality stat: %v", err), request), response)
	} else {
		api.WriteBasicResponse(request.Context(), result, http.StatusOK, response)
	}
}

func (s *Resources) GetADDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		adDataQualityStats       model.ADDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	if order, _, err := parseOrder(queryParams, adDataQualityStats); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if id, hasDomainID := mux.Vars(request)[api.URIPathVariableDomainID]; !hasDomainID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorNoDomainId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetADDataQualityStats(request.Context(), id, start, end, order, limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetAzureDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	if order, _, err := parseOrder(queryParams, azureDataQualityStats); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if id, hasTenantID := mux.Vars(request)[api.URIPathVariableTenantID]; !hasTenantID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoTenantId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetAzureDataQualityStats(request.Context(), id, start, end, order, limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetPlatformAggregateStats(response http.ResponseWriter, request *http.Request) {
	var (
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	// TODO: This is currently using only the Azure stat type, but should check against the appropriate aggregate type for the chosen platform
	// It's safe for now, but should be refactored
	if order, _, err := parseOrder(queryParams, azureDataQualityStats); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if id, hasPlatformID := mux.Vars(request)[api.URIPathVariablePlatformID]; !hasPlatformID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoPlatformId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else {
		var (
			stats any
			count int
		)

		switch id {
		case "ad":
			stats, count, err = s.DB.GetADDataQualityAggregations(request.Context(), start, end, order, limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		case "azure":
			stats, count, err = s.DB.GetAzureDataQualityAggregations(request.Context(), start, end, order, limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(ErrInvalidPlatformId, id), request), response)
			return
		}

		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		ctx                      = request.Context()
		dataQualityStats         model.DataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	if environmentId := queryParams.Get(graphschema.EnvironmentIDKey); environmentId == "" {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, ErrNoEnvironmentId, request), response)
	} else if user, found := auth.GetUserFromAuthCtx(bhctx.FromRequest(request).AuthCtx); !found {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, ErrUnknownUser, request), response)
	} else if ShouldFilterForETAC(s.DogTags, user) {
		hasAccess, err := CheckUserAccessToEnvironments(ctx, s.DB, user, environmentId)
		if err != nil {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		} else if !hasAccess {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusForbidden, ErrNoAccess, request), response)
		}
	} else if environmentExists, err := s.environmentIdExists(ctx, environmentId); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else if !environmentExists {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusNotFound, ErrEnvironmentIdDoesNotExist, request), response)
	} else if _, sort, err := parseOrder(queryParams, dataQualityStats); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else {
		filters := model.Filters{
			"environment_id": []model.Filter{{
				Value:       environmentId,
				Operator:    model.Equals,
				SetOperator: model.FilterAnd,
			}},
			"created_at": []model.Filter{{Value: start.Format(time.RFC3339), Operator: model.GreaterThanOrEquals, SetOperator: model.FilterAnd}, {Value: end.Format(time.RFC3339), Operator: model.LessThanOrEquals, SetOperator: model.FilterAnd}},
		}
		if stats, count, err := s.DB.GetDataQualityStats(ctx, filters, sort, skip, limit); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			api.WriteResponseWrapperWithTimeWindowAndPagination(ctx, stats, start, end, limit, skip, count, http.StatusOK, response)
		}

	}
}

func (s *Resources) environmentIdExists(ctx context.Context, environmentID string) (bool, error) {
	var exists bool

	err := s.Graph.ReadTransaction(ctx, func(tx graph.Transaction) error {
		count, err := tx.Nodes().Filterf(func() graph.Criteria {
			return query.Equals(
				query.NodeProperty(graphschema.EnvironmentIDKey),
				environmentID,
			)
		}).Count()
		if err != nil {
			return err
		}

		exists = count > 0
		return nil
	})

	return exists, err
}

type Sortable interface {
	IsSortable(column string) bool
}

// parseOrder is a helper function which parses any sort_by query params into both the legacy sort string format and the model.Sort format. Returns an error if the columns is not sortable.
func parseOrder(queryParams url.Values, sortable Sortable) (string, model.Sort, error) {
	order := []string{}
	sort := model.Sort{}
	for _, column := range queryParams[api.QueryParameterSortBy] {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !sortable.IsSortable(column) {
			return strings.Join(order, ", "), sort, errors.New(api.ErrorResponseDetailsNotSortable)
		}

		if descending {
			order = append(order, column+" desc")
			sort = append(sort, model.SortItem{Column: column, Direction: model.DescendingSortDirection})

		} else {
			order = append(order, column)
			sort = append(sort, model.SortItem{Column: column, Direction: model.AscendingSortDirection})
		}
	}
	return strings.Join(order, ", "), sort, nil
}
