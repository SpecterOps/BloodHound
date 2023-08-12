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

package config

import (
	"fmt"

	"github.com/specterops/bloodhound/src/serde"
)

// NewDefaultConfiguration returns a new Configuration struct containing all documented
// configuration defaults.
func NewDefaultConfiguration() (Configuration, error) {
	// Generate a new 256-bit key using random bytes converted to Base64 encoding
	if jwtSigningKey, err := GenerateRandomBase64String(32); err != nil {
		return Configuration{}, fmt.Errorf("failed to generate JWT signing key: %w", err)
	} else if generatedPassword, err := GenerateSecureRandomString(32); err != nil {
		return Configuration{}, fmt.Errorf("failed to generate default password: %w", err)
	} else {
		return Configuration{
			Version:     0,
			BindAddress: "127.0.0.1",
			MetricsPort: ":2112",
			RootURL:     serde.MustParseURL("http://localhost"),
			WorkDir:     "/opt/bhe/work",
			LogLevel:    "INFO",
			LogPath:     DefaultLogFilePath,

			// This idle timeout was set to a default of 70 seconds to avoid potential race conditions with the default 60 second
			// idle timeout implemented on the load balancer side.
			NetTimeoutSeconds: 70,
			// This is the threshold (in ms) for marking a query as slow for caching purposes
			SlowQueryThreshold:     100,
			MaxAPICacheSize:        200, // Number of cache items for API utilities
			MaxGraphQueryCacheSize: 100, // Number of cache items for graph queries
			TLS: TLSConfiguration{
				CertFile: "",
				KeyFile:  "",
			},
			Database: DatabaseConfiguration{
				Connection:            "",
				MaxConcurrentSessions: 10,
			},
			Neo4J: DatabaseConfiguration{
				Connection:            "",
				MaxConcurrentSessions: 10,
			},
			Crypto: CryptoConfiguration{
				JWT: JWTConfiguration{
					SigningKey: jwtSigningKey,
				},
				Argon2: Argon2Configuration{
					MemoryKibibytes: 1024 * 1024 * 1, // Minimum recommended memory
					NumIterations:   1,
					NumThreads:      8, // Default recommendation for Backend Server is 8 threads
				},
			},
			DefaultAdmin: DefaultAdminConfiguration{
				PrincipalName: "admin",
				Password:      generatedPassword,
				EmailAddress:  "spam@example.com",
				FirstName:     "Admin",
				LastName:      "User",
				ExpireNow:     true,
			},
			CollectorsBasePath: "/etc/bloodhound/collectors/",
			DatapipeInterval:   60,
			EnableAPILogging:   true,
			DisableAnalysis:    false,
			DisableCypherQC:    false,
			DisableMigrations:  false,
		}, nil
	}
}
