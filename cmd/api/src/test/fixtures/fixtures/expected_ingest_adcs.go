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

package fixtures

import (
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/test"
	"github.com/stretchr/testify/require"
)

var (
	ingestADCSRelationshipAssertionCriteria = []graph.Criteria{
		query.And(
			query.Kind(query.Start(), ad.NTAuthStore),
			query.Equals(query.StartProperty(common.ObjectID.String()), "722A8BB3-AEF5-49C7-9C8C-C1C97A219007"),
			query.Kind(query.Relationship(), ad.NTAuthStoreFor),
			query.Kind(query.End(), ad.Domain),
			query.Equals(query.EndProperty(common.ObjectID.String()), "S-1-5-21-909015691-3030120388-2582151266")),
	}
)

func IngestADCSAssertions(testCtrl test.Controller, tx graph.Transaction) {
	for _, assertionCriteria := range ingestADCSRelationshipAssertionCriteria {
		_, err := tx.Relationships().Filter(assertionCriteria).First()
		require.Nilf(testCtrl, err, "Unable to find an expected relationship: %s", FormatQueryComponent(assertionCriteria))
	}
}
