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
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/api/tools"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
)

// Daemon holds data relevant to the tools API daemon
type Daemon struct {
	cfg    config.Configuration
	server *http.Server
}

func NewDaemon[DBType database.Database](ctx context.Context, connections bootstrap.DatabaseConnections[DBType, *graph.DatabaseSwitch], cfg config.Configuration, graphSchema graph.Schema, extensions ...func(router *chi.Mux)) Daemon {
	var (
		pgMigrator    = tools.NewPGMigrator(ctx, cfg, graphSchema, connections.Graph)
		router        = chi.NewRouter()
		toolContainer = tools.NewToolContainer(connections.RDMS)
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
	router.Put("/pg-migration/start", pgMigrator.MigrationStart)
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

	for _, extension := range extensions {
		extension(router)
	}

	return Daemon{
		cfg: cfg,
		server: &http.Server{
			Addr:     cfg.MetricsPort,
			Handler:  router,
			ErrorLog: log.Adapter(log.LevelError, "ToolAPI", 0),
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
				slog.ErrorContext(ctx, fmt.Sprintf("HTTP server listen error: %v", err))
			}
		}
	} else {
		if err := s.server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				slog.ErrorContext(ctx, fmt.Sprintf("HTTP server listen error: %v", err))
			}
		}
	}
}

// Stop passes in a stop signal to the exit channel, thereby killing the daemon
func (s Daemon) Stop(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}
