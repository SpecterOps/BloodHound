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
package productfeatures_test

import (
	"context"
	"testing"

	"github.com/open-feature/go-sdk/openfeature/isolated"
	"github.com/specterops/bloodhound/cmd/api/src/services/productfeatures"
	featuretest "github.com/specterops/bloodhound/cmd/api/src/services/productfeatures/testing"
	"github.com/stretchr/testify/assert"
)

func TestAllValues_Defaults(t *testing.T) {
	var evaluationAPI = isolated.NewAPI()

	t.Parallel()
	t.Cleanup(func() {
		assert.NoError(t, evaluationAPI.Shutdown(context.Background()))
	})
	assert.Equal(t, map[string]any{
		productfeatures.ETAC_ENABLED:           false,
		productfeatures.PZ_MULTI_TIER_ANALYSIS: false,
		productfeatures.PZ_TIER_LIMIT:          int64(1),
		productfeatures.PZ_LABEL_LIMIT:         int64(0),
	}, productfeatures.AllValues(t.Context(), evaluationAPI.NewClient()))
}

func TestAllValues_ProviderOverrides(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name      string
		overrides featuretest.Overrides
		expected  map[string]any
	}{
		{
			name: "enabled entitlements and increased limits",
			overrides: featuretest.Overrides{
				Bools: map[string]bool{
					productfeatures.ETAC_ENABLED:           true,
					productfeatures.PZ_MULTI_TIER_ANALYSIS: true,
				},
				Ints: map[string]int64{
					productfeatures.PZ_TIER_LIMIT:  5,
					productfeatures.PZ_LABEL_LIMIT: 3,
				},
			},
			expected: map[string]any{
				productfeatures.ETAC_ENABLED:           true,
				productfeatures.PZ_MULTI_TIER_ANALYSIS: true,
				productfeatures.PZ_TIER_LIMIT:          int64(5),
				productfeatures.PZ_LABEL_LIMIT:         int64(3),
			},
		},
		{
			name: "zero overrides remain zero and missing flags use defaults",
			overrides: featuretest.Overrides{
				Bools: map[string]bool{productfeatures.ETAC_ENABLED: false},
				Ints:  map[string]int64{productfeatures.PZ_TIER_LIMIT: 0},
			},
			expected: map[string]any{
				productfeatures.ETAC_ENABLED:           false,
				productfeatures.PZ_MULTI_TIER_ANALYSIS: false,
				productfeatures.PZ_TIER_LIMIT:          int64(0),
				productfeatures.PZ_LABEL_LIMIT:         int64(0),
			},
		},
		{
			name: "type mismatches use defaults",
			overrides: featuretest.Overrides{
				Strings: map[string]string{
					productfeatures.ETAC_ENABLED:  "true",
					productfeatures.PZ_TIER_LIMIT: "5",
				},
			},
			expected: map[string]any{
				productfeatures.ETAC_ENABLED:           false,
				productfeatures.PZ_MULTI_TIER_ANALYSIS: false,
				productfeatures.PZ_TIER_LIMIT:          int64(1),
				productfeatures.PZ_LABEL_LIMIT:         int64(0),
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var client = featuretest.NewClient(t, testCase.overrides)
			assert.Equal(t, testCase.expected, productfeatures.AllValues(t.Context(), client))
		})
	}
}
