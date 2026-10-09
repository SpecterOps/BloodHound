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

package tiering_test

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/analysis/tiering"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/require"
)

func TestIsDecoy(t *testing.T) {
	t.Run("kind", func(t *testing.T) {
		node := graph.NewNode(1, graph.NewProperties(), ad.User, tiering.KindTagDecoy)

		require.True(t, tiering.IsDecoy(node))
	})

	t.Run("system tag", func(t *testing.T) {
		node := graph.NewNode(1, graph.NewProperties().Set(common.SystemTags.String(), ad.Decoy), ad.User)

		require.True(t, tiering.IsDecoy(node))
	})

	t.Run("unmarked", func(t *testing.T) {
		node := graph.NewNode(1, graph.NewProperties(), ad.User)

		require.False(t, tiering.IsDecoy(node))
	})
}
