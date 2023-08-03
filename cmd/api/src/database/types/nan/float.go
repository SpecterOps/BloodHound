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

package nan

import (
	"encoding/json"
	"math"
)

// Float64 is a regular float64 data type intended to be used for floating point
// data types. It comes with extended support for JSON encoding/decoding of
// NaN values, which will be returned as 0 when marshalled or unmarshalled.
type Float64 float64

func (s Float64) MarshalJSON() ([]byte, error) {
	v := float64(s)
	if math.IsNaN(v) {
		v = 0.0
	}
	return json.Marshal(v)
}

func (s *Float64) UnmarshalJSON(input []byte) error {
	var floatValue float64
	if string(input) == "NaN" {
		*s = Float64(math.NaN())
		return nil
	} else {
		if err := json.Unmarshal(input, &floatValue); err != nil {
			return err
		}
		*s = Float64(floatValue)
		return nil
	}
}
