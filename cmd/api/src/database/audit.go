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

	"github.com/specterops/bloodhound/src/ctx"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/src/model"
)

func (s *BloodhoundDB) CreateAuditLog(auditLog *model.AuditLog) error {
	return CheckError(s.db.Create(&auditLog))
}

func (s *BloodhoundDB) AppendAuditLog(context ctx.Context, unused string, model model.Auditable) error {
	if auditLog, err := ctx.NewAuditLogFromContext(context, s.idResolver); err != nil {
		return fmt.Errorf("error creating audit log from context: %w", err)
	} else {
		return s.CreateAuditLog(&auditLog)
	}
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
