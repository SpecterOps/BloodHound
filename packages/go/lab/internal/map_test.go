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

package internal_test

import (
	"testing"

	"github.com/specterops/bloodhound/lab/internal"
	"github.com/stretchr/testify/require"
)

func TestMap(t *testing.T) {
	assert := require.New(t)
	foo := make(internal.Map[any, any])
	assert.NotNil(foo)
	foo.Set("foo", "bar")
	assert.True(foo.Has("foo"))
	assert.Equal("bar", foo["foo"])
	clone := foo.Clone()
	assert.NotSame(clone, foo)
	clone.Delete("foo")
	assert.False(clone.Has("foo"))
	assert.True(foo.Has("foo"))
}
