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

package appdb_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStore_GetEnvironmentsFiltered_Integration(t *testing.T) {
	var (
		ctx         = context.Background()
		store, pool = setupStore(t)
		builtinKind = seedEnvironment(t, pool, "BuiltinExt", "IntgBuiltinEnv", true)
		customKind  = seedEnvironment(t, pool, "CustomExt", "IntgCustomEnv", false)
	)

	t.Run("returns_all_environments", func(t *testing.T) {
		environments, err := store.GetEnvironmentsFiltered(ctx, false)
		require.NoError(t, err)

		var names []string
		for _, environment := range environments {
			names = append(names, environment.EnvironmentKindName)
		}
		assert.Contains(t, names, builtinKind)
		assert.Contains(t, names, customKind)
	})

	t.Run("returns_only_builtin_environments", func(t *testing.T) {
		environments, err := store.GetEnvironmentsFiltered(ctx, true)
		require.NoError(t, err)

		var names []string
		for _, environment := range environments {
			names = append(names, environment.EnvironmentKindName)
		}
		assert.Contains(t, names, builtinKind)
		assert.NotContains(t, names, customKind)
	})
}
