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

package cardinality_test

import (
	"encoding/binary"
	"sync"
	"testing"

	"github.com/axiomhq/hyperloglog"
)

var (
	size4BufferValuePool = &sync.Pool{
		New: func() any {
			// Select a new buffer the size of a single unit32 in bytes
			return make([]byte, 4)
		},
	}
	size4BufferPointerPool = &sync.Pool{
		New: func() any {
			// Select a new buffer the size of a single unit32 in bytes
			buffer := make([]byte, 4)
			return &buffer
		},
	}
)

func generateIDValues(size uint32) []uint32 {
	var (
		iterator uint32
		ids      = make([]uint32, size)
	)

	for iterator = 0; iterator < size; iterator++ {
		ids[iterator] = uint32(iterator)
	}

	return ids
}

type hyperLogLog32 struct {
	sketch *hyperloglog.Sketch
}

func (s hyperLogLog32) AddUint32ValuePool(values ...uint32) {
	buffer := size4BufferValuePool.Get().([]byte)
	defer size4BufferValuePool.Put(buffer)

	for idx := 0; idx < len(values); idx++ {
		binary.LittleEndian.PutUint32(buffer, values[idx])
		s.sketch.Insert(buffer)
	}
}

func (s hyperLogLog32) AddUint32PointerPool(values ...uint32) {
	buffer := size4BufferPointerPool.Get().(*[]byte)
	defer size4BufferPointerPool.Put(buffer)

	for idx := 0; idx < len(values); idx++ {
		binary.LittleEndian.PutUint32(*buffer, values[idx])
		s.sketch.Insert(*buffer)
	}
}

func (s hyperLogLog32) AddUint32ConvertAndStore(values ...uint32) {
	buffer := size4BufferValuePool.Get()
	bufferSlice := buffer.([]byte)
	defer size4BufferValuePool.Put(buffer)

	for idx := 0; idx < len(values); idx++ {
		binary.LittleEndian.PutUint32(bufferSlice, values[idx])
		s.sketch.Insert(bufferSlice)
	}
}

func (s hyperLogLog32) AddUint32NoPool(values ...uint32) {
	buffer := make([]byte, 4)

	for idx := 0; idx < len(values); idx++ {
		binary.LittleEndian.PutUint32(buffer, values[idx])
		s.sketch.Insert(buffer)
	}
}

// 1000 faked ids
func BenchmarkAddUint32_ValuePool_1000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ValuePool(uint32Slice...)
	}
}

func BenchmarkAddUint32_PointerPool_1000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32PointerPool(uint32Slice...)
	}
}

func BenchmarkAddUint32_ConvertAndStore_1000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ConvertAndStore(uint32Slice...)
	}
}

func BenchmarkAddUint32_NoPool_1000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32NoPool(uint32Slice...)
	}
}

// 10000 faked ids
func BenchmarkAddUint32_ValuePool_10000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ValuePool(uint32Slice...)
	}
}

func BenchmarkAddUint32_PointerPool_10000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32PointerPool(uint32Slice...)
	}
}

func BenchmarkAddUint32_ConvertAndStore_10000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ConvertAndStore(uint32Slice...)
	}
}

func BenchmarkAddUint32_NoPool_10000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32NoPool(uint32Slice...)
	}
}

// 100000 faked ids
func BenchmarkAddUint32_ValuePool_100000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(100000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ValuePool(uint32Slice...)
	}
}

func BenchmarkAddUint32_PointerPool_100000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(100000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32PointerPool(uint32Slice...)
	}
}

func BenchmarkAddUint32_ConvertAndStore_100000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(100000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ConvertAndStore(uint32Slice...)
	}
}

func BenchmarkAddUint32_NoPool_100000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(100000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32NoPool(uint32Slice...)
	}
}

// 1000000 faked ids
func BenchmarkAddUint32_ValuePool_1000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ValuePool(uint32Slice...)
	}
}

func BenchmarkAddUint32_PointerPool_1000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32PointerPool(uint32Slice...)
	}
}

func BenchmarkAddUint32_ConvertAndStore_1000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ConvertAndStore(uint32Slice...)
	}
}

func BenchmarkAddUint32_NoPool_1000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(1000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32NoPool(uint32Slice...)
	}
}

// 10000000 faked ids
func BenchmarkAddUint32_ValuePool_10000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ValuePool(uint32Slice...)
	}
}

func BenchmarkAddUint32_PointerPool_10000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32PointerPool(uint32Slice...)
	}
}

func BenchmarkAddUint32_ConvertAndStore_10000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32ConvertAndStore(uint32Slice...)
	}
}

func BenchmarkAddUint32_NoPool_10000000(b *testing.B) {
	var (
		uint32Slice = generateIDValues(10000000)
		sketch      = hyperLogLog32{
			sketch: hyperloglog.New14(),
		}
	)

	for n := 0; n < b.N; n++ {
		sketch.AddUint32NoPool(uint32Slice...)
	}
}
