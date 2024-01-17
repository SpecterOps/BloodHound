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

package traversal

import (
	"context"
	"sync"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
)

type NodeCollector struct {
	Nodes graph.NodeSet
	lock  *sync.Mutex
}

func NewNodeCollector() *NodeCollector {
	return &NodeCollector{
		Nodes: graph.NewNodeSet(),
		lock:  &sync.Mutex{},
	}
}

func (s *NodeCollector) Collect(next *graph.PathSegment) {
	s.Add(next.Node)
}

func (s *NodeCollector) PopulateProperties(ctx context.Context, db graph.Database, propertyNames ...string) error {
	return db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return ops.FetchNodeProperties(tx, s.Nodes, propertyNames)
	})
}

func (s *NodeCollector) Add(node *graph.Node) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Nodes.Add(node)
}

type PathCollector struct {
	Paths graph.PathSet
	lock  *sync.Mutex
}

func NewPathCollector() *PathCollector {
	return &PathCollector{
		lock: &sync.Mutex{},
	}
}

func (s *PathCollector) PopulateNodeProperties(ctx context.Context, db graph.Database, propertyNames ...string) error {
	return db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return ops.FetchNodeProperties(tx, s.Paths.AllNodes(), propertyNames)
	})
}

func (s *PathCollector) Add(path graph.Path) {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.Paths = append(s.Paths, path)
}
