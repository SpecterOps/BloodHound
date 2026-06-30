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

package filters_test

import (
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/packages/go/filters"
	"github.com/stretchr/testify/require"
)

// fakeFilterable is a minimal Filterable used to drive ParseAndValidate without coupling the test to any
// persistence layer.
type fakeFilterable struct {
	fields map[string]filters.FilterableField
}

func (s fakeFilterable) ValidFilters() map[string]filters.FilterableField {
	return s.fields
}

func TestParseFilterOperator(t *testing.T) {
	for _, raw := range []string{"gt", "gte", "lt", "lte", "eq", "neq", "~eq"} {
		t.Run("parses "+raw, func(t *testing.T) {
			operator, err := filters.ParseFilterOperator(raw)
			require.NoError(t, err)
			require.Equal(t, filters.FilterOperator(raw), operator)
		})
	}

	t.Run("rejects an unknown predicate and wraps ErrMalformedFilter", func(t *testing.T) {
		_, err := filters.ParseFilterOperator("bogus")
		require.Error(t, err)
		require.ErrorIs(t, err, filters.ErrMalformedFilter)
		require.Contains(t, err.Error(), "bogus")
	})
}

func TestParseAndValidate_Parsing(t *testing.T) {
	var (
		parser     = filters.NewQueryParameterFilterParser()
		filterable = fakeFilterable{fields: map[string]filters.FilterableField{
			"parameter": {Operators: []filters.FilterOperator{filters.Equals, filters.ApproximatelyEquals}},
		}}
	)

	// These cases mirror the legacy model parser tests to vet behavioral parity.
	cases := []struct {
		name             string
		value            string
		expectedOperator filters.FilterOperator
		expectedValue    string
	}{
		{name: "parses a parameter filter", value: "eq:auth.value", expectedOperator: filters.Equals, expectedValue: "auth.value"},
		{name: "parses a parameter with ~", value: "~eq:auth.value", expectedOperator: filters.ApproximatelyEquals, expectedValue: "auth.value"},
		{name: "parses a parameter filter with spacing", value: "eq:hello world", expectedOperator: filters.Equals, expectedValue: "hello world"},
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
		parsed, err := parser.ParseAndValidate(url.Values{"parameter": {"eq : hello world"}}, filterable)
		require.NoError(t, err)
		require.Empty(t, parsed["parameter"])
	})

	t.Run("parses multiple values for the same field in order", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{"parameter": {"eq:first", "~eq:second"}}, filterable)
		require.NoError(t, err)
		require.Len(t, parsed["parameter"], 2)
		require.Equal(t, "first", parsed["parameter"][0].Value)
		require.Equal(t, filters.ApproximatelyEquals, parsed["parameter"][1].Operator)
		require.Equal(t, "second", parsed["parameter"][1].Value)
	})
}

func TestParseAndValidate_Enrichment(t *testing.T) {
	parser := filters.NewQueryParameterFilterParser()

	t.Run("enriches IsStringData and defaults an empty SetOperator to FilterAnd", func(t *testing.T) {
		filterable := fakeFilterable{fields: map[string]filters.FilterableField{
			"name": {Operators: []filters.FilterOperator{filters.Equals}, IsStringData: true},
		}}

		parsed, err := parser.ParseAndValidate(url.Values{"name": {"eq:value"}}, filterable)
		require.NoError(t, err)
		require.Len(t, parsed["name"], 1)
		require.True(t, parsed["name"][0].IsStringData)
		require.Equal(t, filters.FilterAnd, parsed["name"][0].SetOperator)
	})

	t.Run("carries a declared FilterOr SetOperator through to every filter", func(t *testing.T) {
		filterable := fakeFilterable{fields: map[string]filters.FilterableField{
			"name": {Operators: []filters.FilterOperator{filters.Equals}, SetOperator: filters.FilterOr},
		}}

		parsed, err := parser.ParseAndValidate(url.Values{"name": {"eq:a", "eq:b"}}, filterable)
		require.NoError(t, err)
		require.Len(t, parsed["name"], 2)
		require.Equal(t, filters.FilterOr, parsed["name"][0].SetOperator)
		require.Equal(t, filters.FilterOr, parsed["name"][1].SetOperator)
	})
}

func TestParseAndValidate_IgnoredParameters(t *testing.T) {
	var (
		parser     = filters.NewQueryParameterFilterParser("skip", "limit")
		filterable = fakeFilterable{fields: map[string]filters.FilterableField{
			"name": {Operators: []filters.FilterOperator{filters.Equals}},
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

func TestParseAndValidate_Errors(t *testing.T) {
	var (
		parser     = filters.NewQueryParameterFilterParser()
		filterable = fakeFilterable{fields: map[string]filters.FilterableField{
			"name": {Operators: []filters.FilterOperator{filters.Equals}},
		}}
	)

	t.Run("ErrFieldNotFilterable for an unknown field", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"unknown": {"eq:value"}}, filterable)
		require.ErrorIs(t, err, filters.ErrFieldNotFilterable)

		var validationError *filters.ValidationError
		require.ErrorAs(t, err, &validationError)
		require.Equal(t, "unknown", validationError.Field)
	})

	t.Run("ErrOperatorNotSupported for an unsupported operator", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"name": {"gt:value"}}, filterable)
		require.ErrorIs(t, err, filters.ErrOperatorNotSupported)

		var validationError *filters.ValidationError
		require.ErrorAs(t, err, &validationError)
		require.Equal(t, "name", validationError.Field)
		require.Equal(t, filters.GreaterThan, validationError.Operator)
	})

	t.Run("ErrMalformedFilter for an unknown predicate, preserving the raw predicate", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"name": {"bogus:value"}}, filterable)
		require.ErrorIs(t, err, filters.ErrMalformedFilter)
		require.Contains(t, err.Error(), "bogus")
	})
}
