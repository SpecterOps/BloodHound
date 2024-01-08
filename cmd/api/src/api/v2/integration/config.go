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
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/serde"
	"github.com/specterops/bloodhound/src/test/integration/utils"
)

func (s *Context) loadConfiguration() {
	cfg, err := utils.LoadIntegrationTestConfig()

	if err != nil {
		s.TestCtrl.Fatalf("Failed loading integration test config: %v", err)
	}

	s.cfg = &cfg
}

func (s *Context) GetConfiguration() config.Configuration {
	// Load the configuration if it has not already been pulled in
	if s.cfg == nil {
		s.loadConfiguration()
	}

	return *s.cfg
}

func (s *Context) GetRootURL() *serde.URL {
	cfg := s.GetConfiguration()
	return &cfg.RootURL
}
