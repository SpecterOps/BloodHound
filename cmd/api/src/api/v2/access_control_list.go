// Copyright 2025 Specter Ops, Inc.
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

package v2

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

type UpdateEnvironmentRequest struct {
	EnvironmentID string `json:"environment_id"`
}

type UpdateUserEnvironmentAccessControlRequest struct {
	Environments []UpdateEnvironmentRequest `json:"environments"`
}

func CheckUserAccessToEnvironments(ctx context.Context, db database.EnvironmentAccessControlData, user model.User, environments ...string) (bool, error) {
	if user.AllEnvironments {
		return true, nil
	}

	allowedList, err := db.GetEnvironmentAccessListForUser(ctx, user)

	if err != nil {
		return false, err
	}

	allowedMap := make(map[string]struct{}, len(allowedList))
	for _, envAccess := range allowedList {
		allowedMap[envAccess.EnvironmentID] = struct{}{}
	}

	for _, env := range environments {
		_, ok := allowedMap[env]

		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// ExtractEnvironmentIDsFromUser is a helper function
// to extract a user's environments from their model as a list of strings
func ExtractEnvironmentIDsFromUser(user *model.User) []string {
	list := make([]string, 0, len(user.EnvironmentAccessControl))

	for _, envAccess := range user.EnvironmentAccessControl {
		list = append(list, envAccess.EnvironmentID)
	}

	return list
}
