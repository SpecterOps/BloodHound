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

package services

import (
	"context"
	"fmt"
	"time"

	"github.com/specterops/bloodhound/cache"
	"github.com/specterops/bloodhound/dawgs/graph"
	schema "github.com/specterops/bloodhound/graphschema"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/packages/go/apitoy/app"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/registration"
	"github.com/specterops/bloodhound/src/api/router"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/daemons"
	"github.com/specterops/bloodhound/src/daemons/api/bhapi"
	"github.com/specterops/bloodhound/src/daemons/api/toolapi"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/daemons/gc"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"github.com/specterops/bloodhound/src/queries"
)

// ConnectPostgres initializes a connection to PG, and returns errors if any
func ConnectPostgres(cfg config.Configuration) (*database.BloodhoundDB, error) {
	if db, err := database.OpenDatabase(cfg.Database.PostgreSQLConnectionString()); err != nil {
		return nil, fmt.Errorf("error while attempting to create database connection: %w", err)
	} else {
		return database.NewBloodhoundDB(db, auth.NewIdentityResolver()), nil
	}
}

// ConnectDatabases initializes connections to PG and connection, and returns errors if any
func ConnectDatabases(ctx context.Context, cfg config.Configuration) (bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch], error) {
	connections := bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch]{}

	if db, err := ConnectPostgres(cfg); err != nil {
		return connections, err
	} else if graphDB, err := bootstrap.ConnectGraph(ctx, cfg); err != nil {
		return connections, err
	} else {
		connections.RDMS = db
		connections.Graph = graphDB

		return connections, nil
	}
}

// PreMigrationDaemons Word of caution: These daemons will be launched prior to any migration starting
func PreMigrationDaemons(ctx context.Context, cfg config.Configuration, connections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch]) ([]daemons.Daemon, error) {
	return []daemons.Daemon{
		toolapi.NewDaemon(ctx, connections, cfg, schema.DefaultGraphSchema()),
	}, nil
}

func Entrypoint(ctx context.Context, cfg config.Configuration, connections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch]) ([]daemons.Daemon, error) {
	if !cfg.DisableMigrations {
		if err := bootstrap.MigrateDB(ctx, cfg, connections.RDMS); err != nil {
			return nil, fmt.Errorf("rdms migration error: %w", err)
		} else if err := bootstrap.MigrateGraph(ctx, connections.Graph, schema.DefaultGraphSchema()); err != nil {
			return nil, fmt.Errorf("graph migration error: %w", err)
		}
	} else if err := connections.Graph.SetDefaultGraph(ctx, schema.DefaultGraph()); err != nil {
		return nil, fmt.Errorf("no default graph found but migrations are disabled per configuration: %w", err)
	} else {
		log.Infof("Database migrations are disabled per configuration")
	}

	if apiCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxAPICacheSize}); err != nil {
		return nil, fmt.Errorf("failed to create in-memory cache for API: %w", err)
	} else if graphQueryCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxAPICacheSize}); err != nil {
		return nil, fmt.Errorf("failed to create in-memory cache for graph queries: %w", err)
	} else if collectorManifests, err := cfg.SaveCollectorManifests(); err != nil {
		return nil, fmt.Errorf("failed to save collector manifests: %w", err)
	} else {
		var (
			graphQuery     = queries.NewGraphQuery(connections.Graph, graphQueryCache, cfg)
			authorizer     = auth.NewAuthorizer(connections.RDMS)
			datapipeDaemon = datapipe.NewDaemon(ctx, cfg, connections, graphQueryCache, time.Duration(cfg.DatapipeInterval)*time.Second)
			routerInst     = router.NewRouter(cfg, authorizer, bootstrap.ContentSecurityPolicy)
			ctxInitializer = database.NewContextInitializer(connections.RDMS)
			authenticator  = api.NewAuthenticator(cfg, connections.RDMS, ctxInitializer)
			bhApp          = app.NewBHApp(connections.RDMS, cfg)
		)

		// new app

		registration.RegisterFossGlobalMiddleware(&routerInst, cfg, auth.NewIdentityResolver(), authenticator)
		registration.RegisterFossRoutes(&routerInst, cfg, connections.RDMS, connections.Graph, graphQuery, apiCache, collectorManifests, authenticator, authorizer, bhApp)

		// Set neo4j batch and flush sizes
		neo4jParameters := appcfg.GetNeo4jParameters(ctx, connections.RDMS)
		connections.Graph.SetBatchWriteSize(neo4jParameters.BatchWriteSize)
		connections.Graph.SetWriteFlushSize(neo4jParameters.WriteFlushSize)

		// Trigger analysis on first start
		if err := connections.RDMS.RequestAnalysis(ctx, "init"); err != nil {
			log.Warnf("failed to request init analysis: %v", err)
		}

		return []daemons.Daemon{
			bhapi.NewDaemon(cfg, routerInst.Handler()),
			gc.NewDataPruningDaemon(connections.RDMS),
			datapipeDaemon,
		}, nil
	}
}
