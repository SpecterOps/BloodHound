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
	"math"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/specterops/bloodhound/slicesext"
	"github.com/stretchr/testify/require"
)

func TestFilter(t *testing.T) {
	require.Equal(t, []int{1, 3}, slicesext.Filter([]int{1, 2, 3, 4}, isOdd))
	require.Equal(t, []string{"bbbbbb", "cccccccc"}, slicesext.Filter([]string{"aaaa", "bbbbbb", "cccccccc", "dd"}, isLong))
}

func TestMap(t *testing.T) {
	require.Equal(t, []uint{1, 3, 4, 12}, slicesext.Map([]int{-1, -3, 4, -12}, abs))
	require.Equal(t, []string{"abc", "def", "hij"}, slicesext.Map([]string{"ABC", "DEF", "HIJ"}, strings.ToLower))
	require.Equal(t, []int{3, 6, 9, 12}, slicesext.Map([]int{1, 2, 3, 4}, triple))
}

func TestFlatMap(t *testing.T) {
	require.Equal(t, []string{"a", "a", "b", "b"}, slicesext.FlatMap([]string{"a", "b"}, duplicate[string]))
	require.Equal(t, []int{1, 1, 2, 2}, slicesext.FlatMap([]int{1, 2}, duplicate[int]))
}

func TestUnique(t *testing.T) {
	var (
		in  = []string{"a", "a", "b", "b"}
		out = slicesext.Unique(in)
	)

	require.Equal(t, []string{"a", "b"}, out)
	require.NotSame(t, in, out) // ensure we didn't mutate the original slice
	require.Equal(t, []string{"a", "b"}, slicesext.Unique([]string{"a", "b", "b", "a"}))
	require.Equal(t, []string{"a"}, slicesext.Unique([]string{"a"}))
	require.Equal(t, []int{1, 2, 3}, slicesext.Unique([]int{1, 1, 2, 2, 3}))
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
					a := slicesext.Head(list)
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
					a := slicesext.Tail(list)
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

// Copyright 2021 The Go Authors. All rights reserved.
// TestConcat was ripped from go1.22 source and should be replaced with the stdlib implementation when we move to 1.22
// Original source: https://github.com/golang/go/blob/5c0d0929d3a6378c710376b55a49abd55b31a805/src/slices/slices_test.go#L1228
func TestConcat(t *testing.T) {
	cases := []struct {
		s    [][]int
		want []int
	}{
		{
			s:    [][]int{nil},
			want: nil,
		},
		{
			s:    [][]int{{1}},
			want: []int{1},
		},
		{
			s:    [][]int{{1}, {2}},
			want: []int{1, 2},
		},
		{
			s:    [][]int{{1}, nil, {2}},
			want: []int{1, 2},
		},
	}
	for _, tc := range cases {
		got := slicesext.Concat(tc.s...)
		if !slices.Equal(tc.want, got) {
			t.Errorf("Concat(%v) = %v, want %v", tc.s, got, tc.want)
		}
		var sink []int
		allocs := testing.AllocsPerRun(5, func() {
			sink = slicesext.Concat(tc.s...)
		})
		_ = sink
		if allocs > 1 {
			errorf := t.Errorf
			errorf("Concat(%v) allocated %v times; want 1", tc.s, allocs)
		}
	}
}

// Copyright 2021 The Go Authors. All rights reserved.
// TestConcat was ripped from go1.22 source and should be replaced with the stdlib implementation when we move to 1.22
// Original source: https://github.com/golang/go/blob/5c0d0929d3a6378c710376b55a49abd55b31a805/src/slices/slices_test.go#L1228
func TestConcat_too_large(t *testing.T) {
	// Use zero length element to minimize memory in testing
	type void struct{}
	cases := []struct {
		lengths     []int
		shouldPanic bool
	}{
		{
			lengths:     []int{0, 0},
			shouldPanic: false,
		},
		{
			lengths:     []int{math.MaxInt, 0},
			shouldPanic: false,
		},
		{
			lengths:     []int{0, math.MaxInt},
			shouldPanic: false,
		},
		{
			lengths:     []int{math.MaxInt - 1, 1},
			shouldPanic: false,
		},
		{
			lengths:     []int{math.MaxInt - 1, 1, 1},
			shouldPanic: true,
		},
		{
			lengths:     []int{math.MaxInt, 1},
			shouldPanic: true,
		},
		{
			lengths:     []int{math.MaxInt, math.MaxInt},
			shouldPanic: true,
		},
	}
	for _, tc := range cases {
		var r any
		ss := make([][]void, 0, len(tc.lengths))
		for _, l := range tc.lengths {
			s := make([]void, l)
			ss = append(ss, s)
		}
		func() {
			defer func() {
				r = recover()
			}()
			_ = slicesext.Concat(ss...)
		}()
		if didPanic := r != nil; didPanic != tc.shouldPanic {
			t.Errorf("slices.Concat(lens(%v)) got panic == %v",
				tc.lengths, didPanic)
		}
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
