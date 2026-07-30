// Copyright 2026 Specter Ops, Inc.
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

package handlers

//go:generate go tool mockery

import (
	"context"

	"github.com/specterops/bloodhound/server/graphdb/internal/services"
)

// GraphDB defines the graphdb service boundary for the graphdb handlers package.
type GraphDB interface {
	GetRelationship(ctx context.Context, id int64, includeKindInfo bool) (services.Relationship, error)
	GetNode(ctx context.Context, id int64, includeKindInfo bool) (services.Node, error)
}

// NodeAuthorizer decides whether the caller (provided via ctx) may access a given node.
type NodeAuthorizer interface {
	CanAccessNode(ctx context.Context, node services.Node) bool
}

// Handlers is a dependency injection container for graphdb handlers.
type Handlers struct {
	graphDB        GraphDB
	nodeAuthorizer NodeAuthorizer
}

// NewHandlersContainer initializes the Handlers dependency injection container.
func NewHandlersContainer(graphDB GraphDB, nodeAuthorizer NodeAuthorizer) *Handlers {
	return &Handlers{
		graphDB:        graphDB,
		nodeAuthorizer: nodeAuthorizer,
	}
}
