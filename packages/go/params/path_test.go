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

package params_test

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/stretchr/testify/require"
)

func TestPathUUID(t *testing.T) {
	var (
		identifier = uuid.Must(uuid.FromString("0199b6b8-0000-7000-8000-000000000001"))
		cases      = []struct {
			name     string
			vars     map[string]string
			expected uuid.UUID
			sentinel error
		}{
			{name: "valid UUID", vars: map[string]string{"id": identifier.String()}, expected: identifier},
			{name: "nil UUID", vars: map[string]string{"id": uuid.Nil.String()}, expected: uuid.Nil},
			{name: "no mux variables", sentinel: params.ErrMissingPathParameter},
			{name: "empty mux variables", vars: map[string]string{}, sentinel: params.ErrMissingPathParameter},
			{name: "different parameter", vars: map[string]string{"other": identifier.String()}, sentinel: params.ErrMissingPathParameter},
			{name: "empty value", vars: map[string]string{"id": ""}, sentinel: params.ErrMalformedPathParameter},
			{name: "malformed UUID", vars: map[string]string{"id": "not-a-uuid"}, sentinel: params.ErrMalformedPathParameter},
			{name: "invalid hexadecimal", vars: map[string]string{"id": "0199b6b8-0000-7000-8000-00000000000z"}, sentinel: params.ErrMalformedPathParameter},
			{name: "whitespace", vars: map[string]string{"id": " " + identifier.String()}, sentinel: params.ErrMalformedPathParameter},
		}
	)

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var request = httptest.NewRequest(http.MethodGet, "/", nil)
			if testCase.vars != nil {
				request = mux.SetURLVars(request, testCase.vars)
			}

			identifier, err := params.PathUUID(request, "id")
			require.Equal(t, testCase.expected, identifier)
			if testCase.sentinel == nil {
				require.NoError(t, err)
			} else {
				requirePathValidationError(t, err, testCase.sentinel, testCase.vars["id"])
			}
		})
	}
}

func TestPathIntegers(t *testing.T) {
	var parsers = []struct {
		name    string
		parse   func(*http.Request, string) (int64, error)
		minimum int64
		maximum int64
	}{
		{
			name: "PathInt32",
			parse: func(request *http.Request, parameter string) (int64, error) {
				var value, err = params.PathInt32(request, parameter)
				return int64(value), err
			},
			minimum: math.MinInt32,
			maximum: math.MaxInt32,
		},
		{name: "PathInt64", parse: params.PathInt64, minimum: math.MinInt64, maximum: math.MaxInt64},
	}

	for _, parser := range parsers {
		t.Run(parser.name, func(t *testing.T) {
			var cases = []struct {
				name     string
				vars     map[string]string
				expected int64
				sentinel error
				cause    error
			}{
				{name: "positive", vars: map[string]string{"id": "42"}, expected: 42},
				{name: "zero", vars: map[string]string{"id": "0"}},
				{name: "negative", vars: map[string]string{"id": "-42"}, expected: -42},
				{name: "explicit positive sign", vars: map[string]string{"id": "+42"}, expected: 42},
				{name: "leading zero", vars: map[string]string{"id": "042"}, expected: 42},
				{name: "minimum", vars: map[string]string{"id": strconv.FormatInt(parser.minimum, 10)}, expected: parser.minimum},
				{name: "maximum", vars: map[string]string{"id": strconv.FormatInt(parser.maximum, 10)}, expected: parser.maximum},
				{name: "no mux variables", sentinel: params.ErrMissingPathParameter},
				{name: "empty mux variables", vars: map[string]string{}, sentinel: params.ErrMissingPathParameter},
				{name: "different parameter", vars: map[string]string{"other": "42"}, sentinel: params.ErrMissingPathParameter},
				{name: "empty value", vars: map[string]string{"id": ""}, sentinel: params.ErrMalformedPathParameter, cause: strconv.ErrSyntax},
				{name: "malformed", vars: map[string]string{"id": "abc"}, sentinel: params.ErrMalformedPathParameter, cause: strconv.ErrSyntax},
				{name: "fraction", vars: map[string]string{"id": "1.5"}, sentinel: params.ErrMalformedPathParameter, cause: strconv.ErrSyntax},
				{name: "whitespace", vars: map[string]string{"id": " 42"}, sentinel: params.ErrMalformedPathParameter, cause: strconv.ErrSyntax},
				{name: "hexadecimal", vars: map[string]string{"id": "0x2a"}, sentinel: params.ErrMalformedPathParameter, cause: strconv.ErrSyntax},
				{name: "overflow", vars: map[string]string{"id": strconv.FormatUint(uint64(parser.maximum)+1, 10)}, sentinel: params.ErrMalformedPathParameter, cause: strconv.ErrRange},
				{name: "underflow", vars: map[string]string{"id": "-" + strconv.FormatUint(uint64(parser.maximum)+2, 10)}, sentinel: params.ErrMalformedPathParameter, cause: strconv.ErrRange},
			}

			for _, testCase := range cases {
				t.Run(testCase.name, func(t *testing.T) {
					var request = httptest.NewRequest(http.MethodGet, "/", nil)
					if testCase.vars != nil {
						request = mux.SetURLVars(request, testCase.vars)
					}

					value, err := parser.parse(request, "id")
					require.Equal(t, testCase.expected, value)
					if testCase.sentinel == nil {
						require.NoError(t, err)
					} else {
						requirePathValidationError(t, err, testCase.sentinel, testCase.vars["id"])
						if testCase.cause != nil {
							var parseError *strconv.NumError
							require.ErrorIs(t, err, testCase.cause)
							require.ErrorAs(t, err, &parseError)
							require.Equal(t, testCase.vars["id"], parseError.Num)
						}
					}
				})
			}
		})
	}
}

func TestPathValidationError_PreservesParserCause(t *testing.T) {
	var cases = []struct {
		name      string
		parameter string
		value     string
		parse     func(*http.Request, string) error
		cause     error
	}{
		{
			name:      "UUID syntax",
			parameter: "client_id",
			value:     "not-a-uuid",
			parse: func(request *http.Request, parameter string) error {
				var _, err = params.PathUUID(request, parameter)
				return err
			},
		},
		{
			name:      "int32 overflow",
			parameter: "feature_id",
			value:     "2147483648",
			parse: func(request *http.Request, parameter string) error {
				var _, err = params.PathInt32(request, parameter)
				return err
			},
			cause: strconv.ErrRange,
		},
		{
			name:      "int64 syntax",
			parameter: "node_id",
			value:     "invalid",
			parse: func(request *http.Request, parameter string) error {
				var _, err = params.PathInt64(request, parameter)
				return err
			},
			cause: strconv.ErrSyntax,
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var (
				request         = mux.SetURLVars(httptest.NewRequest(http.MethodGet, "/", nil), map[string]string{testCase.parameter: testCase.value})
				wrapped         = fmt.Errorf("parse request: %w", testCase.parse(request, testCase.parameter))
				validationError *params.PathValidationError
			)

			require.ErrorIs(t, wrapped, params.ErrMalformedPathParameter)
			require.ErrorAs(t, wrapped, &validationError)
			require.Equal(t, testCase.parameter, validationError.Parameter)
			require.Equal(t, testCase.value, validationError.Value)
			require.Error(t, validationError.Cause)
			require.Same(t, validationError.Cause, errors.Unwrap(validationError))
			require.ErrorIs(t, wrapped, validationError.Cause)

			if testCase.cause != nil {
				var parseError *strconv.NumError
				require.ErrorIs(t, wrapped, testCase.cause)
				require.ErrorAs(t, wrapped, &parseError)
				require.Same(t, validationError.Cause, parseError)
				require.Equal(t, "ParseInt", parseError.Func)
				require.Equal(t, testCase.value, parseError.Num)
			} else {
				var _, parseError = uuid.FromString(testCase.value)
				require.EqualError(t, validationError.Cause, parseError.Error())
			}
		})
	}
}

func requirePathValidationError(t *testing.T, err, sentinel error, rawValue string) {
	t.Helper()
	var (
		validationError *params.PathValidationError
		wrapped         = fmt.Errorf("parse request: %w", err)
	)

	require.ErrorIs(t, wrapped, sentinel)
	require.ErrorAs(t, wrapped, &validationError)
	require.Equal(t, "id", validationError.Parameter)
	require.Equal(t, rawValue, validationError.Value)
	if sentinel == params.ErrMissingPathParameter {
		require.NotErrorIs(t, wrapped, params.ErrMalformedPathParameter)
		require.Nil(t, validationError.Cause)
	} else {
		require.NotErrorIs(t, wrapped, params.ErrMissingPathParameter)
		require.Error(t, validationError.Cause)
	}
}
