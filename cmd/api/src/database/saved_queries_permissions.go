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
	CheckUserHasPermissionToSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error)
	GetPermissionsForSavedQuery(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error)
	DeleteSavedQueryPermissionsForUser(ctx context.Context, queryID int64, userID uuid.UUID) error
	DeleteSavedQueryPermissionsForUsers(ctx context.Context, queryID int64, userIDs []uuid.UUID) error
}

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

	return permission, CheckError(s.db.WithContext(ctx).Create(&permission))
}

// CheckUserHasPermissionToSavedQuery returns true or false depending on if the given userID has permission to read the given queryID
// This does not check for ownership
func (s *BloodhoundDB) CheckUserHasPermissionToSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error) {
	rows := int64(0)
	result := s.db.WithContext(ctx).Table("saved_queries_permissions").Select("*").Where("query_id = ? AND shared_to_user_id = ?", queryID, userID).Or("query_id = ? AND public = true", queryID).Limit(1).Count(&rows)

	return rows > 0, CheckError(result)
}

// GetPermissionsForSavedQuery gets all permissions associated with the provided query ID
func (s *BloodhoundDB) GetPermissionsForSavedQuery(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error) {
	queryPermissions := model.SavedQueriesPermissions{QueryID: queryID}
	result := s.db.WithContext(ctx).Where("query_id = ?", queryID).Find(&queryPermissions)
	return queryPermissions, CheckError(result)
}

// DeleteSavedQueryPermissionsForUser deletes all permissions associated with the passed in query id and user id
func (s *BloodhoundDB) DeleteSavedQueryPermissionsForUser(ctx context.Context, queryID int64, userID uuid.UUID) error {
	return CheckError(s.db.WithContext(ctx).Table("saved_queries_permissions").Where("query_id = ? AND shared_to_user_id = ?", queryID, userID).Delete(&model.SavedQueriesPermissions{}))
}

// DeleteSavedQueryPermissionsForUsers batch deletes permissions associated a query id and a list of users
func (s *BloodhoundDB) DeleteSavedQueryPermissionsForUsers(ctx context.Context, queryID int64, userIDs []uuid.UUID) error {
	return CheckError(s.db.WithContext(ctx).Table("saved_queries_permissions").Where("query_id = ? AND shared_to_user_id IN ?", queryID, userIDs).Delete(&model.SavedQueriesPermissions{}))
}
