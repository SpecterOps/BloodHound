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
)

// SavedQueriesPermissionsData methods representing the database interactions pertaining to the saved_queries_permissions model
type SavedQueriesPermissionsData interface {
	CreateSavedQueryPermissionToUser(ctx context.Context, queryID int64, userID uuid.UUID) (model.SavedQueriesPermissions, error)
	CreateSavedQueryPermissionToPublic(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error)
	CreateSavedQueryPermissionsBatch(ctx context.Context, savedQueryPermissions []model.SavedQueriesPermissions) error
	CheckUserHasPermissionToSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error)
	GetPermissionsForSavedQuery(ctx context.Context, queryID int64) ([]model.SavedQueriesPermissions, error)
	GetScopeForSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (SavedQueryScopeMap, error)
}

type savedQueryScope string

const (
	SavedQueryScopeOwned  savedQueryScope = "Owned"
	SavedQueryScopeShared savedQueryScope = "Shared"
	SavedQueryScopePublic savedQueryScope = "Public"
)

// SavedQueryScopeMap holds the information of a saved query's scope [IE: owned, shared, public]
type SavedQueryScopeMap map[savedQueryScope]bool

// CreateSavedQueryPermissionToUser creates a new entry to the SavedQueriesPermissions table granting a provided user id to access a provided query
func (s *BloodhoundDB) CreateSavedQueryPermissionToUser(ctx context.Context, queryID int64, userID uuid.UUID) (model.SavedQueriesPermissions, error) {
	permission := model.SavedQueriesPermissions{
		QueryID:        queryID,
		SharedToUserID: NullUUID(userID),
		Public:         false,
	}

	return permission, CheckError(s.db.WithContext(ctx).Create(&permission))
}

// CreateSavedQueryPermissionToPublic creates a new entry to the SavedQueriesPermissions table granting public read permissions to all users
func (s *BloodhoundDB) CreateSavedQueryPermissionToPublic(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error) {
	permission := model.SavedQueriesPermissions{
		QueryID: queryID,
		Public:  true,
	}

	return permission, CheckError(s.db.WithContext(ctx).Create(&permission))
}

// CreateSavedQueryPermissionsBatch attempts to save the given saved query permissions in batches of 100 in a transaction
func (s *BloodhoundDB) CreateSavedQueryPermissionsBatch(ctx context.Context, savedQueryPermissions []model.SavedQueriesPermissions) error {
	result := s.db.WithContext(ctx).CreateInBatches(savedQueryPermissions, 100)

	return CheckError(result)
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
		SavedQueryScopePublic: false,
		SavedQueryScopeOwned:  false,
		SavedQueryScopeShared: false,
	}

	// Check if the query was shared with the user publicly
	publicCount := int64(0)
	if result := s.db.WithContext(ctx).Select("*").Table("saved_queries_permissions").Where("public = true AND query_id = ?", queryID).Count(&publicCount).Limit(1); result.Error != nil {
		return scopes, result.Error
	} else if publicCount > 0 {
		scopes[SavedQueryScopePublic] = true
	}

	// Check if the user owns the query
	ownedCount := int64(0)
	if result := s.db.WithContext(ctx).Select("*").Table("saved_queries").Where("id = ? AND user_id = ?", queryID, userID).Count(&ownedCount).Limit(1); result.Error != nil {
		return scopes, result.Error
	} else if ownedCount > 0 {
		scopes[SavedQueryScopeOwned] = true
	}

	// Check if the user has had the query shared to them
	sharedCount := int64(0)
	if result := s.db.WithContext(ctx).Select("*").Table("saved_queries_permissions").Where("query_id = ? AND shared_to_user_id = ?", queryID, userID).Count(&sharedCount).Limit(1); result.Error != nil {
		return scopes, result.Error
	} else if sharedCount > 0 {
		scopes[SavedQueryScopeShared] = true
	}

	return scopes, nil
}
