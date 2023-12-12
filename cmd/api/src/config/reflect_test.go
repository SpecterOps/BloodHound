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
	"github.com/stretchr/testify/require"
)

func TestSetValue(t *testing.T) {
	var cfg config.Configuration

	t.Run("basic top level key with underscore", func(t *testing.T) {
		assert.Nil(t, config.SetValue(&cfg, "bind_addr", "0.0.0.0"))
		assert.Equal(t, "0.0.0.0", cfg.BindAddress)
	})

	t.Run("two level path with underscore in both keys", func(t *testing.T) {
		assert.Nil(t, config.SetValue(&cfg, "default_admin_expire_now", "true"))
		assert.Equal(t, true, cfg.DefaultAdmin.ExpireNow)
	})

	t.Run("three level path with underscore in bottom key", func(t *testing.T) {
		assert.Nil(t, config.SetValue(&cfg, "crypto_argon2_memory_kibibytes", "10"))
		assert.Equal(t, uint32(10), cfg.Crypto.Argon2.MemoryKibibytes)
	})

	t.Run("key with two underscores", func(t *testing.T) {
		require.Nil(t, config.SetValue(&cfg, "disable_cypher_qc", "true"))
		assert.Equal(t, true, cfg.DisableCypherQC)
	})

	t.Run("key with three underscores", func(t *testing.T) {
		require.Nil(t, config.SetValue(&cfg, "max_graphdb_cache_size", "0"))
		assert.Equal(t, 0, cfg.MaxGraphQueryCacheSize)
	})

	t.Run("edge cases", func(t *testing.T) {
		assert.NotNil(t, config.SetValue(&cfg, "crypto_argon2_memory_kibibytes", "string"))
		assert.NotNil(t, config.SetValue(&cfg, "", "string"))
		assert.NotNil(t, config.SetValue(cfg, "", "string"))
	})
}
