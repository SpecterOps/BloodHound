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

package v2_test

import (
	"errors"
	"fmt"
	"net/url"
	"testing"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/stretchr/testify/require"
)

func TestParseIntQueryParameter(t *testing.T) {
	t.Parallel()
	type args struct {
		key          string
		defaultValue int
	}
	type want struct {
		res int
		err error
	}
	tests := []struct {
		name     string
		addParam func() url.Values
		args     args
		want     want
	}{
		{
			name: "Error: parameter is not provided",
			args: args{
				key:          "key1",
				defaultValue: 0,
			},
			addParam: func() url.Values { return url.Values{} },
			want: want{
				err: nil,
				res: 0,
			},
		},
		{
			name: "Error: parameter is not an int",
			args: args{
				key:          "key1",
				defaultValue: 0,
			},
			addParam: func() url.Values {
				urlValues := url.Values{}
				urlValues.Add("key1", "value1")
				return urlValues
			},
			want: want{
				err: errors.New("strconv.Atoi: parsing \"value1\": invalid syntax"),
				res: 0,
			},
		},
		{
			name: "Success: query parameter is int",
			args: args{
				key:          "key1",
				defaultValue: 0,
			},
			addParam: func() url.Values {
				urlValues := url.Values{}
				urlValues.Add("key1", "1")
				return urlValues
			},
			want: want{
				err: nil,
				res: 1,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			urlValues := testCase.addParam()

			got, err := v2.ParseIntQueryParameter(urlValues, testCase.args.key, testCase.args.defaultValue)
			if err != nil {
				require.EqualError(t, testCase.want.err, err.Error())
				require.Equal(t, testCase.want.res, got)
			} else {
				require.Equal(t, testCase.want.res, got)
				require.NoError(t, testCase.want.err)
			}
		})
	}
}

func TestParseSkipQueryParameter_BadFormat(t *testing.T) {
	params := url.Values{}
	params.Add("skip", "abcd")
	_, err := v2.ParseSkipQueryParameter(params, 0)
	require.Contains(t, err.Error(), "error converting skip value")
}

func TestParseSkipQueryParameter_Invalid(t *testing.T) {
	params := url.Values{}
	params.Add("skip", "-5")
	_, err := v2.ParseSkipQueryParameter(params, 0)
	require.Equal(t, fmt.Errorf(utils.ErrorInvalidSkip, "-5"), err)
}

func TestParseSkipQueryParameter(t *testing.T) {
	params := url.Values{}
	params.Add("skip", "5")
	skip, err := v2.ParseSkipQueryParameter(params, 0)
	require.Nil(t, err)
	require.Equal(t, 5, skip)
}

func TestParseLimitQueryParameter_BadFormat(t *testing.T) {
	params := url.Values{}
	params.Add("limit", "abcd")
	_, err := v2.ParseLimitQueryParameter(params, 0)
	require.Contains(t, err.Error(), "error converting limit value")
}

func TestParseLimitQueryParameter_Invalid(t *testing.T) {
	params := url.Values{}
	params.Add("limit", "-5")
	_, err := v2.ParseLimitQueryParameter(params, 0)
	require.Equal(t, fmt.Errorf(utils.ErrorInvalidLimit, "-5"), err)
}

func TestParseLimitQueryParameter(t *testing.T) {
	params := url.Values{}
	params.Add("limit", "5")
	limit, err := v2.ParseLimitQueryParameter(params, 0)
	require.Nil(t, err)
	require.Equal(t, 5, limit)
}

func TestParseOptionalLimitQueryParameter(t *testing.T) {
	t.Parallel()
	type args struct {
		defaultInt int
	}
	type want struct {
		res int
		err error
	}
	tests := []struct {
		name     string
		addParam func() url.Values
		args     args
		want     want
	}{
		{
			name: "Success: optional parameter `limit` not present, returns default",
			args: args{
				defaultInt: 1,
			},
			addParam: func() url.Values { return url.Values{} },
			want: want{
				err: nil,
				res: 1,
			},
		},
		{
			name: "Error: optional parameter `limit` invalid",
			args: args{
				defaultInt: 1,
			},
			addParam: func() url.Values {
				urlValues := url.Values{}
				urlValues.Add("limit", "invalid")
				return urlValues
			},
			want: want{
				err: errors.New("error converting limit value invalid to int: strconv.Atoi: parsing \"invalid\": invalid syntax"),
				res: 0,
			},
		},
		{
			name: "Error: optional parameter `limit` negative number",
			args: args{
				defaultInt: 1,
			},
			addParam: func() url.Values {
				urlValues := url.Values{}
				urlValues.Add("limit", "-2")
				return urlValues
			},
			want: want{
				err: errors.New("invalid limit: -2"),
				res: 0,
			},
		},
		{
			name: "Success: optional parameter `limit` negative number",
			args: args{
				defaultInt: 1,
			},
			addParam: func() url.Values {
				urlValues := url.Values{}
				urlValues.Add("limit", "5")
				return urlValues
			},
			want: want{
				err: nil,
				res: 5,
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			urlValues := testCase.addParam()

			got, err := v2.ParseOptionalLimitQueryParameter(urlValues, testCase.args.defaultInt)
			if err != nil {
				require.EqualError(t, testCase.want.err, err.Error())
				require.Equal(t, testCase.want.res, got)
			} else {
				require.Equal(t, testCase.want.res, got)
				require.NoError(t, testCase.want.err)
			}
		})
	}
}

func TestAddWhereClauseToNeo4jQuery_NoWhereClause(t *testing.T) {
	statement := "MATCH (n:Domain) RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"
	filters := "n.name = 'abcd' AND n.objectid = 1"
	expected := "MATCH (n:Domain) WHERE n.name = 'abcd' AND n.objectid = 1 RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"

	actual := v2.AddWhereClauseToNeo4jQuery(statement, filters)
	require.Equal(t, expected, actual)
}

func TestAddWhereClauseToNeo4jQuery_WithWhereClause(t *testing.T) {
	statement := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"
	filters := "n.objectid = 1"
	expected := "MATCH (n:Domain) WHERE n.objectid = 1 AND n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"

	actual := v2.AddWhereClauseToNeo4jQuery(statement, filters)
	require.Equal(t, expected, actual)
}

func TestAddOrderByToNeo4jQuery_NoOrderByClause(t *testing.T) {
	statement := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false);"
	orderBy := "n.name, n.objectid"
	expected := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false) ORDER BY n.name, n.objectid;"

	actual := v2.AddOrderByToNeo4jQuery(statement, orderBy)
	require.Equal(t, expected, actual)
}

func TestAddOrderByToNeo4jQuery_WithOrderByClause(t *testing.T) {
	statement := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false) ORDER BY n.name;"
	orderBy := "n.objectid"
	expected := "MATCH (n:Domain) WHERE n.name = 'abcd' RETURN coalesce(n.name, n.objectid), n.objectid, coalesce(n.collected, false) ORDER BY n.objectid, n.name;"

	actual := v2.AddOrderByToNeo4jQuery(statement, orderBy)
	require.Equal(t, expected, actual)
}
