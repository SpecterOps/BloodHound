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

package integration

import (
	"github.com/specterops/bloodhound/src/test"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
	"github.com/specterops/bloodhound/dawgs/graph"
)

type ReconciliationAssertion func(testCtrl test.Controller, tx graph.Transaction)

func (s *Context) AssertReconciliation(assertion ReconciliationAssertion) {
	graphDB := integration.OpenNeo4jGraphDB(s.TestCtrl)
	defer graphDB.Close()

	require.Nil(s.TestCtrl, graphDB.ReadTransaction(s.Ctx, func(tx graph.Transaction) error {
		assertion(s.TestCtrl, tx)
		return nil
	}), "Unexpected database error during reconciliation assertion")
}
