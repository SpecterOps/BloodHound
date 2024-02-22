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

func sum(a, b int) int {
	return a + b
}

func mult(a, b int) int {
	return a * b
}

func div(a, b int) int {
	return a / b
}

func and(a, b bool) bool {
	return a && b
}

func max(a, b int) int {
	if a > b {
		return a
	} else {
		return b
	}
}

func TestFoldr(t *testing.T) {
	type Reducer[A, B any] func(A, B) B
	type Case[A, B any] struct {
		Name    string
		Reducer Reducer[A, B]
		InitVal B
		List    []A
		Output  B
	}

	cases := []any{
		Case[int, int]{"Foldr(5,list,sum) to be 15", sum, 5, []int{1, 2, 3, 4}, 15},
		Case[int, int]{"Foldr(128,list,div) to be 2", div, 128, []int{1, 2, 4, 8}, 2},
		Case[int, int]{"Foldr(3,list,div) to be 3", div, 3, []int{}, 3},
		Case[bool, bool]{"Foldr(true,list,and) to be false", and, true, []bool{1 > 2, 3 > 2, 5 == 5}, false},
		Case[int, int]{"Foldr(18,list,max) to be 55", max, 18, []int{3, 6, 12, 4, 55, 11}, 55},
		Case[int, int]{"Foldr(111,list,max) to be 111", max, 111, []int{3, 6, 12, 4, 55, 11}, 111},
		Case[int, int]{"Foldr(54,list,(a+b)/2) to be 12", func(a, b int) int { return (a + b) / 2 }, 54, []int{12, 4, 10, 6}, 12},
	}

	for _, c := range cases {
		switch c := c.(type) {
		case Case[int, int]:
			t.Run(c.Name, func(t *testing.T) {
				require.Equal(t, c.Output, slicesext.Foldr(c.InitVal, c.List, c.Reducer))
				require.Equal(t, c.Output, slicesext.FoldrEager(c.InitVal, c.List, c.Reducer))
			})
		case Case[bool, bool]:
			t.Run(c.Name, func(t *testing.T) {
				require.Equal(t, c.Output, slicesext.Foldr(c.InitVal, c.List, c.Reducer))
				require.Equal(t, c.Output, slicesext.FoldrEager(c.InitVal, c.List, c.Reducer))
			})
		}
	}

	t.Run("should handle multiple reducers", func(t *testing.T) {
		// 1*(1+(1*(1+(2*(2+5))))) = 16
		require.Equal(t, 16, slicesext.Foldr(5, []int{1, 1, 2}, sum, mult))
		require.Equal(t, 16, slicesext.FoldrEager(5, []int{1, 1, 2}, sum, mult))
	})
}

func cons(a []byte, b byte) []byte {
	return append(a, b)
}

func FuzzFoldr(f *testing.F) {
	f.Add([]byte{5}, []byte{1, 2, 3, 4})
	f.Fuzz(func(t *testing.T, a []byte, b []byte) {
		out := slicesext.Foldr(a, b, cons)
		require.Equal(t, len(a)+len(b), len(out))
	})
}

func BenchmarkFoldr(b *testing.B) {
	for i := 10; i < 1000000; i = i * 10 {
		list := make([]int, i)
		fill(list, 1)
		b.Run(fmt.Sprintf("lazy_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					slicesext.Foldr(0, list, sum)
				}
			})
		})

		b.Run(fmt.Sprintf("eager_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					slicesext.Foldr(0, list, sum)
				}
			})
		})
	}
}

func fill[T any](s []T, val T) {
	for i := range s {
		s[i] = val
	}
}
