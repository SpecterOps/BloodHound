package database_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

func findReleventAuditLogs(auditLogs model.AuditLogs, action string, user model.User) (model.AuditLog, model.AuditLog) {
	var (
		intentAuditLog, resultAuditLog model.AuditLog
	)

	for _, al := range auditLogs {
		if al.Action == action && al.Fields["principal_name"] == user.PrincipalName {
			if al.Status == string(model.AuditStatusIntent) {
				intentAuditLog = al
			} else {
				resultAuditLog = al
			}
		}
	}

	return intentAuditLog, resultAuditLog
}

func verifyAuditLogs(t *testing.T, dbInst database.Database, action string, user model.User) {
	auditLogs, count, err := dbInst.ListAuditLogs(time.Now(), time.Now().Add(-24*time.Hour), 0, 10, "", model.SQLFilter{})
	if err != nil {
		t.Fatalf("Error getting verifying audit logs: %v", err)
	}
	if count < 2 {
		t.Fatalf("incorrect number of audit logs found. Got: %d Wanted: 2 or more", count)
	}

	intentAuditLog, resultAuditLog := findReleventAuditLogs(auditLogs, action, user)

	if intentAuditLog.Action != action || intentAuditLog.Status != string(model.AuditStatusIntent) || intentAuditLog.Fields["principal_name"] != user.PrincipalName {
		t.Errorf("intent audit log is invalid: %#v", intentAuditLog)
	} else if resultAuditLog.Action != action || resultAuditLog.Status == string(model.AuditStatusIntent) || resultAuditLog.Fields["principal_name"] != user.PrincipalName {
		t.Errorf("result audit log is invalid: %#v", resultAuditLog)
	} else if intentAuditLog.CommitID != resultAuditLog.CommitID {
		t.Errorf("commit IDs on audit logs do not match. Intent log: %s  Result log: %s", intentAuditLog.CommitID, resultAuditLog.CommitID)
	}
}
