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

package slices_test

import (
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/slices"
)

var (
	listEmpty        = []int{}
	listSingle       = []int{0}
	listDuo          = []int{0, 1}
	reversedListDuo  = []int{1, 0}
	listEven         = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}
	reversedListEven = []int{15, 14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
	listOdd          = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14}
	reversedListOdd  = []int{14, 13, 12, 11, 10, 9, 8, 7, 6, 5, 4, 3, 2, 1, 0}
)

func TestFilter(t *testing.T) {
	require.Equal(t, []int{1, 3}, slices.Filter([]int{1, 2, 3, 4}, isOdd))
	require.Equal(t, []string{"bbbbbb", "cccccccc"}, slices.Filter([]string{"aaaa", "bbbbbb", "cccccccc", "dd"}, isLong))
}

func TestMap(t *testing.T) {
	require.Equal(t, []uint{1, 3, 4, 12}, slices.Map([]int{-1, -3, 4, -12}, abs))
	require.Equal(t, []string{"abc", "def", "hij"}, slices.Map([]string{"ABC", "DEF", "HIJ"}, strings.ToLower))
	require.Equal(t, []int{3, 6, 9, 12}, slices.Map([]int{1, 2, 3, 4}, triple))
}

func TestFlatMap(t *testing.T) {
	require.Equal(t, []string{"a", "a", "b", "b"}, slices.FlatMap([]string{"a", "b"}, duplicate[string]))
	require.Equal(t, []int{1, 1, 2, 2}, slices.FlatMap([]int{1, 2}, duplicate[int]))
}

func TestUnique(t *testing.T) {
	var (
		in  = []string{"a", "a", "b", "b"}
		out = slices.Unique(in)
	)

	require.Equal(t, []string{"a", "b"}, out)
	require.NotSame(t, in, out) // ensure we didn't mutate the original slice
	require.Equal(t, []string{"a", "b"}, slices.Unique([]string{"a", "b", "b", "a"}))
	require.Equal(t, []string{"a"}, slices.Unique([]string{"a"}))
	require.Equal(t, []int{1, 2, 3}, slices.Unique([]int{1, 1, 2, 2, 3}))
}

func TestContains(t *testing.T) {
	require.True(t, slices.Contains([]string{"a", "b", "c"}, "c"))
	require.False(t, slices.Contains([]string{"a", "b", "c"}, "d"))
}

func TestReverse(t *testing.T) {
	require.Equal(t, []int{}, slices.Reverse(listEmpty))
	require.Equal(t, []int{0}, slices.Reverse(listSingle))
	require.Equal(t, reversedListDuo, slices.Reverse(listDuo))
	require.Equal(t, reversedListEven, slices.Reverse(listEven))
	require.Equal(t, reversedListOdd, slices.Reverse(listOdd))
}

func BenchmarkReverse(b *testing.B) {
	for i := 10; i < 1000000; i = i * 10 {
		list := make([]int, i)
		for idx := range list {
			list[idx] = idx
		}

		b.Run(fmt.Sprintf("reverse_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					slices.Reverse(list)
				}
			})
		})
	}
}

func BenchmarkHead(b *testing.B) {
	for i := 10; i < 1000000; i = i * 10 {
		list := make([]int, i)
		for idx := range list {
			list[idx] = idx
		}

		b.Run(fmt.Sprintf("head_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					a := slices.Head(list)
					require.IsType(b, int(0), a)
				}
			})
		})

		b.Run(fmt.Sprintf("[0]_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					a := list[0]
					require.IsType(b, int(0), a)
				}
			})
		})
	}
}

func BenchmarkTail(b *testing.B) {
	for i := 10; i < 1000000; i = i * 10 {
		list := make([]int, i)
		for idx := range list {
			list[idx] = idx
		}

		b.Run(fmt.Sprintf("tail_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					a := slices.Tail(list)
					require.IsType(b, []int{}, a)
				}
			})
		})

		b.Run(fmt.Sprintf("[1:]_%d", i), func(b *testing.B) {
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					a := list[1:]
					require.IsType(b, []int{}, a)
				}
			})
		})
	}
}

func abs(n int) uint {
	mask := n >> (strconv.IntSize - 1)
	return uint((n ^ mask) - mask)
}

func duplicate[T any](t T) []T {
	return []T{t, t}
}

func isOdd(n int) bool {
	return n%2 == 1
}

func isLong(s string) bool {
	return len(s) > 5
}

func triple(n int) int {
	return n * 3
}
