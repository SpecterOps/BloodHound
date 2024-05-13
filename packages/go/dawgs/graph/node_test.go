// Copyright 2024 Specter Ops, Inc.
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
	"testing"
)

func Test_NodeSizeOf(t *testing.T) {
	node := graph.Node{ID: graph.ID(1)}
	oldSize := int64(node.SizeOf())

	// ensure that reassignment of the Kinds field affects the size
	node.Kinds = append(node.Kinds, permissionKind, userKind, groupKind)
	newSize := int64(node.SizeOf())
	require.Greater(t, newSize, oldSize)

	// ensure that reassignment of the AddedKinds field affects the size
	oldSize = newSize
	node.AddedKinds = append(node.AddedKinds, permissionKind)
	newSize = int64(node.SizeOf())
	require.Greater(t, newSize, oldSize)

	// ensure that reassignment of the DeletedKinds field affects the size
	oldSize = newSize
	node.DeletedKinds = append(node.DeletedKinds, userKind)
	newSize = int64(node.SizeOf())
	require.Greater(t, newSize, oldSize)

	// ensure that reassignment of the Properties field affects the size
	oldSize = newSize
	node.Properties = graph.NewProperties()
	newSize = int64(node.SizeOf())
	require.Greater(t, newSize, oldSize)
}
