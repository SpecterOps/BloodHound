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

package server

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	iso8601 "github.com/channelmeter/iso8601duration"
	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/registration"
	"github.com/specterops/bloodhound/src/api/router"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/daemons"
	"github.com/specterops/bloodhound/src/daemons/api/bhapi"
	"github.com/specterops/bloodhound/src/daemons/api/toolapi"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/daemons/gc"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/migration"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

const (
	DefaultServerShutdownTimeout = time.Minute
	ContentSecurityPolicy        = "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: blob:; font-src 'self' data:;"
)

// SystemSignalExitChannel is used to shut down the server. It creates a channel that listens for an exit signal from the server.
func SystemSignalExitChannel() chan struct{} {
	exitC := make(chan struct{})

	go func() {
		// Shutdown on SIGINT/SIGTERM
		signalChannel := make(chan os.Signal, 1)
		signal.Notify(signalChannel, syscall.SIGTERM)
		signal.Notify(signalChannel, syscall.SIGINT)

		// Wait for a signal from the OS
		<-signalChannel
		close(exitC)
	}()

	return exitC
}

// MigrateGraph runs migrations for the graph database
func MigrateGraph(cfg config.Configuration, db graph.Database) error {
	if cfg.DisableMigrations {
		log.Infof("Graph migrations are disabled per configuration")
		return nil
	}
	return migrations.NewGraphMigrator(db).Migrate()
}

// MigrateDB runs database migrations on PG
func MigrateDB(cfg config.Configuration, db database.Database, models []any) error {
	if cfg.DisableMigrations {
		log.Infof("Database migrations are disabled per configuration")
		return nil
	}

	if err := db.MigrateModels(models); err != nil {
		return err
	}

	if hasInstallation, err := db.HasInstallation(); err != nil {
		return err
	} else if hasInstallation {
		return nil
	}

	secretDigester := cfg.Crypto.Argon2.NewDigester()

	if roles, err := db.GetAllRoles("", model.SQLFilter{}); err != nil {
		return fmt.Errorf("error while attempting to fetch user roles: %w", err)
	} else if secretDigest, err := secretDigester.Digest(cfg.DefaultAdmin.Password); err != nil {
		return fmt.Errorf("error while attempting to digest secret for user: %w", err)
	} else if adminRole, found := roles.FindByName(auth.RoleAdministrator); !found {
		return fmt.Errorf("unable to find admin role")
	} else {
		var (
			adminUser = model.User{
				Roles: model.Roles{
					adminRole,
				},
				PrincipalName: cfg.DefaultAdmin.PrincipalName,
				EmailAddress:  null.NewString(cfg.DefaultAdmin.EmailAddress, true),
				FirstName:     null.NewString(cfg.DefaultAdmin.FirstName, true),
				LastName:      null.NewString(cfg.DefaultAdmin.LastName, true),
			}

			authSecret = model.AuthSecret{
				Digest:       secretDigest.String(),
				DigestMethod: secretDigester.Method(),
			}
		)

		if cfg.DefaultAdmin.ExpireNow {
			authSecret.ExpiresAt = time.Time{}
		} else if defaultWindow, err := iso8601.FromString(appcfg.DefaultPasswordExpirationWindow); err != nil {
			return fmt.Errorf("unable to parse default password expiration window: %w", err)
		} else {
			authSecret.ExpiresAt = time.Now().Add(defaultWindow.ToDuration())
		}

		if _, err := db.InitializeSecretAuth(adminUser, authSecret); err != nil {
			return fmt.Errorf("error in database while initalizing auth: %w", err)
		} else {
			passwordMsg := fmt.Sprintf("# Initial Password Set To:    %s    #", cfg.DefaultAdmin.Password)
			paddingString := strings.Repeat(" ", len(passwordMsg)-2)
			borderString := strings.Repeat("#", len(passwordMsg))

			log.Infof("%s", borderString)
			log.Infof("#%s#", paddingString)
			log.Infof("%s", passwordMsg)
			log.Infof("#%s#", paddingString)
			log.Infof("%s", borderString)
		}
	}

	return nil
}

// StartServer sets up background daemons, runs the service and waits for an exit signal to shut it down
func StartServer(cfg config.Configuration, exitC chan struct{}) error {
	if err := InitializeLogging(cfg); err != nil {
		return fmt.Errorf("log initialization error: %w", err)
	}

	if db, graphDB, err := ConnectDatabases(cfg); err != nil {
		return fmt.Errorf("db connection error: %w", err)
	} else if err := MigrateDB(cfg, db, migration.ListBHModels()); err != nil {
		return fmt.Errorf("db migration error: %w", err)
	} else if err := MigrateGraph(cfg, graphDB); err != nil {
		return fmt.Errorf("graph db migration error: %w", err)
	} else if apiCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxAPICacheSize}); err != nil {
		return fmt.Errorf("failed to create in-memory cache for API: %w", err)
	} else if graphQueryCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxAPICacheSize}); err != nil {
		return fmt.Errorf("failed to create in-memory cache for graph queries: %w", err)
	} else if collectorManifests, err := cfg.SaveCollectorManifests(); err != nil {
		return fmt.Errorf("failed to save collector manifests: %w", err)
	} else {
		var (
			serviceManager         = daemons.NewManager(DefaultServerShutdownTimeout)
			sessionSweepingService = gc.NewDataPruningDaemon(db)
			routerInst             = router.NewRouter(cfg, auth.NewAuthorizer(), ContentSecurityPolicy)
			toolingService         = toolapi.NewDaemon(cfg, db)
			datapipeDaemon         = datapipe.NewDaemon(cfg, db, graphDB, graphQueryCache, time.Duration(cfg.DatapipeInterval)*time.Second)
			authenticator          = api.NewAuthenticator(cfg, db, database.NewContextInitializer(db))
		)

		registration.RegisterFossGlobalMiddleware(&routerInst, cfg, authenticator)
		registration.RegisterFossRoutes(&routerInst, cfg, db, graphDB, apiCache, graphQueryCache, collectorManifests, authenticator, datapipeDaemon)
		apiDaemon := bhapi.NewDaemon(cfg, routerInst.Handler())

		// Set neo4j batch and flush sizes
		neo4jParameters := appcfg.GetNeo4jParameters(db)
		graphDB.SetBatchWriteSize(neo4jParameters.BatchWriteSize)
		graphDB.SetWriteFlushSize(neo4jParameters.WriteFlushSize)

		// Start daemons
		serviceManager.Start(apiDaemon, toolingService, sessionSweepingService, datapipeDaemon)

		log.Infof("Server started successfully")
		// Wait for a signal to exit
		<-exitC

		log.Infof("Shutting down")
		serviceManager.Stop()

		log.Infof("Server shut down successfully")
	}

	return nil
}
