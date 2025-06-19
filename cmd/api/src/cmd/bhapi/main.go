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
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/bhlog"
	"github.com/specterops/bloodhound/src/bootstrap"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/services"
	"github.com/specterops/bloodhound/src/version"
	"github.com/specterops/dawgs/graph"
)

func printVersion() {
	fmt.Printf("Bloodhound API Version: %s\n", version.GetVersion())
	os.Exit(0)
}

func main() {
	var (
		configFilePath string
		versionFlag    bool
	)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "BloodHound Community Edition API Server\n\nUsage of %s\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.BoolVar(&versionFlag, "version", false, "Get binary version.")
	flag.StringVar(&configFilePath, "configfile", bootstrap.DefaultConfigFilePath(), "Configuration file to load.")
	flag.Parse()

	if versionFlag {
		printVersion()
	}

	// Jump the bootstrap initializer so all logs are configured properly
	if enabled, err := config.GetTextLoggerEnabled(); err != nil {
		bhlog.ConfigureDefaultJSON(os.Stdout)
		slog.Error(fmt.Sprintf("Failed to check text logger enabled: %v", err))
		os.Exit(1)
	} else if enabled {
		bhlog.ConfigureDefaultText(os.Stdout)
	} else {
		bhlog.ConfigureDefaultJSON(os.Stdout)
	}

	if cfg, err := config.GetConfiguration(configFilePath, config.NewDefaultConfiguration); err != nil {
		slog.Error(fmt.Sprintf("Unable to read configuration %s: %v", configFilePath, err))
		os.Exit(1)
	} else {
		initializer := bootstrap.Initializer[*database.BloodhoundDB, *graph.DatabaseSwitch]{
			Configuration:       cfg,
			DBConnector:         services.ConnectDatabases,
			PreMigrationDaemons: services.PreMigrationDaemons,
			Entrypoint:          services.Entrypoint,
		}

		if err := initializer.Launch(context.Background(), true); err != nil {
			slog.Error(fmt.Sprintf("Failed starting the server: %v", err))
			os.Exit(1)
		}
	}
}
