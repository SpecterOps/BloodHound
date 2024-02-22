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
	"os"

	"github.com/specterops/bloodhound/dawgs"
	"github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/drivers/pg"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util/size"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api/tools"
	"github.com/specterops/bloodhound/src/config"
)

func ensureDirectory(path string) error {
	if _, err := os.Stat(path); err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		if err := os.MkdirAll(path, 0755); err != nil {
			return fmt.Errorf("unable to create directory %s: %w", path, err)
		}
	}

	return nil
}

// EnsureServerDirectories checks that all required server directories have been set up.
// If they haven't, it attempts to create them. If creation fails, it returns the error.
func EnsureServerDirectories(cfg config.Configuration) error {
	if err := ensureDirectory(cfg.WorkDir); err != nil {
		return err
	}

	if err := ensureDirectory(cfg.TempDirectory()); err != nil {
		return err
	}

	if err := ensureDirectory(cfg.ClientLogDirectory()); err != nil {
		return err
	}

	if err := ensureDirectory(cfg.CollectorsDirectory()); err != nil {
		return err
	}

	return nil
}

// DefaultConfigFilePath returns the location of the config file
func DefaultConfigFilePath() string {
	return "/etc/bhapi/bhapi.json"
}

func ConnectGraph(ctx context.Context, cfg config.Configuration) (*graph.DatabaseSwitch, error) {
	var connectionString string

	if driverName, err := tools.LookupGraphDriver(ctx, cfg); err != nil {
		return nil, err
	} else {
		switch driverName {
		case neo4j.DriverName:
			log.Infof("Connecting to graph using Neo4j")
			connectionString = cfg.Neo4J.Neo4jConnectionString()

		case pg.DriverName:
			log.Infof("Connecting to graph using PostgreSQL")
			connectionString = cfg.Database.PostgreSQLConnectionString()

		default:
			return nil, fmt.Errorf("unknown graphdb driver name: %s", driverName)
		}

		if connectionString == "" {
			return nil, fmt.Errorf("graph connection requires a connection url to be set")
		} else if graphDatabase, err := dawgs.Open(ctx, driverName, dawgs.Config{
			TraversalMemoryLimit: size.Size(cfg.TraversalMemoryLimit) * size.Gibibyte,
			DriverCfg:            connectionString,
		}); err != nil {
			return nil, err
		} else {
			return graph.NewDatabaseSwitch(ctx, graphDatabase), nil
		}
	}
}

// InitializeLogging sets up output file logging, and returns errors if any
func InitializeLogging(cfg config.Configuration) error {
	var logLevel = log.LevelInfo

	if cfg.LogLevel != "" {
		if parsedLevel, err := log.ParseLevel(cfg.LogLevel); err != nil {
			return err
		} else {
			logLevel = parsedLevel
		}
	}

	log.Configure(log.DefaultConfiguration().WithLevel(logLevel))

	log.Infof("Logging configured")
	return nil
}
