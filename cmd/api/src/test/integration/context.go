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
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/test"
	"github.com/specterops/dawgs/graph"
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

func (s *GraphContext) Begin(ctx test.Context) {
	// Assert the graph schema before continuing
	test.RequireNilErr(ctx, s.Database.AssertSchema(ctx, s.schema))
}

func (s *GraphContext) End(t test.Context) {
	if err := s.Database.Close(t); err != nil {
		t.Fatalf("Error encoutered while closing the database: %v", err)
	}
}

// NewGraphContext creates a new GraphContext
//
// Deprecated: this suite of integration utils is deprecated and should be avoided
// See latest testing guidance for more details.
func NewGraphContext(t *testing.T, ctx test.Context, schema graph.Schema) *GraphContext {
	graphContext := &GraphContext{
		schema:   schema,
		Database: OpenGraphDB(t, schema),
	}

	// Initialize the graph context
	graphContext.Begin(ctx)

	// Ensure that the test cleans up after itself
	ctx.Cleanup(func() {
		graphContext.End(ctx)
	})

	return graphContext
}
