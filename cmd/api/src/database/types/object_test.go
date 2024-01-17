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

package types

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTypes_Scan_InvalidInput(t *testing.T) {
	object := JSONUntypedObject{}
	err := object.Scan(json.RawMessage(`{"a":"b"}`))
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "failed to unmarshal JSONB value")
}

func TestTypes_Scan_Success(t *testing.T) {
	object := JSONUntypedObject{}

	jsonInput := []byte(`{"a":"b"}`)

	err := object.Scan(jsonInput)
	require.Nil(t, err)

	value, err := object.Value()
	require.Nil(t, err)
	require.Equal(t, jsonInput, value)
}

func TestTypes_Value(t *testing.T) {
	object := JSONUntypedObject(map[string]any{"key": "value"})

	value, err := object.Value()
	require.Nil(t, err)

	var result struct {
		Key string `json:"key"`
	}

	err = json.Unmarshal(value.([]byte), &result)
	require.Nil(t, err)
	require.Equal(t, "value", result.Key)
}
