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

// Deprecated: this suite of integration utils is deprecated and should be avoided
// See latest testing guidance for more details.
package integration

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
	"github.com/specterops/bloodhound/cmd/api/src/database/migration"
	"github.com/specterops/bloodhound/cmd/api/src/test/integration/utils"
	"github.com/specterops/bloodhound/packages/go/cache"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"gorm.io/gorm"
)

// OpenDatabase opens a new database connection and returns a BHCE database interface
//
// Deprecated: this suite of integration utils is deprecated and should be avoided
// See latest testing guidance for more details.
func OpenDatabase(t *testing.T) database.Database {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		t.Fatalf("Failed loading integration test config: %v", err)
	} else if db, err := setupPGTestDB(t, cfg); err != nil {
		t.Fatalf("Failed to setup pgtestdb: %v", err)
	} else {
		return database.NewBloodhoundDB(db, auth.NewIdentityResolver())
	}

	return nil
}

func setupPGTestDB(t *testing.T, cfg config.Configuration) (*gorm.DB, error) {
	t.Helper()

	var (
		connConf = pgtestdb.Custom(t, GetPostgresConfig(cfg), pgtestdb.NoopMigrator{})
	)

	return database.OpenDatabase(connConf.URL())
}

// GetPostgresConfig reads key/value pairs from the default integration
// config file and creates a pgtestdb configuration object.
//
// Deprecated: this suite of integration utils is deprecated and should be avoided
// See latest testing guidance for more details.
func GetPostgresConfig(cfg config.Configuration) pgtestdb.Config {
	environmentMap := make(map[string]string)
	for _, entry := range strings.Fields(cfg.Database.Connection) {
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

	return pgtestdb.Config{
		DriverName:                "pgx",
		Host:                      environmentMap["host"],
		Port:                      environmentMap["port"],
		User:                      environmentMap["user"],
		Password:                  environmentMap["password"],
		Database:                  environmentMap["dbname"],
		Options:                   "sslmode=disable",
		ForceTerminateConnections: true,
	}
}

func OpenCache(t *testing.T) cache.Cache {
	if cache, err := cache.NewCache(cache.Config{MaxSize: 200}); err != nil {
		t.Fatalf("Failed creating cache: %e", err)
	} else {
		return cache
	}
	return cache.Cache{}
}

// SetupDB sets up a new database connection and prepares the DB with migrations
//
// Deprecated: this suite of integration utils is deprecated and should be avoided
// See latest testing guidance for more details.
func SetupDB(t *testing.T) database.Database {
	dbInst := OpenDatabase(t)
	if err := Prepare(context.Background(), dbInst); err != nil {
		t.Fatalf("Error preparing DB: %v", err)
	}
	return dbInst
}

func Prepare(ctx context.Context, db database.Database) error {
	if err := db.Wipe(ctx); err != nil {
		return fmt.Errorf("failed to clear database: %v", err)
	} else if err := db.Migrate(ctx); err != nil {
		return fmt.Errorf("failed to migrate database: %v", err)
	}

	return nil
}

// SetupTestMigrator opens a database connection and returns a migrator for testing
//
// Deprecated: this suite of integration utils is deprecated and should be avoided
// See latest testing guidance for more details.
func SetupTestMigrator(t *testing.T, sources ...migration.Source) (*gorm.DB, *migration.Migrator, error) {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		return nil, nil, fmt.Errorf("failed to load integration test config: %w", err)
	} else if db, err := setupPGTestDB(t, cfg); err != nil {
		return nil, nil, fmt.Errorf("failed to setup pgtestdb: %v", err)
	} else {
		OpenGraphDB(t, graphschema.DefaultGraphSchema()).Close(context.Background())
		return db, &migration.Migrator{
			Sources: sources,
			DB:      db,
		}, nil
	}
}
