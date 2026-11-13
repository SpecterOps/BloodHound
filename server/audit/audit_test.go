// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package audit_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/server/audit"
	"github.com/stretchr/testify/assert"
)

func TestRegister_Dependencies(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		pool             *pgxpool.Pool
		identityResolver auth.IdentityResolver
		expectedPanic    string
	}{
		{
			name:             "Success: constructs audit service",
			pool:             new(pgxpool.Pool),
			identityResolver: auth.NewIdentityResolver(),
		},
		{
			name:             "Error: nil pool",
			identityResolver: auth.NewIdentityResolver(),
			expectedPanic:    "audit: Register requires a non-nil Pool",
		},
		{
			name:          "Error: nil identity resolver",
			pool:          new(pgxpool.Pool),
			expectedPanic: "audit: Register requires a non-nil IdentityResolver",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			construct := func() {
				service := audit.Register(testCase.pool, testCase.identityResolver)
				assert.NotNil(t, service)
			}

			if testCase.expectedPanic != "" {
				assert.PanicsWithValue(t, testCase.expectedPanic, construct)
			} else {
				assert.NotPanics(t, construct)
			}
		})
	}
}
