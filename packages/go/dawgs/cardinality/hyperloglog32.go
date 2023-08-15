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

package cardinality

import (
	"encoding/binary"
	"sync"

	"github.com/axiomhq/hyperloglog"
)

var (
	size4BufferPool = &sync.Pool{
		New: func() any {
			// Select a new buffer the size of a single unit32 in bytes
			return make([]byte, 4)
		},
	}
)

type hyperLogLog32 struct {
	sketch *hyperloglog.Sketch
}

func NewHyperLogLog32() Simplex[uint32] {
	return hyperLogLog32{
		sketch: hyperloglog.New14(),
	}
}

func (s hyperLogLog32) Clone() Simplex[uint32] {
	return hyperLogLog32{
		sketch: s.sketch.Clone(),
	}
}

func (s hyperLogLog32) Clear() {
	s.sketch = hyperloglog.New14()
}

func (s hyperLogLog32) Add(values ...uint32) {
	buffer := size4BufferPool.Get()
	byteBuffer := buffer.([]byte)
	defer size4BufferPool.Put(buffer)

	for idx := 0; idx < len(values); idx++ {
		binary.LittleEndian.PutUint32(byteBuffer, values[idx])
		s.sketch.Insert(byteBuffer)
	}
}

func (s hyperLogLog32) Or(provider Provider[uint32]) {
	switch typedProvider := provider.(type) {
	case hyperLogLog32:
		s.sketch.Merge(typedProvider.sketch)

	case Duplex[uint32]:
		typedProvider.Each(func(nextValue uint32) (bool, error) {
			s.Add(nextValue)
			return true, nil
		})
	}
}

func (s hyperLogLog32) Cardinality() uint64 {
	return s.sketch.Estimate()
}
