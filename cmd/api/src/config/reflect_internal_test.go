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

package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Test_setRawValue(t *testing.T) {
	var (
		uintValue    uint
		uint8Value   uint8
		uint16Value  uint16
		uint32Value  uint32
		uint64Value  uint64
		intValue     int
		int8Value    int8
		int16Value   int16
		int32Value   int32
		int64Value   int64
		float32Value float32
		float64Value float64
		boolValue    bool
		stringValue  string
		timeDuration time.Duration
	)

	assert.NotNil(t, setRawValue(&uintValue, "not a number"))

	assert.NotNil(t, setRawValue(&uintValue, "not a number"))
	assert.NotNil(t, setRawValue(&uint8Value, "not a number"))
	assert.NotNil(t, setRawValue(&uint16Value, "not a number"))
	assert.NotNil(t, setRawValue(&uint32Value, "not a number"))
	assert.NotNil(t, setRawValue(&uint64Value, "not a number"))
	assert.NotNil(t, setRawValue(&intValue, "not a number"))
	assert.NotNil(t, setRawValue(&int8Value, "not a number"))
	assert.NotNil(t, setRawValue(&int16Value, "not a number"))
	assert.NotNil(t, setRawValue(&int32Value, "not a number"))
	assert.NotNil(t, setRawValue(&int64Value, "not a number"))
	assert.NotNil(t, setRawValue(&float32Value, "not a number"))
	assert.NotNil(t, setRawValue(&float64Value, "not a number"))
	assert.NotNil(t, setRawValue(&boolValue, "not a boolean"))
	assert.NotNil(t, setRawValue(&timeDuration, "not a time duration"))

	// Test the error case where the value is not addressable
	assert.NotNil(t, setRawValue(timeDuration, "not a time duration"))

	assert.Nil(t, setRawValue(&uintValue, "100000"))
	assert.Nil(t, setRawValue(&uint8Value, "128"))
	assert.Nil(t, setRawValue(&uint16Value, "4500"))
	assert.Nil(t, setRawValue(&uint32Value, "100000"))
	assert.Nil(t, setRawValue(&uint64Value, "5000000000"))
	assert.Nil(t, setRawValue(&intValue, "100000"))
	assert.Nil(t, setRawValue(&int8Value, "127"))
	assert.Nil(t, setRawValue(&int16Value, "4500"))
	assert.Nil(t, setRawValue(&int32Value, "100000"))
	assert.Nil(t, setRawValue(&int64Value, "5000000000"))
	assert.Nil(t, setRawValue(&float32Value, "128.821"))
	assert.Nil(t, setRawValue(&float64Value, "5000000000.1"))
	assert.Nil(t, setRawValue(&boolValue, "false"))
	assert.Nil(t, setRawValue(&stringValue, "string"))
	assert.Nil(t, setRawValue(&timeDuration, "5000ns"))

	assert.Equal(t, uint(100000), uintValue)
	assert.Equal(t, uint8(128), uint8Value)
	assert.Equal(t, uint16(4500), uint16Value)
	assert.Equal(t, uint32(100000), uint32Value)
	assert.Equal(t, uint64(5000000000), uint64Value)
	assert.Equal(t, int(100000), intValue)
	assert.Equal(t, int8(127), int8Value)
	assert.Equal(t, int16(4500), int16Value)
	assert.Equal(t, int32(100000), int32Value)
	assert.Equal(t, int64(5000000000), int64Value)
	assert.Equal(t, float32(128.821), float32Value)
	assert.Equal(t, float64(5000000000.1), float64Value)
	assert.Equal(t, false, boolValue)
	assert.Equal(t, "string", stringValue)
	assert.Equal(t, time.Nanosecond*5000, timeDuration)
}
