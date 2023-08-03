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

package graph

import (
	"sort"

	"github.com/RoaringBitmap/roaring/roaring64"
)

func NodeSetToBitmap(nodes NodeSet) *roaring64.Bitmap {
	bitmap := roaring64.NewBitmap()

	for _, node := range nodes {
		bitmap.Add(node.ID.Uint64())
	}

	return bitmap
}
func SortAndSliceNodeSet(set NodeSet, skip, limit int) NodeSet {
	nodes := SortNodeSetById(set)
	if skip == 0 && limit == 0 {
		return NewNodeSet(nodes...)
	} else if limit == 0 {
		return NewNodeSet(nodes[skip:]...)
	} else if skip == 0 {
		return NewNodeSet(nodes[:limit]...)
	} else {
		return NewNodeSet(nodes[skip : skip+limit]...)
	}
}

func SortNodeSetById(set NodeSet) []*Node {
	ids := set.IDs()
	SortIDSlice(ids)
	results := make([]*Node, set.Len())
	for idx, id := range ids {
		results[idx] = set.Get(id)
	}

	return results
}

func SortIDSlice(ids []ID) {
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].Int64() < ids[j].Int64()
	})
}

func CopyIDSlice(ids []ID) []ID {
	sliceCopy := make([]ID, len(ids))
	copy(sliceCopy, ids)

	return sliceCopy
}
