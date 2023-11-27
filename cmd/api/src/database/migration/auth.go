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
	"strings"

	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func preload(tx *gorm.DB, associationSpecs []string) *gorm.DB {
	cursor := tx
	for _, associationSpec := range associationSpecs {
		cursor = cursor.Preload(associationSpec)
	}

	return cursor
}

func getAllPermissions(tx *gorm.DB) (model.Permissions, error) {
	var existingPermissions model.Permissions
	return existingPermissions, tx.Find(&existingPermissions).Error
}

func getAllRoles(tx *gorm.DB) (model.Roles, error) {
	var roles model.Roles
	return roles, preload(tx, model.RoleAssociations()).Find(&roles).Error
}

func (s *Migrator) updatePermissions() error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if existingPermissions, err := getAllPermissions(tx); err != nil {
			return err
		} else {
			for _, expectedPermission := range auth.Permissions().All() {
				if !existingPermissions.Has(expectedPermission) {
					if result := tx.Create(&expectedPermission); result.Error != nil {
						return result.Error
					}

					log.Infof("Permission %s created during migration", expectedPermission)
				}
			}
		}

		return nil
	})
}

func (s *Migrator) checkUserEmailAddresses() error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		var users model.Users

		for _, userAssociation := range model.UserAssociations() {
			tx.Preload(userAssociation)
		}

		if result := tx.Find(&users); result.Error != nil {
			return result.Error
		} else {
			seenAddresses := make(map[string]struct{})

			for _, user := range users {
				if !user.EmailAddress.Valid || len(user.EmailAddress.String) == 0 {
					log.Errorf("UPNTE Error: user %s is missing a valid email address.", user.ID)
				} else {
					emailAddress := strings.ToLower(user.EmailAddress.String)

					if _, alreadySawAddress := seenAddresses[emailAddress]; alreadySawAddress {
						log.Errorf("UPNTE Error: user %s contains a non-unique email address.", user.ID)
					}

					seenAddresses[emailAddress] = struct{}{}
				}
			}
		}

		return nil
	})
}

func (s *Migrator) updateRoles() error {
	return s.DB.Transaction(func(tx *gorm.DB) error {
		if permissions, err := getAllPermissions(tx); err != nil {
			return err
		} else if existingRoles, err := getAllRoles(tx); err != nil {
			return err
		} else {
			for _, expectedRoleTemplate := range auth.Roles() {

				if expectedRole, err := expectedRoleTemplate.Build(permissions); err != nil {
					return err

				} else if existingRole, found := existingRoles.FindByName(expectedRole.Name); !found {
					// If cannot find by name, lookup by permissions set
					if existingRole, ok := existingRoles.FindByPermissions(expectedRole.Permissions); !ok {
						// If no Role exists w/ expectedPermissions, create new Role
						if result := tx.Create(&expectedRole); result.Error != nil {
							return result.Error
						}
						log.Infof("Role %s created during migration", expectedRole.Name)

					} else {
						// A role with the required permission set exists but has changed, update the preexisting role
						existingRole.Name = expectedRole.Name
						existingRole.Description = expectedRole.Description

						if result := s.DB.Save(existingRole); result.Error != nil {
							return result.Error
						}

						log.Infof("Role %s updated during migration", expectedRole.Name)
					}

				} else if !expectedRole.Permissions.Equals(existingRole.Permissions) || expectedRole.Description != existingRole.Description {
					// The role for the associated name has changed, update the preexisting role
					existingRole.Permissions = expectedRole.Permissions
					existingRole.Description = expectedRole.Description

					if result := s.DB.Session(&gorm.Session{FullSaveAssociations: true}).Updates(existingRole); result.Error != nil {
						return result.Error
					}

					log.Infof("Role %s updated during migration", expectedRole.Name)
				}
			}
		}

		return nil
	})
}
