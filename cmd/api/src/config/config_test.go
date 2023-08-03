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

package config_test

import (
	"testing"

	"github.com/specterops/bloodhound/src/config"
	"github.com/stretchr/testify/assert"
)

func TestSetValuesFromEnv(t *testing.T) {
	const envPrefix = "test"
	var cfg config.Configuration

	assert.NotNil(t, config.SetValuesFromEnv(envPrefix, &cfg, []string{
		"test_crypto_argon2_memory_kibibytes=not a number value",
	}))

	assert.Nil(t, config.SetValuesFromEnv(envPrefix, &cfg, []string{
		"broken_env_var",
		"test_root_url=https://example.com/?q=query_test",
		"WAYLAND_DISPLAY=wayland-1",
		"XCURSOR_SIZE=24",
		"XDG_CONFIG_DIRS=/etc/xdg",
		"XDG_DATA_DIRS=/usr/local/share:/usr/share",
		"test_bind_addr=0.0.0.0",
		"XDG_RUNTIME_DIR=/run/user/1000",
		"XDG_SEAT=seat0",
		"XDG_SESSION_CLASS=user",
		"XDG_SESSION_ID=1",
		"XDG_SESSION_TYPE=wayland",
		"test_crypto_argon2_memory_kibibytes= 10 ",
		"test_this_path_does_not_exist=10",
	}))

	assert.Equal(t, "https://example.com/?q=query_test", cfg.RootURL.String())
	assert.Equal(t, "0.0.0.0", cfg.BindAddress)
	assert.Equal(t, uint32(10), cfg.Crypto.Argon2.MemoryKibibytes)
}

func TestWritableConfiguration_SetValue(t *testing.T) {
	var cfg config.Configuration

	assert.Nil(t, config.SetValue(&cfg, "bind_addr", "0.0.0.0"))
	assert.Equal(t, "0.0.0.0", cfg.BindAddress)

	assert.Nil(t, config.SetValue(&cfg, "crypto_argon2_memory_kibibytes", "10"))
	assert.Equal(t, uint32(10), cfg.Crypto.Argon2.MemoryKibibytes)

	assert.NotNil(t, config.SetValue(&cfg, "crypto_argon2_memory_kibibytes", "string"))
	assert.Nil(t, config.SetValue(&cfg, "crypto_fake", "string"))
	assert.NotNil(t, config.SetValue(&cfg, "", "string"))
	assert.NotNil(t, config.SetValue(cfg, "", "string"))
}

func TestDatabaseConfiguration(t *testing.T) {
	const envPrefix = "bhe"
	var cfg config.Configuration

	assert.Nil(t, config.SetValuesFromEnv(envPrefix, &cfg, []string{
		"bhe_neo4j_addr=localhost:7070",
		"bhe_neo4j_database=neo4j",
		"bhe_neo4j_username=neo4j",
		"bhe_neo4j_secret=neo4jj",

		"bhe_database_addr=localhost:5432",
		"bhe_database_database=bhe",
		"bhe_database_username=bhe",
		"bhe_database_secret=bhe4eva",
	}))

	assert.Equal(t, "neo4j://neo4j:neo4jj@localhost:7070/neo4j", cfg.Neo4J.Neo4jConnectionString())
	assert.Equal(t, "postgresql://bhe:bhe4eva@localhost:5432/bhe", cfg.Database.PostgreSQLConnectionString())
}
