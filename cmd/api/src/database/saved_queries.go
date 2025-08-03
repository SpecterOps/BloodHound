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
	"errors"
	"strings"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type SavedQueriesData interface {
	GetSavedQuery(ctx context.Context, savedQueryID int64) (model.SavedQuery, error)
	ListSavedQueries(ctx context.Context, scope string, userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) ([]model.ScopedSavedQuery, int, error)
	CreateSavedQuery(ctx context.Context, userID uuid.UUID, name string, query string, description string) (model.SavedQuery, error)
	UpdateSavedQuery(ctx context.Context, savedQuery model.SavedQuery) (model.SavedQuery, error)
	DeleteSavedQuery(ctx context.Context, savedQueryID int64) error
	SavedQueryBelongsToUser(ctx context.Context, userID uuid.UUID, savedQueryID int64) (bool, error)
	GetSharedSavedQueries(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error)
	GetPublicSavedQueries(ctx context.Context) (model.SavedQueries, error)
	CreateSavedQueries(ctx context.Context, savedQueries model.SavedQueries) error
	GetAllSavedQueriesByUser(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error)
	GetOwnedSavedQueriesByUser(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error)
}

func (s *BloodhoundDB) GetSavedQuery(ctx context.Context, savedQueryID int64) (model.SavedQuery, error) {
	savedQuery := model.SavedQuery{}
	result := s.db.WithContext(ctx).First(&savedQuery, savedQueryID)
	return savedQuery, CheckError(result)
}

func (s *BloodhoundDB) ListSavedQueries(ctx context.Context, scope string, userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) ([]model.ScopedSavedQuery, int, error) {
	var (
		queries []model.ScopedSavedQuery
		// cant chain scope + cursor after declaration so must declare twice
		cursor         = s.db.WithContext(ctx).Select("DISTINCT sq.*, CASE WHEN (sqp.public = TRUE AND sq.user_id <> ?) THEN 'public' WHEN sqp.shared_to_user_id = ? THEN 'shared' ELSE 'owned' END AS scope", userID, userID).Table("saved_queries sq").Joins("LEFT JOIN public.saved_queries_permissions sqp ON sq.id = sqp.query_id")
		countCursor    = s.Scope(Paginate(skip, limit)).WithContext(ctx).Select("DISTINCT sq.*, CASE WHEN (sqp.public = TRUE AND sq.user_id <> ?) THEN 'public' WHEN sqp.shared_to_user_id = ? THEN 'shared' ELSE 'owned' END AS scope", userID, userID).Table("saved_queries sq").Joins("LEFT JOIN public.saved_queries_permissions sqp ON sq.id = sqp.query_id")
		orderReplacer  = strings.NewReplacer("id", "sq.id", "created_at", "sq.created_at", "updated_at", "sq.updated_at")
		filterReplacer = strings.NewReplacer("id", "sq.id")
		count          int64
	)

	// replace ambiguous identifiers with the desired table prefix
	order = orderReplacer.Replace(order)
	filter.SQLString = filterReplacer.Replace(filter.SQLString)

	switch strings.ToLower(scope) {
	case string(model.SavedQueryScopeOwned):
		cursor = cursor.Where("sq.user_id = ?", userID)
		countCursor = countCursor.Where("sq.user_id = ?", userID)
	case string(model.SavedQueryScopeShared):
		cursor = cursor.Where("sqp.shared_to_user_id = ?", userID)
		countCursor = countCursor.Where("sqp.shared_to_user_id = ?", userID)
	case string(model.SavedQueryScopePublic):
		cursor = cursor.Where("sqp.public = TRUE", userID)
		countCursor = countCursor.Where("sqp.public = TRUE", userID)
	case string(model.SavedQueryScopeAll):
		cursor = cursor.Where("sqp.public = TRUE OR sq.user_id = ? OR sqp.shared_to_user_id = ?", userID, userID)
		countCursor = countCursor.Where("sqp.public = TRUE OR sq.user_id = ? OR sqp.shared_to_user_id = ?", userID, userID)
	default:
		return nil, 0, errors.New("invalid scope parameter")
	}

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params...)
		countCursor = countCursor.Where(filter.SQLString, filter.Params...)
	} else if order != "" {
		cursor = cursor.Order(order)
		countCursor = countCursor.Order(order)
	}

	result := countCursor.Count(&count)
	if result.Error != nil {
		return nil, 0, CheckError(result)
	}

	cursor.Find(&queries)
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
	}
	return savedQuery.UserID == userID.String(), nil
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

func (s *BloodhoundDB) GetOwnedSavedQueriesByUser(ctx context.Context, userID uuid.UUID) (model.SavedQueries, error) {
	savedQueries := model.SavedQueries{}
	result := s.db.WithContext(ctx).Where("user_id = ?", userID).Find(&savedQueries)
	return savedQueries, CheckError(result)
}

// CreateSavedQueries - inserts saved queries records in batches
func (s *BloodhoundDB) CreateSavedQueries(ctx context.Context, savedQueries model.SavedQueries) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).CreateInBatches(&savedQueries, 100)
		return CheckError(result)
	})
}
