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
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/stretchr/testify/require"
)

// fakeFilterable is a minimal Filterable used to drive ParseAndValidate without coupling the test to any
// persistence layer.
type fakeFilterable struct {
	fields map[string]params.FilterableField
}

func (s fakeFilterable) ValidFilters() map[string]params.FilterableField {
	return s.fields
}

func TestParseFilterOperator(t *testing.T) {
	for _, raw := range []string{"gt", "gte", "lt", "lte", "eq", "neq", "~eq"} {
		t.Run("parses "+raw, func(t *testing.T) {
			operator, err := params.ParseFilterOperator(raw)
			require.NoError(t, err)
			require.Equal(t, params.FilterOperator(raw), operator)
		})
	}

	t.Run("rejects an unknown predicate and wraps ErrMalformedFilter", func(t *testing.T) {
		_, err := params.ParseFilterOperator("bogus")
		require.Error(t, err)
		require.ErrorIs(t, err, params.ErrMalformedFilter)
		require.Contains(t, err.Error(), "bogus")
	})
}

func TestFilterParseAndValidate_Parsing(t *testing.T) {
	var (
		parser     = params.NewQueryParameterFilterParser()
		filterable = fakeFilterable{fields: map[string]params.FilterableField{
			"parameter": {Operators: []params.FilterOperator{params.Equals, params.ApproximatelyEquals}},
		}}
	)

	// These cases mirror the legacy model parser tests to vet behavioral parity.
	cases := []struct {
		name             string
		value            string
		expectedOperator params.FilterOperator
		expectedValue    string
	}{
		{name: "parses a parameter filter", value: "eq:auth.value", expectedOperator: params.Equals, expectedValue: "auth.value"},
		{name: "parses a parameter with ~", value: "~eq:auth.value", expectedOperator: params.ApproximatelyEquals, expectedValue: "auth.value"},
		{name: "parses a parameter filter with spacing", value: "eq:hello world", expectedOperator: params.Equals, expectedValue: "hello world"},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			parsed, err := parser.ParseAndValidate(url.Values{"parameter": {testCase.value}}, filterable)
			require.NoError(t, err)
			require.Len(t, parsed["parameter"], 1)
			require.Equal(t, "parameter", parsed["parameter"][0].Field)
			require.Equal(t, testCase.expectedOperator, parsed["parameter"][0].Operator)
			require.Equal(t, testCase.expectedValue, parsed["parameter"][0].Value)
		})
	}

	t.Run("silently skips a non-matching value", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{"parameter": {"hello world"}}, filterable)
		require.NoError(t, err)
		require.Empty(t, parsed["parameter"])
	})

	t.Run("supports subsequent ':' in value", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{"parameter": {"eq:foo:bar"}}, filterable)
		require.NoError(t, err)
		require.Len(t, parsed["parameter"], 1)
		require.Equal(t, "foo:bar", parsed["parameter"][0].Value)
	})

	t.Run("parses multiple values for the same field in order", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{"parameter": {"eq:first", "~eq:second"}}, filterable)
		require.NoError(t, err)
		require.Len(t, parsed["parameter"], 2)
		require.Equal(t, "first", parsed["parameter"][0].Value)
		require.Equal(t, params.ApproximatelyEquals, parsed["parameter"][1].Operator)
		require.Equal(t, "second", parsed["parameter"][1].Value)
	})
}

func TestFilterParseAndValidate_Enrichment(t *testing.T) {
	parser := params.NewQueryParameterFilterParser()

	t.Run("enriches IsStringData and defaults an empty SetOperator to FilterAnd", func(t *testing.T) {
		filterable := fakeFilterable{fields: map[string]params.FilterableField{
			"name": {Operators: []params.FilterOperator{params.Equals}, IsStringData: true},
		}}

		parsed, err := parser.ParseAndValidate(url.Values{"name": {"eq:value"}}, filterable)
		require.NoError(t, err)
		require.Len(t, parsed["name"], 1)
		require.True(t, parsed["name"][0].IsStringData)
		require.Equal(t, params.FilterAnd, parsed["name"][0].SetOperator)
	})

	t.Run("carries a declared FilterOr SetOperator through to every filter", func(t *testing.T) {
		filterable := fakeFilterable{fields: map[string]params.FilterableField{
			"name": {Operators: []params.FilterOperator{params.Equals}, SetOperator: params.FilterOr},
		}}

		parsed, err := parser.ParseAndValidate(url.Values{"name": {"eq:a", "eq:b"}}, filterable)
		require.NoError(t, err)
		require.Len(t, parsed["name"], 2)
		require.Equal(t, params.FilterOr, parsed["name"][0].SetOperator)
		require.Equal(t, params.FilterOr, parsed["name"][1].SetOperator)
	})
}

func TestFilterParseAndValidate_IgnoredParameters(t *testing.T) {
	var (
		parser     = params.NewQueryParameterFilterParser("skip", "limit")
		filterable = fakeFilterable{fields: map[string]params.FilterableField{
			"name": {Operators: []params.FilterOperator{params.Equals}},
		}}
	)

	// Ignored parameters are skipped before field validation, so a value that would otherwise fail the
	// filterable lookup must not produce an error.
	parsed, err := parser.ParseAndValidate(url.Values{
		"skip":  {"eq:0"},
		"limit": {"eq:10"},
		"name":  {"eq:value"},
	}, filterable)
	require.NoError(t, err)
	require.Len(t, parsed["name"], 1)
	require.NotContains(t, parsed, "skip")
	require.NotContains(t, parsed, "limit")
}

func TestFilterParseAndValidate_Errors(t *testing.T) {
	var (
		parser     = params.NewQueryParameterFilterParser()
		filterable = fakeFilterable{fields: map[string]params.FilterableField{
			"name": {Operators: []params.FilterOperator{params.Equals}},
		}}
	)

	t.Run("ErrFieldNotFilterable for an unknown field", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"unknown": {"eq:value"}}, filterable)
		require.ErrorIs(t, err, params.ErrFieldNotFilterable)

		var validationError *params.FilterValidationError
		require.ErrorAs(t, err, &validationError)
		require.Equal(t, "unknown", validationError.Field)
	})

	t.Run("ErrOperatorNotSupported for an unsupported operator", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"name": {"gt:value"}}, filterable)
		require.ErrorIs(t, err, params.ErrOperatorNotSupported)

		var validationError *params.FilterValidationError
		require.ErrorAs(t, err, &validationError)
		require.Equal(t, "name", validationError.Field)
		require.Equal(t, params.GreaterThan, validationError.Operator)
	})

	t.Run("ErrMalformedFilter for an unknown predicate, preserving the raw predicate", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"name": {"bogus:value"}}, filterable)
		require.ErrorIs(t, err, params.ErrMalformedFilter)
		require.Contains(t, err.Error(), "bogus")
	})
}
