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
	schema "github.com/specterops/bloodhound/graphschema"
	"log"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/bootstrap"
)

var GraphDBFixture = NewGraphDBFixture()

func NewGraphDBFixture() *lab.Fixture[*graph.DatabaseSwitch] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (*graph.DatabaseSwitch, error) {
		if config, ok := lab.Unpack(harness, ConfigFixture); !ok {
			return nil, fmt.Errorf("unable to unpack ConfigFixture")
		} else if graphdb, err := bootstrap.ConnectGraph(context.TODO(), config); err != nil {
			return nil, err
		} else if err := bootstrap.MigrateGraph(context.Background(), graphdb, schema.DefaultGraphSchema()); err != nil {
			return nil, fmt.Errorf("failed migrating Graph database: %v", err)
		} else {
			return graph.NewDatabaseSwitch(context.Background(), graphdb), nil
		}
	}, nil)

	if err := lab.SetDependency(fixture, ConfigFixture); err != nil {
		log.Fatalln(err)
	}

	return fixture
}
