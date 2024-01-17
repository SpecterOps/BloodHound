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
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/utils"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/errors"
)

type DataType int

func ErrBadQueryParameter(request *http.Request, key string, err error) *api.ErrorWrapper {
	return api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("query parameter \"%s\" is malformed: %v", key, err), request)
}

func ParseIntQueryParameter(params url.Values, key string, defaultValue int) (int, error) {
	if param := params.Get(key); param != "" {
		return strconv.Atoi(param)
	}

	return defaultValue, nil
}

func ParseSkipQueryParameter(params url.Values, defaultValue int) (int, error) {
	if param := params.Get(model.PaginationQueryParameterSkip); param == "" {
		return defaultValue, nil
	} else if skip, err := strconv.Atoi(param); err != nil {
		return 0, fmt.Errorf("error converting skip value %v to int: %v", param, err)
	} else if skip < 0 {
		return 0, fmt.Errorf(utils.ErrorInvalidSkip, skip)
	} else {
		return skip, nil
	}
}

func ParseLimitQueryParameter(params url.Values, defaultValue int) (int, error) {
	if param := params.Get(model.PaginationQueryParameterLimit); param == "" {
		return defaultValue, nil
	} else if limit, err := strconv.Atoi(param); err != nil {
		return 0, fmt.Errorf("error converting limit value %v to int: %v", param, err)
	} else if limit < 0 {
		return 0, fmt.Errorf(utils.ErrorInvalidLimit, limit)
	} else {
		return limit, nil
	}
}

func ParseTimeQueryParameter(params url.Values, key string, defaultValue time.Time) (time.Time, error) {
	if param := params.Get(key); param != "" {
		return time.Parse(time.RFC3339Nano, param)
	}

	return defaultValue, nil
}

type QueryWrapper struct {
	BaseQuery  string
	ListQuery  string
	GraphQuery string
	CountQuery string
}

func (s QueryWrapper) GetQuery(dataType model.DataType) string {
	switch dataType {
	case model.DataTypeGraph:
		if s.BaseQuery != "" {
			return fmt.Sprintf("%v RETURN apoc.agg.graph(p)", s.BaseQuery)
		} else {
			return fmt.Sprintf("%v RETURN apoc.agg.graph(p)", s.GraphQuery)
		}

	case model.DataTypeList:
		if s.BaseQuery != "" {
			return fmt.Sprintf("%v WITH distinct(n) RETURN n as node ORDER BY coalesce(n.name, n.objectid) SKIP $skip LIMIT $limit", s.BaseQuery)
		} else {
			return fmt.Sprintf("%v WITH distinct(n) RETURN n as node ORDER BY coalesce(n.name, n.objectid) SKIP $skip LIMIT $limit", s.ListQuery)
		}

	case model.DataTypeCount:
		// Count queries may contain significant optimization work that makes formatting them inappropriate
		if s.CountQuery != "" {
			return s.CountQuery
		}

		if s.BaseQuery != "" {
			return fmt.Sprintf("%v RETURN count(distinct(n))", s.BaseQuery)
		} else {
			return fmt.Sprintf("%v RETURN count(distinct(n))", s.ListQuery)
		}
	}

	return ""
}

func GetEntityObjectIDFromRequestPath(req *http.Request) (string, error) {
	if id, hasID := mux.Vars(req)["object_id"]; !hasID {
		return "", errors.Error("no object ID found in request")
	} else {
		return id, nil
	}
}

func AddWhereClauseToNeo4jQuery(statement string, filters string) string {
	var (
		statement1      string
		statement2      string
		fullWhereClause string
		whereIndex      = strings.Index(strings.ToUpper(statement), "WHERE")
	)

	if whereIndex >= 0 {
		// WHERE clause already exists; split the statement at 'WHERE', include 'WHERE' in statement1
		statement1 = statement[0 : whereIndex+6]
		statement2 = statement[whereIndex+6:]

		// add filters to the existing 'WHERE' clauses
		fullWhereClause = fmt.Sprintf("%s AND ", filters)
	} else {
		// no 'WHERE' clause; it will have to be added before 'RETURN'
		// split the statement at RETURN, include 'RETURN' in statement2
		returnIndex := strings.Index(strings.ToUpper(statement), "RETURN")
		statement1 = statement[0:returnIndex]
		statement2 = statement[returnIndex:]

		// create a standalone 'WHERE' clause with the filters
		fullWhereClause = fmt.Sprintf("WHERE %s ", filters)
	}

	statement = fmt.Sprintf("%s%s%s", statement1, fullWhereClause, statement2)
	return statement
}

func AddOrderByToNeo4jQuery(statement string, orderBy string) string {
	orderByIndex := strings.Index(strings.ToUpper(statement), "ORDER BY")

	if orderByIndex >= 0 {
		// insert orderBy after the existing 'ORDER BY'
		statement1 := statement[0 : orderByIndex+9]
		statement2 := statement[orderByIndex+9:]

		statement = fmt.Sprintf("%s%s, %s", statement1, orderBy, statement2)
	} else {
		// strip the trailing semicolon (if it exists) and add 'ORDER BY' clause at the end
		if string(statement[len(statement)-1]) == ";" {
			statement = statement[0 : len(statement)-1]
		}

		statement = statement + " ORDER BY " + orderBy + ";"
	}
	return statement
}
