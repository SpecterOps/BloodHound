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
	"testing"

	"github.com/stretchr/testify/require"
)

// This is a test that serves as a documented example of how HLL works. This test was designed to use the 14 register
// HLL implementation.
//
// For more information on HLL see: https://en.wikipedia.org/wiki/HyperLogLog
//
func TestHyperLogLog32(t *testing.T) {
	const cardinalityMax = 10_000_000

	sketch := NewHyperLogLog32()

	for i := uint32(0); i < cardinalityMax; i++ {
		sketch.Add(i)
	}

	var (
		estimatedCardinality = sketch.Cardinality()
		deviation            = 100 - cardinalityMax/float64(estimatedCardinality)*100
	)

	// We expect the HLL sketch to have a cardinality that does not deviate more than 0.5% from reality
	require.Truef(t, deviation < 0.5, "Expected a cardinality less than 0.5%% but got %.2f%%", deviation)

	for i := 0; i < 100; i++ {
		previous := sketch.Cardinality()

		sketch.Add(0)
		after := sketch.Cardinality()

		require.Equal(t, previous, after, "Expected cardinality to remain the same after encoding the same ID twice")
	}
}

func TestHyperLogLog32_Add(t *testing.T) {
	sketch := NewHyperLogLog32()

	sketch.Add(1)
	require.Equal(t, uint64(1), sketch.Cardinality())

	sketch.Add(1)
	require.Equal(t, uint64(1), sketch.Cardinality())

	sketch.Add(2)
	require.Equal(t, uint64(2), sketch.Cardinality())

	sketch.Add(2)
	require.Equal(t, uint64(2), sketch.Cardinality())
}
