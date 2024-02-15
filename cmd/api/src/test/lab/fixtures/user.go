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

package fixtures

import (
	"context"
	"fmt"
	"log"

	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

func NewUserFixture(db *lab.Fixture[*database.BloodhoundDB]) *lab.Fixture[*model.User] {

	fixture := lab.NewFixture(
		func(h *lab.Harness) (*model.User, error) {
			if tx, ok := lab.Unpack(h, db); !ok {
				return nil, fmt.Errorf("unable to unpack BloodhoundDB")
			} else if user, err := tx.CreateUser(context.Background(), model.User{}); err != nil {
				return nil, fmt.Errorf("unable to create user: %w", err)
			} else {
				return &user, nil
			}
		},
		func(h *lab.Harness, user *model.User) error {
			if tx, ok := lab.Unpack(h, PostgresFixture); !ok {
				return fmt.Errorf("unable to unpack BloodhoundDB")
			} else {
				if err := tx.DeleteUser(context.Background(), *user); err != nil {
					return fmt.Errorf("failure deleting user: %w", err)
				}
			}
			return nil
		},
	)

	if err := lab.SetDependency(fixture, db); err != nil {
		log.Fatalf("UserFixture dependency error: %v", err)
	} else if err := lab.SetDependency(fixture, BHAdminApiClientFixture); err != nil {
		log.Fatalf("UserFixture dependency error: %v", err)
	}

	return fixture
}
