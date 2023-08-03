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
	"crypto/rand"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/crypto"
)

func usageExit() {
	flag.Usage()
	os.Exit(1)
}

func fatalf(format string, args ...any) {
	fmt.Printf(format, args...)

	if !strings.HasSuffix(format, "\n") {
		fmt.Println()
	}

	os.Exit(1)
}
func newJWTSigningKey() ([]byte, error) {
	signingKey := make([]byte, api.JWTSigningKeyByteLength)

	if _, err := rand.Read(signingKey); err != nil {
		return nil, err
	}

	return signingKey, nil
}

func writeNewConfiguration(path string, skipArgon2 bool) error {
	cfg, err := config.NewDefaultConfiguration()
	if err != nil {
		return fmt.Errorf("failed to create default configuration: %w", err)
	}

	if !skipArgon2 {
		fmt.Println("Tuning argon2 parameters...")

		argon2Digester, err := crypto.Tune(time.Millisecond * 500)
		if err != nil {
			return fmt.Errorf("failed tuning argon2: %w", err)
		}

		cfg.Crypto.Argon2 = config.Argon2Configuration{
			MemoryKibibytes: argon2Digester.MemoryKibibytes,
			NumIterations:   argon2Digester.NumIterations,
			NumThreads:      argon2Digester.NumThreads,
		}
	} else {
		// These are unsafe defaults and should be noted as such in the configuration
		cfg.Crypto.Argon2 = config.Argon2Configuration{
			MemoryKibibytes: 262144,
			NumIterations:   2,
			NumThreads:      1,
		}
	}

	// Set a new random JWT signing key
	if jwtSigningKeyBytes, err := newJWTSigningKey(); err != nil {
		return err
	} else {
		cfg.Crypto.JWT.SetSigningKeyBytes(jwtSigningKeyBytes)
	}

	if err := config.WriteConfigurationFile(path, cfg); err != nil {
		return fmt.Errorf("error writing config: %v", err)
	}

	return nil
}

func updateConfig(path string, cfg config.Configuration) {
	backupCfgPath := fmt.Sprintf("%s.bak", path)

	if err := os.Rename(path, backupCfgPath); err != nil {
		fatalf("Failed to move old configuration %s to %s as a backup: %v", path, backupCfgPath, err)
	}

	if err := config.WriteConfigurationFile(path, cfg); err != nil {
		fatalf("Error writing config: %v", err)
	}
}

func migrate(path string) (config.Configuration, error) {
	if cfg, err := migrateConfiguration(path); err != nil {
		return config.Configuration{}, fmt.Errorf("failed migration configuration %s: %v", path, err)
	} else {
		return cfg, nil
	}
}

func main() {
	var (
		confPath   string
		skipArgon2 bool
	)

	flag.StringVar(&confPath, "f", "", "Path to the configuration file.")
	flag.BoolVar(&skipArgon2, "skip-argon2", false, "Offset automatic argon2 tuning. Only applies to creating new configurations.")
	flag.Parse()

	if confPath == "" {
		fmt.Println("Configuration file path is required.")
		usageExit()
	}

	if _, err := os.Stat(confPath); err != nil {
		if os.IsNotExist(err) {
			if err := writeNewConfiguration(confPath, skipArgon2); err != nil {
				fatalf("Error writing new config: %v", err)
			}
		} else {
			fatalf("Failed checking if configuration %s exists: %v", confPath, err)
		}
	} else if newCfg, err := migrate(confPath); err != nil {
		fatalf("Error during migration: %v", err)
	} else {
		updateConfig(confPath, newCfg)
	}
}
