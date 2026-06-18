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

import (
	"context"
)

//go:generate go run go.uber.org/mock/mockgen -destination=../mocks/appcfg.go -package=mocks . Appcfg

// Appcfg defines the service boundary for the appcfg handlers package.
type Appcfg interface {
	// Method stubs will be added as we define them
}

// Handlers holds the HTTP handlers for application configuration endpoints.
type Handlers struct {
	appcfg Appcfg
}

// NewHandlersContainer returns a new Handlers instance backed by the given service.
func NewHandlersContainer(svc Appcfg) *Handlers {
	return &Handlers{appcfg: svc}
}
