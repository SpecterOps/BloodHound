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

package test

import (
	"sync/atomic"

	"github.com/specterops/bloodhound/dawgs/graph"
)

var (
	idSequence int64 = 0
)

func nextID() graph.ID {
	return graph.ID(atomic.AddInt64(&idSequence, 1))
}

func Node(kinds ...graph.Kind) *graph.Node {
	return graph.NewNode(nextID(), graph.NewProperties(), kinds...)
}

func Edge(start, end *graph.Node, kind graph.Kind) *graph.Relationship {
	return graph.NewRelationship(nextID(), start.ID, end.ID, graph.NewProperties(), kind)
}
