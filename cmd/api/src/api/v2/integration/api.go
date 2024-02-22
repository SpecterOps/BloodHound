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
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/daemons"
	"github.com/specterops/bloodhound/src/services"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/test/integration/utils"
)

func (s *Context) APIServerURL(paths ...string) string {
	var (
		cfg           = s.GetConfiguration()
		fullPath, err = api.NewJoinedURL(cfg.RootURL.String(), paths...)
	)

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
func (s *Context) EnableAPI() {
	if cfg, err := utils.LoadIntegrationTestConfig(); err != nil {
		s.TestCtrl.Fatalf("Failed loading integration test config: %v", err)
	} else {
		s.waitGroup.Add(1)

		go func() {
			defer s.waitGroup.Done()

			initializer := bootstrap.Initializer[*database.BloodhoundDB, *graph.DatabaseSwitch]{
				Configuration: cfg,
				DBConnector:   services.ConnectDatabases,
				Entrypoint: func(ctx context.Context, cfg config.Configuration, databaseConnections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch]) ([]daemons.Daemon, error) {
					if err := databaseConnections.RDMS.Wipe(); err != nil {
						return nil, err
					}

					return services.Entrypoint(ctx, cfg, databaseConnections)
				},
			}

			if err := initializer.Launch(s.ctx, false); err != nil {
				log.Errorf("Failed launching API server: %v", err)
			}
		}()
	}

	// Wait, at most, 30 seconds for the API to boot
	s.WaitForAPI(time.Second * 30)
}
