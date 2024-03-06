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
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"time"
)

func findRelevantAuditLogs(auditLogs model.AuditLogs, action model.AuditLogAction, fieldKey string, fieldData string) (model.AuditLog, model.AuditLog) {
	var (
		intentAuditLog, resultAuditLog model.AuditLog
	)

	for _, al := range auditLogs {
		if al.Action == action && al.Fields[fieldKey] == fieldData {
			if al.Status == string(model.AuditLogStatusIntent) {
				intentAuditLog = al
			} else {
				resultAuditLog = al
			}
		}
	}

	return intentAuditLog, resultAuditLog
}

func VerifyAuditLogs(dbInst database.Database, expectedAction model.AuditLogAction, expectedFieldKey string, expectedFieldData string) error {
	auditLogs, count, err := dbInst.ListAuditLogs(context.Background(), time.Now(), time.Now().Add(-24*time.Hour), 0, 10, "", model.SQLFilter{})
	if err != nil {
		return fmt.Errorf("error getting verifying audit logs: %v", err)
	}
	if count < 2 {
		return fmt.Errorf("incorrect number of audit logs found. Got: %d Wanted: 2 or more", count)
	}

	intentAuditLog, resultAuditLog := findRelevantAuditLogs(auditLogs, expectedAction, expectedFieldKey, expectedFieldData)

	if intentAuditLog.ID == 0 || resultAuditLog.ID == 0 {
		return fmt.Errorf("unable to find audit logs matching the provided data. expectedAction: %s expectedFieldKey: %s expectedFieldData: %v", expectedAction, expectedFieldKey, expectedFieldData)
	} else if intentAuditLog.Action != expectedAction || intentAuditLog.Status != string(model.AuditLogStatusIntent) || intentAuditLog.Fields[expectedFieldKey] != expectedFieldData {
		return fmt.Errorf("intent audit log is invalid: %#v", intentAuditLog)
	} else if resultAuditLog.Action != expectedAction || resultAuditLog.Status == string(model.AuditLogStatusIntent) || resultAuditLog.Fields[expectedFieldKey] != expectedFieldData {
		return fmt.Errorf("result audit log is invalid: %#v", resultAuditLog)
	} else if intentAuditLog.CommitID != resultAuditLog.CommitID {
		return fmt.Errorf("commit IDs on audit logs do not match. Intent log: %s  Result log: %s", intentAuditLog.CommitID, resultAuditLog.CommitID)
	}

	return nil
}
