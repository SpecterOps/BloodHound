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
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
	azureSchema "github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

const (
	ErrNoTenantId        string = "no tenant id specified in url"
	ErrNoPlatformId      string = "no platform id specified in url"
	ErrInvalidPlatformId string = "invalid platform id specified in url: %v"

	QueryParameterEnvironmentKind string = "environment_kind"
	QueryParameterEnvironmentID   string = "environment_id"
	QueryParameterSourceKind      string = "source_kind"
	QueryParameterIncludeBuiltin  string = "include_builtin"
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

type dataQualityKindReader interface {
	GetKindsByIDs(ctx context.Context, ids ...int32) ([]model.Kind, error)
}

type schemaEnvironmentKey struct {
	schemaExtensionID       int32
	schemaEnvironmentKindID int32
}

func dataQualityEnvironmentType(environmentKind string) string {
	switch environmentKind {
	case adSchema.Domain.String():
		return "active-directory"
	case azureSchema.Tenant.String():
		return "azure"
	default:
		return environmentKind
	}
}

func dataQualityEnvironmentCollected(node *graph.Node) bool {
	if !node.Properties.Exists(common.Collected.String()) {
		return false
	}

	collected, _ := node.Properties.Get(common.Collected.String()).Bool()
	if node.Kinds.ContainsOneOf(azureSchema.Tenant, adSchema.Domain) {
		return collected
	}

	return true
}

func dataQualitySourceKinds(ctx context.Context, db dataQualityKindReader, schemaEnvironments []model.SchemaEnvironment) (map[int32]model.Kind, error) {
	var sourceKindIDs []int32

	for _, schemaEnvironment := range schemaEnvironments {
		sourceKindIDs = append(sourceKindIDs, schemaEnvironment.SourceKindId)
	}

	kinds, err := db.GetKindsByIDs(ctx, sourceKindIDs...)
	if err != nil {
		return nil, err
	}

	sourceKindByID := make(map[int32]model.Kind, len(kinds))
	for _, kind := range kinds {
		sourceKindByID[kind.ID] = kind
	}

	return sourceKindByID, nil
}

func dataQualitySchemaEnvironmentByKind(schemaEnvironments []model.SchemaEnvironment, sourceKindByID map[int32]model.Kind) map[string]model.SchemaEnvironment {
	environmentByKind := make(map[string]model.SchemaEnvironment)

	for _, schemaEnvironment := range schemaEnvironments {
		if _, found := sourceKindByID[schemaEnvironment.SourceKindId]; !found {
			continue
		}

		existingEnvironment, found := environmentByKind[schemaEnvironment.EnvironmentKindName]
		if !found || (!existingEnvironment.IsBuiltin && schemaEnvironment.IsBuiltin) {
			environmentByKind[schemaEnvironment.EnvironmentKindName] = schemaEnvironment
		}
	}

	return environmentByKind
}

func BuildDataQualityEnvironmentSelectors(nodes []*graph.Node, schemaEnvironments []model.SchemaEnvironment, sourceKindByID map[int32]model.Kind) model.DataQualityEnvironmentSelectors {
	var (
		environmentByKind = dataQualitySchemaEnvironmentByKind(schemaEnvironments, sourceKindByID)
		selectors         = make(model.DataQualityEnvironmentSelectors, 0, len(nodes))
	)

	for _, node := range nodes {
		for environmentKind, schemaEnvironment := range environmentByKind {
			if !node.Kinds.ContainsOneOf(graph.StringKind(environmentKind)) {
				continue
			}

			sourceKind := sourceKindByID[schemaEnvironment.SourceKindId]
			name, _ := node.Properties.GetOrDefault(common.Name.String(), graphschema.DefaultMissingName).String()
			objectID, _ := node.Properties.GetOrDefault(common.ObjectID.String(), graphschema.DefaultMissingObjectId).String()

			selectors = append(selectors, model.DataQualityEnvironmentSelector{
				Type:              dataQualityEnvironmentType(environmentKind),
				Name:              name,
				ObjectID:          objectID,
				Collected:         dataQualityEnvironmentCollected(node),
				IsBuiltin:         schemaEnvironment.IsBuiltin,
				EnvironmentKindID: schemaEnvironment.EnvironmentKindId,
				EnvironmentKind:   environmentKind,
				SourceKindID:      sourceKind.ID,
				SourceKind:        sourceKind.Name,
			})
			break
		}
	}

	return selectors
}

func collectedDataQualityEnvironmentIDs(ctx context.Context, graphDB graph.Database, environmentKind string) ([]string, error) {
	var environmentIDs []string

	err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Kind(query.Node(), graph.StringKind(environmentKind))
		}))
		if err != nil {
			return err
		}

		for _, node := range nodes {
			if !dataQualityEnvironmentCollected(node) {
				continue
			}

			if environmentID, err := node.Properties.Get(common.ObjectID.String()).String(); err == nil {
				environmentIDs = append(environmentIDs, environmentID)
			}
		}

		return nil
	})

	return environmentIDs, err
}

func dataQualityFilters(start string, end string, environmentIDs []string, schemaEnvironments []model.SchemaEnvironment) model.Filters {
	filters := model.Filters{
		"metric_type": []model.Filter{{
			Operator:     model.Equals,
			Value:        string(model.DataQualityMetricTypeNode),
			IsStringData: true,
		}},
		"created_at": []model.Filter{
			{Operator: model.GreaterThanOrEquals, Value: start, IsStringData: true},
			{Operator: model.LessThanOrEquals, Value: end, IsStringData: true},
		},
	}

	for _, environmentID := range environmentIDs {
		filters["environment_id"] = append(filters["environment_id"], model.Filter{
			Operator:     model.Equals,
			Value:        environmentID,
			IsStringData: true,
			SetOperator:  model.FilterOr,
		})
	}

	for _, schemaEnvironment := range schemaEnvironments {
		filters["schema_extension_id"] = append(filters["schema_extension_id"], model.Filter{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(schemaEnvironment.SchemaExtensionId), 10),
			SetOperator: model.FilterOr,
		})
		filters["schema_environment_kind_id"] = append(filters["schema_environment_kind_id"], model.Filter{
			Operator:    model.Equals,
			Value:       strconv.FormatInt(int64(schemaEnvironment.EnvironmentKindId), 10),
			SetOperator: model.FilterOr,
		})
	}

	return filters
}

func matchingDataQualitySchemaEnvironments(schemaEnvironments []model.SchemaEnvironment, sourceKindByID map[int32]model.Kind, environmentKind string, sourceKind string, includeBuiltin bool) []model.SchemaEnvironment {
	var matches []model.SchemaEnvironment

	for _, schemaEnvironment := range schemaEnvironments {
		if schemaEnvironment.EnvironmentKindName != environmentKind {
			continue
		}

		if !includeBuiltin && schemaEnvironment.IsBuiltin {
			continue
		}

		if sourceKind != "" {
			if kind, found := sourceKindByID[schemaEnvironment.SourceKindId]; !found || kind.Name != sourceKind {
				continue
			}
		}

		matches = append(matches, schemaEnvironment)
	}

	return matches
}

func enrichDataQualityNodeKindStats(stats model.DataQualityStats, schemaEnvironments []model.SchemaEnvironment, sourceKindByID map[int32]model.Kind, kindByID map[int32]model.Kind) model.DataQualityNodeKindStats {
	var (
		enriched               = make(model.DataQualityNodeKindStats, 0, len(stats))
		schemaEnvironmentByKey = make(map[schemaEnvironmentKey]model.SchemaEnvironment, len(schemaEnvironments))
	)

	for _, schemaEnvironment := range schemaEnvironments {
		schemaEnvironmentByKey[schemaEnvironmentKey{
			schemaExtensionID:       schemaEnvironment.SchemaExtensionId,
			schemaEnvironmentKindID: schemaEnvironment.EnvironmentKindId,
		}] = schemaEnvironment
	}

	for _, stat := range stats {
		var (
			kindName          = stat.MetricName
			sourceKindName    string
			schemaEnvironment = schemaEnvironmentByKey[schemaEnvironmentKey{
				schemaExtensionID:       stat.SchemaExtensionID,
				schemaEnvironmentKindID: stat.SchemaEnvironmentKindID,
			}]
		)

		if stat.KindID.Valid {
			if kind, found := kindByID[stat.KindID.Int32]; found {
				kindName = kind.Name
			}
		}

		if sourceKind, found := sourceKindByID[schemaEnvironment.SourceKindId]; found {
			sourceKindName = sourceKind.Name
		}

		enriched = append(enriched, model.DataQualityNodeKindStat{
			Serial:                  stat.Serial,
			RunID:                   stat.RunID,
			IsBuiltin:               schemaEnvironment.IsBuiltin,
			SchemaEnvironmentKindID: stat.SchemaEnvironmentKindID,
			EnvironmentKind:         schemaEnvironment.EnvironmentKindName,
			SourceKind:              sourceKindName,
			EnvironmentID:           stat.EnvironmentID,
			MetricType:              stat.MetricType,
			MetricName:              stat.MetricName,
			MetricValue:             stat.MetricValue,
			KindID:                  stat.KindID,
			KindName:                kindName,
		})
	}

	return enriched
}

func (s *Resources) ListDataQualityEnvironments(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	sortItems, err := api.ParseGraphSortParameters(model.DataQualityEnvironmentSelectors{}, request.URL.Query())
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
		return
	}

	schemaEnvironments, err := s.DB.GetEnvironments(ctx)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	} else if len(schemaEnvironments) == 0 {
		api.WriteBasicResponse(ctx, model.DataQualityEnvironmentSelectors{}, http.StatusOK, response)
		return
	}

	sourceKindByID, err := dataQualitySourceKinds(ctx, s.DB, schemaEnvironments)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	environmentKinds := make(graph.Kinds, 0, len(schemaEnvironments))
	for _, schemaEnvironment := range schemaEnvironments {
		environmentKinds = environmentKinds.Add(graph.StringKind(schemaEnvironment.EnvironmentKindName))
	}

	nodes, err := s.GraphQuery.GetFilteredAndSortedNodes(sortItems, query.KindIn(query.Node(), environmentKinds...))
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	api.WriteBasicResponse(ctx, BuildDataQualityEnvironmentSelectors(nodes, schemaEnvironments, sourceKindByID), http.StatusOK, response)
}

func (s *Resources) GetDataQualityNodeKindStats(response http.ResponseWriter, request *http.Request) {
	var (
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	environmentKind := queryParams.Get(QueryParameterEnvironmentKind)
	if environmentKind == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "no environment kind specified", request), response)
		return
	}

	includeBuiltin := true
	if rawIncludeBuiltin := queryParams.Get(QueryParameterIncludeBuiltin); rawIncludeBuiltin != "" {
		if parsedIncludeBuiltin, err := strconv.ParseBool(rawIncludeBuiltin); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("invalid include_builtin specified: %s", rawIncludeBuiltin), request), response)
			return
		} else {
			includeBuiltin = parsedIncludeBuiltin
		}
	}

	if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if sort, err := api.ParseSortParameters(model.DataQualityStats{}, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
	} else if schemaEnvironments, err := s.DB.GetEnvironments(request.Context()); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if sourceKindByID, err := dataQualitySourceKinds(request.Context(), s.DB, schemaEnvironments); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		var (
			environmentIDs = []string{}
			sourceKind     = queryParams.Get(QueryParameterSourceKind)
		)

		if environmentID := queryParams.Get(QueryParameterEnvironmentID); environmentID != "" {
			environmentIDs = append(environmentIDs, environmentID)
		} else if collectedEnvironmentIDs, err := collectedDataQualityEnvironmentIDs(request.Context(), s.Graph, environmentKind); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
			return
		} else {
			environmentIDs = collectedEnvironmentIDs
		}

		if len(environmentIDs) == 0 {
			api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), model.DataQualityNodeKindStats{}, start, end, limit, skip, 0, http.StatusOK, response)
			return
		}

		matchingSchemaEnvironments := matchingDataQualitySchemaEnvironments(schemaEnvironments, sourceKindByID, environmentKind, sourceKind, includeBuiltin)
		if len(matchingSchemaEnvironments) == 0 {
			api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), model.DataQualityNodeKindStats{}, start, end, limit, skip, 0, http.StatusOK, response)
			return
		}

		stats, count, err := s.DB.GetDataQualityStats(
			request.Context(),
			dataQualityFilters(start.Format(time.RFC3339Nano), end.Format(time.RFC3339Nano), environmentIDs, matchingSchemaEnvironments),
			sort,
			skip,
			limit,
		)
		if err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		}

		var kindIDs []int32
		for _, stat := range stats {
			if stat.KindID.Valid {
				kindIDs = append(kindIDs, stat.KindID.Int32)
			}
		}

		kindByID := map[int32]model.Kind{}
		if len(kindIDs) > 0 {
			if kinds, err := s.DB.GetKindsByIDs(request.Context(), kindIDs...); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			} else {
				for _, kind := range kinds {
					kindByID[kind.ID] = kind
				}
			}
		}

		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), enrichDataQualityNodeKindStats(stats, schemaEnvironments, sourceKindByID, kindByID), start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetADDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		order                    []string
		adDataQualityStats       model.ADDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	for _, column := range queryParams[api.QueryParameterSortBy] {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !adDataQualityStats.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	if id, hasDomainID := mux.Vars(request)[api.URIPathVariableDomainID]; !hasDomainID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorNoDomainId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetADDataQualityStats(request.Context(), id, start, end, strings.Join(order, ", "), limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetAzureDataQualityStats(response http.ResponseWriter, request *http.Request) {
	var (
		order                    []string
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	for _, column := range queryParams[api.QueryParameterSortBy] {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !azureDataQualityStats.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	if id, hasTenantID := mux.Vars(request)[api.URIPathVariableTenantID]; !hasTenantID {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, ErrNoTenantId, request), response)
	} else if start, err := ParseTimeQueryParameter(queryParams, "start", defaultStart); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["start"]), request), response)
	} else if end, err := ParseTimeQueryParameter(queryParams, "end", defaultEnd); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(api.ErrorInvalidRFC3339, queryParams["end"]), request), response)
	} else if limit, err := ParseLimitQueryParameter(queryParams, 1000); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidLimit, queryParams["limit"]), request), response)
	} else if skip, err := ParseSkipQueryParameter(queryParams, 0); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf(utils.ErrorInvalidSkip, queryParams["skip"]), request), response)
	} else if stats, count, err := s.DB.GetAzureDataQualityStats(request.Context(), id, start, end, strings.Join(order, ", "), limit, skip); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else {
		api.WriteResponseWrapperWithTimeWindowAndPagination(request.Context(), stats, start, end, limit, skip, count, http.StatusOK, response)
	}
}

func (s *Resources) GetPlatformAggregateStats(response http.ResponseWriter, request *http.Request) {
	var (
		order                    []string
		azureDataQualityStats    model.AzureDataQualityStats
		queryParams              = request.URL.Query()
		defaultEnd, defaultStart = DefaultTimeRange()
	)

	// TODO: This is currently using only the Azure stat type, but should check against the appropriate aggregate type for the chosen platform
	// It's safe for now, but should be refactored
	for _, column := range queryParams[api.QueryParameterSortBy] {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		if !azureDataQualityStats.IsSortable(column) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsNotSortable, request), response)
			return
		}

		if descending {
			order = append(order, column+" desc")
		} else {
			order = append(order, column)
		}
	}

	if id, hasPlatformID := mux.Vars(request)[api.URIPathVariablePlatformID]; !hasPlatformID {
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
			stats, count, err = s.DB.GetADDataQualityAggregations(request.Context(), start, end, strings.Join(order, ", "), limit, skip)
			if err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}
		case "azure":
			stats, count, err = s.DB.GetAzureDataQualityAggregations(request.Context(), start, end, strings.Join(order, ", "), limit, skip)
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
