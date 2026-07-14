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

package graphify

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConvertGenericNode(t *testing.T) {
	t.Run("flag off: objectid is uppercased", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "objectid",
				Kinds: []string{"SomeKind"},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, false)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "OBJECTID", converted.NodeProps[0].ObjectID)
	})

	t.Run("flag on: objectid preserves original case", func(t *testing.T) {
		var (
			entity = ein.GenericNode{
				ID:    "ObjectId",
				Kinds: []string{"SomeKind"},
			}
			converted = &ConvertedData{}
		)

		err := ConvertGenericNode(entity, converted, true)
		require.NoError(t, err)
		require.Len(t, converted.NodeProps, 1)
		assert.Equal(t, "ObjectId", converted.NodeProps[0].ObjectID)
	})
}
