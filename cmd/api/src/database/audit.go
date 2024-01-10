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
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
)

const (
	ErrAuthContextInvalid = errors.Error("auth context is invalid")
)

func newAuditLog(ctx ctx.Context, action string, data model.Auditable, idResolver auth.IdentityResolver) (model.AuditLog, error) {
	auditLog := model.AuditLog{
		Action:    action,
		Fields:    types.JSONUntypedObject(data.AuditData()),
		RequestID: ctx.RequestID,
		Status:    "success", // TODO: parameterize this so we can pass the actual status instead of hard-coding
	}

	authContext := ctx.AuthCtx
	if !authContext.Authenticated() {
		return auditLog, ErrAuthContextInvalid
	} else if identity, err := idResolver.GetIdentity(ctx.AuthCtx); err != nil {
		return auditLog, ErrAuthContextInvalid
	} else {
		auditLog.ActorID = identity.ID.String()
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

func (s *BloodhoundDB) ListAuditLogs(before, after time.Time, offset, limit int, order string, filter model.SQLFilter) (model.AuditLogs, int, error) {
	var (
		auditLogs model.AuditLogs
		result    *gorm.DB
		cursor    = s.Scope(Paginate(offset, limit)).Where("created_at between ? and ?", after, before).Order("created_at desc")
		actors    = make(map[string]model.User)
		count     int64
	)

	if order != "" && filter.SQLString == "" {
		result = cursor.Order(order).Find(&auditLogs).Count(&count)
	} else if order != "" && filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params).Order(order).Find(&auditLogs).Count(&count)
	} else if order == "" && filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params).Find(&auditLogs).Count(&count)
	} else {
		result = cursor.Find(&auditLogs).Count(&count)
	}

	// Populate the actor name and email from the users table. Uses a map as a cache to avoid looking up the same user multiple times.
	// NOTE: This is intended to be a temporary solution for including this data directly in the returned AuditLog. This is due to a lack
	// of flexibility with gorm's associations/JOINs.
	// TODO: When gorm is removed, select these values with a JOIN on the users table.
	for i, log := range auditLogs {
		var actor model.User
		actor, ok := actors[log.ActorID]
		if !ok {
			if userId, err := uuid.FromString(log.ActorID); err != nil {
				return auditLogs, int(count), fmt.Errorf("error parsing actor_id: %w", err)
			} else if user, err := s.GetUser(userId); err != nil {
				return auditLogs, int(count), fmt.Errorf("error retrieving actor: %w", err)
			} else {
				actors[user.ID.String()] = user
				actor = user
			}
		}

		auditLogs[i].ActorName = actor.PrincipalName
		auditLogs[i].ActorEmail = actor.EmailAddress.ValueOrZero()
	}

	return auditLogs, int(count), CheckError(result)
}
