package database_test

import (
	"fmt"
	"time"

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

func findReleventAuditLogs(auditLogs model.AuditLogs, action string, fieldKey string, fieldData string) (model.AuditLog, model.AuditLog) {
	var (
		intentAuditLog, resultAuditLog model.AuditLog
	)

	for _, al := range auditLogs {
		if al.Action == action && al.Fields[fieldKey] == fieldData {
			if al.Status == string(model.AuditStatusIntent) {
				intentAuditLog = al
			} else {
				resultAuditLog = al
			}
		}
	}

	return intentAuditLog, resultAuditLog
}

func verifyAuditLogs(dbInst database.Database, action string, fieldKey string, fieldData string) error {
	auditLogs, count, err := dbInst.ListAuditLogs(time.Now(), time.Now().Add(-24*time.Hour), 0, 10, "", model.SQLFilter{})
	if err != nil {
		return fmt.Errorf("Error getting verifying audit logs: %v", err)
	}
	if count < 2 {
		return fmt.Errorf("incorrect number of audit logs found. Got: %d Wanted: 2 or more", count)
	}

	intentAuditLog, resultAuditLog := findReleventAuditLogs(auditLogs, action, fieldKey, fieldData)

	if intentAuditLog.ID == 0 || resultAuditLog.ID == 0 {
		return fmt.Errorf("unable to find audit logs matching the provided data. action: %s fieldKey: %s fieldData: %v", action, fieldKey, fieldData)
	} else if intentAuditLog.Action != action || intentAuditLog.Status != string(model.AuditStatusIntent) || intentAuditLog.Fields[fieldKey] != fieldData {
		return fmt.Errorf("intent audit log is invalid: %#v", intentAuditLog)
	} else if resultAuditLog.Action != action || resultAuditLog.Status == string(model.AuditStatusIntent) || resultAuditLog.Fields[fieldKey] != fieldData {
		return fmt.Errorf("result audit log is invalid: %#v", resultAuditLog)
	} else if intentAuditLog.CommitID != resultAuditLog.CommitID {
		return fmt.Errorf("commit IDs on audit logs do not match. Intent log: %s  Result log: %s", intentAuditLog.CommitID, resultAuditLog.CommitID)
	}

	return nil
}
