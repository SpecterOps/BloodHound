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
	"io"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/cmd/api/src/bootstrap"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/services"
	"github.com/specterops/bloodhound/cmd/api/src/version"
	"github.com/specterops/bloodhound/packages/go/bhlog"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/level"
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

	// Eagerly set logging format if valid environment variable is set
	bhlog.ConfigureDefaultJSON(os.Stdout)
	if config.GetTextLoggerEnabled() {
		bhlog.ConfigureDefaultText(os.Stdout)
	}

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

	cfg, err := config.GetConfiguration(configFilePath, config.NewDefaultConfiguration)
	if err != nil {
		slog.Error(fmt.Sprintf("Unable to read configuration %s: %v", configFilePath, err))
		os.Exit(1)
	}

	// Initialize logging
	var (
		logFile *os.File

		logLevel            = slog.LevelInfo
		logWriter io.Writer = os.Stdout
	)

	if cfg.LogPath != "" {
		logFile, err = os.OpenFile(cfg.LogPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			slog.Error(
				"Failed to configure logging to file",
				slog.String("path", cfg.LogPath),
				attr.Error(err),
			)
		} else {
			defer logFile.Close()
			slog.Info("Additionally logging to file", slog.String("log_file", cfg.LogPath))
			logWriter = io.MultiWriter(logWriter, logFile)
		}
	}

	if cfg.LogLevel != "" {
		if parsedLevel, err := bhlog.ParseLevel(cfg.LogLevel); err != nil {
			slog.Warn("Configured log level is invalid. Ignoring.", slog.String("requested_log_level", cfg.LogLevel))
		} else {
			logLevel = parsedLevel
		}
	}

	if cfg.EnableTextLogger {
		bhlog.ConfigureDefaultText(logWriter)
	} else {
		bhlog.ConfigureDefaultJSON(logWriter)
	}

	level.SetGlobalLevel(logLevel)
	slog.Info("Logging configured", slog.String("log_level", logLevel.String()))

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
