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
	"time"

	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func (s *Context) GetLatestAuditLogs() model.AuditLogs {
	auditLogsResponse, err := s.AdminClient().GetLatestAuditLogs()

	require.Nilf(s.TestCtrl, err, "Failed fetching latest audit logs: %v", err)
	return auditLogsResponse.Logs
}
func (s *Context) ListAuditLogs(after, before time.Time, offset, limit int) model.AuditLogs {
	auditLogsResponse, err := s.AdminClient().ListAuditLogs(after, before, offset, limit)

	require.Nilf(s.TestCtrl, err, "Failed fetching audit logs: %v", err)
	return auditLogsResponse.Logs
}

func (s *Context) AssetAuditLog(auditLog model.AuditLog, expectedAction string, expectedFields map[string]any) {
	assert.Equal(s.TestCtrl, auditLog.Action, expectedAction)

	for expectedFieldName, expectedFieldValue := range expectedFields {
		actualFieldValue, hasField := auditLog.Fields[expectedFieldName]

		assert.True(s.TestCtrl, hasField)
		assert.Equal(s.TestCtrl, expectedFieldValue, actualFieldValue)
	}
}

func (s *Context) AssertAuditLogHasAction(action string, expectedFields map[string]any) {
	found := false

	for _, auditLog := range s.GetLatestAuditLogs() {
		if auditLog.Action == action {
			s.AssetAuditLog(auditLog, action, expectedFields)

			found = true
			break
		}
	}

	require.Truef(s.TestCtrl, found, "Expected to find audit log entry by name %s but was unable to", action)
}
