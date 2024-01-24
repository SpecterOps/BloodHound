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

package integration

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/test"
)

type GraphContext struct {
	Database graph.Database
	schema   graph.Schema
}

func (s *GraphContext) BatchOperation(ctx test.Context, delegate graph.BatchDelegate) {
	test.RequireNilErr(ctx, s.Database.BatchOperation(ctx, delegate))
}

func (s *GraphContext) ReadTransaction(ctx test.Context, delegate graph.TransactionDelegate) {
	test.RequireNilErr(ctx, s.Database.WriteTransaction(ctx, delegate))
}

func (s *GraphContext) WriteTransaction(ctx test.Context, delegate graph.TransactionDelegate) {
	test.RequireNilErr(ctx, s.Database.WriteTransaction(ctx, delegate))
}

func (s *GraphContext) wipe(ctx test.Context) {
	s.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if nodeCount, err := tx.Nodes().Count(); err != nil {
			return err
		} else if nodeCount > 0 {
			return tx.Nodes().Delete()
		}

		return nil
	})
}

func (s *GraphContext) Begin(ctx test.Context) {
	// Clear the graph to ensure a clean slate
	s.wipe(ctx)

	// Assert the graph schema before continuing
	test.RequireNilErr(ctx, s.Database.AssertSchema(ctx, s.schema))
}

func (s *GraphContext) End(t test.Context) {
	if err := s.Database.Close(t); err != nil {
		t.Fatalf("Error encoutered while closing the database: %v", err)
	}
}

func NewGraphContext(ctx test.Context, schema graph.Schema) *GraphContext {
	graphContext := &GraphContext{
		schema:   schema,
		Database: OpenGraphDB(ctx),
	}

	// Initialize the graph context
	graphContext.Begin(ctx)

	// Ensure that the test cleans up after itself
	ctx.Cleanup(func() {
		graphContext.End(ctx)
	})

	return graphContext
}
