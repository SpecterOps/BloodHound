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
)

const UserInterfacePath = "/ui"

// Auth defines the service boundary for the auth handlers package.
type Auth interface {
	Logout(ctx context.Context, sessionID int64) error
}

// Handlers is a dependency injection container for auth handlers
type Handlers struct {
	service Auth
}

// NewHandlersContainer initializes the Handlers dependency injection container
func NewHandlersContainer(auth Auth) *Handlers {
	return &Handlers{
		service: auth,
	}
}
