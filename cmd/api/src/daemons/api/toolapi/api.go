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

package toolapi

import (
	"context"
	"crypto/tls"
	"errors"

	"log"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/api/tools"
	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/storage"
	"github.com/specterops/dawgs/graph"
)

// Daemon holds data relevant to the tools API daemon
type Daemon struct {
	cfg    config.Configuration
	server *http.Server
}

func NewDaemon[DBType database.Database](ctx context.Context, connections bootstrap.DatabaseConnections[DBType, *graph.DatabaseSwitch], cfg config.Configuration, graphSchema graph.Schema, retainedFileService storage.FileService, extensions ...func(router *chi.Mux)) Daemon {
	var (
		pgMigrator    = tools.NewPGMigrator(ctx, cfg, graphSchema, connections.Graph)
		router        = chi.NewRouter()
		toolContainer = tools.NewToolContainer(connections.RDMS)
		ingestControl = tools.NewIngestControlTool(connections.RDMS, retainedFileService)
	)

	router.Mount("/metrics", promhttp.Handler())

	// Support normal pprof endpoints for easier consumption with standard tools
	router.Mount("/debug/pprof/allocs", pprof.Handler("allocs"))
	router.Mount("/debug/pprof/block", pprof.Handler("block"))
	router.Mount("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	router.Mount("/debug/pprof/heap", pprof.Handler("heap"))
	router.Mount("/debug/pprof/mutex", pprof.Handler("mutex"))
	router.Mount("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	router.Get("/debug/pprof/profile", pprof.Profile)
	router.Get("/debug/pprof/trace", pprof.Trace)

	// TODO: remove old trace handler when we can wire up acumen to handle the above pprof endpoints instead
	router.Get("/trace", tools.NewTraceHandler())

	router.Put("/graph-db/switch/pg", pgMigrator.SwitchPostgreSQL)
	router.Put("/graph-db/switch/neo4j", pgMigrator.SwitchNeo4j)
	router.Put("/pg-migration/pg-to-neo", pgMigrator.MigrationStartPGToNeo)
	router.Put("/pg-migration/neo-to-pg", pgMigrator.MigrationStartNeoToPG)
	router.Get("/pg-migration/status", pgMigrator.MigrationStatus)
	router.Put("/pg-migration/cancel", pgMigrator.MigrationCancel)

	// Allow query of datapipe status for infrastructure tooling
	router.Get("/datapipe/status", func(w http.ResponseWriter, r *http.Request) {
		if dpStatus, err := connections.RDMS.GetDatapipeStatus(ctx); err != nil {
			api.HandleDatabaseError(r, w, err)
		} else {
			api.WriteJSONResponse(r.Context(), dpStatus, http.StatusOK, w)
		}
	})

	// Health endpoint that is online even during migrations
	router.Get("/health", func(response http.ResponseWriter, _ *http.Request) {
		response.WriteHeader(http.StatusOK)
	})

	router.Get("/logging", tools.GetLoggingDetails)
	router.Put("/logging", tools.PutLoggingDetails)

	router.Get("/features", toolContainer.GetFlags)
	router.Put("/features/{feature_id:[0-9]+}/toggle", toolContainer.ToggleFlag)

	router.Get("/analysis/schedule", toolContainer.GetScheduledAnalysisConfiguration)
	router.Put("/analysis/schedule", toolContainer.SetScheduledAnalysisConfiguration)
	router.Get("/parameters", toolContainer.GetApplicationConfigurations)
	router.Put("/parameters", toolContainer.SetApplicationParameter)

	router.Put("/ingest/retention/enable", ingestControl.EnableIngestFileRetention)
	router.Put("/ingest/retention/disable", ingestControl.DisableIngestFileRetention)
	router.Get("/ingest/retention/fetch", ingestControl.FetchRetainedIngestFiles)

	for _, extension := range extensions {
		extension(router)
	}

	return Daemon{
		cfg: cfg,
		server: &http.Server{
			Addr:     cfg.MetricsPort,
			Handler:  router,
			ErrorLog: log.Default(),
			TLSConfig: &tls.Config{
				MinVersion: tls.VersionTLS12,
				CipherSuites: []uint16{
					tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
					tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
					tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
					tls.TLS_AES_128_GCM_SHA256,
					tls.TLS_AES_256_GCM_SHA384,
					tls.TLS_CHACHA20_POLY1305_SHA256,
				},
			},
		},
	}
}

// Name returns the name of the daemon
func (s Daemon) Name() string {
	return "Tools API"
}

// Start begins the daemon and waits for a stop signal in the exit channel
func (s Daemon) Start(ctx context.Context) {
	if s.cfg.TLS.Enabled() {
		if err := s.server.ListenAndServeTLS(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.ErrorContext(ctx, "HTTP server listen error", attr.Error(err))
			}
		}
	} else {
		if err := s.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.ErrorContext(ctx, "HTTP server listen error", attr.Error(err))
			}
		}
	}
}

// Stop passes in a stop signal to the exit channel, thereby killing the daemon
func (s Daemon) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
