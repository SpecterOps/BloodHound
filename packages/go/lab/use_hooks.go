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

package lab

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func UseHooks[Ctx TBConstraint, Fixture any](ctx Ctx, setup func() (Fixture, error), teardown func(Fixture) error) (Assertions, Fixture) {
	assert := require.New(testing.TB(ctx))
	assert.NotNil(setup)
	t, err := setup()
	assert.NoError(err)
	if teardown != nil {
		testing.TB(ctx).Cleanup(func() {
			assert.NoError(teardown(t))
		})
	}
	return assert, t
}
