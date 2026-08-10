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

package servertest

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/peterldowns/pgtestdb"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/stretchr/testify/require"
)

// NewDatabase creates an isolated test database with all migrations applied. The
// database is automatically closed when the test ends.
func NewDatabase(t *testing.T) *database.BloodhoundDB {
	t.Helper()

	var (
		ctx      = context.Background()
		connConf = pgtestdb.Custom(t, PostgresConfig(t), pgtestdb.NoopMigrator{})
	)

	cfg, err := config.NewDefaultConnectionConfiguration(connConf.URL())
	require.NoError(t, err)

	gormDB, dbPool, err := database.OpenDatabase(cfg.Database)
	require.NoError(t, err)

	db := database.NewBloodhoundDB(gormDB, dbPool, auth.NewIdentityResolver(), cfg)
	require.NoError(t, db.Migrate(ctx))

	t.Cleanup(func() { db.Close(ctx) })

	return db
}

// PostgresConfig reads the integration test configuration from the environment
// and returns a pgtestdb.Config suitable for spinning up isolated databases.
// Both TCP and unix-socket host values are supported.
func PostgresConfig(t *testing.T) pgtestdb.Config {
	t.Helper()

	var (
		environmentMap = make(map[string]string)
		options        string
	)

	cfg, err := utils.LoadIntegrationTestConfig()
	require.NoError(t, err)

	for entry := range strings.FieldsSeq(cfg.Database.Connection) {
		if parts := strings.SplitN(entry, "=", 2); len(parts) == 2 {
			environmentMap[parts[0]] = parts[1]
		}
	}

	if strings.HasPrefix(environmentMap["host"], "/") {
		return pgtestdb.Config{
			DriverName: "pgx",
			User:       environmentMap["user"],
			Password:   environmentMap["password"],
			Database:   environmentMap["dbname"],
			Options:    fmt.Sprintf("host=%s", url.PathEscape(environmentMap["host"])),
			TestRole: &pgtestdb.Role{
				Username:     environmentMap["user"],
				Password:     environmentMap["password"],
				Capabilities: "NOSUPERUSER NOCREATEROLE",
			},
		}
	}

	options, err = tlsOptions(environmentMap)
	require.NoError(t, err)

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   options,
		ForceTerminateConnections: true,
	}
}

// authenticatedSSLModes are the sslmode values that both encrypt the connection
// and verify the server certificate, protecting against cleartext transmission
// and man-in-the-middle attacks (CWE-319).
var authenticatedSSLModes = map[string]bool{
	"verify-ca":   true,
	"verify-full": true,
}

// tlsOptions returns the TLS-related connection options for a TCP host,
// URL-query encoded as required by pgtestdb.Config.Options (for example
// "sslmode=verify-full&sslrootcert=%2Fpath"). Local hosts default to
// sslmode=disable because the local test database is not configured for TLS.
// Non-local hosts must specify an authenticated sslmode (verify-ca or
// verify-full); otherwise an error is returned so credentials and data are never
// transmitted in cleartext or over an unverified connection.
func tlsOptions(environmentMap map[string]string) (string, error) {
	if isLocalHost(environmentMap["host"]) {
		return "sslmode=disable", nil
	}

	if !authenticatedSSLModes[environmentMap["sslmode"]] {
		return "", fmt.Errorf("non-local database host %q requires an authenticated sslmode (verify-ca or verify-full), got %q", environmentMap["host"], environmentMap["sslmode"])
	}

	options := url.Values{}
	for _, key := range []string{"sslmode", "sslrootcert", "sslcert", "sslkey"} {
		if value := environmentMap[key]; value != "" {
			options.Set(key, value)
		}
	}
	return options.Encode(), nil
}

// isLocalHost reports whether host refers to the local machine.
func isLocalHost(host string) bool {
	switch host {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
