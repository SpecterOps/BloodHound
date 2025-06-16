// Copyright 2024 Specter Ops, Inc.
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

package pgsql

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIdentifierSet_CombinedKey(t *testing.T) {
	t.Parallel()
	identifiers := AsIdentifierSet("1", "4", "3", "2")
	assert.Equal(t, Identifier("1234"), identifiers.CombinedKey())
}

func TestIdentifierSet_CheckedAdd(t *testing.T) {
	t.Parallel()
	identifiers := AsIdentifierSet("1", "4", "3", "2")
	assert.True(t, identifiers.CheckedAdd("5"))
	assert.False(t, identifiers.CheckedAdd("5"))
}

func TestIdentifierSet_Matches(t *testing.T) {
	t.Parallel()
	identifiers := AsIdentifierSet("1", "4", "3", "2")
	assert.True(t, identifiers.Matches(identifiers.Copy()))
	assert.False(t, identifiers.Matches(AsIdentifierSet("55")))
}

func TestIdentifierSet_Remove(t *testing.T) {
	t.Parallel()
	identifiers := AsIdentifierSet("1", "4", "3", "2")
	identifiers.Remove("4", "5")

	assert.True(t, identifiers.Matches(AsIdentifierSet("1", "3", "2")))

	identifiers.RemoveSet(AsIdentifierSet("3", "4"))
	assert.True(t, identifiers.Matches(AsIdentifierSet("1", "2")))
}

func TestIdentifierSet_MergeSet(t *testing.T) {
	t.Parallel()
	identifiers := AsIdentifierSet("1", "4")
	identifiers.MergeSet(AsIdentifierSet("3", "2"))

	assert.True(t, identifiers.Matches(AsIdentifierSet("1", "4", "3", "2")))
}

func TestIdentifierSet_Slice(t *testing.T) {
	t.Parallel()
	identifiers := AsIdentifierSet("1", "4", "3", "2").Slice()

	assert.Equal(t, 4, len(identifiers))

	// Should be sorted for stable outputs
	assert.Equal(t, Identifier("1"), identifiers[0])
	assert.Equal(t, Identifier("2"), identifiers[1])
	assert.Equal(t, Identifier("3"), identifiers[2])
	assert.Equal(t, Identifier("4"), identifiers[3])

	// Test string slices as well
	stringIdentifiers := AsIdentifierSet("1", "4", "3", "2").Strings()

	assert.Equal(t, 4, len(stringIdentifiers))

	// Should be sorted for stable outputs
	assert.Equal(t, "1", stringIdentifiers[0])
	assert.Equal(t, "2", stringIdentifiers[1])
	assert.Equal(t, "3", stringIdentifiers[2])
	assert.Equal(t, "4", stringIdentifiers[3])
}

func TestIdentifierSet_Copy(t *testing.T) {
	t.Parallel()
	identifiers := AsIdentifierSet("1", "4", "3", "2")
	assert.True(t, identifiers.Matches(identifiers.Copy()))
}
