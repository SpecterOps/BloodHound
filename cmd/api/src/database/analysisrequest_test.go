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

package database_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestAnalysisRequest(t *testing.T) {
	var (
		testCtx = context.Background()
		dbInst  = integration.SetupDB(t)
	)

	err := dbInst.RequestAnalysis(testCtx, "test")
	require.Nil(t, err)

	analReq, err := dbInst.GetAnalysisRequest(testCtx)
	require.Nil(t, err)
	require.Equal(t, analReq.RequestType, model.AnalysisRequestAnalysis)
	require.Equal(t, analReq.RequestedBy, "test")
	require.False(t, analReq.RequestedAt.IsZero())

	err = dbInst.DeleteAnalysisRequest(testCtx)
	require.Nil(t, err)

	_, err = dbInst.GetAnalysisRequest(testCtx)
	require.ErrorIs(t, err, sql.ErrNoRows)
}
