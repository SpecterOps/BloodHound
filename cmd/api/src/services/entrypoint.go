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

	"github.com/prometheus/client_golang/prometheus"
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
	storageService "github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/cache"
	schema "github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/metricsregistration"
	"github.com/specterops/bloodhound/packages/go/storage"
	"github.com/specterops/bloodhound/server/modules"
	"github.com/specterops/dawgs/graph"
)

var requiredFileServices = []storage.FileServiceName{
	storage.FileServiceIngest,
	storage.FileServiceRetained,
	storage.FileServiceCollectors,
	storage.FileServiceWork,
}

func ensureFileServices(
	fileServiceResolver storageService.FileServiceResolver,
	requiredFileServices ...storage.FileServiceName,
) error {
	for _, serviceName := range requiredFileServices {
		if _, err := fileServiceResolver.Resolve(serviceName); err != nil {
			return fmt.Errorf("failed to resolve %s file service: %w", serviceName, err)
		}
	}

	return nil
}

// ConnectPostgres initializes a connection to PG, and returns errors if any
func ConnectPostgres(cfg config.Configuration) (*database.BloodhoundDB, error) {
	if db, dbPool, err := database.OpenDatabase(cfg.Database); err != nil {
		return nil, fmt.Errorf("error while attempting to create database connection: %w", err)
	} else {
		return database.NewBloodhoundDB(db, dbPool, auth.NewIdentityResolver(), cfg), nil
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

// CreateRuntimeDependencies creates the needed dependencies prior to migration. For instance, the FileService is needed for
// IngestControl which occurs prior to migration. This function can be used to make the struct to contain the services that
// are necessary for the application.
func CreateRuntimeDependencies(ctx context.Context, cfg config.Configuration, connections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch]) (bootstrap.RuntimeDependencies, error) {
	dependencies := bootstrap.RuntimeDependencies{}
	if fileServices, err := storageService.NewDefaultFileServices(cfg); err != nil {
		return dependencies, fmt.Errorf("failed to initialize file services: %w", err)
	} else if fileServiceResolver, err := storageService.NewFileServiceResolver(fileServices); err != nil {
		return dependencies, fmt.Errorf("failed to initialize file service resolver: %w", err)
		// Multiple file services are required at runtime. Checking it here to ensure that they were properly registered
		// to fail fast if there were any required services that were not registered.
	} else if err := ensureFileServices(fileServiceResolver, requiredFileServices...); err != nil {
		return dependencies, fmt.Errorf("failed to resolve required file service: %w", err)
	} else {
		dependencies.FileServiceResolver = fileServiceResolver
		return dependencies, nil
	}
}

// PreMigrationDaemons Word of caution: These daemons will be launched prior to any migration starting
func PreMigrationDaemons(ctx context.Context, cfg config.Configuration, connections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch], dependencies bootstrap.RuntimeDependencies) ([]daemons.Daemon, error) {
	if dependencies.FileServiceResolver == nil {
		return nil, fmt.Errorf("file service resolver is required")
	}

	if retainedFileService, err := dependencies.FileServiceResolver.Resolve(storage.FileServiceRetained); err != nil {
		return nil, fmt.Errorf("error resolving FileServiceRetained: %w", err)
	} else {
		return []daemons.Daemon{
			toolapi.NewDaemon(ctx, connections, cfg, schema.DefaultGraphSchema(), retainedFileService),
		}, nil
	}
}

func Entrypoint(ctx context.Context, cfg config.Configuration, connections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch], dependencies bootstrap.RuntimeDependencies) ([]daemons.Daemon, error) {
	if dependencies.FileServiceResolver == nil {
		return nil, fmt.Errorf("file service resolver is required")
	}

	dogtagsService := dogtags.NewDefaultService()

	slog.InfoContext(ctx, "DogTags provider initialized",
		slog.String("namespace", "dogtags"),
		slog.String("provider", dogtagsService.ProviderName()))

	flags := dogtagsService.GetAllDogTags()
	slog.InfoContext(ctx, "DogTags Configuration",
		slog.String("namespace", "dogtags"),
		slog.Any("flags", flags))

	if !cfg.DisableMigrations {
		if err := bootstrap.MigrateDB(ctx, cfg, connections.RDMS, config.NewDefaultAdminConfiguration); err != nil {
			return nil, fmt.Errorf("rdms migration error: %w", err)
		} else if err := migrations.NewGraphMigrator(connections.Graph, migrations.WithSchemalessKindBackfill(connections.RDMS)).Migrate(ctx); err != nil {
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

	// Remove authentication tokens if the APITokens parameter is disabled
	if !appcfg.GetAPITokensParameter(ctx, connections.RDMS) {
		slog.WarnContext(ctx, "APITokens parameter is disabled")
		if dErr := connections.RDMS.DeleteAllAuthTokens(ctx); dErr != nil {
			return nil, fmt.Errorf("failed to delete all auth tokens at startup: %w", dErr)
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
			pipeline               = datapipe.NewPipeline(ctx, cfg, connections.RDMS, connections.Graph, graphQueryCache, ingestSchema, dependencies.FileServiceResolver, cl)
			graphQuery             = queries.NewGraphQuery(connections.Graph, graphQueryCache, cfg)
			authorizer             = auth.NewAuthorizer(connections.RDMS)
			datapipeDaemon         = datapipe.NewDaemon(pipeline, startDelay, time.Duration(cfg.DatapipeInterval)*time.Second, connections.RDMS)
			routerInst             = router.NewRouter(cfg, authorizer, fmt.Sprintf(bootstrap.ContentSecurityPolicy, "", "", "", "", "", ""))
			authenticator          = api.NewAuthenticator(cfg, connections.RDMS, api.NewAuthExtensions(cfg, connections.RDMS))
			openGraphSchemaService = opengraphschema.NewOpenGraphSchemaService(connections.RDMS, connections.Graph)
		)

		registration.RegisterFossGlobalMiddleware(&routerInst, cfg, auth.NewIdentityResolver(), authenticator, connections.RDMS)
		registration.RegisterFossRoutes(&routerInst, cfg, connections.RDMS, connections.Graph, graphQuery, apiCache, collectorManifests, authenticator, authorizer, ingestSchema, dependencies.FileServiceResolver, dogtagsService, openGraphSchemaService)

		modules.Register(modules.Deps{
			Router: &routerInst,
			Pool:   connections.RDMS.Pool(),
		})

		// Set neo4j batch and flush sizes
		neo4jParameters := appcfg.GetNeo4jParameters(ctx, connections.RDMS)
		connections.Graph.SetBatchWriteSize(neo4jParameters.BatchWriteSize)
		connections.Graph.SetWriteFlushSize(neo4jParameters.WriteFlushSize)

		// Register all metrics into a single registry before exposing it via the
		// default registerer so /metrics never observes a partially-initialized state.
		promRegistry := prometheus.NewRegistry()
		if err := metricsregistration.RegisterBHCEMetrics(promRegistry); err != nil {
			return nil, fmt.Errorf("failed to register prometheus metrics: %w", err)
		} else if err := prometheus.DefaultRegisterer.Register(promRegistry); err != nil {
			return nil, fmt.Errorf("failed to expose prometheus registry: %w", err)
		} else if err := connections.RDMS.RequestAnalysis(ctx, "init"); err != nil {
			slog.WarnContext(ctx, "Failed to request init analysis", attr.Error(err))
		}

		return []daemons.Daemon{
			bhapi.NewDaemon(cfg, routerInst.Handler()),
			gc.NewDataPruningDaemon(connections.RDMS),
			cl,
			datapipeDaemon,
		}, nil
	}
}
