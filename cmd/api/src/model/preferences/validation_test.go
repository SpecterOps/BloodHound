// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package preferences

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestLoadSchema(t *testing.T) {
	t.Run("success: loads valid schema without error", func(t *testing.T) {
		schema, err := loadSchema("preferences_schema.json")
		require.NoError(t, err)
		require.NotNil(t, schema)
	})
	t.Run("failure: returns an error for a non-existent schema file", func(t *testing.T) {
		schema, err := loadSchema("bad_schema.json")
		require.Error(t, err)
		require.Nil(t, schema)
	})
}

func TestTransformPreferences(t *testing.T) {
	testPreference1 := model.Preferences{
		"dark_mode": model.PreferenceItem{Value: true},
	}
	testPreference2 := model.Preferences{
		"dark_mode":        model.PreferenceItem{Value: true},
		"preferred_domain": model.PreferenceItem{Value: "S-1-4-22-01234"},
	}
	emptyPreferences := model.Preferences{}

	expected1 := map[string]any{
		"dark_mode": map[string]any{"value": true},
	}
	expected2 := map[string]any{
		"dark_mode":        map[string]any{"value": true},
		"preferred_domain": map[string]any{"value": "S-1-4-22-01234"},
	}
	expectedEmpty := map[string]any{}

	t.Run("transforms a single preference into map[string]any", func(t *testing.T) {
		result, err := transformPreferences(testPreference1)
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.Equal(t, expected1, result)
	})
	t.Run("transforms multiple preferences into map[string]any", func(t *testing.T) {
		result, err := transformPreferences(testPreference2)
		require.NoError(t, err)
		require.Len(t, result, 2)
		require.Equal(t, expected2, result)
	})
	t.Run("transforms empty preferences into empty map", func(t *testing.T) {
		result, err := transformPreferences(emptyPreferences)
		require.NoError(t, err)
		require.Empty(t, result)
		require.Equal(t, expectedEmpty, result)
	})
}
