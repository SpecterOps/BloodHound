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

//go:build serial_integration
// +build serial_integration

package v2_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/stretchr/testify/require"
)

func Test_ListAuditLogs(t *testing.T) {
	testCtx := integration.NewFOSSContext(t)

	t.Run("Test Getting Latest Audit Logs", func(t *testing.T) {
		var (
			newAssetGroup          = testCtx.CreateAssetGroup("Testing Audit Logs", "test_get_auditLogs")
			expectedAuditLogFields = map[string]any{
				"asset_group_name": newAssetGroup.Name,
				"asset_group_tag":  newAssetGroup.Tag,
			}
		)

		testCtx.DeleteAssetGroup(newAssetGroup.ID)

		testCtx.AssertAuditLogHasAction("CreateAssetGroup", expectedAuditLogFields)
		testCtx.AssertAuditLogHasAction("DeleteAssetGroup", expectedAuditLogFields)
	})

	t.Run("Test Getting Audit Logs by Time Range", func(t *testing.T) {
		newAssetGroup := testCtx.CreateAssetGroup("Testing Audit Logs", "test_get_auditLogs")

		// Sleep just a moment to give the API some time to avoid audit log jitter
		time.Sleep(time.Second / 2)

		deletionTimestamp := time.Now()

		testCtx.DeleteAssetGroup(newAssetGroup.ID)

		// Expect one audit log entry from the deletion
		auditLogs := testCtx.ListAuditLogs(deletionTimestamp, time.Now(), 0, 1000)
		require.Equal(t, 2, len(auditLogs), "Expected exactly 2 audit log entries but saw %d", len(auditLogs))

		// Make sure these two actions are from the same request
		require.Equal(t, auditLogs[0].RequestID, auditLogs[1].RequestID)

		// Makes sure these two actions are from the same two phase commit
		require.Equal(t, auditLogs[0].CommitID, auditLogs[1].CommitID)

		// Audit logs are in LIFO order
		require.Equal(t, auditLogs[0].Status, "success")
		require.Equal(t, auditLogs[1].Status, "intent")

		testCtx.AssetAuditLog(auditLogs[0], "DeleteAssetGroup", map[string]any{
			"asset_group_name": newAssetGroup.Name,
			"asset_group_tag":  newAssetGroup.Tag,
		})

		testCtx.AssetAuditLog(auditLogs[1], "DeleteAssetGroup", map[string]any{
			"asset_group_name": newAssetGroup.Name,
			"asset_group_tag":  newAssetGroup.Tag,
		})
	})
}
