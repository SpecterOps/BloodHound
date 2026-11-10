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
package users

//go:generate go tool mockery

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// AuthenticatedUser contains the authentication attributes commonly needed to identify
// and authorize the caller of a request. This type is intended to stay small.
// feature-specific user data belongs in the owning slice's domain types rather
// than being added here.
type AuthenticatedUser struct {
	ID              uuid.UUID
	PrincipalName   string
	EmailAddress    string
	Administrator   bool
	AllEnvironments bool
}

func (s AuthenticatedUser) GetID() uuid.UUID {
	return s.ID
}

func (s AuthenticatedUser) IsAdministrator() bool {
	return s.Administrator
}

func (s AuthenticatedUser) HasAllEnvironments() bool {
	return s.AllEnvironments
}

// User is the minimal interface for a user in the BHE system.
type User interface {
	GetID() uuid.UUID
	HasAllEnvironments() bool
}

// Resolver extracts a Principal from a request context. Implementations are responsible
// for reading authentication context, validating credentials, and returning the
// authenticated user.
type Resolver interface {
	Resolve(ctx context.Context) (AuthenticatedUser, bool)
}

// RequestContextResolver is the production implementation of Resolver that
// extracts the user from the BloodHound request context.
type RequestContextResolver struct{}

// NewResolver returns a new RequestContextResolver.
func NewResolver() *RequestContextResolver {
	return &RequestContextResolver{}
}

// Resolve extracts a Principal from the request's BloodHound context.
func (s *RequestContextResolver) Resolve(ctx context.Context) (AuthenticatedUser, bool) {
	var (
		bhContext = bhctx.Get(ctx)
		modelUser model.User
		ok        bool
	)

	modelUser, ok = auth.GetUserFromAuthCtx(bhContext.AuthCtx)
	if !ok {
		return AuthenticatedUser{}, false
	}

	return AuthenticatedUser{
		ID:              modelUser.ID,
		PrincipalName:   modelUser.PrincipalName,
		EmailAddress:    modelUser.EmailAddress.ValueOrZero(),
		Administrator:   modelUser.Roles.Has(model.Role{Name: auth.RoleAdministrator}),
		AllEnvironments: modelUser.HasAllEnvironments(),
	}, true
}
