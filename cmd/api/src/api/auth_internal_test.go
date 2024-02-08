package api

import (
	"context"
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/ctx"
	dbMocks "github.com/specterops/bloodhound/src/database/mocks"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"go.uber.org/mock/gomock"
)

var (
	commitId    = uuid.FromStringOrNil("11111111-1111-1111-1111-111111111111")
	testyUserId = uuid.FromStringOrNil("22222222-2222-2222-2222-222222222222")
	testyUser   = model.User{
		Unique: model.Unique{
			ID: testyUserId,
		},
		PrincipalName: "testy",
		EmailAddress:  null.StringFrom("test@email.com"),
	}
)

func setupRequest(user model.User) (context.Context, LoginRequest) {
	bhCtx := ctx.Context{
		RequestID: "12345",
		RequestIP: "1.2.3.4",
	}
	testCtx := context.Background()
	testCtx = ctx.Set(testCtx, &bhCtx)

	loginRequest := LoginRequest{Username: user.PrincipalName}

	return testCtx, loginRequest
}

func buildAuditLogs(testCtx context.Context, user model.User, loginRequest LoginRequest) model.AuditLog {
	bhCtx := ctx.Get(testCtx)

	auditLog := model.AuditLog{
		Action:          "LoginAttempt",
		ActorName:       user.PrincipalName,
		ActorEmail:      user.EmailAddress.ValueOrZero(),
		Fields:          types.JSONUntypedObject{"username": loginRequest.Username},
		RequestID:       bhCtx.RequestID,
		SourceIpAddress: bhCtx.RequestIP,
		Status:          string(model.AuditStatusSuccess),
		CommitID:        commitId,
	}

	if user.ID.String() != "00000000-0000-0000-0000-000000000000" {
		auditLog.ActorID = user.ID.String()
	}
	fmt.Printf("****** auditLog is %+v\n ****** user.ID.String is %s", auditLog, user.ID.String())
	return auditLog
}

func TestAuditLogin(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockDB := dbMocks.NewMockDatabase(mockCtrl)
	a := authenticator{
		db: mockDB,
	}
	testCtx, loginRequest := setupRequest(testyUser)
	expectedAuditLog := buildAuditLogs(testCtx, testyUser, loginRequest)

	mockDB.EXPECT().CreateAuditLog(expectedAuditLog)
	a.auditLogin(testCtx, commitId, testyUser, loginRequest, string(model.AuditStatusSuccess), nil)
}

func TestAuditLogin_UserNotFound(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	mockDB := dbMocks.NewMockDatabase(mockCtrl)
	a := authenticator{
		db: mockDB,
	}
	testCtx, loginRequest := setupRequest(model.User{})
	expectedAuditLog := buildAuditLogs(testCtx, model.User{}, loginRequest)

	mockDB.EXPECT().CreateAuditLog(expectedAuditLog)
	a.auditLogin(testCtx, commitId, model.User{}, loginRequest, string(model.AuditStatusSuccess), nil)
}
