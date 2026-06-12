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

package dataquality

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestExcludeSourceKindsFromOpenGraphNodeKinds(t *testing.T) {
	nodeKinds := model.GraphSchemaNodeKinds{
		{Serial: model.Serial{ID: 1}, Name: "Base"},
		{Serial: model.Serial{ID: 2}, Name: "GitHubUser"},
		{Serial: model.Serial{ID: 3}, Name: "AZBase"},
		{Serial: model.Serial{ID: 4}, Name: "GitHubRepository"},
	}
	sourceKinds := []model.Kind{
		{Name: "Base"},
		{Name: "AZBase"},
	}

	result := excludeSourceKindsFromOpenGraphNodeKinds(nodeKinds, sourceKinds)

	require.Equal(t, model.GraphSchemaNodeKinds{
		{Serial: model.Serial{ID: 2}, Name: "GitHubUser"},
		{Serial: model.Serial{ID: 4}, Name: "GitHubRepository"},
	}, result)
}
