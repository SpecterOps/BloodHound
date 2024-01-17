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
	t.Run("simulated env", func(t *testing.T) {
		const envPrefix = "test"
		var cfg config.Configuration

		assert.NotNil(t, config.SetValuesFromEnv(envPrefix, &cfg, []string{
			"test_crypto_argon2_memory_kibibytes=not a number value",
		}))

		assert.Nil(t, config.SetValuesFromEnv(envPrefix, &cfg, []string{
			"broken_env_var",
			"TEST_ROOT_URL=https://example.com/?q=query_test",
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
	})

	t.Run("database configuration", func(t *testing.T) {
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
			"bhe_database_secret=supersecretpassword",
		}))

		assert.Equal(t, "neo4j://neo4j:neo4jj@localhost:7070/neo4j", cfg.Neo4J.Neo4jConnectionString())
		assert.Equal(t, "postgresql://bhe:supersecretpassword@localhost:5432/bhe", cfg.Database.PostgreSQLConnectionString())
	})

	// This test ensures that fields that could be considered sensitive are configurable through expected environment
	// variables. Not all fields in a given struct are necessarily sensitive and should not be included here.
	t.Run("all sensitive fields are configurable", func(t *testing.T) {
		const (
			envPrefix = "bhe"

			SAMLSPCERT        = "bhe_saml_sp_cert"
			SAMLSPKEY         = "bhe_saml_sp_key"
			TLSCERTFILE       = "bhe_tls_cert_file"
			TLSKEYFILE        = "bhe_tls_key_file"
			DBCONN            = "bhe_database_connection"
			DBADDR            = "bhe_database_addr"
			DBUSER            = "bhe_database_username"
			DBSECRET          = "bhe_database_secret"
			DBDB              = "bhe_database_database"
			NEOCONN           = "bhe_neo4j_connection"
			NEOADDR           = "bhe_neo4j_addr"
			NEOUSER           = "bhe_neo4j_username"
			NEOSECRET         = "bhe_neo4j_secret"
			NEODB             = "bhe_neo4j_database"
			JWTSIGNKEY        = "bhe_crypto_jwt_signing_key"
			DEFADMINPRINCNAME = "bhe_default_admin_principal_name"
			DEFADMINPASS      = "bhe_default_admin_password"
			DEFADMINEMAIL     = "bhe_default_admin_email_address"
			DEFADMINFIRST     = "bhe_default_admin_first_name"
			DEFADMINLAST      = "bhe_default_admin_last_name"
		)

		var (
			cfg        config.Configuration
			envOptions []string

			options = map[string]string{
				SAMLSPCERT:        "thisisacert",
				SAMLSPKEY:         "thisisacertkey",
				TLSCERTFILE:       "thisisacertfilepath",
				TLSKEYFILE:        "thisisakeyfilepath",
				DBCONN:            "databaseconnectionstring",
				DBADDR:            "databaseaddress",
				DBUSER:            "databaseuser",
				DBSECRET:          "databasesecret",
				DBDB:              "databasedatabase",
				NEOCONN:           "neo4jconnectionstring",
				NEOADDR:           "neo4jaddress",
				NEOUSER:           "neo4juser",
				NEOSECRET:         "neo4jsecret",
				NEODB:             "neo4jdatabase",
				JWTSIGNKEY:        "jwtsigningkey",
				DEFADMINPRINCNAME: "defaultadminprincipalname",
				DEFADMINPASS:      "defaultadminpassword",
				DEFADMINEMAIL:     "defaultadminemailaddress",
				DEFADMINFIRST:     "defaultadminfirstname",
				DEFADMINLAST:      "defaultadminlastname",
			}
		)

		envOptions = make([]string, 0, len(options))

		for k, v := range options {
			envOptions = append(envOptions, k+"="+v)
		}

		assert.Nil(t, config.SetValuesFromEnv(envPrefix, &cfg, envOptions))

		t.Run("saml", func(t *testing.T) {
			assert.Equal(t, options[SAMLSPCERT], cfg.SAML.ServiceProviderCertificate)
			assert.Equal(t, options[SAMLSPKEY], cfg.SAML.ServiceProviderKey)
		})

		t.Run("tls", func(t *testing.T) {
			assert.Equal(t, options[TLSCERTFILE], cfg.TLS.CertFile)
			assert.Equal(t, options[TLSKEYFILE], cfg.TLS.KeyFile)
		})

		t.Run("database", func(t *testing.T) {
			assert.Equal(t, options[DBCONN], cfg.Database.Connection)
			assert.Equal(t, options[DBADDR], cfg.Database.Address)
			assert.Equal(t, options[DBUSER], cfg.Database.Username)
			assert.Equal(t, options[DBSECRET], cfg.Database.Secret)
			assert.Equal(t, options[DBDB], cfg.Database.Database)
		})

		t.Run("neo4j", func(t *testing.T) {
			assert.Equal(t, options[NEOCONN], cfg.Neo4J.Connection)
			assert.Equal(t, options[NEOADDR], cfg.Neo4J.Address)
			assert.Equal(t, options[NEOUSER], cfg.Neo4J.Username)
			assert.Equal(t, options[NEOSECRET], cfg.Neo4J.Secret)
			assert.Equal(t, options[NEODB], cfg.Neo4J.Database)
		})

		t.Run("crypto", func(t *testing.T) {
			assert.Equal(t, options[JWTSIGNKEY], cfg.Crypto.JWT.SigningKey)
		})

		t.Run("default admin", func(t *testing.T) {
			assert.Equal(t, options[DEFADMINPRINCNAME], cfg.DefaultAdmin.PrincipalName)
			assert.Equal(t, options[DEFADMINPASS], cfg.DefaultAdmin.Password)
			assert.Equal(t, options[DEFADMINEMAIL], cfg.DefaultAdmin.EmailAddress)
			assert.Equal(t, options[DEFADMINFIRST], cfg.DefaultAdmin.FirstName)
			assert.Equal(t, options[DEFADMINLAST], cfg.DefaultAdmin.LastName)
		})
	})
}
