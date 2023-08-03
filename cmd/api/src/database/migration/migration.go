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

package migration

import (
	"embed"
	"fmt"

	"gorm.io/gorm"
)

//go:embed migrations

var migrations embed.FS

type Migrator struct {
	migrations embed.FS
	db         *gorm.DB
}

func NewMigrator(db *gorm.DB) *Migrator {
	return &Migrator{
		migrations: migrations,
		db:         db,
	}
}

func (s *Migrator) Migrate(models []any) error {
	if err := s.executeStepwiseMigrations(models); err != nil {
		return fmt.Errorf("failed to execute stepwise migrations: %w", err)
	}

	if err := s.cleanupIngest(); err != nil {
		return err
	}

	if err := s.updatePermissions(); err != nil {
		return err
	}

	if err := s.updateRoles(); err != nil {
		return err
	}

	if err := s.updateAssetGroups(); err != nil {
		return err
	}

	if err := s.setAppConfigDefaults(); err != nil {
		return err
	}

	if err := s.checkUserEmailAddresses(); err != nil {
		return err
	}

	return nil
}
