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

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/specterops/bloodhound/src/database/migration"

	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/migrations"
	"github.com/specterops/bloodhound/src/server"
	"github.com/specterops/bloodhound/src/version"
	"github.com/specterops/bloodhound/log"

	// This import is required by swaggo
	_ "github.com/specterops/bloodhound/src/docs"
)

func printVersion() {
	fmt.Printf("Bloodhound API Version: %s\n", version.GetVersion())
	os.Exit(0)
}

func performMigrationsOnly(cfg config.Configuration) {
	if db, graphDB, err := server.ConnectDatabases(cfg); err != nil {
		log.Fatalf("Failed connecting to databases: %v", err)
	} else if err := db.MigrateModels(migration.ListBHModels()); err != nil {
		log.Fatalf("Migrations failed: %v", err)
	} else {
		var migrator = migrations.NewGraphMigrator(graphDB)
		if err := migrator.Migrate(); err != nil {
			log.Fatalf("Error running migrations for graph db: %v", err)
		}
	}

	fmt.Println("Migrations executed successfully")
}

func main() {
	var (
		configFilePath string
		logFilePath    string
		migrationFlag  bool
		versionFlag    bool
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "BloodHound Community Edition API Server\n\nUsage of %s\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&migrationFlag, "migrate", false, "Only perform database migrations. Do not start the server.")
	flag.BoolVar(&versionFlag, "version", false, "Get binary version.")
	flag.StringVar(&configFilePath, "configfile", server.DefaultConfigFilePath(), "Configuration file to load.")
	flag.StringVar(&logFilePath, "logfile", config.DefaultLogFilePath, "Log file to write to.")
	flag.Parse()

	if versionFlag {
		printVersion()
	}

	// Initialize basic logging facilities while we start up
	log.ConfigureDefaults()

	if cfg, err := config.GetConfiguration(configFilePath); err != nil {
		log.Fatalf("Unable to read configuration %s: %v", configFilePath, err)
	} else if err := server.EnsureServerDirectories(cfg); err != nil {
		log.Fatalf("Fatal error while attempting to ensure working directories: %v", err)
	} else if migrationFlag {
		performMigrationsOnly(cfg)
	} else if err := server.StartServer(cfg, server.SystemSignalExitChannel()); err != nil {
		log.Fatalf("Server start error: %v", err)
	}
}
