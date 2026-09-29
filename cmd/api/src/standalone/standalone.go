// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package standalone

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/serde"
	"github.com/specterops/dawgs/drivers/pg"
)

const configurationFileName = "bhapi.json"

type Options struct {
	DataDirectory        string
	PostgresBinDirectory string
	APIPort              uint16
}

type Instance struct {
	Configuration config.Configuration
	Postgres      Postgres
}

func DefaultDataDirectory() (string, error) {
	userConfigDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolving user configuration directory: %w", err)
	}

	return filepath.Join(userConfigDirectory, "BloodHound"), nil
}

func Prepare(options Options) (Instance, error) {
	if options.DataDirectory == "" {
		return Instance{}, fmt.Errorf("standalone data directory is required")
	}
	if options.APIPort == 0 {
		return Instance{}, fmt.Errorf("standalone API port is required")
	}

	if err := os.MkdirAll(options.DataDirectory, 0o700); err != nil {
		return Instance{}, fmt.Errorf("creating standalone data directory: %w", err)
	}

	configurationPath := filepath.Join(options.DataDirectory, configurationFileName)
	configuration, err := loadConfiguration(configurationPath)
	if err != nil {
		return Instance{}, err
	}

	postgresPort, err := AllocatePort()
	if err != nil {
		return Instance{}, err
	}

	if configuration.Database.Secret == "" {
		configuration.Database.Secret, err = config.GenerateSecureRandomString(32)
		if err != nil {
			return Instance{}, fmt.Errorf("generating local PostgreSQL password: %w", err)
		}
	}

	configuration.BindAddress = fmt.Sprintf("127.0.0.1:%d", options.APIPort)
	configuration.RootURL = serde.MustParseURL(fmt.Sprintf("http://127.0.0.1:%d/", options.APIPort))
	configuration.WorkDir = filepath.Join(options.DataDirectory, "work")
	configuration.CollectorsBasePath = filepath.Join(options.DataDirectory, "collectors")
	configuration.EmbeddedExtensionsBasePath = filepath.Join(options.DataDirectory, "extensions")
	configuration.GraphDriver = pg.DriverName
	configuration.Database.Address = fmt.Sprintf("127.0.0.1:%d", postgresPort)
	configuration.Database.Database = databaseName
	configuration.Database.Username = databaseUser
	configuration.Database.Connection = localConnectionString(configuration.Database.Secret, postgresPort)

	if err := config.WriteConfigurationFile(configurationPath, configuration); err != nil {
		return Instance{}, fmt.Errorf("persisting standalone configuration: %w", err)
	}
	if err := os.Chmod(configurationPath, 0o600); err != nil {
		return Instance{}, fmt.Errorf("protecting standalone configuration: %w", err)
	}

	postgresBinDirectory := options.PostgresBinDirectory
	if postgresBinDirectory == "" {
		postgresRuntimeDirectory, err := EnsureEmbeddedPostgresRuntime(options.DataDirectory)
		if err != nil {
			return Instance{}, err
		}
		postgresBinDirectory = filepath.Join(postgresRuntimeDirectory, "bin")
	}

	return Instance{
		Configuration: configuration,
		Postgres: NewPostgres(
			postgresBinDirectory,
			filepath.Join(options.DataDirectory, "postgres", "data"),
			postgresPort,
			configuration.Database.Secret,
		),
	}, nil
}

func loadConfiguration(configurationPath string) (config.Configuration, error) {
	if _, err := os.Stat(configurationPath); os.IsNotExist(err) {
		configuration, defaultConfigurationErr := config.NewDefaultConfiguration()
		if defaultConfigurationErr != nil {
			return config.Configuration{}, fmt.Errorf("creating standalone configuration: %w", defaultConfigurationErr)
		}
		return configuration, nil
	} else if err != nil {
		return config.Configuration{}, fmt.Errorf("checking standalone configuration: %w", err)
	}

	configuration, err := config.ReadConfigurationFile(configurationPath)
	if err != nil {
		return config.Configuration{}, fmt.Errorf("reading standalone configuration: %w", err)
	}
	return configuration, nil
}

func localConnectionString(password string, port uint16) string {
	return fmt.Sprintf("postgresql://%s:%s@127.0.0.1:%d/%s?sslmode=disable", databaseUser, url.QueryEscape(password), port, databaseName)
}
