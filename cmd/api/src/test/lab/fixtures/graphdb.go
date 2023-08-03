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
	"fmt"
	"log"

	"github.com/specterops/bloodhound/src/server"
	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
)

var GraphDBFixture = NewGraphDBFixture()

func NewGraphDBFixture() *lab.Fixture[graph.Database] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (graph.Database, error) {
		if config, ok := lab.Unpack(harness, ConfigFixture); !ok {
			return nil, fmt.Errorf("unable to unpack ConfigFixture")
		} else if graphdb, err := dawgs.Open(neo4j.DriverName, config.Neo4J.Neo4jConnectionString()); err != nil {
			return graphdb, err
		} else if err := server.MigrateGraph(graphdb); err != nil {
			return graphdb, fmt.Errorf("failed migrating Graph database: %v", err)
		} else {
			return graphdb, nil
		}
	}, nil)
	if err := lab.SetDependency(fixture, ConfigFixture); err != nil {
		log.Fatalln(err)
	}
	return fixture
}
