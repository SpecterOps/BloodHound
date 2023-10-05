package config_test

import (
	"testing"

	"github.com/specterops/bloodhound/src/config"
	"github.com/stretchr/testify/assert"
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

	t.Run("attempting to set a value to an unknown field should not fail", func(t *testing.T) {
		assert.Nil(t, config.SetValue(&cfg, "crypto_fake", "string"))
	})

	t.Run("edge cases", func(t *testing.T) {
		assert.NotNil(t, config.SetValue(&cfg, "crypto_argon2_memory_kibibytes", "string"))
		assert.NotNil(t, config.SetValue(&cfg, "", "string"))
		assert.NotNil(t, config.SetValue(cfg, "", "string"))
	})
}
