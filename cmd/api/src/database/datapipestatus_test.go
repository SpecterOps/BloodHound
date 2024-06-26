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
	"testing"

	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
)

func TestDatapipeStatus(t *testing.T) {
	var (
		testCtx = context.Background()
		db      = integration.SetupDB(t)
	)

	status, err := db.GetDatapipeStatus(testCtx)
	require.Nil(t, err)
	require.Equal(t, model.DatapipeStatusIdle, status.Status)

	err = db.SetDatapipeStatus(testCtx, model.DatapipeStatusAnalyzing, false)
	require.Nil(t, err)

	status, err = db.GetDatapipeStatus(testCtx)
	require.Nil(t, err)
	require.Equal(t, model.DatapipeStatusAnalyzing, status.Status)

	// when `SetDatapipeStatus` is called with `true` for the `updateAnalysisTime` parameter, assert that the time is no longer null
	require.True(t, status.LastCompleteAnalysisAt.IsZero())
	err = db.SetDatapipeStatus(testCtx, model.DatapipeStatusAnalyzing, true)
	require.Nil(t, err)
	status, err = db.GetDatapipeStatus(testCtx)
	require.Nil(t, err)
	require.True(t, !status.LastCompleteAnalysisAt.IsZero())

}
