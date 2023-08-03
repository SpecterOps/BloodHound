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

package integration

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/daemons/gc"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/migration"
	"github.com/specterops/bloodhound/src/server"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/bloodhound/src/test/integration/utils"
)

type APIServerContext struct {
	Context         context.Context
	DB              *database.BloodhoundDB
	GraphDB         graph.Database
	Configuration   config.Configuration
	APICache        cache.Cache
	GraphQueryCache cache.Cache
}

type APIStartFunc func(ctx APIServerContext) error

func StartBHServer(apiServerContext APIServerContext) error {
	if err := server.InitializeLogging(apiServerContext.Configuration); err != nil {
		return fmt.Errorf("log initialization error: %w", err)
	}

	var (
		serviceManager         = daemons.NewManager(server.DefaultServerShutdownTimeout)
		sessionSweepingService = gc.NewDataPruningDaemon(apiServerContext.DB)
		routerInst             = router.NewRouter(apiServerContext.Configuration, auth.NewAuthorizer(), server.ContentSecurityPolicy)
		fakeManifests          = config.CollectorManifests{}
		datapipeDaemon         = datapipe.NewDaemon(apiServerContext.Configuration, apiServerContext.DB, apiServerContext.GraphDB, apiServerContext.GraphQueryCache, time.Second)
		authenticator          = api.NewAuthenticator(apiServerContext.Configuration, apiServerContext.DB, database.NewContextInitializer(apiServerContext.DB))
	)

	registration.RegisterFossGlobalMiddleware(&routerInst, apiServerContext.Configuration, authenticator)
	registration.RegisterFossRoutes(
		&routerInst,
		apiServerContext.Configuration,
		apiServerContext.DB,
		apiServerContext.GraphDB,
		apiServerContext.APICache,
		apiServerContext.GraphQueryCache,
		fakeManifests,
		authenticator,
		datapipeDaemon,
	)
	apiDaemon := bhapi.NewDaemon(apiServerContext.Configuration, routerInst.Handler())

	// Start daemons
	serviceManager.Start(apiDaemon, sessionSweepingService, datapipeDaemon)

	// Wait for a signal to exit
	<-apiServerContext.Context.Done()
	serviceManager.Stop()

	return nil
}

func (s *Context) APIServerURL(paths ...string) string {
	fullPath, err := api.NewJoinedURL(s.cfg.RootURL.String(), paths...)

	if err != nil {
		s.TestCtrl.Fatalf("Bad API server URL paths specified: %v. Paths: %v", err, paths)
	}

	return fullPath
}

func (s *Context) WaitForAPI(timeout time.Duration) {
	var (
		started    = time.Now()
		httpClient = http.Client{
			Timeout: time.Second,
		}
	)

	for time.Since(started) < timeout {
		if resp, err := httpClient.Get(s.APIServerURL("health")); err != nil {
			time.Sleep(time.Second)
		} else {
			// Close the response right away, we don't need the body
			resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				break
			}
		}
	}

	if time.Since(started) >= timeout {
		s.TestCtrl.Fatalf("timed out waiting for HTTP API to come online")
	}
}

// EnableAPI loads all dependencies and starts up a new API server
func (s *Context) EnableAPI(startFunc APIStartFunc) {
	log.Infof("Starting up integration test harness")

	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		s.TestCtrl.Fatalf("Failed loading integration test config: %v", err)
	} else if err := server.EnsureServerDirectories(cfg); err != nil {
		s.TestCtrl.Fatalf("Failed ensuring integration test directories: %v", err)
	} else if db, graphDB, err := server.ConnectDatabases(cfg); err != nil {
		s.TestCtrl.Fatalf("Failed connecting to databases: %v", err)
	} else if err := integration.Prepare(db); err != nil {
		s.TestCtrl.Fatalf("Failed ensuring database: %v", err)
	} else if err := server.MigrateDB(cfg, db, migration.ListBHModels()); err != nil {
		s.TestCtrl.Fatalf("Failed migrating database: %v", err)
	} else if err := server.MigrateGraph(graphDB); err != nil {
		s.TestCtrl.Fatalf("Failed migrating Graph database: %v", err)
	} else if apiCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxAPICacheSize}); err != nil {
		s.TestCtrl.Fatalf("Failed to create in-memory cache for API: %v", err)
	} else if graphQueryCache, err := cache.NewCache(cache.Config{MaxSize: cfg.MaxGraphQueryCacheSize}); err != nil {
		s.TestCtrl.Fatalf("Failed to create in-memory cache for graphDB: %v", err)
	} else {
		s.DB = db
		s.Graph = graphDB
		s.cfg = &cfg
		// Start the HTTP API
		s.WaitGroup.Add(1)

		go func() {
			defer s.WaitGroup.Done()

			if err := startFunc(APIServerContext{
				Context:         s.Ctx,
				DB:              db,
				GraphDB:         graphDB,
				Configuration:   cfg,
				APICache:        apiCache,
				GraphQueryCache: graphQueryCache,
			}); err != nil {
				fmt.Printf("Error running HTTP API: %v", err)
			}
		}()

	}

	// Wait, at most, 30 seconds for the API to boot
	s.WaitForAPI(time.Second * 30)
}
