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

package size

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validateMapSizes[K comparable, V any](t *testing.T, mapInst map[K]V, expectedKeySize, expectedElementSize uint8) {
	var (
		// Zero size key and element map definition
		structStructMap                           = map[struct{}]struct{}{}
		structStructMapType, structStructHmapInst = mapTypeAndValue(structStructMap)
		structStructHmapSize                      = Of(*structStructHmapInst)

		// Pull the runtime type definition for the map as well as its backing struct
		mapType, hmapInst = mapTypeAndValue(mapInst)

		// The expected bucket size is calculated as the size of the key and element types multiplied by the number of
		// key/element pairs in a bucket
		expectedBucketSize = uint16(mapType.keysize+mapType.elemsize)*bucketCnt + structStructMapType.bucketsize
	)

	assert.Equal(t, expectedKeySize, mapType.keysize)
	assert.Equal(t, expectedElementSize, mapType.elemsize)
	assert.Equal(t, expectedBucketSize, mapType.bucketsize)
	assert.Equal(t, structStructHmapSize, Of(*hmapInst))
	assert.Equal(t, Size(mapType.bucketsize)+structStructHmapSize, OfMap(mapInst))
}

func TestMapSizes(t *testing.T) {
	validateMapSizes(t, map[struct{}]struct{}{}, 0, 0)
	validateMapSizes(t, map[int8]struct{}{}, 1, 0)
	validateMapSizes(t, map[int16]struct{}{}, 2, 0)
	validateMapSizes(t, map[int32]struct{}{}, 4, 0)
	validateMapSizes(t, map[int64]struct{}{}, 8, 0)
	validateMapSizes(t, map[string]struct{}{}, 16, 0)

	validateMapSizes(t, map[struct{}]string{}, 0, 16)
	validateMapSizes(t, map[int8]string{}, 1, 16)
	validateMapSizes(t, map[int16]string{}, 2, 16)
	validateMapSizes(t, map[int32]string{}, 4, 16)
	validateMapSizes(t, map[int64]string{}, 8, 16)
	validateMapSizes(t, map[string]string{}, 16, 16)

	validateMapSizes(t, map[struct{}]any{}, 0, 16)
	validateMapSizes(t, map[int8]any{}, 1, 16)
	validateMapSizes(t, map[int16]any{}, 2, 16)
	validateMapSizes(t, map[int32]any{}, 4, 16)
	validateMapSizes(t, map[int64]any{}, 8, 16)
	validateMapSizes(t, map[string]any{}, 16, 16)

	var (
		int8StringMap = map[int8]string{
			0: "test",
			1: "test",
			2: "test",
			3: "test",
		}

		int8StringMapSize = OfMap(int8StringMap)

		// Subtract the stand-alone map size from the size of the map with value contents taken into consideration to
		// get the raw content size of the four strings
		int8StringMapValuesSize = OfMapValues(int8StringMap) - int8StringMapSize
	)

	validateMapSizes(t, int8StringMap, 1, 16)

	// Each string contains 16 bytes for the header and 4 additional bytes of data
	require.Equal(t, Size(4)*4, int8StringMapValuesSize)
}

func TestOfValueSlice(t *testing.T) {
	var (
		slice        = make([]int64, 0, 32)
		expectedSize = cap(slice)*8 + 24
	)

	require.Equal(t, expectedSize, int(OfSlice(slice)))
}
