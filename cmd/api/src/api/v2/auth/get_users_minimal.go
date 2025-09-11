// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/gofrs/uuid"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type UsersMinimalResponse struct {
	Users []UserMinimal `json:"users"`
}

type UserMinimal struct {
	ID            uuid.UUID `json:"id"`
	PrincipalName string    `json:"principal_name"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
}

// ListActiveUsersMinimal - Returns a list of Users without any sensitive data. At the time, this is used in the saved queries
// workflow to return a list of users with whom a query can be shared with.
func (s ManagementResource) ListActiveUsersMinimal(response http.ResponseWriter, request *http.Request) {
	var (
		users       UserMinimal
		queryParams = request.URL.Query()
		order       = make([]string, 0)
	)

	orderBy, err := api.ParseSortParameters(UserMinimal{}, queryParams)
	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}
	// ensure deterministic ordering if not provided
	if len(orderBy) == 0 {
		orderBy = append(orderBy, model.SortItem{
			Column: "id",
		})
	}
	for _, column := range orderBy {
		if column.Direction == model.DescendingSortDirection {
			order = append(order, column.Column+" desc")
		} else {
			order = append(order, column.Column)
		}
	}
	queryParameterFilterParser := model.NewQueryParameterFilterParser()
	if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		return
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(users.GetFilterableColumns(), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := users.GetValidFilterPredicatesAsStrings(name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}

					queryFilters[name][i].IsStringData = users.IsStringColumn(filter.Name)
				}
			}
		}

		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "error building SQL for filter", request), response)
			return
		} else if activeUsers, err := s.db.GetAllActiveUsers(request.Context(), strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			usersMinimal := make([]UserMinimal, 0)
			for _, user := range activeUsers {
				usersMinimal = append(usersMinimal, UserMinimal{
					ID:            user.ID,
					PrincipalName: user.PrincipalName,
					FirstName:     user.FirstName.String,
					LastName:      user.LastName.String,
				})
			}
			api.WriteBasicResponse(request.Context(), UsersMinimalResponse{Users: usersMinimal}, http.StatusOK, response)
		}
	}
}

// Below is needed to allow sorting and filtering on the ListActiveUsersMinimal endpoint.
// Using the model.User columns is not ideal as that would allow users to filter/sort on columns they may not have access to.

// IsSortable - determines if the passed column can be sorted on or not
func (s UserMinimal) IsSortable(column string) bool {
	switch column {
	case "first_name",
		"last_name",
		"principal_name",
		"id":
		return true
	default:
		return false
	}
}

// ValidFilters - returns a map of columns and their valid filters
func (s UserMinimal) ValidFilters() map[string][]model.FilterOperator {
	return map[string][]model.FilterOperator{
		"first_name":     {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"last_name":      {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"principal_name": {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"id":             {model.Equals, model.NotEquals},
	}
}

// IsStringColumn - determines if the passed column is a string or not
func (s UserMinimal) IsStringColumn(column string) bool {
	switch column {
	case "first_name",
		"last_name",
		"principal_name":
		return true
	default:
		return false
	}
}

// GetFilterableColumns - returns a list of filterable columns
func (s UserMinimal) GetFilterableColumns() []string {
	columns := make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

// GetValidFilterPredicatesAsStrings - returns a list of predicates that a column can be filtered on
func (s UserMinimal) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		stringPredicates := make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}
