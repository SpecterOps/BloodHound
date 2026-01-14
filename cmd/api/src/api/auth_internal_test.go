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

package api

import (
	"context"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	dbMocks "github.com/specterops/bloodhound/cmd/api/src/database/mocks"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
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

	var loginRequest LoginRequest
	if user.PrincipalName == "" {
		loginRequest.Username = "nonExistentUser"
	} else {
		loginRequest.Username = user.PrincipalName
	}

	return testCtx, loginRequest
}

func buildAuditLog(testCtx context.Context, status model.AuditLogEntryStatus, user model.User, fields types.JSONUntypedObject) model.AuditLog {
	bhCtx := ctx.Get(testCtx)

	auditLog := model.AuditLog{
		Action:          model.AuditLogActionLoginAttempt,
		ActorName:       user.PrincipalName,
		ActorEmail:      user.EmailAddress.ValueOrZero(),
		Fields:          fields,
		RequestID:       bhCtx.RequestID,
		SourceIpAddress: bhCtx.RequestIP,
		Status:          status,
		CommitID:        commitId,
	}

	if user.ID.String() != "00000000-0000-0000-0000-000000000000" {
		auditLog.ActorID = user.ID.String()
	}

	return auditLog
}

func TestAuditLogin(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		mockDB   = dbMocks.NewMockDatabase(mockCtrl)
		a        = AuthenticatorBase{db: mockDB}
	)

	testCtx, loginRequest := setupRequest(testyUser)
	fields := types.JSONUntypedObject{"username": loginRequest.Username, "auth_type": auth.ProviderTypeSecret}
	expectedAuditLog := buildAuditLog(testCtx, model.AuditLogStatusSuccess, testyUser, fields)

	mockDB.EXPECT().CreateAuditLog(testCtx, expectedAuditLog)
	a.auditLogin(testCtx, commitId, model.AuditLogStatusSuccess, testyUser, fields)
}

func TestAuditLogin_UserNotFound(t *testing.T) {
	var (
		mockCtrl = gomock.NewController(t)
		mockDB   = dbMocks.NewMockDatabase(mockCtrl)
		a        = AuthenticatorBase{db: mockDB}
	)
	testCtx, loginRequest := setupRequest(model.User{})
	fields := types.JSONUntypedObject{"username": loginRequest.Username, "auth_type": auth.ProviderTypeSecret, "error": ErrInvalidAuth}
	expectedAuditLog := buildAuditLog(testCtx, model.AuditLogStatusFailure, model.User{}, fields)

	mockDB.EXPECT().CreateAuditLog(testCtx, expectedAuditLog)
	a.auditLogin(testCtx, commitId, model.AuditLogStatusFailure, model.User{}, fields)
}
