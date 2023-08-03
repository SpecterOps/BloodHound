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
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/utils"
	"github.com/stretchr/testify/require"
)

func TestParseSkipQueryParameter_BadFormat(t *testing.T) {
	params := url.Values{}
	params.Add("skip", "abcd")
	_, err := ParseSkipQueryParameter(params, 0)
	require.Contains(t, err.Error(), "error converting skip value")
}

func TestParseSkipQueryParameter_Invalid(t *testing.T) {
	params := url.Values{}
	params.Add("skip", "-5")
	_, err := ParseSkipQueryParameter(params, 0)
	require.Equal(t, fmt.Errorf(utils.ErrorInvalidSkip, "-5"), err)
}

func TestParseSkipQueryParameter(t *testing.T) {
	params := url.Values{}
	params.Add("skip", "5")
	skip, err := ParseSkipQueryParameter(params, 0)
	require.Nil(t, err)
	require.Equal(t, 5, skip)
}

func TestParseLimitQueryParameter_BadFormat(t *testing.T) {
	params := url.Values{}
	params.Add("limit", "abcd")
	_, err := ParseLimitQueryParameter(params, 0)
	require.Contains(t, err.Error(), "error converting limit value")
}

func TestParseLimitQueryParameter_Invalid(t *testing.T) {
	params := url.Values{}
	params.Add("limit", "-5")
	_, err := ParseLimitQueryParameter(params, 0)
	require.Equal(t, fmt.Errorf(utils.ErrorInvalidLimit, "-5"), err)
}

func TestParseLimitQueryParameter(t *testing.T) {
	params := url.Values{}
	params.Add("limit", "5")
	limit, err := ParseLimitQueryParameter(params, 0)
	require.Nil(t, err)
	require.Equal(t, 5, limit)
}

func TestAddWhereClauseToNeo4jQuery_NoWhereClause(t *testing.T) {
	statement := "MATCH (n:Domain) RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"
	filters := "n.name = 'abcd' AND n.objectid = 1"
	expected := "MATCH (n:Domain) WHERE n.name = 'abcd' AND n.objectid = 1 RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"

	actual := AddWhereClauseToNeo4jQuery(statement, filters)
	require.Equal(t, expected, actual)
}

func TestAddWhereClauseToNeo4jQuery_WithWhereClause(t *testing.T) {
	statement := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"
	filters := "n.objectid = 1"
	expected := "MATCH (n:Domain) WHERE n.objectid = 1 AND n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"

	actual := AddWhereClauseToNeo4jQuery(statement, filters)
	require.Equal(t, expected, actual)
}

func TestAddOrderByToNeo4jQuery_NoOrderByClause(t *testing.T) {
	statement := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"
	orderBy := "n.name, n.objectid"
	expected := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false) ORDER BY n.name, n.objectid;"

	actual := AddOrderByToNeo4jQuery(statement, orderBy)
	require.Equal(t, expected, actual)
}

func TestAddOrderByToNeo4jQuery_WithOrderByClause(t *testing.T) {
	statement := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false) ORDER BY n.name;"
	orderBy := "n.objectid"
	expected := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false) ORDER BY n.objectid, n.name;"

	actual := AddOrderByToNeo4jQuery(statement, orderBy)
	require.Equal(t, expected, actual)
}
