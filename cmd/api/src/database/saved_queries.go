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
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func (s *BloodhoundDB) ListSavedQueries(userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) (model.SavedQueries, int, error) {
	var (
		queries model.SavedQueries
		result  *gorm.DB
		count   int64
		cursor  = s.Scope(Paginate(skip, limit)).Where("user_id = ?", userID)
	)

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params)
		result = s.db.Model(&queries).Where("user_id = ?", userID).Where(filter.SQLString, filter.Params).Count(&count)
	} else {
		result = s.db.Model(&queries).Where("user_id = ?", userID).Count(&count)
	}

	if result.Error != nil {
		return queries, 0, result.Error
	}

	if order != "" {
		cursor = cursor.Order(order)
	}
	result = cursor.Find(&queries)

	return queries, int(count), CheckError(result)
}

func (s *BloodhoundDB) CreateSavedQuery(userID uuid.UUID, name string, query string) (model.SavedQuery, error) {
	savedQuery := model.SavedQuery{
		UserID: userID.String(),
		Name:   name,
		Query:  query,
	}

	return savedQuery, CheckError(s.db.Create(&savedQuery))
}

func (s *BloodhoundDB) DeleteSavedQuery(id int) error {
	return CheckError(s.db.Delete(&model.SavedQuery{}, id))
}

func (s *BloodhoundDB) SavedQueryBelongsToUser(userID uuid.UUID, savedQueryID int) (bool, error) {
	var savedQuery model.SavedQuery
	if result := s.db.First(&savedQuery, savedQueryID); result.Error != nil {
		return false, CheckError(result)
	} else if savedQuery.UserID == userID.String() {
		return true, nil
	} else {
		return false, nil
	}
}
