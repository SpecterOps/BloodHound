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

package util_test

import (
	"testing"

	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/stretchr/testify/require"
)

func TestIsNeoTimeoutError(t *testing.T) {
	err := neo4j.Neo4jError{
		Code: "Neo.ClientError.Transaction.TransactionTimedOut",
		Msg:  "The transaction has been terminated. Retry your operation in a new transaction, and you should see a successful result. The transaction has not completed within the specified timeout (dbms.transaction.timeout). You may want to retry with a longer timeout.",
	}

	require.True(t, util.IsNeoTimeoutError(&err))

	err = neo4j.Neo4jError{
		Code: "This.Is.A.Test",
		Msg:  "Blah",
	}

	require.False(t, util.IsNeoTimeoutError(&err))
}
