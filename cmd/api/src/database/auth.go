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

package database

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

// NewClientAuthToken creates a new Client AuthToken row using the details provided
// INSERT INTO auth_tokens (client_id, hmac_method, last_access) VALUES (...)
func NewClientAuthToken(ownerID uuid.UUID, hmacMethod string) (model.AuthToken, error) {
	var (
		authToken = model.AuthToken{
			ClientID:   NullUUID(ownerID),
			HmacMethod: hmacMethod,
			LastAccess: time.Now().UTC(),
		}

		tokenBytes = make([]byte, 40)
	)

	if hmacMethod != auth.HMAC_SHA2_256 {
		return authToken, fmt.Errorf("HMAC method %s is not supported", hmacMethod)
	}

	if id, err := uuid.NewV4(); err != nil {
		return authToken, err
	} else {
		authToken.ID = id
	}

	if _, err := rand.Read(tokenBytes); err != nil {
		return authToken, nil
	}

	authToken.Key = base64.StdEncoding.EncodeToString(tokenBytes)
	return authToken, nil
}

type AuthContextInitializer interface {
	InitContextFromToken(authToken model.AuthToken) (auth.Context, error)
}

type contextInitializer struct {
	db Database
}

func NewContextInitializer(db Database) AuthContextInitializer {
	return contextInitializer{db: db}
}

func (s contextInitializer) InitContextFromToken(authToken model.AuthToken) (auth.Context, error) {
	if authToken.UserID.Valid {
		if user, err := s.db.GetUser(authToken.UserID.UUID); err != nil {
			return auth.Context{}, err
		} else {
			return auth.Context{
				Owner: user,
			}, nil
		}
	}

	return auth.Context{}, ErrNotFound
}

func (s *BloodhoundDB) CreateRole(role model.Role) (model.Role, error) {
	var (
		updatedRole = role
		result      = s.db.Create(&updatedRole)
	)

	return updatedRole, CheckError(result)
}

// UpdateRole updates permissions for the row matching the provided Role struct
// UPDATE roles SET permissions=.... WHERE role_id = ...
func (s *BloodhoundDB) UpdateRole(role model.Role) error {
	// Update permissions first
	if err := s.db.Model(&role).Association("Permissions").Replace(&role.Permissions); err != nil {
		return err
	}

	result := s.db.Save(&role)
	return CheckError(result)
}

// GetAllRoles retrieves all available roles in the db
// SELECT * FROM roles
func (s *BloodhoundDB) GetAllRoles(order string, filter model.SQLFilter) (model.Roles, error) {
	var (
		roles  model.Roles
		result *gorm.DB
	)

	if order == "" && filter.SQLString == "" {
		result = s.preload(model.RoleAssociations()).Find(&roles)
	} else if order == "" && filter.SQLString != "" {
		result = s.preload(model.RoleAssociations()).Where(filter.SQLString, filter.Params).Find(&roles)
	} else if order != "" && filter.SQLString == "" {
		result = s.preload(model.RoleAssociations()).Order(order).Find(&roles)
	} else {
		result = s.preload(model.RoleAssociations()).Where(filter.SQLString, filter.Params).Order(order).Find(&roles)
	}

	return roles, CheckError(result)
}

// GetRoles retrieves all rows in the Roles table corresponding to the provided list of IDs
// SELECT * FROM roles where ID in (...)
func (s *BloodhoundDB) GetRoles(ids []int32) (model.Roles, error) {
	var (
		roles  model.Roles
		result = s.preload(model.RoleAssociations()).Where("id in ?", ids).Find(&roles)
	)

	return roles, CheckError(result)
}

// GetRolesByName retrieves all rows in the Roles table corresponding to the provided list of role names
// SELECT * FROM roles WHERE role_name IN (..)
func (s *BloodhoundDB) GetRolesByName(names []string) (model.Roles, error) {
	var (
		roles  model.Roles
		result = s.preload(model.RoleAssociations()).Where("name in ?", names).Find(&roles)
	)

	return roles, CheckError(result)
}

// GetRole retrieves the role associated with the provided ID
// SELECT * FROM roles WHERE role_id = ....
func (s *BloodhoundDB) GetRole(id int32) (model.Role, error) {
	var (
		role   model.Role
		result = s.preload(model.RoleAssociations()).First(&role, id)
	)

	return role, CheckError(result)
}

// LookupRoleByName retrieves a row from the Roles table corresponding to the role name provided
// SELECT * FROM roles WHERE role_name = ....
func (s *BloodhoundDB) LookupRoleByName(name string) (model.Role, error) {
	var (
		role   model.Role
		result = s.preload(model.RoleAssociations()).Where("name = ?", name).First(&role)
	)

	return role, CheckError(result)
}

// GetAllPermissions retrieves all rows from the Permissions table
// SELECT * FROM permissions
func (s *BloodhoundDB) GetAllPermissions(order string, filter model.SQLFilter) (model.Permissions, error) {
	var (
		permissions model.Permissions
		result      *gorm.DB
	)

	if order == "" && filter.SQLString == "" {
		result = s.db.Find(&permissions)
	} else if order != "" && filter.SQLString == "" {
		result = s.db.Order(order).Find(&permissions)
	} else if order == "" && filter.SQLString != "" {
		result = s.db.Where(filter.SQLString, filter.Params).Find(&permissions)
	} else {
		result = s.db.Where(filter.SQLString, filter.Params).Order(order).Find(&permissions)
	}

	return permissions, CheckError(result)
}

// GetPermission retrieves a row in the Permissions table corresponding to the ID provided
// SELECT * FROM permissions WHERE permission_id = ...
func (s *BloodhoundDB) GetPermission(id int) (model.Permission, error) {
	var (
		permission model.Permission
		result     = s.db.First(&permission, id)
	)

	return permission, CheckError(result)
}

// CreatePermission creates a new permission row with the struct provided
// INSERT INTO permissions (id, authority, name) VALUES (ID, authority, name)
func (s *BloodhoundDB) CreatePermission(permission model.Permission) (model.Permission, error) {
	var (
		updatedPermission = permission
		result            = s.db.Create(&updatedPermission)
	)

	return updatedPermission, CheckError(result)
}

// InitializeSAMLAuth creates new SAMLProvider, User and Installation entries based on the input provided
func (s *BloodhoundDB) InitializeSAMLAuth(adminUser model.User, samlProvider model.SAMLProvider) (model.SAMLProvider, model.Installation, error) {
	var (
		updatedAdminUser    = adminUser
		updatedSAMLProvider = samlProvider
		newInstallation     model.Installation
	)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if newInstallationID, err := uuid.NewV4(); err != nil {
			return err
		} else {
			newInstallation.ID = newInstallationID

			if result := tx.Create(&newInstallation); result.Error != nil {
				return CheckError(result)
			}
		}

		if result := tx.Create(&updatedSAMLProvider); result.Error != nil {
			return CheckError(result)
		}

		if newUserID, err := uuid.NewV4(); err != nil {
			return err
		} else {
			updatedAdminUser.ID = newUserID
			updatedAdminUser.SAMLProvider = &updatedSAMLProvider
			updatedAdminUser.SAMLProviderID = null.Int32From(updatedSAMLProvider.ID)

			if result := tx.Create(&updatedAdminUser); result.Error != nil {
				return CheckError(result)
			}
		}

		return nil
	})

	return updatedSAMLProvider, newInstallation, err
}

// InitializeSecretAuth creates new AuthSecret, User and Installation entries based on the input provided
func (s *BloodhoundDB) InitializeSecretAuth(adminUser model.User, authSecret model.AuthSecret) (model.Installation, error) {
	var (
		updatedAdminUser  = adminUser
		updatedAuthSecret = authSecret
		newInstallation   model.Installation
	)

	err := s.db.Transaction(func(tx *gorm.DB) error {
		if newInstallationID, err := uuid.NewV4(); err != nil {
			return err
		} else {
			newInstallation.ID = newInstallationID

			if result := tx.Create(&newInstallation); result.Error != nil {
				return CheckError(result)
			}
		}

		if newUserID, err := uuid.NewV4(); err != nil {
			return err
		} else {
			updatedAdminUser.ID = newUserID

			if result := tx.Create(&updatedAdminUser); result.Error != nil {
				return CheckError(result)
			}
		}

		updatedAuthSecret.UserID = updatedAdminUser.ID

		if result := tx.Create(&updatedAuthSecret); result.Error != nil {
			return CheckError(result)
		}

		return nil
	})

	return newInstallation, err
}

// CreateInstallation creates a new Installation row
// INSERT INTO installations(....) VALUES (...)
func (s *BloodhoundDB) CreateInstallation() (model.Installation, error) {
	if newID, err := uuid.NewV4(); err != nil {
		return model.Installation{}, err
	} else {
		installation := model.Installation{
			Unique: model.Unique{
				ID: newID,
			},
		}

		result := s.db.Create(&installation)
		return installation, CheckError(result)
	}
}

// GetInstallation retrieves the first row from installations
// SELECT TOP 1 * FROM installations
func (s *BloodhoundDB) GetInstallation() (model.Installation, error) {
	var (
		installation model.Installation
		result       = s.db.First(&installation)
	)

	return installation, CheckError(result)
}

// HasInstallation checks if an installation exists
// SELECT CASE WHEN EXISTS (SELECT 1 FROM installations) THEN true ELSE false END
func (s *BloodhoundDB) HasInstallation() (bool, error) {
	if _, err := s.GetInstallation(); err != nil {
		if err == ErrNotFound {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

// CreateUser creates a new user
// INSERT INTO users (...) VALUES (...)
func (s *BloodhoundDB) CreateUser(ctx context.Context, user model.User) (model.User, error) {
	updatedUser := user

	if newID, err := uuid.NewV4(); err != nil {
		return updatedUser, err
	} else if updatedUser.AuthSecret != nil {
		updatedUser.ID = newID
		updatedUser.AuthSecret.UserID = newID
	} else {
		updatedUser.ID = newID
	}

	auditEntry := model.AuditEntry{
		Action: "CreateUser",
		Model:  &updatedUser,
	}
	return updatedUser, s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Create(&updatedUser))
	})
}

// UpdateUser updates the roles associated with the user according to the input struct
// UPDATE users SET roles = ....
func (s *BloodhoundDB) UpdateUser(ctx context.Context, user model.User) error {
	var (
		auditEntry = model.AuditEntry{
			Action: "UpdateUser",
			Model:  &user, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		// Update roles first
		if err := tx.Model(&user).Association("Roles").Replace(&user.Roles); err != nil {
			return err
		}

		result := tx.Save(&user)
		return CheckError(result)
	})
}

func (s *BloodhoundDB) GetAllUsers(order string, filter model.SQLFilter) (model.Users, error) {
	var (
		users  model.Users
		result *gorm.DB
	)

	if order != "" && filter.SQLString == "" {
		result = s.preload(model.UserAssociations()).Order(order).Find(&users)
	} else if order != "" && filter.SQLString != "" {
		result = s.preload(model.UserAssociations()).Where(filter.SQLString, filter.Params).Order(order).Find(&users)
	} else if order == "" && filter.SQLString != "" {
		result = s.preload(model.UserAssociations()).Where(filter.SQLString, filter.Params).Find(&users)
	} else {
		result = s.preload(model.UserAssociations()).Find(&users)
	}

	return users, CheckError(result)
}

// GetUser returns the user associated with the provided ID
// SELECT * FROM users WHERE id = ...
func (s *BloodhoundDB) GetUser(id uuid.UUID) (model.User, error) {
	var (
		user   model.User
		result = s.preload(model.UserAssociations()).First(&user, id)
	)

	return user, CheckError(result)
}

// DeleteUser removes all roles for a given user, thereby revoking all permissions
// UPDATE users SET roles = nil WHERE user_id = ....
func (s *BloodhoundDB) DeleteUser(ctx context.Context, user model.User) error {
	auditEntry := model.AuditEntry{
		Action: "DeleteUser",
		Model:  &user,
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		// Clear associations first
		if err := tx.Model(&user).Association("Roles").Clear(); err != nil {
			return err
		}

		return CheckError(tx.Delete(&user))
	})
}

// LookupUser retrieves the User row associated with the provided name. The name is matched against both the
// principal_name and email address fields of a user.
//
// SELECT * FROM users WHERE lower(principal_name) = ... or lower(email_address) = ...
func (s *BloodhoundDB) LookupUser(name string) (model.User, error) {
	var (
		user          model.User
		formattedName = strings.ToLower(name)
		result        = s.preload(model.UserAssociations()).Where("principal_name = ? or lower(email_address) = ?", name, formattedName).First(&user)
	)

	return user, CheckError(result)
}

// CreateAuthToken creates a new AuthToken row using the provided struct
// INSERT INTO auth_tokens (...) VALUES (....)
func (s *BloodhoundDB) CreateAuthToken(ctx context.Context, authToken model.AuthToken) (model.AuthToken, error) {
	auditEntry := model.AuditEntry{
		Action: "CreateAuthToken",
		Model:  &authToken,
	}

	return authToken, s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Create(&authToken))
	})
}

// UpdateAuthToken updates all fields in the AuthToken row as specified in the provided struct
// UPDATE auth_tokens SET key = ..., hmac_method = ..., last_access = ...
// WHERE user_id = ... AND client_id = ...
func (s *BloodhoundDB) UpdateAuthToken(authToken model.AuthToken) error {
	result := s.db.Save(&authToken)
	return CheckError(result)
}

// GetAuthToken retrieves the AuthToken row associated with the provided ID
// SELECT * FROM auth_tokens WHERE id = ....
func (s *BloodhoundDB) GetAuthToken(id uuid.UUID) (model.AuthToken, error) {
	var (
		authToken model.AuthToken
		result    = s.db.First(&authToken, id)
	)

	return authToken, CheckError(result)
}

func (s *BloodhoundDB) GetAllAuthTokens(order string, filter model.SQLFilter) (model.AuthTokens, error) {
	var (
		tokens model.AuthTokens
		result *gorm.DB
	)

	if order != "" && filter.SQLString == "" {
		result = s.db.Order(order).Find(&tokens)
	} else if order != "" && filter.SQLString != "" {
		result = s.db.Where(filter.SQLString, filter.Params).Order(order).Find(&tokens)
	} else if order == "" && filter.SQLString != "" {
		result = s.db.Where(filter.SQLString, filter.Params).Find(&tokens)
	} else {
		result = s.db.Find(&tokens)
	}

	return tokens, CheckError(result)
}

func (s *BloodhoundDB) ListUserTokens(userID uuid.UUID, order string, filter model.SQLFilter) (model.AuthTokens, error) {
	var (
		authTokens model.AuthTokens
		result     *gorm.DB
	)

	if order != "" && filter.SQLString == "" {
		result = s.db.Where("user_id = ?", userID).Order(order).Find(&authTokens)
	} else if order == "" && filter.SQLString == "" {
		result = s.db.Where("user_id = ?", userID).Find(&authTokens)
	} else if order == "" && filter.SQLString != "" {
		result = s.db.Where("user_id = ?", userID).Where(filter.SQLString, filter.Params).Find(&authTokens)
	} else {
		result = s.db.Where("user_id = ?", userID).Where(filter.SQLString, filter.Params).Order(order).Find(&authTokens)
	}

	return authTokens, CheckError(result)
}

func (s *BloodhoundDB) GetUserToken(userId, tokenId uuid.UUID) (model.AuthToken, error) {
	var (
		authToken model.AuthToken
		result    = s.db.First(&authToken, "id = ? AND user_id = ?", tokenId, userId)
	)
	return authToken, CheckError(result)
}

// DeleteAuthToken deletes the provided AuthToken row
// DELETE FROM auth_tokens WHERE id = ...
func (s *BloodhoundDB) DeleteAuthToken(ctx context.Context, authToken model.AuthToken) error {
	auditEntry := model.AuditEntry{
		Action: "DeleteAuthToken",
		Model:  &authToken,
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Where("id = ?", authToken.ID).Delete(&authToken))
	})
}

// CreateAuthSecret creates a new AuthSecret row
// INSERT INTO auth_secrets (...) VALUES (....)
func (s *BloodhoundDB) CreateAuthSecret(ctx context.Context, authSecret model.AuthSecret) (model.AuthSecret, error) {
	auditEntry := model.AuditEntry{
		Action: "CreateAuthSecret",
		Model:  &authSecret,
	}

	return authSecret, s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Create(&authSecret))
	})
}

// GetAuthSecret retrieves the AuthSecret row associated with the provided ID
// SELECT * FROM auth_secrets WHERE id = ....
func (s *BloodhoundDB) GetAuthSecret(id int32) (model.AuthSecret, error) {
	var (
		authSecret model.AuthSecret
		result     = s.db.Find(&authSecret, id)
	)

	return authSecret, CheckError(result)
}

// UpdateAuthSecret updates the auth secret with the input struct specified
// UPDATE auth_secrets SET digest = .., hmac_method = ..., expires_at = ...
// WHERE user_id = ....
func (s *BloodhoundDB) UpdateAuthSecret(ctx context.Context, authSecret model.AuthSecret) error {
	auditEntry := model.AuditEntry{
		Action: "UpdateAuthSecret",
		Model:  &authSecret,
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Save(&authSecret))
	})
}

// DeleteAuthSecret deletes the auth secret row corresponding to the struct specified
// DELETE FROM auth_secrets WHERE user_id = ...
func (s *BloodhoundDB) DeleteAuthSecret(ctx context.Context, authSecret model.AuthSecret) error {
	auditEntry := model.AuditEntry{
		Action: "DeleteAuthSecret",
		Model:  &authSecret,
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Delete(&authSecret))
	})
}

// CreateSAMLProvider creates a new saml_providers row using the data in the input struct
// INSERT INTO saml_identity_providers (...) VALUES (...)
func (s *BloodhoundDB) CreateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider) (model.SAMLProvider, error) {
	var (
		auditEntry = model.AuditEntry{
			Action: "CreateSAMLIdentityProvider",
			Model:  &samlProvider, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Create(&samlProvider))
	})

	return samlProvider, err
}

// CreateSAMLProvider updates a saml_providers row using the data in the input struct
// UPDATE saml_identity_providers SET (...) VALUES (...) WHERE id = ...
func (s *BloodhoundDB) UpdateSAMLIdentityProvider(ctx context.Context, provider model.SAMLProvider) error {
	var (
		auditEntry = model.AuditEntry{
			Action: "UpdateSAMLIdentityProvider",
			Model:  &provider, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Save(&provider))
	})
}

// LookupSAMLProviderByName returns a SAML provider corresponding to the name provided
// SELECT * FROM saml_providers WHERE name = ....
func (s *BloodhoundDB) LookupSAMLProviderByName(name string) (model.SAMLProvider, error) {
	var (
		samlProvider model.SAMLProvider
		result       = s.db.Where("name = ?", name).Find(&samlProvider)
	)

	return samlProvider, CheckError(result)
}

// GetAllSAMLProviders returns all SAML providers
// SELECT * FROM saml_providers
func (s *BloodhoundDB) GetAllSAMLProviders() (model.SAMLProviders, error) {
	var (
		samlProviders model.SAMLProviders
		result        = s.db.Find(&samlProviders)
	)

	return samlProviders, CheckError(result)
}

// GetSAMLProvider returns a SAML provider corresponding to the ID provided
// SELECT * FOM saml_providers WHERE id = ..
func (s *BloodhoundDB) GetSAMLProvider(id int32) (model.SAMLProvider, error) {
	var (
		samlProvider model.SAMLProvider
		result       = s.db.First(&samlProvider, id)
	)

	return samlProvider, CheckError(result)
}

func (s *BloodhoundDB) DeleteSAMLProvider(ctx context.Context, provider model.SAMLProvider) error {
	var (
		auditEntry = model.AuditEntry{
			Action: "DeleteSAMLProvider",
			Model:  &provider, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Delete(&provider))
	})
}

// GetSAMLProviderUsers returns all users that are bound to the SAML provider ID provided
// SELECT * FROM users WHERE saml_provider_id = ..
func (s *BloodhoundDB) GetSAMLProviderUsers(id int32) (model.Users, error) {
	var users model.Users
	return users, CheckError(s.preload(model.UserAssociations()).Where("saml_provider_id = ?", id).Find(&users))
}

// CreateUserSession creates a new UserSession row
// INSERT INTO user_sessions (...) VALUES (..)
func (s *BloodhoundDB) CreateUserSession(userSession model.UserSession) (model.UserSession, error) {
	var (
		newUserSession = userSession
		result         = s.db.Create(&newUserSession)
	)

	return newUserSession, CheckError(result)
}

// EndUserSession terminates the provided session
// UPDATE user_sessions SET expires_at = <now> WHERE user_id = ...
func (s *BloodhoundDB) EndUserSession(userSession model.UserSession) {
	s.db.Model(&userSession).Update("expires_at", gorm.Expr("NOW()"))
}

func (s *BloodhoundDB) LookupActiveSessionsByUser(user model.User) ([]model.UserSession, error) {
	var userSessions []model.UserSession

	result := s.db.Where("expires_at >= NOW() AND user_id = ?", user.ID).Find(&userSessions)
	return userSessions, CheckError(result)
}

// GetUserSession retrieves the UserSession row associated with the provided ID
// SELECT * FROM user_sessions WHERE id = ...
func (s *BloodhoundDB) GetUserSession(id int64) (model.UserSession, error) {
	var (
		userSession model.UserSession
		result      = s.preload(model.UserSessionAssociations()).Find(&userSession, id)
	)

	return userSession, CheckError(result)
}

// SweepSessions deletes all sessions that have already expired
func (s *BloodhoundDB) SweepSessions() {
	s.db.Where("expires_at < NOW()").Delete(&model.UserSession{})
}
