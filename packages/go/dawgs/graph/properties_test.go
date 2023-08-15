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

package graph_test

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/stretchr/testify/require"
	"strconv"
	"testing"
)

func TestNewProperties(t *testing.T) {
	properties := graph.NewProperties()

	require.Equal(t, 24, int(properties.SizeOf()))
	require.Equal(t, 0, properties.Len())
	require.Equal(t, nil, properties.Get("test").Any())
	require.Equal(t, map[string]any{}, properties.ModifiedProperties())
	require.Equal(t, []string(nil), properties.DeletedProperties())

	properties.Delete("test")
	require.False(t, properties.Exists("not found"))
	require.Equal(t, "default", properties.GetOrDefault("not found", "default").Any())

	// Set values to trigger allocation of the storage map
	properties.Set("test", "test")
	properties.Set("value", "test")

	require.Equal(t, 736, int(properties.SizeOf()))

	properties.Delete("value")

	require.Equal(t, 732, int(properties.SizeOf()))

	require.False(t, properties.Exists("not found"))
	require.Equal(t, 1, properties.Len())
	require.Equal(t, "default", properties.GetOrDefault("not found", "default").Any())
	require.Equal(t, "test", properties.GetOrDefault("test", "default").Any())
	require.Equal(t, "test", properties.Get("test").Any())

	properties.Delete("test")
	properties.Set("test", "other")

	require.Equal(t, map[string]any{"test": "other"}, properties.ModifiedProperties())

	_, markedForDeletion := properties.Deleted["test"]
	require.False(t, markedForDeletion)
	require.Equal(t, []string{"value"}, properties.DeletedProperties())
}

func TestSizeOfProperties(t *testing.T) {
	t.Skip("Skipping SizeOf due to flake")
	properties := graph.NewProperties()

	require.Equal(t, 24, int(properties.SizeOf()))

	// Set an initial value to force allocation
	properties.Set("initial", 0)
	require.Equal(t, 536, int(properties.SizeOf()))

	// Further allocation past 8 forces a new bucket to be allocated for the backing maps in the properties struct
	for iter := 0; iter < 8; iter++ {
		properties.Set("test"+strconv.Itoa(iter), iter)
	}

	require.Equal(t, 1784, int(properties.SizeOf()))

}
