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
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type SavedQueriesData interface {
	GetSavedQuery(ctx context.Context, savedQueryID int64) (model.SavedQuery, error)
	ListSavedQueries(ctx context.Context, userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) (model.SavedQueries, int, error)
	CreateSavedQuery(ctx context.Context, userID uuid.UUID, name string, query string, description string) (model.SavedQuery, error)
	UpdateSavedQuery(ctx context.Context, savedQuery model.SavedQuery) (model.SavedQuery, error)
	DeleteSavedQuery(ctx context.Context, savedQueryID int64) error
	SavedQueryBelongsToUser(ctx context.Context, userID uuid.UUID, savedQueryID int64) (bool, error)
	GetSharedSavedQueries(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error)
	GetPublicSavedQueries(ctx context.Context) (model.SavedQueries, error)
	CreateSavedQueries(ctx context.Context, savedQueries model.SavedQueries) error
	GetAllSavedQueriesByUser(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error)
}

func (s *BloodhoundDB) GetSavedQuery(ctx context.Context, savedQueryID int64) (model.SavedQuery, error) {
	savedQuery := model.SavedQuery{}
	result := s.db.WithContext(ctx).First(&savedQuery, savedQueryID)
	return savedQuery, CheckError(result)
}

func (s *BloodhoundDB) ListSavedQueries(ctx context.Context, userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) (model.SavedQueries, int, error) {
	var (
		queries model.SavedQueries
		result  *gorm.DB
		count   int64
		cursor  = s.Scope(Paginate(skip, limit)).WithContext(ctx).Where("user_id = ?", userID)
	)

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params...)
		result = s.db.Model(&queries).WithContext(ctx).Where("user_id = ?", userID).Where(filter.SQLString, filter.Params...).Count(&count)
	} else {
		result = s.db.Model(&queries).WithContext(ctx).Where("user_id = ?", userID).Count(&count)
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

func (s *BloodhoundDB) CreateSavedQuery(ctx context.Context, userID uuid.UUID, name string, query string, description string) (model.SavedQuery, error) {
	savedQuery := model.SavedQuery{
		UserID:      userID.String(),
		Name:        name,
		Query:       query,
		Description: description,
	}

	return savedQuery, CheckError(s.db.WithContext(ctx).Create(&savedQuery))
}

func (s *BloodhoundDB) UpdateSavedQuery(ctx context.Context, savedQuery model.SavedQuery) (model.SavedQuery, error) {
	return savedQuery, CheckError(s.db.WithContext(ctx).Save(&savedQuery))
}

func (s *BloodhoundDB) DeleteSavedQuery(ctx context.Context, savedQueryID int64) error {
	return CheckError(s.db.WithContext(ctx).Delete(&model.SavedQuery{}, savedQueryID))
}

func (s *BloodhoundDB) SavedQueryBelongsToUser(ctx context.Context, userID uuid.UUID, savedQueryID int64) (bool, error) {
	var savedQuery model.SavedQuery
	if result := s.db.WithContext(ctx).First(&savedQuery, savedQueryID); result.Error != nil {
		return false, CheckError(result)
	} else if savedQuery.UserID == userID.String() {
		return true, nil
	} else {
		return false, nil
	}
}

// GetSharedSavedQueries returns all the saved queries that the given userID has access to, including global queries
func (s *BloodhoundDB) GetSharedSavedQueries(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error) {
	savedQueries := model.SavedQueries{}

	result := s.db.WithContext(ctx).Select("saved_queries.*").Joins("JOIN saved_queries_permissions sqp ON sqp.query_id = saved_queries.id").Where("sqp.shared_to_user_id = ? ", userID).Find(&savedQueries)

	return savedQueries, CheckError(result)
}

// GetPublicSavedQueries returns all the queries that were shared publicly
func (s *BloodhoundDB) GetPublicSavedQueries(ctx context.Context) (model.SavedQueries, error) {
	savedQueries := model.SavedQueries{}

	result := s.db.WithContext(ctx).Select("saved_queries.*").Joins("JOIN saved_queries_permissions sqp ON sqp.query_id = saved_queries.id").Where("sqp.public = true").Find(&savedQueries)

	return savedQueries, CheckError(result)
}

// GetAllSavedQueriesByUser - Returns queries that are public, owned by, or shared to the user.
func (s *BloodhoundDB) GetAllSavedQueriesByUser(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error) {
	savedQueries := model.SavedQueries{}
	results := s.db.WithContext(ctx).Select("DISTINCT saved_queries.*").Joins("LEFT JOIN saved_queries_permissions sqp ON sqp.query_id = saved_queries.id").Where("sqp.public = true OR saved_queries.user_id = ? OR sqp.shared_to_user_id = ?", userID, userID).Find(&savedQueries)
	return savedQueries, CheckError(results)
}

// CreateSavedQueries - inserts saved queries records in batches
func (s *BloodhoundDB) CreateSavedQueries(ctx context.Context, savedQueries model.SavedQueries) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).CreateInBatches(&savedQueries, 100)
		return CheckError(result)
	})
}
