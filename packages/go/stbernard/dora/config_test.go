// Copyright 2026 Specter Ops, Inc.
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

package dora

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.GitHub.Owner == "" {
		t.Error("Expected default GitHub owner to be set")
	}
	if config.GitHub.Repo == "" {
		t.Error("Expected default GitHub repo to be set")
	}
	if config.Storage.Type != "sqlite" {
		t.Errorf("Expected storage type to be 'sqlite', got '%s'", config.Storage.Type)
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "valid config",
			config:  DefaultConfig(),
			wantErr: false,
		},
		{
			name: "missing github owner",
			config: Config{
				GitHub: GitHubConfig{
					Repo: "test-repo",
				},
			},
			wantErr: true,
		},
		{
			name: "missing github repo",
			config: Config{
				GitHub: GitHubConfig{
					Owner: "test-owner",
				},
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.wantErr && err == nil {
				t.Error("Expected validation error, got nil")
			}
			if !tc.wantErr && err != nil {
				t.Errorf("Expected no validation error, got: %v", err)
			}
		})
	}
}

func TestConfigLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, ".dora.yaml")

	// Create a config and save it
	config := DefaultConfig()
	config.GitHub.Owner = "test-owner"
	config.GitHub.Repo = "test-repo"

	if err := config.SaveToFile(configPath); err != nil {
		t.Fatalf("Failed to save config: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		t.Fatal("Config file was not created")
	}

	// Load the config back
	loadedConfig, err := LoadConfigFromFile(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if loadedConfig.GitHub.Owner != "test-owner" {
		t.Errorf("Expected owner 'test-owner', got '%s'", loadedConfig.GitHub.Owner)
	}
	if loadedConfig.GitHub.Repo != "test-repo" {
		t.Errorf("Expected repo 'test-repo', got '%s'", loadedConfig.GitHub.Repo)
	}
}

func TestLoadConfigWithOverride(t *testing.T) {
	tempDir := t.TempDir()
	baseConfigPath := filepath.Join(tempDir, ".dora.yaml")
	localConfigPath := filepath.Join(tempDir, ".dora.local.yaml")

	// Create base config
	baseConfig := DefaultConfig()
	baseConfig.GitHub.Owner = "base-owner"
	baseConfig.GitHub.Repo = "base-repo"
	if err := baseConfig.SaveToFile(baseConfigPath); err != nil {
		t.Fatalf("Failed to save base config: %v", err)
	}

	// Create local override config
	localConfig := Config{
		GitHub: GitHubConfig{
			Owner: "override-owner",
		},
	}
	if err := localConfig.SaveToFile(localConfigPath); err != nil {
		t.Fatalf("Failed to save local config: %v", err)
	}

	// Load with override
	merged, err := LoadConfig(tempDir)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	if merged.GitHub.Owner != "override-owner" {
		t.Errorf("Expected overridden owner 'override-owner', got '%s'", merged.GitHub.Owner)
	}
	if merged.GitHub.Repo != "base-repo" {
		t.Errorf("Expected base repo 'base-repo', got '%s'", merged.GitHub.Repo)
	}
}
