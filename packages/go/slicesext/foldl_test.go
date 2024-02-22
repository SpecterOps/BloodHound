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

package slicesext_test

import (
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/slicesext"
	"github.com/stretchr/testify/require"
)

func TestFoldl(t *testing.T) {
	type Accumulator[A, B any] func(A, B) B
	type Case[A, B any] struct {
		Name        string
		Accumulator Accumulator[A, B]
		InitVal     B
		List        []A
		Output      B
	}

	cases := []any{
		Case[int, int]{"Foldl(64,list,Div) to be 2", div, 64, []int{4, 2, 4}, 2},
		Case[int, int]{"Foldl(3,list,Div) to be 3", div, 3, []int{}, 3},
		Case[int, int]{"Foldl(5,list,Max) to be 5", max, 5, []int{1, 2, 3, 4}, 5},
		Case[int, int]{"Foldl(5,list,Max) to be 7", max, 5, []int{1, 2, 3, 4, 5, 6, 7}, 7},
		Case[int, int]{"Foldl(4,list,2a+b) to be 43", func(a, b int) int { return 2*a + b }, 4, []int{1, 2, 3}, 43},
	}

	for _, c := range cases {
		switch c := c.(type) {
		case Case[int, int]:
			t.Run(c.Name, func(t *testing.T) {
				require.Equal(t, c.Output, slicesext.Foldl(c.InitVal, c.List, c.Accumulator))
				require.Equal(t, c.Output, slicesext.FoldlLazy(c.InitVal, c.List, c.Accumulator))
			})
		}
	}

	t.Run("should handle multiple reducers", func(t *testing.T) {
		// ((((5 +1)*1)+1)*1)+2)*2 = 18
		require.Equal(t, 18, slicesext.Foldl(5, []int{1, 1, 2}, sum, mult))
		require.Equal(t, 18, slicesext.FoldlLazy(5, []int{1, 1, 2}, sum, mult))
	})
}

func FuzzFoldl(f *testing.F) {
	f.Add([]byte{5}, []byte{1, 2, 3, 4})
	f.Fuzz(func(t *testing.T, a []byte, b []byte) {
		out := slicesext.Foldl(a, b, func(x []byte, y byte) []byte {
			return cons(x, y)
		})
		require.Equal(t, len(a)+len(b), len(out))
	})
}

func FuzzFoldlLazy(f *testing.F) {
	f.Add([]byte{5}, []byte{1, 2, 3, 4})
	f.Fuzz(func(t *testing.T, a []byte, b []byte) {
		out := slicesext.FoldlLazy(a, b, func(x []byte, y byte) []byte {
			return cons(x, y)
		})
		require.Equal(t, len(a)+len(b), len(out))
	})
}

func BenchmarkFoldl(b *testing.B) {
	for i := 10; i < 1000000; i = i * 10 {
		list := make([]int, i)
		fill(list, 1)
		b.Run(fmt.Sprintf("lazy_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					slicesext.Foldl(0, list, sum)
				}
			})
		})

		b.Run(fmt.Sprintf("eager_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					slicesext.FoldlLazy(0, list, sum)
				}
			})
		})
	}
}
