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
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	FirstName string    `json:"first_name"`
	LastName  string    `json:"last_name"`
}

// ListActiveUsersMinimal - Returns a list of Users without any sensitive data. At the time, this is used in the saved queries
// workflow to return a list of users with whom a query can be shared with.
func (s ManagementResource) ListActiveUsersMinimal(response http.ResponseWriter, request *http.Request) {
	var (
		usersFilter                UserMinimal
		queryParams                = request.URL.Query()
		queryParameterFilterParser = model.NewQueryParameterFilterParser()
	)

	if orderBy, err := api.ParseSortParameters(UserMinimal{}, queryParams); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if order, err := api.BuildSQLSort(orderBy, model.SortItem{Column: "email"}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
	} else if queryFilters, err := queryParameterFilterParser.ParseQueryParameterFilters(request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
	} else {
		for name, filters := range queryFilters {
			if valid := slices.Contains(api.GetFilterableColumns(usersFilter), name); !valid {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			}

			if validPredicates, err := api.GetValidFilterPredicatesAsStrings(usersFilter, name); err != nil {
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s", api.ErrorResponseDetailsColumnNotFilterable, name), request), response)
				return
			} else {
				for i, filter := range filters {
					if !slices.Contains(validPredicates, string(filter.Operator)) {
						api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("%s: %s %s", api.ErrorResponseDetailsFilterPredicateNotSupported, filter.Name, filter.Operator), request), response)
						return
					}
					queryFilters[name][i].IsStringData = usersFilter.IsStringColumn(filter.Name)
				}
			}
		}
		if sqlFilter, err := queryFilters.BuildSQLFilter(); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsBadQueryParameterFilters, request), response)
		} else if activeUsers, err := s.db.GetAllUsers(request.Context(), strings.Join(order, ", "), sqlFilter); err != nil {
			api.HandleDatabaseError(request, response, err)
		} else {
			usersMinimal := make([]UserMinimal, 0)
			for _, user := range activeUsers {
				if !user.IsDisabled {
					usersMinimal = append(usersMinimal, UserMinimal{
						ID:        user.ID,
						Email:     user.EmailAddress.String,
						FirstName: user.FirstName.String,
						LastName:  user.LastName.String,
					})
				}
			}
			api.WriteBasicResponse(request.Context(), UsersMinimalResponse{Users: usersMinimal}, http.StatusOK, response)
		}
	}
}

// Below is needed to allow sorting and filtering on the ListActiveUsersMinimal endpoint.
// Using the same filter as model.User is not ideal as that would allow users to filter/sort on columns they may not have access to.

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
		"first_name": {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"last_name":  {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"email":      {model.Equals, model.NotEquals, model.ApproximatelyEquals},
		"id":         {model.Equals, model.NotEquals},
	}
}

// IsStringColumn - determines if the passed column is a string or not
func (s UserMinimal) IsStringColumn(column string) bool {
	switch column {
	case "first_name",
		"last_name",
		"email":
		return true
	default:
		return false
	}
}
