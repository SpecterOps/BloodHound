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
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

const (
	ConfigFileName      = ".dora.yaml"
	LocalConfigFileName = ".dora.local.yaml"
	DoraDataDir         = ".dora"
	TokensDir           = "tokens"
	DefaultDBName       = "dora.db"
)

var (
	ErrConfigNotFound    = errors.New("config file not found")
	ErrInvalidConfig     = errors.New("invalid configuration")
	ErrMissingGitHubRepo = errors.New("github repository configuration is required")
)

// Config represents the complete DORA metrics configuration
type Config struct {
	GitHub  GitHubConfig  `yaml:"github"`
	JIRA    JIRAConfig    `yaml:"jira,omitempty"`
	Storage StorageConfig `yaml:"storage"`
	Metrics MetricsConfig `yaml:"metrics"`
}

// GitHubConfig holds GitHub-specific settings
type GitHubConfig struct {
	Owner      string           `yaml:"owner"`
	Repo       string           `yaml:"repo"`
	Production ProductionConfig `yaml:"production"`
}

// ProductionConfig defines how to identify production deployments
type ProductionConfig struct {
	Workflow    string `yaml:"workflow"`
	Environment string `yaml:"environment"`
}

// JIRAConfig holds JIRA-specific settings
type JIRAConfig struct {
	Domain      string   `yaml:"domain"`
	ProjectKeys []string `yaml:"project_keys"`
}

// StorageConfig defines data storage settings
type StorageConfig struct {
	Type string `yaml:"type"` // Currently only "sqlite"
	Path string `yaml:"path"` // Relative to workspace root
}

// MetricsConfig holds metrics calculation settings
type MetricsConfig struct {
	DefaultPeriod string `yaml:"default_period"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() Config {
	return Config{
		GitHub: GitHubConfig{
			Owner: "SpecterOps",
			Repo:  "bloodhound-enterprise",
			Production: ProductionConfig{
				Workflow:    "cicd-distroless.yml",
				Environment: "production",
			},
		},
		Storage: StorageConfig{
			Type: "sqlite",
			Path: filepath.Join(DoraDataDir, DefaultDBName),
		},
		Metrics: MetricsConfig{
			DefaultPeriod: "30d",
		},
	}
}

// Validate checks if the configuration is valid
func (s Config) Validate() error {
	if s.GitHub.Owner == "" || s.GitHub.Repo == "" {
		return fmt.Errorf("%w: github.owner and github.repo are required", ErrMissingGitHubRepo)
	}
	return nil
}

// SaveToFile writes the configuration to a YAML file
func (s Config) SaveToFile(path string) error {
	data, err := yaml.Marshal(s)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("creating config directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing config file: %w", err)
	}

	return nil
}

// LoadConfigFromFile reads a configuration from a YAML file
func LoadConfigFromFile(path string) (Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return config, fmt.Errorf("%w: %s", ErrConfigNotFound, path)
		}
		return config, fmt.Errorf("reading config file: %w", err)
	}

	if err := yaml.Unmarshal(data, &config); err != nil {
		return config, fmt.Errorf("parsing config file: %w", err)
	}

	return config, nil
}

// LoadConfig loads configuration from workspace root with optional local overrides
func LoadConfig(workspaceRoot string) (Config, error) {
	var (
		baseConfigPath  = filepath.Join(workspaceRoot, ConfigFileName)
		localConfigPath = filepath.Join(workspaceRoot, LocalConfigFileName)
	)

	// Load base config (required)
	config, err := LoadConfigFromFile(baseConfigPath)
	if err != nil {
		return config, err
	}

	// Attempt to load local override (optional)
	if localConfig, err := LoadConfigFromFile(localConfigPath); err == nil {
		// Merge local config over base config
		config = mergeConfigs(config, localConfig)
	} else if !errors.Is(err, ErrConfigNotFound) {
		// Only error if it's not a "file not found" error
		return config, fmt.Errorf("loading local config: %w", err)
	}

	return config, config.Validate()
}

// mergeConfigs merges override config into base config
// Non-zero values in override take precedence
func mergeConfigs(base, override Config) Config {
	result := base

	// Merge GitHub config
	if override.GitHub.Owner != "" {
		result.GitHub.Owner = override.GitHub.Owner
	}
	if override.GitHub.Repo != "" {
		result.GitHub.Repo = override.GitHub.Repo
	}
	if override.GitHub.Production.Workflow != "" {
		result.GitHub.Production.Workflow = override.GitHub.Production.Workflow
	}
	if override.GitHub.Production.Environment != "" {
		result.GitHub.Production.Environment = override.GitHub.Production.Environment
	}

	// Merge JIRA config
	if override.JIRA.Domain != "" {
		result.JIRA.Domain = override.JIRA.Domain
	}
	if len(override.JIRA.ProjectKeys) > 0 {
		result.JIRA.ProjectKeys = override.JIRA.ProjectKeys
	}

	// Merge Storage config
	if override.Storage.Type != "" {
		result.Storage.Type = override.Storage.Type
	}
	if override.Storage.Path != "" {
		result.Storage.Path = override.Storage.Path
	}

	// Merge Metrics config
	if override.Metrics.DefaultPeriod != "" {
		result.Metrics.DefaultPeriod = override.Metrics.DefaultPeriod
	}

	return result
}

// ApplyEnvironmentOverrides applies environment variable overrides to the config
func (s *Config) ApplyEnvironmentOverrides() {
	if owner := os.Getenv("DORA_GITHUB_OWNER"); owner != "" {
		s.GitHub.Owner = owner
	}
	if repo := os.Getenv("DORA_GITHUB_REPO"); repo != "" {
		s.GitHub.Repo = repo
	}
	if workflow := os.Getenv("DORA_GITHUB_WORKFLOW"); workflow != "" {
		s.GitHub.Production.Workflow = workflow
	}
	if domain := os.Getenv("DORA_JIRA_DOMAIN"); domain != "" {
		s.JIRA.Domain = domain
	}
	if dbPath := os.Getenv("DORA_STORAGE_PATH"); dbPath != "" {
		s.Storage.Path = dbPath
	}
}

// GetStoragePath returns the absolute path to the storage file
func (s Config) GetStoragePath(workspaceRoot string) string {
	if filepath.IsAbs(s.Storage.Path) {
		return s.Storage.Path
	}
	return filepath.Join(workspaceRoot, s.Storage.Path)
}

// GetTokensDir returns the absolute path to the tokens directory
func (s Config) GetTokensDir(workspaceRoot string) string {
	return filepath.Join(workspaceRoot, DoraDataDir, TokensDir)
}
