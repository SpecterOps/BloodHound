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
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/server"
	"github.com/specterops/bloodhound/lab"
)

var BHApiFixture = NewApiFixture(integration.StartBHServer)

func NewApiFixture(startFn integration.APIStartFunc) *lab.Fixture[integration.APIServerContext] {
	var (
		ctx       context.Context
		cancel    context.CancelFunc
		wg        *sync.WaitGroup
		serverErr error

		dependencyErrs = make([]error, 0)
		fixture        = lab.NewFixture(func(harness *lab.Harness) (integration.APIServerContext, error) {
			ctx, cancel = context.WithCancel(context.Background())
			wg = &sync.WaitGroup{}
			out := integration.APIServerContext{}
			if config, ok := lab.Unpack(harness, ConfigFixture); !ok {
				return out, fmt.Errorf("unable to unpack ConfigFixture")
			} else if err := server.EnsureServerDirectories(config); err != nil {
				return out, err
			} else if pgdb, ok := lab.Unpack(harness, PostgresFixture); !ok {
				return out, fmt.Errorf("unable to unpack PostgresFixture")
			} else if graphdb, ok := lab.Unpack(harness, GraphDBFixture); !ok {
				return out, fmt.Errorf("unable to unpack GraphDBFixture")
			} else if graphcache, ok := lab.Unpack(harness, GraphCacheFixture); !ok {
				return out, fmt.Errorf("unable to unpack GraphCacheFixture")
			} else if apicache, ok := lab.Unpack(harness, ApiCacheFixture); !ok {
				return out, fmt.Errorf("unable to unpack GraphCacheFixture")
			} else {
				out.Context = ctx
				out.DB = pgdb
				out.GraphDB = graphdb
				out.Configuration = config
				out.APICache = apicache
				out.GraphQueryCache = graphcache
				// Start the server
				wg.Add(1)
				go func() {
					defer wg.Done()
					serverErr = startFn(out)
				}()

				if err := waitForAPI(30*time.Second, config.RootURL.String()); err != nil {
					return out, err
				} else {
					return out, nil
				}
			}
		}, func(harness *lab.Harness, apiServerCtx integration.APIServerContext) error {
			cancel()
			wg.Wait()
			return serverErr
		})
	)

	dependencyErrs = append(dependencyErrs, lab.SetDependency(fixture, ConfigFixture))
	dependencyErrs = append(dependencyErrs, lab.SetDependency(fixture, PostgresFixture))
	dependencyErrs = append(dependencyErrs, lab.SetDependency(fixture, GraphDBFixture))
	dependencyErrs = append(dependencyErrs, lab.SetDependency(fixture, ApiCacheFixture))
	dependencyErrs = append(dependencyErrs, lab.SetDependency(fixture, GraphCacheFixture))

	if err := errors.Join(dependencyErrs...); err != nil {
		log.Fatalf("Errors encountered while setting up dependencies:\n%v\n", err)
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
