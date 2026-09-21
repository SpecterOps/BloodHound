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
	"github.com/stretchr/testify/require"
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
		options, err := tlsOptions(environmentMap)
		require.NoError(t, err)
		assert.Equal(t, "sslmode=disable", options)
	})

	t.Run("remote host with authenticated sslmode preserves TLS settings in URL query format", func(t *testing.T) {
		environmentMap := map[string]string{
			"host":        "db.example.com",
			"sslmode":     "verify-full",
			"sslrootcert": "/etc/ssl/root.crt",
		}
		options, err := tlsOptions(environmentMap)
		require.NoError(t, err)
		assert.Equal(t, "sslmode=verify-full&sslrootcert=%2Fetc%2Fssl%2Froot.crt", options)
	})

	t.Run("remote host rejects insecure or unverified sslmode values", func(t *testing.T) {
		for _, sslmode := range []string{"", "disable", "allow", "prefer", "require"} {
			environmentMap := map[string]string{"host": "db.example.com", "sslmode": sslmode}
			_, err := tlsOptions(environmentMap)
			assert.Errorf(t, err, "sslmode=%q should be rejected for a remote host", sslmode)
		}
	})
}
