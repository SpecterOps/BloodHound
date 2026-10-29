// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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

	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
)

// OpenGraphSchema defines the schema service boundary for the OpenGraph schema handlers package.
type OpenGraphSchema interface {
	GetNodeKind(ctx context.Context, id int32) (services.NodeKind, error)
	GetRelationshipKind(ctx context.Context, id int32) (services.RelationshipKind, error)
}

// Handlers is a dependency injection container for OpenGraph schema handlers.
type Handlers struct {
	schema OpenGraphSchema
}

// NewHandlersContainer initializes the Handlers dependency injection container.
func NewHandlersContainer(schema OpenGraphSchema) *Handlers {
	return &Handlers{
		schema: schema,
	}
}
