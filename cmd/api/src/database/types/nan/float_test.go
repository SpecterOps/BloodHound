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

package nan_test

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/specterops/bloodhound/src/database/types/nan"
	"github.com/stretchr/testify/require"
)

func TestJSONFloat_MarshalJSON_NaN(t *testing.T) {
	nan := nan.Float64(math.NaN())
	marshaledNan, err := json.Marshal(nan)
	require.Nil(t, err)

	floatMarshaledNan, err := strconv.ParseFloat(string(marshaledNan), 64)
	require.Nil(t, err)
	require.Equal(t, float64(0), floatMarshaledNan)
}

func TestJSONFloat_MarshalJSON(t *testing.T) {
	values := []float64{12.34, -12.34, 0}
	for _, value := range values {
		marshaledValue, err := json.Marshal(value)
		require.Nil(t, err)

		floatValue, err := strconv.ParseFloat(string(marshaledValue), 64)
		require.Nil(t, err)
		require.Equal(t, value, floatValue)
	}
}

func TestJSONFloat_UnmarshalJSON_NaN(t *testing.T) {
	type data struct {
		Value nan.Float64 `json:"value"`
	}

	input := data{Value: nan.Float64(math.NaN())}
	// We know the marshalling works well from the tests above
	marshaledInput, err := json.Marshal(input)
	require.Nil(t, err)

	output := data{}
	err = json.Unmarshal(marshaledInput, &output)
	require.Nil(t, err)
	require.Equal(t, nan.Float64(0), output.Value)
}

func TestJSONFloat_UnmarshalJSON(t *testing.T) {
	type data struct {
		Value nan.Float64 `json:"value"`
	}

	values := []data{
		{Value: 23.45},
		{Value: -23.45},
		{Value: 0},
	}

	for _, input := range values {
		// We know the marshalling works well from the tests above
		marshaledInput, err := json.Marshal(input)
		require.Nil(t, err)

		output := data{}
		err = json.Unmarshal(marshaledInput, &output)
		require.Nil(t, err)
		require.Equal(t, input.Value, output.Value)
	}
}
