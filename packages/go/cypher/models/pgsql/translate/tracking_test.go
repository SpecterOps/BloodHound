// Copyright 2025 Specter Ops, Inc.
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

package translate

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScope(t *testing.T) {
	var (
		scope = NewScope()
	)

	grandparent, err := scope.PushFrame()
	require.Nil(t, err)

	parent, err := scope.PushFrame()
	require.Nil(t, err)

	child, err := scope.PushFrame()
	require.Nil(t, err)

	require.Equal(t, 0, grandparent.id)
	require.Equal(t, 1, parent.id)
	require.Equal(t, 2, child.id)

	require.Nil(t, scope.UnwindToFrame(parent))
	require.Equal(t, parent.id, scope.CurrentFrame().id)
}
