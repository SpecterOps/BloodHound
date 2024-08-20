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
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SavedQueriesPermissionsData methods representing the database interactions pertaining to the saved_queries_permissions model
type SavedQueriesPermissionsData interface {
	CreateSavedQueryPermissionToUser(ctx context.Context, queryID int64, userID uuid.UUID) (model.SavedQueriesPermissions, error)
	CreateSavedQueryPermissionToPublic(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error)
	CreateSavedQueryPermissionsBatch(ctx context.Context, savedQueryPermissions []model.SavedQueriesPermissions) ([]model.SavedQueriesPermissions, error)
	CheckUserHasPermissionToSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error)
	GetPermissionsForSavedQuery(ctx context.Context, queryID int64) ([]model.SavedQueriesPermissions, error)
	GetScopeForSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (SavedQueryScopeMap, error)
	DeleteSavedQueryPermissionPublic(ctx context.Context, queryID int64) error
	DeleteSavedQueryPermissionsForUser(ctx context.Context, queryID int64, userID uuid.UUID) error
	IsSavedQueryShared(ctx context.Context, queryID int64) (bool, error)
}

// SavedQueryScopeMap holds the information of a saved query's scope [IE: owned, shared, public]
type SavedQueryScopeMap map[model.SavedQueryScope]bool

// CreateSavedQueryPermissionToUser creates a new entry to the SavedQueriesPermissions table granting a provided user id to access a provided query
func (s *BloodhoundDB) CreateSavedQueryPermissionToUser(ctx context.Context, queryID int64, userID uuid.UUID) (model.SavedQueriesPermissions, error) {
	permission := model.SavedQueriesPermissions{
		QueryID:        queryID,
		SharedToUserID: NullUUID(userID),
		Public:         false,
	}

	return permission, CheckError(s.db.WithContext(ctx).Create(&permission))
}

// CreateSavedQueryPermissionToPublic creates a new entry to the SavedQueriesPermissions table granting global read permissions to all users
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

// CreateSavedQueryPermissionsBatch attempts to save the given saved query permissions in batches of 100 in a transaction
func (s *BloodhoundDB) CreateSavedQueryPermissionsBatch(ctx context.Context, savedQueryPermissions []model.SavedQueriesPermissions) ([]model.SavedQueriesPermissions, error) {
	result := s.db.WithContext(ctx).Clauses(clause.OnConflict{
		DoNothing: true,
	}).CreateInBatches(&savedQueryPermissions, 100)

	return savedQueryPermissions, CheckError(result)
}

// CheckUserHasPermissionToSavedQuery returns true or false depending on if the given userID has permission to read the given queryID
func (s *BloodhoundDB) CheckUserHasPermissionToSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error) {
	rows := int64(0)
	result := s.db.WithContext(ctx).Table("saved_queries_permissions").Select("*").Where("query_id = ? AND shared_to_user_id = ?", queryID, userID).Or("query_id = ? AND public = true", queryID).Limit(1).Count(&rows)

	return rows > 0, CheckError(result)
}

// GetPermissionsForSavedQuery gets all permissions associated with the provided query ID
func (s *BloodhoundDB) GetPermissionsForSavedQuery(ctx context.Context, queryID int64) ([]model.SavedQueriesPermissions, error) {
	queryPermissions := make([]model.SavedQueriesPermissions, 0)
	result := s.db.WithContext(ctx).Where("query_id = ?", queryID).Find(&queryPermissions)

	return queryPermissions, CheckError(result)
}

// GetScopeForSavedQuery will return a map of the possible scopes given a query id and a user id
func (s *BloodhoundDB) GetScopeForSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (SavedQueryScopeMap, error) {
	scopes := SavedQueryScopeMap{
		model.SavedQueryScopePublic: false,
		model.SavedQueryScopeOwned:  false,
		model.SavedQueryScopeShared: false,
	}

	// Check if the query was shared with the user publicly
	publicCount := int64(0)
	if result := s.db.WithContext(ctx).Select("*").Table("saved_queries_permissions").Where("public = true AND query_id = ?", queryID).Count(&publicCount).Limit(1); result.Error != nil {
		return scopes, CheckError(result)
	} else if publicCount > 0 {
		scopes[model.SavedQueryScopePublic] = true
	}

	// Check if the user owns the query
	ownedCount := int64(0)
	if result := s.db.WithContext(ctx).Select("*").Table("saved_queries").Where("id = ? AND user_id = ?", queryID, userID).Count(&ownedCount).Limit(1); result.Error != nil {
		return scopes, CheckError(result)
	} else if ownedCount > 0 {
		scopes[model.SavedQueryScopeOwned] = true
	}

	// Check if the user has had the query shared to them
	sharedCount := int64(0)
	if result := s.db.WithContext(ctx).Select("*").Table("saved_queries_permissions").Where("query_id = ? AND shared_to_user_id = ?", queryID, userID).Count(&sharedCount).Limit(1); result.Error != nil {
		return scopes, CheckError(result)
	} else if sharedCount > 0 {
		scopes[model.SavedQueryScopeShared] = true
	}

	return scopes, nil
}

// DeleteSavedQueryPermissionPublic deletes the saved query permission associated with the passed in query id and public = true
func (s *BloodhoundDB) DeleteSavedQueryPermissionPublic(ctx context.Context, queryID int64) error {
	return CheckError(s.db.WithContext(ctx).Table("saved_queries_permissions").Where("public = true AND query_id = ?", queryID).Delete(&model.SavedQueriesPermissions{}))
}

// DeleteSavedQueryPermissionsForUser deletes all permissions associated with the passed in query id and user id
func (s *BloodhoundDB) DeleteSavedQueryPermissionsForUser(ctx context.Context, queryID int64, userID uuid.UUID) error {
	return CheckError(s.db.WithContext(ctx).Table("saved_queries_permissions").Where("query_id = ? AND shared_to_user_id = ?", queryID, userID).Delete(&model.SavedQueriesPermissions{}))
}

// IsSavedQueryShared returns true or false depending on if the saved query is being shared with users
func (s *BloodhoundDB) IsSavedQueryShared(ctx context.Context, queryID int64) (bool, error) {
	sharedCount := int64(0)
	if result := s.db.WithContext(ctx).Select("*").Table("saved_queries_permissions").Where("query_id = ? AND shared_to_user_id IS NOT NULL", queryID).Count(&sharedCount).Limit(1); result.Error != nil {
		return false, CheckError(result)
	} else if sharedCount > 0 {
		return true, nil
	} else {
		return false, nil
	}
}
