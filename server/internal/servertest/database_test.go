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

//go:build integration

package servertest

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsLocalHost(t *testing.T) {
	cases := map[string]bool{
		"localhost":      true,
		"127.0.0.1":      true,
		"::1":            true,
		"db.example.com": false,
		"10.0.0.5":       false,
		"":               false,
	}

	for host, want := range cases {
		assert.Equalf(t, want, isLocalHost(host), "isLocalHost(%q)", host)
	}
}

func TestTLSOptions(t *testing.T) {
	t.Run("local host forces sslmode=disable regardless of configured TLS", func(t *testing.T) {
		environmentMap := map[string]string{"host": "localhost", "sslmode": "require"}
		assert.Equal(t, "sslmode=disable", tlsOptions(environmentMap))
	})

	t.Run("remote host preserves configured TLS settings in URL query format", func(t *testing.T) {
		environmentMap := map[string]string{
			"host":        "db.example.com",
			"sslmode":     "require",
			"sslrootcert": "/etc/ssl/root.crt",
		}
		assert.Equal(t, "sslmode=require&sslrootcert=%2Fetc%2Fssl%2Froot.crt", tlsOptions(environmentMap))
	})

	t.Run("remote host without TLS settings yields empty options", func(t *testing.T) {
		environmentMap := map[string]string{"host": "db.example.com"}
		assert.Empty(t, tlsOptions(environmentMap))
	})
}
