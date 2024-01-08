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
	"context"
	"flag"
	"fmt"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/services"
	"github.com/specterops/bloodhound/src/version"
	"os"

	// This import is required by swaggo
	_ "github.com/specterops/bloodhound/src/docs"
)

func printVersion() {
	fmt.Printf("Bloodhound API Version: %s\n", version.GetVersion())
	os.Exit(0)
}

func main() {
	var (
		configFilePath string
		logFilePath    string
		versionFlag    bool
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "BloodHound Community Edition API Server\n\nUsage of %s\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&versionFlag, "version", false, "Get binary version.")
	flag.StringVar(&configFilePath, "configfile", bootstrap.DefaultConfigFilePath(), "Configuration file to load.")
	flag.StringVar(&logFilePath, "logfile", config.DefaultLogFilePath, "Log file to write to.")
	flag.Parse()

	if versionFlag {
		printVersion()
	}

	// Initialize basic logging facilities while we start up
	log.ConfigureDefaults()

	if cfg, err := config.GetConfiguration(configFilePath, config.NewDefaultConfiguration); err != nil {
		log.Fatalf("Unable to read configuration %s: %v", configFilePath, err)
	} else {
		initializer := bootstrap.Initializer[*database.BloodhoundDB, *graph.DatabaseSwitch]{
			Configuration: cfg,
			DBConnector:   services.ConnectDatabases,
			Entrypoint:    services.Entrypoint,
		}

		if err := initializer.Launch(context.Background(), true); err != nil {
			log.Fatalf("Failed starting the server: %v", err)
		}
	}
}
