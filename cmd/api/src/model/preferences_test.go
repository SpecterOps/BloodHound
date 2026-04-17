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

package model_test

import (
	"encoding/json"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestPreferences_Scan_InvalidInput(t *testing.T) {
	preferences := model.Preferences{}
	err := preferences.Scan(json.RawMessage(`{"random":"random"}`))
	require.Error(t, err)
	require.Contains(t, err.Error(), "expected []byte representation of JSONB")
}

func TestPreferences_Scan_InvalidJSON(t *testing.T) {
	preferences := model.Preferences{}
	err := preferences.Scan([]byte(`{not valid json}`))
	require.Error(t, err)
}

func TestPreferences_Scan_EmptyObject(t *testing.T) {
	preferences := model.Preferences{}
	err := preferences.Scan([]byte(`{}`))
	require.NoError(t, err)
	require.Empty(t, preferences)
}

func TestPreferences_Scan_Success(t *testing.T) {
	preferences := model.Preferences{}
	jsonbInput := []byte(`{
		"dark_mode": {
			"value": true,
			"enterprise": false
		},
		"preferred_domain": {
			"value": "S-1-5-24-1234567890",
			"enterprise": true
		}
	}`)

	err := preferences.Scan(jsonbInput)
	require.NoError(t, err)
	require.Len(t, preferences, 2)

	darkMode, exists := preferences["dark_mode"]
	require.True(t, exists)
	require.Equal(t, true, darkMode.Value)
	require.Equal(t, false, darkMode.Enterprise)

	preferredDomain, exists := preferences["preferred_domain"]
	require.True(t, exists)
	require.Equal(t, "S-1-5-24-1234567890", preferredDomain.Value)
	require.Equal(t, true, preferredDomain.Enterprise)
}

func TestPreferences_Value_Success(t *testing.T) {
	preferences := model.Preferences{
		"dark_mode": model.PreferenceItem{
			Value:      true,
			Enterprise: false,
		},
	}

	value, err := preferences.Value()
	require.NoError(t, err)

	var result model.Preferences
	err = json.Unmarshal(value.([]byte), &result)
	require.NoError(t, err)
	require.Len(t, result, 1)

	darkMode, exists := result["dark_mode"]
	require.True(t, exists)
	require.Equal(t, true, darkMode.Value)
	require.Equal(t, false, darkMode.Enterprise)
}

func TestPreferences_Value_EmptyPreferences(t *testing.T) {
	preferences := model.Preferences{}

	value, err := preferences.Value()
	require.NoError(t, err)
	require.Equal(t, []byte(`{}`), value)
}

func TestPreferences_Value_NilPreferences(t *testing.T) {
	var preferences model.Preferences

	value, err := preferences.Value()
	require.NoError(t, err)
	require.Equal(t, []byte("{}"), value)
}

func TestPreferences_Scan_Value_RoundTrip(t *testing.T) {
	originalJSONB := []byte(`{
		"dark_mode": {
			"value": true,
			"enterprise": false
		},
		"preferred_domain": {
			"value": "S-1-5-24-1234567890",
			"enterprise": true
		}
	}`)
	preferences := model.Preferences{}

	// unmarshal original jsonb into preferences
	err := preferences.Scan(originalJSONB)
	require.NoError(t, err)

	// marshal newly-created preferences map back to jsonb
	value, err := preferences.Value()
	require.NoError(t, err)

	// final unmarshaling after roundtrip to verify success
	var roundTrip model.Preferences
	err = json.Unmarshal(value.([]byte), &roundTrip)
	require.NoError(t, err)

	require.Equal(t, preferences, roundTrip)
}
