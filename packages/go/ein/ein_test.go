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
package ein

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func validKinds() graph.Kinds {
	return slicesext.Concat(ad.NodeKinds(), ad.Relationships(), azure.NodeKinds(), azure.Relationships())
}

func TestParseKind(t *testing.T) {
	t.Run("all known strings map to their graph.Kind", func(t *testing.T) {
		for _, k := range validKinds() {
			res, err := ParseKind(k.String())
			require.Nil(t, err)
			assert.Equal(t, k, res, "expect string to map back to original kind")
		}
	})

	t.Run("unknown kind strings cause an error", func(t *testing.T) {
		_, err := ParseKind("unsupported")
		assert.Contains(t, err.Error(), "unsupported", "error contains unsupported kind string")
	})
}
