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

	"github.com/specterops/bloodhound/server/extensions/internal/services"
)

// Extensions defines the extensions service boundary for the extensions handlers package.
type Extensions interface {
	GetNodeKind(ctx context.Context, id int32) (services.NodeKind, error)
}

// Handlers is a dependency injection container for extensions handlers.
type Handlers struct {
	extensions Extensions
}

// NewHandlersContainer initializes the Handlers dependency injection container.
func NewHandlersContainer(extensions Extensions) *Handlers {
	return &Handlers{
		extensions: extensions,
	}
}
