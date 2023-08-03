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

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/database/migration"
	"github.com/specterops/bloodhound/src/server"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
)

func (s *Context) initDatabase() {
	cfg := s.GetConfiguration()

	if db, err := server.ConnectPostgres(cfg); err != nil {
		s.TestCtrl.Fatalf("Failed connecting to databases: %v", err)
	} else if err := integration.Prepare(db); err != nil {
		s.TestCtrl.Fatalf("Failed preparing DB: %v", err)
	} else if err := server.MigrateDB(cfg, db, migration.ListBHModels()); err != nil {
		s.TestCtrl.Fatalf("Failed migrating DB: %v", err)
	} else {
		s.DB = db
	}
}

func (s *Context) GetDatabase() database.Database {
	// If the database has not been initialized, bring it up first
	if s.DB == nil {
		s.initDatabase()
	}

	return s.DB
}

func (s *Context) ClearGraphDB() error {
	return s.Graph.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		return tx.Nodes().Filterf(func() graph.Criteria {
			return query.KindIn(query.Node(), ad.Entity, azure.Entity)
		}).Delete()
	})
}
