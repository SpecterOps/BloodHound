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

// CreateSavedQueryPermissionToUser creates a new entry to the SavedQueriesPermissions table granting a provided user id to access a provided query
func (s *BloodhoundDB) CreateSavedQueryPermissionToUser(ctx context.Context, queryID int64, userID uuid.UUID) (model.SavedQueriesPermissions, error) {
	permission := model.SavedQueriesPermissions{
		QueryID:        queryID,
		SharedToUserID: userID,
		Global:         false,
	}

	return permission, CheckError(s.db.WithContext(ctx).Create(&permission))

}

// CreateSavedQueryPermissionToGlobal creates a new entry to the SavedQueriesPermissions table granting global read permissions to all users
func (s *BloodhoundDB) CreateSavedQueryPermissionToGlobal(ctx context.Context, queryID int64) (model.SavedQueriesPermissions, error) {
	permission := model.SavedQueriesPermissions{
		QueryID: queryID,
		Global:  true,
	}

	return permission, CheckError(s.db.WithContext(ctx).Create(&permission))
}

func (s *BloodhoundDB) GetSavedQueriesSharedWithUser(ctx context.Context, userID int64) (model.SavedQueries, error) {
	savedQueries := model.SavedQueries{}

	result := s.db.WithContext(ctx).Where("shared_to_user_id = ?", userID).Or("global = true").Find(&savedQueries)

	return savedQueries, CheckError(result)
}

func (s *BloodhoundDB) CheckUserHasPermissionToSavedQuery(ctx context.Context, queryID int64, userID uuid.UUID) (bool, error) {
	userHasPermission := false
	result := s.db.WithContext(ctx).First(&userHasPermission, "user_id = ? AND query_id = ?", userID, queryID)

	return userHasPermission, CheckError(result)
}
