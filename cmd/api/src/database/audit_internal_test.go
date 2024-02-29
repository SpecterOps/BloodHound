package database

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
)

var (
	auditData = model.AuditData{"test": "message"}
	commitId  = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
	entry     = model.AuditEntry{
		CommitID: commitId,
		Action:   "TestAction",
		Model:    auditData,
		Status:   model.AuditStatusSuccess,
	}
	idResolver  = auth.NewIdentityResolver()
	requestID   = "12345"
	requestIP   = "1.2.3.4"
	testyUserId = uuid.FromStringOrNil("22222222-2222-2222-2222-222222222222")
	testyUser   = model.User{
		Unique: model.Unique{
			ID: testyUserId,
		},
		PrincipalName: "testy",
		EmailAddress:  null.StringFrom("test@email.com"),
	}
)

func setupRequest(user model.User) context.Context {
	bhCtx := ctx.Context{
		AuthCtx: auth.Context{
			Owner: user,
		},
		RequestID: requestID,
		RequestIP: requestIP,
	}
	testCtx := context.Background()
	testCtx = ctx.Set(testCtx, &bhCtx)

	return testCtx
}

func TestNewAuditLog(t *testing.T) {
	testCtx := setupRequest(testyUser)

	auditLog, err := newAuditLog(testCtx, entry, idResolver)
	if err != nil {
		t.Errorf("error creating audit log: %s", err.Error())
	}

	require.Equal(t, auditLog.Action, "TestAction")
	require.Equal(t, testyUser.EmailAddress.ValueOrZero(), auditLog.ActorEmail)
	require.Equal(t, testyUser.ID.String(), auditLog.ActorID)
	require.Equal(t, testyUser.PrincipalName, auditLog.ActorName)
	require.Equal(t, commitId, auditLog.CommitID)
	require.Equal(t, fmt.Sprintf("%s", auditData), fmt.Sprintf("%s", auditLog.Fields))
	require.Equal(t, requestID, auditLog.RequestID)
	require.Equal(t, requestIP, auditLog.SourceIpAddress)
	require.Equal(t, string(model.AuditStatusSuccess), auditLog.Status)
}

func TestNewAuditLog_Error(t *testing.T) {
	testCtx := setupRequest(testyUser)
	errorEntry := entry
	errorEntry.Status = model.AuditStatusFailure
	errorEntry.ErrorMsg = "this is a test error message"

	auditLog, err := newAuditLog(testCtx, errorEntry, idResolver)
	if err != nil {
		t.Errorf("error creating audit log: %s", err.Error())
	}

	require.Equal(t, string(model.AuditStatusFailure), auditLog.Status)
	require.Equal(t, auditLog.Fields, types.JSONUntypedObject{"error": "this is a test error message", "test": "message"})
}

func TestNewAuditLog_BadAuthContext(t *testing.T) {
	bhCtx := ctx.Context{
		RequestID: requestID,
		RequestIP: requestIP,
	}
	testCtx := context.Background()
	testCtx = ctx.Set(testCtx, &bhCtx)

	auditData := model.AuditData{"test": "message"}
	entry := model.AuditEntry{
		CommitID: commitId,
		Action:   "TestAction",
		Model:    auditData,
		Status:   model.AuditStatusFailure,
		ErrorMsg: "this is a test error message",
	}

	_, err := newAuditLog(testCtx, entry, idResolver)
	require.Equal(t, ErrAuthContextInvalid, err)
}
