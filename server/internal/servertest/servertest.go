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

//go:build integration

// Package servertest provides shared end-to-end test scaffolding for the
// vertical-slice feature modules under server/. It centralizes database
// provisioning, production router and middleware wiring, and JWT session
// creation so that each module's *_e2e_test.go file can register its own routes
// and focus on behavior rather than boilerplate.
package servertest

import (
	"context"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/registration"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

// testJWTSigningKey is a fixed signing key used for e2e tests. It must be at
// least 32 bytes long to satisfy the JWT configuration.
const testJWTSigningKey = "test-secret-key-that-is-at-least-32-bytes-long"

// Harness bundles the infrastructure produced by NewHarness for a single e2e
// test: an isolated database, a running httptest server routed through the
// production router and middleware, and the authenticator used to mint tokens.
type Harness struct {
	DB     *database.BloodhoundDB
	Server *httptest.Server
	Auther api.Authenticator
}

// NewHarness provisions an isolated, migrated database, wires the production
// router with the FOSS global middleware, invokes registerRoutes to mount the
// module under test, and starts an httptest.Server. The database and server are
// closed automatically when the test ends.
func NewHarness(t *testing.T, registerRoutes func(routerInst *router.Router, db *database.BloodhoundDB)) *Harness {
	t.Helper()

	db := NewDatabase(t)

	cfg, err := config.NewDefaultConfiguration()
	require.NoError(t, err)

	// Set up JWT signing key before constructing auth components, which copy cfg by value.
	cfg.Crypto.JWT.SetSigningKeyBytes([]byte(testJWTSigningKey))

	var (
		authExt    = api.NewAuthExtensions(cfg, db)
		auther     = api.NewAuthenticator(cfg, db, authExt)
		authorizer = auth.NewAuthorizer(db)
		resolver   = auth.NewIdentityResolver()
		routerInst = router.NewRouter(cfg, authorizer, "")
	)

	// Register global middleware (required for auth to work).
	registration.RegisterFossGlobalMiddleware(&routerInst, cfg, resolver, auther, db)

	registerRoutes(&routerInst, db)

	server := httptest.NewServer(routerInst.Handler())
	t.Cleanup(server.Close)

	return &Harness{DB: db, Server: server, Auther: auther}
}

// MintJWT creates a signed JWT token for the given user using the
// authenticator. The user is persisted with an auth secret (and whatever roles
// were preset on it), reloaded to populate associations, and a session is
// created in the database. The returned token is a valid bearer token.
func MintJWT(t *testing.T, ctx context.Context, db *database.BloodhoundDB, auther api.Authenticator, user model.User) string {
	t.Helper()

	authSecret := model.AuthSecret{
		Digest:       "dummy-digest-for-e2e-test",
		DigestMethod: "argon2",
		ExpiresAt:    time.Now().Add(24 * time.Hour).UTC(),
	}
	user.AuthSecret = &authSecret

	dbUser, err := db.CreateUser(ctx, user)
	require.NoError(t, err)

	// Reload user to populate the AuthSecret ID and any role/permission associations
	// that the session validation middleware relies on.
	dbUser, err = db.GetUser(ctx, dbUser.ID)
	require.NoError(t, err)
	require.NotNil(t, dbUser.AuthSecret, "User should have an AuthSecret")

	token, err := auther.CreateSession(ctx, dbUser, *dbUser.AuthSecret)
	require.NoError(t, err)
	return token
}

// AdminRole returns the Administrator role from the database, which carries all
// permissions. Callers assign it to a user (user.Roles) before minting a token
// for endpoints that require elevated permissions.
func AdminRole(t *testing.T, ctx context.Context, db *database.BloodhoundDB) model.Role {
	t.Helper()

	allRoles, err := db.GetAllRoles(ctx, "", model.SQLFilter{})
	require.NoError(t, err)

	adminRole, found := allRoles.FindByName(auth.RoleAdministrator)
	require.True(t, found, "Administrator role should exist in database")
	require.Greater(t, len(adminRole.Permissions), 0, "Administrator role should have permissions")

	return adminRole
}
