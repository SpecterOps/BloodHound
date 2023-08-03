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

package database

import (
	"time"

	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/errors"
)

const (
	ErrAuthContextInvalid = errors.Error("auth context is invalid")
)

func newAuditLog(ctx ctx.Context, action string, data model.Auditable, idResolver auth.IdentityResolver) (model.AuditLog, error) {
	auditLog := model.AuditLog{
		Action:    action,
		Fields:    types.JSONUntypedObject(data.AuditData()),
		RequestID: ctx.RequestID,
	}

	authContext := ctx.AuthCtx
	if !authContext.Authenticated() {
		return auditLog, ErrAuthContextInvalid
	} else if identity, err := idResolver.GetIdentity(ctx.AuthCtx); err != nil {
		return auditLog, ErrAuthContextInvalid
	} else {
		auditLog.ActorID = identity.ID.String()
		auditLog.ActorName = identity.Name
	}

	return auditLog, nil
}

func (s *BloodhoundDB) AppendAuditLog(ctx ctx.Context, action string, data model.Auditable) error {
	if auditLog, err := newAuditLog(ctx, action, data, s.idResolver); err != nil {
		return err
	} else {
		return CheckError(s.db.Create(&auditLog))
	}
}

func (s *BloodhoundDB) GetAuditLogs(offset, limit int) (model.AuditLogs, error) {
	var auditLogs model.AuditLogs
	return auditLogs, CheckError(s.Scope(Paginate(offset, limit)).Order("created_at desc").Find(&auditLogs))
}

func (s *BloodhoundDB) GetAuditLogsBetween(before, after time.Time, offset, limit int, order string, filter model.SQLFilter) (model.AuditLogs, error) {
	var (
		auditLogs model.AuditLogs
		result    *gorm.DB
		cursor    = s.Scope(Paginate(offset, limit)).Where("created_at between ? and ?", after, before).Order("created_at desc")
	)

	if order != "" && filter.SQLString == "" {
		result = cursor.Order(order).Find(&auditLogs)
	} else if order != "" && filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params).Order(order).Find(&auditLogs)
	} else if order == "" && filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params).Find(&auditLogs)
	} else {
		result = cursor.Find(&auditLogs)
	}

	return auditLogs, CheckError(result)
}

func (s *BloodhoundDB) GetAuditLogsCount() (int, error) {
	var count int64
	result := s.db.Find(&model.AuditLogs{}).Count(&count)
	if result.Error != nil {
		return 0, result.Error
	}
	return int(count), nil
}
