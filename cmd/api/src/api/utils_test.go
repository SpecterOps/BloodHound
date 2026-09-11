// Copyright 2024 Specter Ops, Inc.
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

package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseOptionalBool(t *testing.T) {
	result, err := ParseOptionalBool("true", false)
	require.NoError(t, err)
	assert.Equal(t, true, result)
}

func TestParseOptionalBoolEmptyValue(t *testing.T) {
	result, _ := ParseOptionalBool("", true)
	assert.Equal(t, true, result)
}

func TestParseOptionalBoolMisspelledValue(t *testing.T) {
	result, err := ParseOptionalBool("trueee", false)
	assert.Error(t, err)
	assert.Equal(t, false, result)
}

func TestFilterStructSlice(t *testing.T) {
	type TestStruct struct {
		val int
	}

	var (
		structs  = []TestStruct{{val: 0}, {val: 1}, {val: 2}, {val: 3}}
		expected = []TestStruct{{val: 0}, {val: 1}}
	)

	result := FilterStructSlice(structs, func(testStruct TestStruct) bool {
		return testStruct.val < 2
	})
	assert.ElementsMatch(t, expected, result)
}
