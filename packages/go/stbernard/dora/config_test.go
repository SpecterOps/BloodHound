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

// TestGetStoragePath tests storage path resolution
func TestGetStoragePath(t *testing.T) {
	tests := []struct {
		name          string
		configPath    string
		workspaceRoot string
		expectedPath  string
	}{
		{
			name:          "absolute_path",
			configPath:    "/absolute/path/dora.db",
			workspaceRoot: "/workspace",
			expectedPath:  "/absolute/path/dora.db",
		},
		{
			name:          "relative_path",
			configPath:    ".dora/metrics.db",
			workspaceRoot: "/workspace",
			expectedPath:  "/workspace/.dora/metrics.db",
		},
		{
			name:          "relative_with_parent_ref",
			configPath:    "../data/dora.db",
			workspaceRoot: "/workspace/project",
			expectedPath:  "/workspace/data/dora.db", // filepath.Join normalizes the path
		},
		{
			name:          "simple_filename",
			configPath:    "dora.db",
			workspaceRoot: "/workspace",
			expectedPath:  "/workspace/dora.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := Config{
				Storage: StorageConfig{
					Path: tt.configPath,
				},
			}
			result := config.GetStoragePath(tt.workspaceRoot)
			if result != tt.expectedPath {
				t.Errorf("GetStoragePath(%s, %s) = %s, expected %s",
					tt.configPath, tt.workspaceRoot, result, tt.expectedPath)
			}
		})
	}
}

// TestApplyEnvironmentOverrides tests environment variable overrides
func TestApplyEnvironmentOverrides(t *testing.T) {
	// Save original env vars and restore them after test
	originalEnvVars := map[string]string{
		"DORA_GITHUB_OWNER":    os.Getenv("DORA_GITHUB_OWNER"),
		"DORA_GITHUB_REPO":     os.Getenv("DORA_GITHUB_REPO"),
		"DORA_GITHUB_WORKFLOW": os.Getenv("DORA_GITHUB_WORKFLOW"),
		"DORA_STORAGE_PATH":    os.Getenv("DORA_STORAGE_PATH"),
	}
	defer func() {
		for key, value := range originalEnvVars {
			if value == "" {
				os.Unsetenv(key)
			} else {
				os.Setenv(key, value)
			}
		}
	}()

	tests := []struct {
		name         string
		envVars      map[string]string
		initialOwner string
		initialRepo  string
		expectOwner  string
		expectRepo   string
		expectPath   string
	}{
		{
			name: "override_owner_and_repo",
			envVars: map[string]string{
				"DORA_GITHUB_OWNER": "env-owner",
				"DORA_GITHUB_REPO":  "env-repo",
			},
			initialOwner: "config-owner",
			initialRepo:  "config-repo",
			expectOwner:  "env-owner",
			expectRepo:   "env-repo",
			expectPath:   ".dora/metrics.db",
		},
		{
			name: "override_storage_path",
			envVars: map[string]string{
				"DORA_STORAGE_PATH": "/custom/path/db",
			},
			initialOwner: "owner",
			initialRepo:  "repo",
			expectOwner:  "owner",
			expectRepo:   "repo",
			expectPath:   "/custom/path/db",
		},
		{
			name:         "no_env_vars_set",
			envVars:      map[string]string{},
			initialOwner: "owner",
			initialRepo:  "repo",
			expectOwner:  "owner",
			expectRepo:   "repo",
			expectPath:   ".dora/metrics.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all env vars first
			os.Unsetenv("DORA_GITHUB_OWNER")
			os.Unsetenv("DORA_GITHUB_REPO")
			os.Unsetenv("DORA_GITHUB_WORKFLOW")
			os.Unsetenv("DORA_STORAGE_PATH")

			// Set test env vars
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			config := Config{
				GitHub: GitHubConfig{
					Owner: tt.initialOwner,
					Repo:  tt.initialRepo,
				},
				Storage: StorageConfig{
					Path: ".dora/metrics.db",
				},
			}

			config.ApplyEnvironmentOverrides()

			if config.GitHub.Owner != tt.expectOwner {
				t.Errorf("Expected owner %s, got %s", tt.expectOwner, config.GitHub.Owner)
			}
			if config.GitHub.Repo != tt.expectRepo {
				t.Errorf("Expected repo %s, got %s", tt.expectRepo, config.GitHub.Repo)
			}
			if config.Storage.Path != tt.expectPath {
				t.Errorf("Expected path %s, got %s", tt.expectPath, config.Storage.Path)
			}
		})
	}
}
