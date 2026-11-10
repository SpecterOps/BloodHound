// Copyright 2026 Specter Ops, Inc.
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

package appdb

import (
	"context"
	"fmt"

	"github.com/huandu/go-sqlbuilder"
)

const (
	TableUserSessions = "user_sessions"

	userSessionColumnID        = "id"
	userSessionColumnExpiresAt = "expires_at"
	userSessionColumnUpdatedAt = "updated_at"
)

// Logout ends the user session identified by sessionID by expiring it immediately.
func (s *Store) Logout(ctx context.Context, sessionID int64) error {
	updateBuilder := sqlbuilder.PostgreSQL.NewUpdateBuilder()
	updateBuilder.Update(TableUserSessions)
	updateBuilder.Set(
		updateBuilder.Assign(userSessionColumnExpiresAt, sqlbuilder.Raw(sqlCurrentTimestamp)),
		updateBuilder.Assign(userSessionColumnUpdatedAt, sqlbuilder.Raw(sqlCurrentTimestamp)),
	)
	updateBuilder.Where(updateBuilder.Equal(userSessionColumnID, sessionID))

	sqlQuery, args := updateBuilder.Build()

	if _, err := s.db.Exec(ctx, sqlQuery, args...); err != nil {
		return fmt.Errorf("ending user session: %s", err)
	}

	return nil
}
