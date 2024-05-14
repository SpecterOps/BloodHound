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
	"unsafe"
)

func Test_RelationshipSizeOf(t *testing.T) {
	relationship := graph.Relationship{ID: graph.ID(1)}
	initialSize := int64(relationship.SizeOf())

	// ensure that the initial size accounts for all the struct fields
	sizeOfIDs := 3 * int64(unsafe.Sizeof(relationship.StartID))
	sizeOfKind := int64(unsafe.Sizeof(relationship.Kind))
	sizeOfProperties := int64(unsafe.Sizeof(relationship.Properties))
	require.Greater(t, initialSize, sizeOfIDs+sizeOfKind+sizeOfProperties)

	// ID, StartID, EndID and Kind have zero-value sizes that aren't impacted by
	// reassignment to a non-zero value. Therefore, skipping testing of those fields.

	// ensure that reassignment of the Properties field affects size
	relationship.Properties = graph.NewProperties()
	newSize := int64(relationship.SizeOf())
	require.Greater(t, newSize, initialSize)
}
