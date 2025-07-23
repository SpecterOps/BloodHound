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

package database

import (
	"context"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// SavedQueriesPermissionsData methods representing the database interactions pertaining to the saved_queries_permissions model
type SavedQueriesPermissionsData interface {
	GetSavedQueryPermissions(ctx context.Context, queryID int64) ([]model.SavedQueriesPermissions, error)
	CreateSavedQueryPermissionToPublic(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error)
	CreateSavedQueryPermissionsToUsers(ctx context.Context, queryID int64, userIDs ...uuid.UUID) ([]model.SavedQueriesPermissions, error)
	DeleteSavedQueryPermissionsForUsers(ctx context.Context, queryID int64, userIDs ...uuid.UUID) error
	GetScopeForSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (SavedQueryScopeMap, error)
	IsSavedQueryPublic(ctx context.Context, savedQueryID int64) (bool, error)
	IsSavedQuerySharedToUser(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error)
	IsSavedQuerySharedToUserOrPublic(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error)
}

// SavedQueryScopeMap holds the information of a saved query's scope [IE: owned, shared, public]
type SavedQueryScopeMap map[model.SavedQueryScope]bool

// GetSavedQueryPermissions - returns permission data if the user owns the query or the query is public
func (s *BloodhoundDB) GetSavedQueryPermissions(ctx context.Context, queryID int64) ([]model.SavedQueriesPermissions, error) {
	var rows []model.SavedQueriesPermissions
	result := s.db.WithContext(ctx).Select("*").Table("saved_queries_permissions sqp").Where("sqp.query_id = ?", queryID).Find(&rows)
	return rows, CheckError(result)
}

// CreateSavedQueryPermissionToPublic creates a new entry to the SavedQueriesPermissions table granting public read permissions to all users
func (s *BloodhoundDB) CreateSavedQueryPermissionToPublic(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error) {
	permission := model.SavedQueriesPermissions{
		QueryID: queryID,
		Public:  true,
	}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := CheckError(tx.Create(&permission)); err != nil {
			return err
		} else if err := CheckError(tx.Table("saved_queries_permissions").Where("query_id = ? AND public = false", queryID).Delete(&model.SavedQueriesPermissions{})); err != nil {
			return err
		}

		return nil
	})

	return permission, err
}

// CreateSavedQueryPermissionsToUsers - attempts to save the given saved query permissions in batches of 100 in a transaction
func (s *BloodhoundDB) CreateSavedQueryPermissionsToUsers(ctx context.Context, queryID int64, userIDs ...uuid.UUID) ([]model.SavedQueriesPermissions, error) {
	var newPermissions []model.SavedQueriesPermissions
	for _, sharedUserID := range userIDs {
		newPermissions = append(newPermissions, model.SavedQueriesPermissions{
			QueryID:        queryID,
			SharedToUserID: NullUUID(sharedUserID),
			Public:         false,
		})
	}

	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(&newPermissions, 100)

	return newPermissions, CheckError(result)
}

// DeleteSavedQueryPermissionsForUsers batch deletes permissions associated with a query id and a list of users
// If no user ids are supplied, all records for query id are deleted
func (s *BloodhoundDB) DeleteSavedQueryPermissionsForUsers(ctx context.Context, queryID int64, userIDs ...uuid.UUID) error {
	result := s.db.WithContext(ctx).Table("saved_queries_permissions").Where("query_id = ?", queryID)
	if len(userIDs) > 0 {
		result = result.Where("shared_to_user_id IN ?", userIDs)
	}

	return CheckError(result.Delete(&model.SavedQueriesPermissions{}))
}

// GetScopeForSavedQuery will return a map of the possible scopes given a query id and a user id
func (s *BloodhoundDB) GetScopeForSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (SavedQueryScopeMap, error) {
	var (
		err    error
		scopes = SavedQueryScopeMap{
			model.SavedQueryScopePublic: false,
			model.SavedQueryScopeOwned:  false,
			model.SavedQueryScopeShared: false,
		}
	)

	// Check if the query was shared with the user publicly
	if scopes[model.SavedQueryScopePublic], err = s.IsSavedQueryPublic(ctx, queryID); err != nil {
		return scopes, err
	}

	// Check if the user owns the query
	if scopes[model.SavedQueryScopeOwned], err = s.SavedQueryBelongsToUser(ctx, userID, queryID); err != nil {
		return scopes, err
	}

	// Check if the user has had the query shared to them
	if scopes[model.SavedQueryScopeShared], err = s.IsSavedQuerySharedToUser(ctx, queryID, userID); err != nil {
		return scopes, err
	}

	return scopes, nil
}

// IsSavedQueryPublic returns true or false whether a provided saved query is public
func (s *BloodhoundDB) IsSavedQueryPublic(ctx context.Context, queryID int64) (bool, error) {
	rows := int64(0)
	result := s.db.WithContext(ctx).Table("saved_queries_permissions").Where("public = true AND query_id = ?", queryID).Count(&rows)

	return rows > 0, CheckError(result)
}

// IsSavedQuerySharedToUser returns true or false whether a provided saved query is shared with a provided user
func (s *BloodhoundDB) IsSavedQuerySharedToUser(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error) {
	rows := int64(0)
	result := s.db.WithContext(ctx).Table("saved_queries_permissions").Where("query_id = ? AND shared_to_user_id = ?", queryID, userID).Count(&rows)

	return rows > 0, CheckError(result)
}

func (s *BloodhoundDB) IsSavedQuerySharedToUserOrPublic(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error) {
	rows := int64(0)
	result := s.db.WithContext(ctx).Table("saved_queries_permissions").Where("query_id = ? AND (shared_to_user_id = ? or public = true)", queryID, userID).Count(&rows)
	return rows > 0, CheckError(result)
}
