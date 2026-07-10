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
package users

//go:generate go tool mockery

import (
	"net/http"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// User is the minimal interface for a user in the BHE system.
type User interface {
	GetID() uuid.UUID
	HasAllEnvironments() bool
}

// Resolver extracts a User from an HTTP request. Implementations are responsible
// for reading authentication context, validating credentials, and returning the
// authenticated user.
type Resolver interface {
	Resolve(r *http.Request) (User, bool)
}

// RequestContextResolver is the production implementation of Resolver that
// extracts the user from the BloodHound request context.
type RequestContextResolver struct{}

// NewResolver returns a new RequestContextResolver.
func NewResolver() *RequestContextResolver {
	return &RequestContextResolver{}
}

// Resolve extracts a User from the request's BloodHound context.
func (s *RequestContextResolver) Resolve(r *http.Request) (User, bool) {
	var (
		bhContext = bhctx.FromRequest(r)
		modelUser model.User
		ok        bool
	)

	modelUser, ok = auth.GetUserFromAuthCtx(bhContext.AuthCtx)
	return &modelUser, ok
}
