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

package graphify_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/services/graphify"
	"github.com/stretchr/testify/assert"
)

func TestNormalizeEinNodeProperties(t *testing.T) {
	t.Parallel()
	var (
		nowUTC     = time.Now().UTC()
		objectID   = "objectid"
		properties = map[string]any{
			graphify.ReconcileProperty:      false,
			common.Name.String():            "name",
			common.OperatingSystem.String(): "temple",
			ad.DistinguishedName.String():   "distinguished-name",
		}
		normalizedProperties = graphify.NormalizeEinNodeProperties(properties, objectID, nowUTC)
	)

	assert.Nil(t, normalizedProperties[graphify.ReconcileProperty])
	assert.NotNil(t, normalizedProperties[common.LastSeen.String()])
	assert.Equal(t, "OBJECTID", normalizedProperties[common.ObjectID.String()])
	assert.Equal(t, "NAME", normalizedProperties[common.Name.String()])
	assert.Equal(t, "DISTINGUISHED-NAME", normalizedProperties[ad.DistinguishedName.String()])
	assert.Equal(t, "TEMPLE", normalizedProperties[common.OperatingSystem.String()])
}
