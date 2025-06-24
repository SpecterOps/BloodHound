// Copyright 2025 Specter Ops, Inc.
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

package api_test

import (
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/query"
	"github.com/stretchr/testify/require"
)

func Test_ParseGraphSortParameters_InvalidSortColumn(t *testing.T) {
	domains := model.DomainSelectors{}
	params := url.Values{}
	params.Add("sort_by", "invalidColumn")

	_, err := api.ParseGraphSortParameters(domains, params)
	require.ErrorIs(t, err, api.ErrResponseDetailsCriteriaNotSortable)
}

func Test_ParseGraphSortParameters_Success(t *testing.T) {
	domains := model.DomainSelectors{}
	params := url.Values{}
	params.Add("sort_by", "objectid")
	params.Add("sort_by", "-name")

	orderCriteria, err := api.ParseGraphSortParameters(domains, params)
	require.Nil(t, err)
	require.Equal(t, orderCriteria[0].SortCriteria.(*cypher.PropertyLookup), query.NodeProperty("objectid"))
	require.Equal(t, orderCriteria[0].Direction, query.SortDirectionAscending)

	require.Equal(t, orderCriteria[1].SortCriteria.(*cypher.PropertyLookup), query.NodeProperty("name"))
	require.Equal(t, orderCriteria[1].Direction, query.SortDirectionDescending)
}

func Test_ParseSortParameters(t *testing.T) {
	domains := model.DomainSelectors{}

	t.Run("invalid column", func(t *testing.T) {
		params := url.Values{}
		params.Add("sort_by", "invalidColumn")

		_, err := api.ParseSortParameters(domains, params)
		require.ErrorIs(t, err, api.ErrResponseDetailsColumnNotSortable)
	})

	t.Run("successful sort ascending", func(t *testing.T) {
		params := url.Values{}
		params.Add("sort_by", "objectid")

		sortItems, err := api.ParseSortParameters(domains, params)
		require.Nil(t, err)
		require.Equal(t, sortItems[0].Direction, model.AscendingSortDirection)
		require.Equal(t, sortItems[0].Column, "objectid")
	})

	t.Run("successful sort descending", func(t *testing.T) {
		params := url.Values{}
		params.Add("sort_by", "-objectid")

		sortItems, err := api.ParseSortParameters(domains, params)
		require.Nil(t, err)
		require.Equal(t, sortItems[0].Direction, model.DescendingSortDirection)
		require.Equal(t, sortItems[0].Column, "objectid")
	})
}
