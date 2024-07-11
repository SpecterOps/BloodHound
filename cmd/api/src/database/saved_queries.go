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
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func (s *BloodhoundDB) ListSavedQueries(ctx context.Context, userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) (model.SavedQueries, int, error) {
	var (
		queries model.SavedQueries
		result  *gorm.DB
		count   int64
		cursor  = s.Scope(Paginate(skip, limit)).WithContext(ctx).Where("user_id = ?", userID)
	)

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params)
		result = s.db.Model(&queries).WithContext(ctx).Where("user_id = ?", userID).Where(filter.SQLString, filter.Params).Count(&count)
	} else {
		result = s.db.Model(&queries).WithContext(ctx).Where("user_id = ?", userID).Count(&count)
	}

	if result.Error != nil {
		return queries, 0, result.Error
	}

	if order != "" {
		cursor = cursor.Order(order)
	}

	result = cursor.Joins("JOIN saved_queries_permissions sqp ON sqp.query_id = saved_queries.id AND (sqp.global OR saved_queries.user_id = ?)", userID).Find(&queries)

	return queries, int(count), CheckError(result)
}

func (s *BloodhoundDB) CreateSavedQuery(ctx context.Context, userID uuid.UUID, name string, query string, description string) (model.SavedQuery, error) {
	savedQuery := model.SavedQuery{
		UserID:      userID.String(),
		Name:        name,
		Description: description,
		Query:       query,
	}

	// TODO make this a transaction
	if result := s.db.WithContext(ctx).Create(&savedQuery); result.Error != nil {
		return model.SavedQuery{}, result.Error
	} else if _, err := s.CreateSavedQueryPermissionToUser(ctx, savedQuery.ID, userID); err != nil {
		return savedQuery, err
	}
	return savedQuery, nil
}

func (s *BloodhoundDB) DeleteSavedQuery(ctx context.Context, id int) error {
	return CheckError(s.db.WithContext(ctx).Delete(&model.SavedQuery{}, id))
}

func (s *BloodhoundDB) SavedQueryBelongsToUser(ctx context.Context, userID uuid.UUID, savedQueryID int) (bool, error) {
	var savedQuery model.SavedQuery
	if result := s.db.WithContext(ctx).First(&savedQuery, savedQueryID); result.Error != nil {
		return false, CheckError(result)
	} else if savedQuery.UserID == userID.String() {
		return true, nil
	} else {
		return false, nil
	}
}

func (s *BloodhoundDB) SearchSavedQueries(ctx context.Context, userID uuid.UUID, filter model.SQLFilter, skip, limit int, order string, name string, query string, description string) (model.SavedQueries, int, error) {
	var (
		queries model.SavedQueries
		result  *gorm.DB
		count   int64
		cursor  = s.Scope(Paginate(skip, limit)).WithContext(ctx).Where("user_id = ?", userID)
	)

	// if filter.SQLString != "" {
	// 	cursor = cursor.Where(filter.SQLString, filter.Params)
	// 	result = s.db.Joins("JOIN saved_queries_permissions sqp ON sqp.query_id = saved_queries.id AND (sqp.global OR saved_queries.user_id = ?)", userID).
	// 		Model(&queries).
	// 		WithContext(ctx).
	// 		Where("user_id = ?", userID).
	// 		Where(filter.SQLString, filter.Params).
	// 		Count(&count)
	// } else {
	result = s.db.Joins("JOIN saved_queries_permissions sqp ON sqp.query_id = saved_queries.id", userID).
		Model(&queries).
		WithContext(ctx).
		Where("user_id = ?", userID).
		Where("sqp.global = ? OR sqp.shared_to_user_id = ?", true, userID).
		Count(&count)
	//}

	if result.Error != nil {
		return queries, 0, result.Error
	}

	if order != "" {
		cursor = cursor.Order(order)
	}

	if description != "" {
		cursor = cursor.Where("description = ?", description)
	}

	if query != "" {
		cursor = cursor.Where("query = ?", query)
	}

	if name != "" {
		cursor = cursor.Where("name = ?", name)
	}

	result = cursor.Find(&queries)

	return queries, int(count), CheckError(result)
}

func (s *BloodhoundDB) GetSavedQueryPermissions(ctx context.Context, queryID int64) model.SavedQueriesPermissions {
	sqp := model.SavedQueriesPermissions{QueryID: queryID}
	s.db.WithContext(ctx).Where("query_id = ?", queryID).Find(&sqp)
	return sqp
}
