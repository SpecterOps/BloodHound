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
package v2_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/services/productfeatures"
	featuretest "github.com/specterops/bloodhound/cmd/api/src/services/productfeatures/testing"
	"github.com/stretchr/testify/assert"
)

func TestGetProductFeatures_LegacyResponse(t *testing.T) {
	var (
		resources = v2.Resources{FeatureFlags: featuretest.NewClient(t, featuretest.Overrides{
			Bools: map[string]bool{productfeatures.ETAC_ENABLED: true},
			Ints:  map[string]int64{productfeatures.PZ_TIER_LIMIT: 5},
		})}
		request  = httptest.NewRequest(http.MethodGet, "/api/v2/dog-tags", nil)
		response = httptest.NewRecorder()
	)

	resources.GetProductFeatures(response, request)

	assert.Equal(t, http.StatusOK, response.Code)
	assert.JSONEq(t, `{
        "data": {
            "auth.environment_targeted_access_control": true,
            "privilege_zones.multi_tier_analysis": false,
            "privilege_zones.tier_limit": 5,
            "privilege_zones.label_limit": 0
        }
    }`, response.Body.String())
}
