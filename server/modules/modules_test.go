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

package modules

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/wireup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// recorder captures every invocation of a wireup.Module produced by its
// module helper so tests can assert that the registry walked the slice in
// order and forwarded the supplied Deps.
type recorder struct {
	name   string
	order  *[]string
	called []wireup.Deps
}

func (s *recorder) module() wireup.Module {
	return func(deps wireup.Deps) {
		if s.order != nil {
			*s.order = append(*s.order, s.name)
		}
		s.called = append(s.called, deps)
	}
}

func TestRegister_InvokesEachModuleOnceWithTheSuppliedDeps(t *testing.T) {
	var (
		first  = &recorder{name: "first"}
		second = &recorder{name: "second"}
		// Non-zero pointer fields so the assertion exercises identity rather
		// than zero-value equality.
		deps = wireup.Deps{Router: &router.Router{}, Pool: &pgxpool.Pool{}}
	)

	register(deps, []wireup.Module{first.module(), second.module()})

	require.Len(t, first.called, 1)
	require.Len(t, second.called, 1)
	assert.Equal(t, deps, first.called[0])
	assert.Equal(t, deps, second.called[0])
}

func TestRegister_PreservesSliceOrder(t *testing.T) {
	var (
		order   []string
		first   = &recorder{name: "first", order: &order}
		second  = &recorder{name: "second", order: &order}
		third   = &recorder{name: "third", order: &order}
		modules = []wireup.Module{first.module(), second.module(), third.module()}
	)

	register(wireup.Deps{}, modules)

	assert.Equal(t, []string{"first", "second", "third"}, order)
}

func TestRegister_EmptySliceIsANoop(t *testing.T) {
	require.NotPanics(t, func() {
		register(wireup.Deps{}, nil)
		register(wireup.Deps{}, []wireup.Module{})
	})
}

func TestDefaultRegistry_IsNotEmpty(t *testing.T) {
	// Catches the regression where a refactor accidentally empties the slice
	// and silently drops every feature from the server.
	assert.NotEmpty(t, all)
}
