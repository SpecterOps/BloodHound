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
	size8BufferPool = &sync.Pool{
		New: func() any {
			// Select a new buffer the size of a single uint64 in bytes
			return make([]byte, 8)
		},
	}
)

type hyperLogLog64 struct {
	sketch *hyperloglog.Sketch
}

func NewHyperLogLog64() Simplex[uint64] {
	return hyperLogLog64{
		sketch: hyperloglog.New16(),
	}
}

func (s hyperLogLog64) Clone() Simplex[uint64] {
	return hyperLogLog64{
		sketch: s.sketch.Clone(),
	}
}

func (s hyperLogLog64) Clear() {
	s.sketch = hyperloglog.New16()
}

func (s hyperLogLog64) Add(values ...uint64) {
	buffer := size8BufferPool.Get()
	byteBuffer := buffer.([]byte)
	defer size8BufferPool.Put(buffer)

	for idx := 0; idx < len(values); idx++ {
		binary.LittleEndian.PutUint64(byteBuffer, values[idx])
		s.sketch.Insert(byteBuffer)
	}
}

func (s hyperLogLog64) Or(provider Provider[uint64]) {
	switch typedProvider := provider.(type) {
	case hyperLogLog64:
		s.sketch.Merge(typedProvider.sketch)

	case Duplex[uint64]:
		typedProvider.Each(func(nextValue uint64) (bool, error) {
			s.Add(nextValue)
			return true, nil
		})
	}
}

func (s hyperLogLog64) Cardinality() uint64 {
	return s.sketch.Estimate()
}
