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
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/stretchr/testify/require"
)

func TestPagingParseAndValidate_Parsing(t *testing.T) {
	var cases = []struct {
		name     string
		config   params.PagingConfig
		values   url.Values
		expected params.Paging
	}{
		{
			name:     "absent parameters yield zero skip and default limit",
			config:   params.PagingConfig{},
			values:   url.Values{},
			expected: params.Paging{Skip: 0, Limit: params.DefaultLimit},
		},
		{
			name:     "empty parameters yield zero skip and default limit",
			config:   params.PagingConfig{},
			values:   url.Values{"skip": []string{""}, "limit": []string{""}},
			expected: params.Paging{Skip: 0, Limit: params.DefaultLimit},
		},
		{
			name:     "parses skip",
			config:   params.PagingConfig{},
			values:   url.Values{"skip": []string{"25"}},
			expected: params.Paging{Skip: 25, Limit: params.DefaultLimit},
		},
		{
			name:     "parses limit",
			config:   params.PagingConfig{},
			values:   url.Values{"limit": []string{"250"}},
			expected: params.Paging{Limit: 250},
		},
		{
			name:     "parses skip and limit together",
			config:   params.PagingConfig{},
			values:   url.Values{"skip": []string{"10"}, "limit": []string{"50"}},
			expected: params.Paging{Skip: 10, Limit: 50},
		},
		{
			name:     "applies configured default limit",
			config:   params.PagingConfig{DefaultLimit: 42},
			values:   url.Values{},
			expected: params.Paging{Limit: 42},
		},
		{
			name:     "reads skip from configured skip key",
			config:   params.PagingConfig{SkipKey: "offset"},
			values:   url.Values{"offset": []string{"30"}, "skip": []string{"99"}},
			expected: params.Paging{Skip: 30, Limit: params.DefaultLimit},
		},
		{
			name:     "falls back to default skip key when configured key is not in the request",
			config:   params.PagingConfig{SkipKey: "offset"},
			values:   url.Values{"skip": []string{"99"}},
			expected: params.Paging{Skip: 99, Limit: params.DefaultLimit},
		},
		{
			name:     "does not fall back when configured key is in the request but empty",
			config:   params.PagingConfig{SkipKey: "offset"},
			values:   url.Values{"offset": []string{""}, "skip": []string{"99"}},
			expected: params.Paging{Skip: 0, Limit: params.DefaultLimit},
		},
		{
			name:     "allows unlimited results when configured",
			config:   params.PagingConfig{AllowUnlimited: true},
			values:   url.Values{"limit": []string{"-1"}},
			expected: params.Paging{Limit: params.UnlimitedResults},
		},
		{
			name:     "ignores unrelated parameters",
			config:   params.PagingConfig{},
			values:   url.Values{"name": []string{"eq:alice"}, "sort_by": []string{"-name"}},
			expected: params.Paging{Limit: params.DefaultLimit},
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			paging, err := params.NewPagingParser(testCase.config).ParseAndValidate(testCase.values)

			require.NoError(t, err)
			require.Equal(t, testCase.expected, paging)
		})
	}
}

func TestPagingParseAndValidate_Errors(t *testing.T) {
	var cases = []struct {
		name          string
		config        params.PagingConfig
		values        url.Values
		sentinel      error
		parameterName string
		rawValue      string
	}{
		{
			name:          "rejects non-numeric skip",
			config:        params.PagingConfig{},
			values:        url.Values{"skip": []string{"invalid"}},
			sentinel:      params.ErrInvalidInteger,
			parameterName: "skip",
			rawValue:      "invalid",
		},
		{
			name:          "rejects negative skip",
			config:        params.PagingConfig{},
			values:        url.Values{"skip": []string{"-5"}},
			sentinel:      params.ErrInvalidInteger,
			parameterName: "skip",
			rawValue:      "-5",
		},
		{
			name:          "rejects malformed skip under configured skip key",
			config:        params.PagingConfig{SkipKey: "offset"},
			values:        url.Values{"offset": []string{"invalid"}},
			sentinel:      params.ErrInvalidInteger,
			parameterName: "offset",
			rawValue:      "invalid",
		},
		{
			name:          "rejects malformed skip under fallback skip key",
			config:        params.PagingConfig{SkipKey: "offset"},
			values:        url.Values{"skip": []string{"invalid"}},
			sentinel:      params.ErrInvalidInteger,
			parameterName: "skip",
			rawValue:      "invalid",
		},
		{
			name:          "rejects non-numeric limit",
			config:        params.PagingConfig{},
			values:        url.Values{"limit": []string{"invalid"}},
			sentinel:      params.ErrInvalidInteger,
			parameterName: "limit",
			rawValue:      "invalid",
		},
		{
			name:          "rejects unlimited results by default",
			config:        params.PagingConfig{},
			values:        url.Values{"limit": []string{"-1"}},
			sentinel:      params.ErrInvalidInteger,
			parameterName: "limit",
			rawValue:      "-1",
		},
		{
			name:          "rejects negative limit even when unlimited is allowed",
			config:        params.PagingConfig{AllowUnlimited: true},
			values:        url.Values{"limit": []string{"-2"}},
			sentinel:      params.ErrInvalidInteger,
			parameterName: "limit",
			rawValue:      "-2",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var validationErr *params.PagingValidationError

			_, err := params.NewPagingParser(testCase.config).ParseAndValidate(testCase.values)

			require.Error(t, err)
			require.ErrorIs(t, err, testCase.sentinel)
			require.True(t, errors.As(err, &validationErr))
			require.Equal(t, testCase.parameterName, validationErr.Parameter)
			require.Equal(t, testCase.rawValue, validationErr.Value)
		})
	}
}
