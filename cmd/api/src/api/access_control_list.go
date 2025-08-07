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

package api

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
)

func CheckUserAccessToEnvironments(ctx context.Context, db database.EnvironmentAccessControlData, userUuid uuid.UUID, environments ...string) (bool, error) {
	allowedList, err := db.GetEnvironmentAccessListForUser(ctx, userUuid)

	if err != nil {
		return false, err
	}

	allowedMap := make(map[string]struct{}, 10)
	for _, envAccess := range allowedList {
		allowedMap[envAccess.Environment] = struct{}{}
	}

	for _, env := range environments {
		_, ok := allowedMap[env]

		if !ok {
			return false, nil
		}
	}

	return true, nil
}
