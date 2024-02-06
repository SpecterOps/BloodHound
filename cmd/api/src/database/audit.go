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
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

const (
	ErrAuthContextInvalid = errors.Error("auth context is invalid")
)

func newAuditLog(context context.Context, entry model.AuditEntry, idResolver auth.IdentityResolver) (model.AuditLog, error) {
	bheCtx := ctx.Get(context)

	auditLog := model.AuditLog{
		Action:          entry.Action,
		Fields:          types.JSONUntypedObject(entry.Model.AuditData()),
		RequestID:       bheCtx.RequestID,
		SourceIpAddress: bheCtx.RequestIP,
		Status:          string(entry.Status),
		CommitID:        entry.CommitID,
	}

	if entry.ErrorMsg != "" {
		auditLog.Fields["error"] = entry.ErrorMsg
	}

	authContext := bheCtx.AuthCtx
	if !authContext.Authenticated() {
		return auditLog, ErrAuthContextInvalid
	} else if identity, err := idResolver.GetIdentity(bheCtx.AuthCtx); err != nil {
		return auditLog, ErrAuthContextInvalid
	} else {
		auditLog.ActorID = identity.ID.String()
		auditLog.ActorName = identity.Name
		auditLog.ActorEmail = identity.Email
	}

	return auditLog, nil
}

func (s *BloodhoundDB) AppendAuditLog(ctx context.Context, entry model.AuditEntry) error {
	if auditLog, err := newAuditLog(ctx, entry, s.idResolver); err != nil && err != ErrAuthContextInvalid {
		return fmt.Errorf("audit log append: %w", err)
	} else {
		return s.CreateAuditLog(auditLog)
	}
}

func (s *BloodhoundDB) CreateAuditLog(auditLog model.AuditLog) error {
	return CheckError(s.db.Create(&auditLog))
}

func (s *BloodhoundDB) ListAuditLogs(before, after time.Time, offset, limit int, order string, filter model.SQLFilter) (model.AuditLogs, int, error) {
	var (
		auditLogs model.AuditLogs
		result    *gorm.DB
		cursor    = s.Scope(Paginate(offset, limit)).Where("created_at between ? and ?", after, before).Order("created_at desc")
		count     int64
	)

	// This code went through a partial refactor when adding support for new fields.
	// See the comments here for more information: https://github.com/SpecterOps/BloodHound/pull/297#issuecomment-1887640827

	if filter.SQLString != "" {
		result = s.db.Model(&auditLogs).Where(filter.SQLString, filter.Params).Count(&count)
	} else {
		result = s.db.Model(&auditLogs).Count(&count)
	}

	if result.Error != nil {
		return nil, 0, CheckError(result)
	}

	if order != "" && filter.SQLString == "" {
		result = cursor.Order(order).Find(&auditLogs)
	} else if order != "" && filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params).Order(order).Find(&auditLogs)
	} else if order == "" && filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params).Find(&auditLogs)
	} else {
		result = cursor.Find(&auditLogs)
	}

	return auditLogs, int(count), CheckError(result)
}

func (s *BloodhoundDB) MaybeAuditableTransaction(ctx context.Context, auditDisabled bool, auditEntry model.AuditEntry, f func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	if auditDisabled {
		return s.db.Transaction(f, opts...)
	} else {
		return s.AuditableTransaction(ctx, auditEntry, f, opts...)
	}
}

func (s *BloodhoundDB) AuditableTransaction(ctx context.Context, auditEntry model.AuditEntry, f func(tx *gorm.DB) error, opts ...*sql.TxOptions) error {
	var (
		commitID, err = uuid.NewV4()
	)

	if err != nil {
		return fmt.Errorf("commitID could not be created: %w", err)
	}

	auditEntry.CommitID = commitID
	auditEntry.Status = model.AuditStatusIntent

	if err := s.AppendAuditLog(ctx, auditEntry); err != nil {
		return fmt.Errorf("could not append intent to audit log: %w", err)
	}

	err = s.db.Transaction(f, opts...)

	if err != nil {
		auditEntry.Status = model.AuditStatusFailure
		auditEntry.ErrorMsg = err.Error()
	} else {
		auditEntry.Status = model.AuditStatusSuccess
	}

	if err := s.AppendAuditLog(ctx, auditEntry); err != nil {
		return fmt.Errorf("could not append %s to audit log: %w", auditEntry.Status, err)
	}

	return err
}
