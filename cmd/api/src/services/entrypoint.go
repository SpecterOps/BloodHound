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
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/registration"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/daemons"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/api/bhapi"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/api/toolapi"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/datapipe"
	"github.com/specterops/bloodhound/cmd/api/src/daemons/gc"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/migrations"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/queries"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/cmd/api/src/services/opengraphschema"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/cache"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
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

	dogtagsService := dogtags.NewDefaultService()

	flags := dogtagsService.GetAllDogTags()
	slog.InfoContext(ctx, "DogTags Configuration",
		slog.String("namespace", "dogtags"),
		slog.Any("flags", flags))

	if !cfg.DisableMigrations {
		if err := bootstrap.MigrateDB(ctx, cfg, connections.RDMS, config.NewDefaultAdminConfiguration); err != nil {
			return nil, fmt.Errorf("rdms migration error: %w", err)
		} else if err := migrations.NewGraphMigrator(connections.Graph).Migrate(ctx); err != nil {
			return nil, fmt.Errorf("graph migration error: %w", err)
		} else if err := bootstrap.PopulateExtensionData(ctx, connections.RDMS); err != nil {
			return nil, fmt.Errorf("extensions data population error: %w", err)
		}
	} else if err := connections.Graph.SetDefaultGraph(ctx, schema.DefaultGraph()); err != nil {
		return nil, fmt.Errorf("no default graph found but migrations are disabled per configuration: %w", err)
	} else {
		slog.InfoContext(ctx, "Database migrations are disabled per configuration")
	}

	// Allow recreating the default admin account to help with lockouts/loading database dumps
	if cfg.RecreateDefaultAdmin {
		slog.InfoContext(ctx, "Recreating default admin user")
		if err := bootstrap.CreateDefaultAdmin(ctx, cfg, connections.RDMS, config.NewDefaultAdminConfiguration); err != nil {
			return nil, err
		}
	}

	if apiCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxAPICacheSize}); err != nil {
		return nil, fmt.Errorf("failed to create in-memory cache for API: %w", err)
	} else if graphQueryCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxAPICacheSize}); err != nil {
		return nil, fmt.Errorf("failed to create in-memory cache for graph queries: %w", err)
	} else if collectorManifests, err := cfg.SaveCollectorManifests(); err != nil {
		return nil, fmt.Errorf("failed to save collector manifests: %w", err)
	} else if ingestSchema, err := upload.LoadIngestSchema(); err != nil {
		return nil, fmt.Errorf("failed to load OpenGraph schema: %w", err)
	} else {
		startDelay := 0 * time.Second

		var (
			cl                     = changelog.NewChangelog(connections.Graph, connections.RDMS, changelog.DefaultOptions())
			pipeline               = datapipe.NewPipeline(ctx, cfg, connections.RDMS, connections.Graph, graphQueryCache, ingestSchema, cl)
			graphQuery             = queries.NewGraphQuery(connections.Graph, graphQueryCache, cfg)
			authorizer             = auth.NewAuthorizer(connections.RDMS)
			datapipeDaemon         = datapipe.NewDaemon(pipeline, startDelay, time.Duration(cfg.DatapipeInterval)*time.Second, connections.RDMS)
			routerInst             = router.NewRouter(cfg, authorizer, fmt.Sprintf(bootstrap.ContentSecurityPolicy, "", "", "", "", "", ""))
			authenticator          = api.NewAuthenticator(cfg, connections.RDMS, api.NewAuthExtensions(cfg, connections.RDMS))
			openGraphSchemaService = opengraphschema.NewOpenGraphSchemaService(connections.RDMS, connections.Graph)
		)

		registration.RegisterFossGlobalMiddleware(&routerInst, cfg, auth.NewIdentityResolver(), authenticator, connections.RDMS)
		registration.RegisterFossRoutes(&routerInst, cfg, connections.RDMS, connections.Graph, graphQuery, apiCache, collectorManifests, authenticator, authorizer, ingestSchema, dogtagsService, openGraphSchemaService)

		// Set neo4j batch and flush sizes
		neo4jParameters := appcfg.GetNeo4jParameters(ctx, connections.RDMS)
		connections.Graph.SetBatchWriteSize(neo4jParameters.BatchWriteSize)
		connections.Graph.SetWriteFlushSize(neo4jParameters.WriteFlushSize)

		// Trigger analysis on first start
		if err := connections.RDMS.RequestAnalysis(ctx, "init"); err != nil {
			slog.WarnContext(ctx, fmt.Sprintf("failed to request init analysis: %v", err))
		}

		return []daemons.Daemon{
			bhapi.NewDaemon(cfg, routerInst.Handler()),
			gc.NewDataPruningDaemon(connections.RDMS),
			cl,
			datapipeDaemon,
		}, nil
	}
}
