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

package model

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/cypher/model"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
)

func TestDomainSelectors_TestIsSortable(t *testing.T) {
	domains := DomainSelectors{}
	require.True(t, domains.IsSortable("objectid"))
	require.False(t, domains.IsSortable("foo"))
}

func TestDomainSelectors_GetFilterableColumns(t *testing.T) {
	domains := DomainSelectors{}
	columns := domains.GetFilterableColumns()
	require.Equal(t, 3, len(columns))
}

func TestDomainSelectors_GetValidFilterPredicatesAsStrings(t *testing.T) {
	domains := DomainSelectors{}
	_, err := domains.GetValidFilterPredicatesAsStrings("foo")
	require.Equal(t, ErrorResponseDetailsColumnNotFilterable, err.Error())

	columns := []string{"name", "objectid", "collected"}

	for _, column := range columns {
		predicates, err := domains.GetValidFilterPredicatesAsStrings(column)
		require.Nil(t, err)
		require.Equal(t, 2, len(predicates))
		require.Equal(t, "eq", predicates[0])
		require.Equal(t, "neq", predicates[1])
	}
}

func TestDomainSelectors_GetOrderCriteria_InvalidSortColumn(t *testing.T) {
	domains := DomainSelectors{}
	params := url.Values{}
	params.Add("sort_by", "invalidColumn")

	_, err := domains.GetOrderCriteria(params)
	require.Equal(t, ErrorResponseDetailsColumnNotSortable, err.Error())
}

func TestDomainSelectors_GetOrderCriteria_Success(t *testing.T) {
	domains := DomainSelectors{}
	params := url.Values{}
	params.Add("sort_by", "objectid")
	params.Add("sort_by", "-name")

	orderCriteria, err := domains.GetOrderCriteria(params)
	require.Nil(t, err)
	require.Equal(t, orderCriteria[0].Property, "objectid")
	require.True(t, orderCriteria[0].Order.(model.SortOrder) == query.Ascending())

	require.Equal(t, orderCriteria[1].Property, "name")
	require.True(t, orderCriteria[0].Order.(model.SortOrder) == query.Ascending())
}

func TestDomainSelectors_GetFilterCriteria_InvalidFilterColumn(t *testing.T) {
	request, err := http.NewRequest("GET", "endpoint", nil)
	require.Nil(t, err)
	q := url.Values{}
	q.Add("invalidColumn", "eq:foo")

	request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	request.URL.RawQuery = q.Encode()
	domains := DomainSelectors{}

	_, err = domains.GetFilterCriteria(request)
	require.Equal(t, ErrorResponseDetailsColumnNotFilterable, err.Error())
}

func TestDomainSelectors_GetFilterCriteria_InvalidFilterPredicate(t *testing.T) {
	request, err := http.NewRequest("GET", "endpoint", nil)
	require.Nil(t, err)
	q := url.Values{}
	q.Add("objectid", "gt:foo")

	request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	request.URL.RawQuery = q.Encode()
	domains := DomainSelectors{}

	_, err = domains.GetFilterCriteria(request)
	require.Equal(t, ErrorResponseDetailsFilterPredicateNotSupported, err.Error())
}

func TestDomainSelectors_GetFilterCriteria_Success(t *testing.T) {
	request, err := http.NewRequest("GET", "endpoint", nil)
	require.Nil(t, err)
	q := url.Values{}
	q.Add("objectid", "eq:foo")

	request.Header.Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	request.URL.RawQuery = q.Encode()
	domains := DomainSelectors{}

	filterCriteria, err := domains.GetFilterCriteria(request)
	require.Nil(t, err)
	require.NotNil(t, filterCriteria)
}
