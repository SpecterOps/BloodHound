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

package fixtures

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/daemons"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/services"
	"github.com/specterops/bloodhound/packages/go/lab"
	"github.com/specterops/dawgs/graph"
)

var BHApiFixture = NewApiFixture()

func NewApiFixture() *lab.Fixture[bool] {
	return NewCustomApiFixture(ConfigFixture)
}

func NewCustomApiFixture(cfgFixture *lab.Fixture[config.Configuration]) *lab.Fixture[bool] {
	var (
		ctx, cancel = context.WithCancel(context.Background())
		wg          = &sync.WaitGroup{}
		serverErr   error
	)

	fixture := lab.NewFixture(func(harness *lab.Harness) (bool, error) {
		if cfg, ok := lab.Unpack(harness, cfgFixture); !ok {
			return false, fmt.Errorf("unable to unpack cfgFixture")
		} else {
			// Start the server
			wg.Add(1)

			go func() {
				defer wg.Done()

				initializer := bootstrap.Initializer[*database.BloodhoundDB, *graph.DatabaseSwitch]{
					Configuration: cfg,
					DBConnector:   services.ConnectDatabases,
					Entrypoint: func(ctx context.Context, cfg config.Configuration, databaseConnections bootstrap.DatabaseConnections[*database.BloodhoundDB, *graph.DatabaseSwitch]) ([]daemons.Daemon, error) {
						if err := databaseConnections.RDMS.Wipe(ctx); err != nil {
							return nil, err
						}

						return services.Entrypoint(ctx, cfg, databaseConnections)
					},
				}

				if err := initializer.Launch(ctx, false); err != nil {
					serverErr = err
				}
			}()

			if err := waitForAPI(30*time.Second, cfg.RootURL.String()); err != nil {
				return false, err
			} else {
				return true, nil
			}
		}
	}, func(harness *lab.Harness, started bool) error {
		cancel()
		wg.Wait()

		return serverErr
	})

	if err := lab.SetDependency(fixture, cfgFixture); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("BHApiFixture dependency error: %v", err))
		os.Exit(1)
	}

	return fixture
}

func waitForAPI(timeout time.Duration, serverUrl string) error {
	var (
		started    = time.Now()
		httpClient = http.Client{
			Timeout: time.Second,
		}
	)

	for time.Since(started) < timeout {
		if url, err := url.JoinPath(serverUrl, "health"); err != nil {
			return err
		} else if resp, err := httpClient.Get(url); err != nil {
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
		return fmt.Errorf("timed out waiting for HTTP API to come online")
	}
	return nil
}
