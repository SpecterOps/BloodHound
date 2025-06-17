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

package bootstrap

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/specterops/dawgs/graph"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/daemons"
	"github.com/specterops/bloodhound/src/database"
)

type DatabaseConnections[DBType database.Database, GraphType graph.Database] struct {
	RDMS  DBType
	Graph GraphType
}

type DatabaseConstructor[DBType database.Database, GraphType graph.Database] func(ctx context.Context, cfg config.Configuration) (DatabaseConnections[DBType, GraphType], error)
type InitializerLogic[DBType database.Database, GraphType graph.Database] func(ctx context.Context, cfg config.Configuration, databaseConnections DatabaseConnections[DBType, GraphType]) ([]daemons.Daemon, error)

type Initializer[DBType database.Database, GraphType graph.Database] struct {
	Configuration       config.Configuration
	PreMigrationDaemons InitializerLogic[DBType, GraphType]
	Entrypoint          InitializerLogic[DBType, GraphType]
	DBConnector         DatabaseConstructor[DBType, GraphType]
}

func (s Initializer[DBType, GraphType]) Launch(parentCtx context.Context, handleSignals bool) error {
	var (
		ctx                 = parentCtx
		daemonManager       = daemons.NewManager(DefaultServerShutdownTimeout)
		databaseConnections DatabaseConnections[DBType, GraphType]
		err                 error
	)

	if handleSignals {
		ctx = NewDaemonContext(parentCtx)
	}

	if err := InitializeLogging(s.Configuration); err != nil {
		return fmt.Errorf("log initialization error: %w", err)
	}

	if err := EnsureServerDirectories(s.Configuration); err != nil {
		return fmt.Errorf("failed to ensure server directories: %w", err)
	}

	if databaseConnections, err = s.DBConnector(ctx, s.Configuration); err != nil {
		return fmt.Errorf("failed to connect to databases: %w", err)
	}
	// Ensure that the database instances are closed once we're ready to exit regardless
	defer databaseConnections.RDMS.Close(ctx)
	defer databaseConnections.Graph.Close(ctx)

	// Daemons that start prior to blocking db migration
	if s.PreMigrationDaemons != nil {
		if daemonInstances, err := s.PreMigrationDaemons(ctx, s.Configuration, databaseConnections); err != nil {
			return fmt.Errorf("failed to start services: %w", err)
		} else {
			daemonManager.Start(ctx, daemonInstances...)
		}
	}

	// Daemons that start after blocking db migration
	if daemonInstances, err := s.Entrypoint(ctx, s.Configuration, databaseConnections); err != nil {
		return fmt.Errorf("failed to start services: %w", err)
	} else {
		daemonManager.Start(ctx, daemonInstances...)
	}

	// Log successful start and wait for a signal to exit
	slog.InfoContext(ctx, "Server started successfully")
	<-ctx.Done()

	slog.InfoContext(ctx, "Shutting down")
	// TODO: Refactor this pattern in favor of context handling
	daemonManager.Stop()

	return nil
}
