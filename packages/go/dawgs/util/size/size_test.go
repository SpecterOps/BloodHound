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

package size_test

import (
	"github.com/specterops/bloodhound/dawgs/util/size"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestOfAny(t *testing.T) {
	var value = "123"

	// Size of pointer
	require.Equal(t, size.Size(0x10), size.OfAny(&value))

	// Size of empty struct
	require.Equal(t, size.Size(0x00), size.OfAny(struct{}{}))

	// Size of variable types
	require.Equal(t, size.Size(0x01), size.OfAny(true))
	require.Equal(t, size.Size(0x08), size.OfAny(uintptr(0)))
	require.Equal(t, size.Size(0x08), size.OfAny(complex64(1)))
	require.Equal(t, size.Size(0x10), size.OfAny(complex128(2)))

	require.Equal(t, size.Size(0x40), size.OfAny("test"))
	require.Equal(t, size.Size(0x08), size.OfAny(uint(0)))
	require.Equal(t, size.Size(0x01), size.OfAny(uint8(1)))
	require.Equal(t, size.Size(0x02), size.OfAny(uint16(2)))
	require.Equal(t, size.Size(0x04), size.OfAny(uint32(3)))
	require.Equal(t, size.Size(0x08), size.OfAny(uint64(4)))

	require.Equal(t, size.Size(0x08), size.OfAny(int(0)))
	require.Equal(t, size.Size(0x01), size.OfAny(int8(1)))
	require.Equal(t, size.Size(0x02), size.OfAny(int16(2)))
	require.Equal(t, size.Size(0x04), size.OfAny(int32(3)))
	require.Equal(t, size.Size(0x08), size.OfAny(int64(4)))

	require.Equal(t, size.Size(0x04), size.OfAny(float32(6.6)))
	require.Equal(t, size.Size(0x08), size.OfAny(float64(7.7)))

	require.Equal(t, size.Size(0x30), size.OfAny([]int{1, 2, 3}))
	require.Equal(t, size.Size(0x1b), size.OfAny([]int8{1, 2, 3}))
	require.Equal(t, size.Size(0x1e), size.OfAny([]int16{1, 2, 3}))
	require.Equal(t, size.Size(0x24), size.OfAny([]int32{1, 2, 3}))
	require.Equal(t, size.Size(0x30), size.OfAny([]int64{1, 2, 3}))

	require.Equal(t, size.Size(0x30), size.OfAny([]uint{1, 2, 3}))
	require.Equal(t, size.Size(0x1b), size.OfAny([]uint8{1, 2, 3}))
	require.Equal(t, size.Size(0x1e), size.OfAny([]uint16{1, 2, 3}))
	require.Equal(t, size.Size(0x24), size.OfAny([]uint32{1, 2, 3}))
	require.Equal(t, size.Size(0x30), size.OfAny([]uint64{1, 2, 3}))

	require.Equal(t, size.Size(0x24), size.OfAny([]float32{1, 2, 3}))
	require.Equal(t, size.Size(0x30), size.OfAny([]float64{1, 2, 3}))
	require.Equal(t, size.Size(0x30), size.OfAny([]uintptr{1, 2, 3}))
	require.Equal(t, size.Size(0x30), size.OfAny([]complex64{1, 2, 3}))
	require.Equal(t, size.Size(0x48), size.OfAny([]complex128{1, 2, 3}))

	require.Equal(t, size.Size(0x108), size.OfAny([]string{"a", "baa", "long string"}))
	require.Equal(t, size.Size(0x1b), size.OfAny([]bool{true, false, true}))

	require.Equal(t, size.Size(0x44), size.OfAny([]any{"aa", 123, false}))
}

func TestOfValueSlice(t *testing.T) {
	var (
		slice        = make([]int64, 0, 32)
		expectedSize = cap(slice)*8 + 24
	)

	require.Equal(t, expectedSize, int(size.OfSlice(slice)))
}
