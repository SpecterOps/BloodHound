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

package sorts_test

import (
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/packages/go/sorts"
	"github.com/stretchr/testify/require"
)

// fakeSortable is a minimal Sortable used to drive ParseAndValidate without coupling the test to any
// persistence layer.
type fakeSortable struct {
	sortableFields map[string]bool
}

func (s fakeSortable) IsSortable(field string) bool {
	return s.sortableFields[field]
}

func TestParseAndValidate_Parsing(t *testing.T) {
	var (
		parser   = sorts.NewQueryParameterSortParser()
		sortable = fakeSortable{sortableFields: map[string]bool{"objectid": true, "name": true}}
	)

	// These cases mirror the legacy api sort parser tests to vet behavioral parity.
	t.Run("parses an ascending field", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{"sort_by": {"objectid"}}, sortable)
		require.NoError(t, err)
		require.Len(t, parsed, 1)
		require.Equal(t, "objectid", parsed[0].Field)
		require.Equal(t, sorts.Ascending, parsed[0].Direction)
	})

	t.Run("parses a descending field denoted by a leading -", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{"sort_by": {"-objectid"}}, sortable)
		require.NoError(t, err)
		require.Len(t, parsed, 1)
		require.Equal(t, "objectid", parsed[0].Field)
		require.Equal(t, sorts.Descending, parsed[0].Direction)
	})

	t.Run("parses multiple fields in order", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{"sort_by": {"objectid", "-name"}}, sortable)
		require.NoError(t, err)
		require.Len(t, parsed, 2)
		require.Equal(t, "objectid", parsed[0].Field)
		require.Equal(t, sorts.Ascending, parsed[0].Direction)
		require.Equal(t, "name", parsed[1].Field)
		require.Equal(t, sorts.Descending, parsed[1].Direction)
	})

	t.Run("returns an empty result when no sort_by parameter is supplied", func(t *testing.T) {
		parsed, err := parser.ParseAndValidate(url.Values{}, sortable)
		require.NoError(t, err)
		require.Empty(t, parsed)
	})
}

func TestParseAndValidate_Errors(t *testing.T) {
	var (
		parser   = sorts.NewQueryParameterSortParser()
		sortable = fakeSortable{sortableFields: map[string]bool{"objectid": true}}
	)

	t.Run("ErrFieldNotSortable for an unsortable field, preserving the offending field", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"sort_by": {"invalidField"}}, sortable)
		require.ErrorIs(t, err, sorts.ErrFieldNotSortable)

		var validationError *sorts.ValidationError
		require.ErrorAs(t, err, &validationError)
		require.Equal(t, "invalidField", validationError.Field)
	})

	t.Run("ErrFieldNotSortable applies to the field with its descending prefix stripped", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"sort_by": {"-invalidField"}}, sortable)
		require.ErrorIs(t, err, sorts.ErrFieldNotSortable)

		var validationError *sorts.ValidationError
		require.ErrorAs(t, err, &validationError)
		require.Equal(t, "invalidField", validationError.Field)
	})

	t.Run("ErrFieldEmpty for an empty sort_by value", func(t *testing.T) {
		_, err := parser.ParseAndValidate(url.Values{"sort_by": {""}}, sortable)
		require.ErrorIs(t, err, sorts.ErrFieldEmpty)
	})
}
