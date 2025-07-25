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

package test

import (
	"context"
	"fmt"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func findRelevantAuditLogs(auditLogs model.AuditLogs, expectedAction model.AuditLogAction, expectedResultStatus model.AuditLogEntryStatus, expectedData model.AuditData) (intentAuditLog, resultAuditLog model.AuditLog) {
	for _, al := range auditLogs {
		// Assume true unless we encounter a deviation
		matchingData := true
		if al.Action == expectedAction {
			for expectedDataKey, expectedDataValue := range expectedData {
				actualDataValue, hasData := al.Fields[expectedDataKey]
				if !hasData || expectedDataValue != actualDataValue {
					matchingData = false
					break
				}
			}

			if !matchingData {
				continue
			}
			switch al.Status {
			case model.AuditLogStatusIntent:
				intentAuditLog = al
			case expectedResultStatus:
				resultAuditLog = al
			}
		}
	}

	return intentAuditLog, resultAuditLog
}

// VerifyAuditLogs Assumes success status for audit log
func VerifyAuditLogs(dbInst database.Database, expectedAction model.AuditLogAction, expectedField, expectedFieldValue string) error {
	auditLogs, count, err := dbInst.ListAuditLogs(context.Background(), time.Now(), time.Now().Add(-24*time.Hour), 0, 10, "", model.SQLFilter{})
	if err != nil {
		return fmt.Errorf("error getting verifying audit logs: %v", err)
	}
	if count < 2 {
		return fmt.Errorf("incorrect number of audit logs found. Got: %d Wanted: 2 or more", count)
	}
	return AssertAuditLogs(auditLogs, expectedAction, model.AuditLogStatusSuccess, model.AuditData{expectedField: expectedFieldValue})
}

func AssertAuditLogs(auditLogs model.AuditLogs, expectedAction model.AuditLogAction, expectedResultStatus model.AuditLogEntryStatus, expectedData model.AuditData) error {
	intentAuditLog, resultAuditLog := findRelevantAuditLogs(auditLogs, expectedAction, expectedResultStatus, expectedData)

	if intentAuditLog.ID == 0 || resultAuditLog.ID == 0 {
		return fmt.Errorf("unable to find audit logs matching the provided data. expectedAction: %s expectedResultStatus %s expectedData: %v", expectedAction, expectedResultStatus, expectedData)
	} else if intentAuditLog.CommitID != resultAuditLog.CommitID {
		return fmt.Errorf("commit IDs on audit logs do not match. Intent log: %s  Result log: %s", intentAuditLog.CommitID, resultAuditLog.CommitID)
	}

	return nil
}
