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

package server

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/dawgs"
	_ "github.com/specterops/bloodhound/dawgs/drivers/neo4j"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
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

func mustGetWorkingDirectory() string {
	workingDirectory, err := os.Getwd()

	if err != nil {
		fmt.Printf("Unable to lookup working directory: %v", err)
		os.Exit(1)
	}

	return workingDirectory
}

// DefaultConfigFilePath returns the location of the config file
func DefaultConfigFilePath() string {
	return "/etc/bhapi/bhapi.json"
}

// DefaultWorkDirPath returns the default location of the  working directory
func DefaultWorkDirPath() string {
	return filepath.Join(mustGetWorkingDirectory(), "work")
}

// ConnectPostgres initializes a connection to PG, and returns errors if any
func ConnectPostgres(cfg config.Configuration) (*database.BloodhoundDB, error) {
	if db, err := database.OpenDatabase(cfg.Database.PostgreSQLConnectionString()); err != nil {
		return nil, fmt.Errorf("error while attempting to create database connection: %w", err)
	} else {
		return database.NewBloodhoundDB(db, auth.NewIdentityResolver()), nil
	}
}

// ConnectDatabases initializes connections to PG and connection, and returns errors if any
func ConnectDatabases(cfg config.Configuration) (*database.BloodhoundDB, graph.Database, error) {
	if db, err := ConnectPostgres(cfg); err != nil {
		return nil, nil, err
	} else if graphDatabase, err := dawgs.Open("neo4j", cfg.Neo4J.Neo4jConnectionString()); err != nil {
		return nil, nil, err
	} else {
		return db, graphDatabase, nil
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
