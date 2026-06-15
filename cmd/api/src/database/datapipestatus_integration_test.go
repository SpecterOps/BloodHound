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

//go:build integration

package database_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

func TestDatabase_GetDatapipeStatus(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	status, err := testSuite.BHDatabase.GetDatapipeStatus(testSuite.Context)
	require.NoError(t, err)
	require.Equal(t, model.DatapipeStatusIdle, status.Status)
}

func TestDatabase_SetDatapipeStatus(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	err := testSuite.BHDatabase.SetDatapipeStatus(testSuite.Context, model.DatapipeStatusAnalyzing)
	require.NoError(t, err)

	status, err := testSuite.BHDatabase.GetDatapipeStatus(testSuite.Context)
	require.NoError(t, err)
	require.Equal(t, model.DatapipeStatusAnalyzing, status.Status)
}

func TestDatabase_SetLastAnalysisStartTime(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	status, err := testSuite.BHDatabase.GetDatapipeStatus(testSuite.Context)
	require.NoError(t, err)
	require.True(t, status.LastAnalysisRunAt.IsZero())

	err = testSuite.BHDatabase.SetLastAnalysisStartTime(testSuite.Context)
	require.NoError(t, err)

	status, err = testSuite.BHDatabase.GetDatapipeStatus(testSuite.Context)
	require.NoError(t, err)
	require.False(t, status.LastAnalysisRunAt.IsZero())
}

func TestDatabase_UpdateLastAnalysisCompleteTime(t *testing.T) {
	testSuite := setupIntegrationTestSuite(t)
	defer teardownIntegrationTestSuite(t, &testSuite)

	status, err := testSuite.BHDatabase.GetDatapipeStatus(testSuite.Context)
	require.NoError(t, err)
	require.True(t, status.LastCompleteAnalysisAt.IsZero())

	err = testSuite.BHDatabase.UpdateLastAnalysisCompleteTime(testSuite.Context)
	require.NoError(t, err)

	status, err = testSuite.BHDatabase.GetDatapipeStatus(testSuite.Context)
	require.NoError(t, err)
	require.False(t, status.LastCompleteAnalysisAt.IsZero())
}
