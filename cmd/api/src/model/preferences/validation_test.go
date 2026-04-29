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
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestValidatePreferences(t *testing.T) {
	// add new preferences here as they're added to the preferences schema
	validKeyValues := map[string]any{
		"dark_mode":        true,
		"preferred_domain": "S-1-4-29-98765",
	}

	// test each preference validation individually
	for key, value := range validKeyValues {
		t.Run("valid "+key, func(t *testing.T) {
			input := model.Preferences{
				key: model.PreferenceItem{Value: value},
			}
			err := ValidatePreferences(input)
			require.NoError(t, err)
		})
	}

	t.Run("all valid keys together", func(t *testing.T) {
		input := model.Preferences{}
		for key, value := range validKeyValues {
			input[key] = model.PreferenceItem{Value: value}
		}
		err := ValidatePreferences(input)
		require.NoError(t, err)
	})

	t.Run("empty preferences succeeds", func(t *testing.T) {
		input := model.Preferences{}
		err := ValidatePreferences(input)
		require.NoError(t, err)
	})

	t.Run("unknown key fails to validate", func(t *testing.T) {
		input := model.Preferences{
			"random_pref": model.PreferenceItem{Value: true},
		}
		err := ValidatePreferences(input)
		require.Error(t, err)
	})

	t.Run("capitalized key fails to validate", func(t *testing.T) {
		input := model.Preferences{
			"Dark_Mode": model.PreferenceItem{Value: true},
		}
		err := ValidatePreferences(input)
		require.Error(t, err)
	})

	t.Run("empty string key fails to validate", func(t *testing.T) {
		input := model.Preferences{
			"": model.PreferenceItem{Value: true},
		}
		err := ValidatePreferences(input)
		require.Error(t, err)
	})
}

func TestValidatePreferences_Value(t *testing.T) {
	invalidTypes := []struct {
		name  string
		key   string
		value any
	}{
		{
			name:  "dark_mode with string value",
			key:   "dark_mode",
			value: "not bool",
		},
		{
			name:  "dark_mode with int value",
			key:   "dark_mode",
			value: 00,
		},
		{
			name:  "preferred_domain with bool value",
			key:   "preferred_domain",
			value: false,
		},
		{
			name:  "preferred_domain with int value",
			key:   "preferred_domain",
			value: 23752756192,
		},
	}

	for _, testCase := range invalidTypes {
		t.Run(testCase.name, func(t *testing.T) {
			input := model.Preferences{
				testCase.key: model.PreferenceItem{Value: testCase.value},
			}
			err := ValidatePreferences(input)
			require.Error(t, err)
		})
	}

	t.Run("missing value field fails validation", func(t *testing.T) {
		input := model.Preferences{
			"preferred_domain": model.PreferenceItem{},
		}
		err := ValidatePreferences(input)
		require.Error(t, err)
	})

	t.Run("nil value fails validation", func(t *testing.T) {
		input := model.Preferences{
			"dark_mode": model.PreferenceItem{Value: nil},
		}
		err := ValidatePreferences(input)
		fmt.Println(err)
		require.Error(t, err)
	})
}

func TestLoadSchema(t *testing.T) {
	t.Run("loads valid schema without error", func(t *testing.T) {
		schema, err := loadSchema("preferences_schema.json")
		require.NoError(t, err)
		require.NotNil(t, schema)
	})
	t.Run("returns error for a non-existent schema file", func(t *testing.T) {
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
