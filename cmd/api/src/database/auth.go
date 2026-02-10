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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	"gorm.io/gorm"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
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

// GetAllRoles retrieves all available roles in the db
// SELECT * FROM roles
func (s *BloodhoundDB) GetAllRoles(ctx context.Context, order string, filter model.SQLFilter) (model.Roles, error) {
	var (
		roles  model.Roles
		cursor = s.preload(model.RoleAssociations()).WithContext(ctx)
	)

	if order != "" && filter.SQLString == "" {
		cursor = cursor.Order(order)
	}
	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params...)
	}

	return roles, CheckError(cursor.Find(&roles))
}

// GetRoles retrieves all rows in the Roles table corresponding to the provided list of IDs
// SELECT * FROM roles where ID in (...)
func (s *BloodhoundDB) GetRoles(ctx context.Context, ids []int32) (model.Roles, error) {
	var (
		roles  model.Roles
		result = s.preload(model.RoleAssociations()).WithContext(ctx).Where("id in ?", ids).Find(&roles)
	)

	return roles, CheckError(result)
}

// GetRole retrieves the role associated with the provided ID
// SELECT * FROM roles WHERE role_id = ....
func (s *BloodhoundDB) GetRole(ctx context.Context, id int32) (model.Role, error) {
	var (
		role   model.Role
		result = s.preload(model.RoleAssociations()).WithContext(ctx).First(&role, id)
	)

	return role, CheckError(result)
}

// GetAllPermissions retrieves all rows from the Permissions table
// SELECT * FROM permissions
func (s *BloodhoundDB) GetAllPermissions(ctx context.Context, order string, filter model.SQLFilter) (model.Permissions, error) {
	var (
		permissions model.Permissions
		cursor      = s.db.WithContext(ctx)
	)

	if order != "" {
		cursor = cursor.Order(order)
	}

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params...)
	}

	return permissions, CheckError(cursor.Find(&permissions))
}

// GetPermission retrieves a row in the Permissions table corresponding to the ID provided
// SELECT * FROM permissions WHERE permission_id = ...
func (s *BloodhoundDB) GetPermission(ctx context.Context, id int) (model.Permission, error) {
	var (
		permission model.Permission
		result     = s.db.WithContext(ctx).First(&permission, id)
	)

	return permission, CheckError(result)
}

// InitializeSecretAuth creates new AuthSecret, User and Installation entries based on the input provided
func (s *BloodhoundDB) InitializeSecretAuth(ctx context.Context, adminUser model.User, authSecret model.AuthSecret) (model.Installation, error) {
	var (
		updatedAdminUser  = adminUser
		updatedAuthSecret = authSecret
		newInstallation   model.Installation
	)

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
func (s *BloodhoundDB) CreateInstallation(ctx context.Context) (model.Installation, error) {
	if newID, err := uuid.NewV4(); err != nil {
		return model.Installation{}, err
	} else {
		installation := model.Installation{
			Unique: model.Unique{
				ID: newID,
			},
		}

		result := s.db.WithContext(ctx).Create(&installation)
		return installation, CheckError(result)
	}
}

// GetInstallation retrieves the first row from installations
// SELECT TOP 1 * FROM installations
func (s *BloodhoundDB) GetInstallation(ctx context.Context) (model.Installation, error) {
	var (
		installation model.Installation
		result       = s.db.WithContext(ctx).First(&installation)
	)

	return installation, CheckError(result)
}

// HasInstallation checks if an installation exists
// SELECT CASE WHEN EXISTS (SELECT 1 FROM installations) THEN true ELSE false END
func (s *BloodhoundDB) HasInstallation(ctx context.Context) (bool, error) {
	if _, err := s.GetInstallation(ctx); err != nil {
		if errors.Is(err, ErrNotFound) {
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
		// Ensure lowercase emails
		updatedUser.EmailAddress = null.StringFrom(strings.ToLower(updatedUser.EmailAddress.ValueOrZero()))
		updatedUser.ID = newID
	}

	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionCreateUser,
		Model:  &updatedUser,
	}
	return updatedUser, s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		result := tx.WithContext(ctx).Create(&updatedUser)

		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint \"users_principal_name_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateUserPrincipal, tx.Error)
			} else if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint \"users_email_address_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateEmail, tx.Error)
			}
		}

		return CheckError(result)
	})
}

// UpdateUser updates the roles associated with the user according to the input struct
// UPDATE users SET roles = ....
func (s *BloodhoundDB) UpdateUser(ctx context.Context, user model.User) error {
	// Ensure lowercase emails
	user.EmailAddress = null.StringFrom(strings.ToLower(user.EmailAddress.ValueOrZero()))

	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionUpdateUser,
		Model:  &user, // Pointer is required to ensure success log contains updated fields after transaction
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		// Update roles first
		if err := tx.Model(&user).WithContext(ctx).Association("Roles").Replace(&user.Roles); err != nil {
			return err
		}

		// Clear a user's etac list before applying their new one when saving the user model
		if user.AllEnvironments || user.EnvironmentTargetedAccessControl != nil {
			bhdb := NewBloodhoundDB(tx, s.idResolver)
			if err := bhdb.DeleteEnvironmentTargetedAccessControlForUser(ctx, user); err != nil {
				return fmt.Errorf("error deleting user's environment list: %w", err)
			}
		}

		// AuthSecret must be manually retrieved and deleted
		if user.AuthSecret == nil {
			var authSecret model.AuthSecret
			if err := tx.Raw("SELECT * FROM auth_secrets WHERE user_id = ?", user.ID).First(&authSecret).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			} else if authSecret.ID > 0 {
				bhdb := NewBloodhoundDB(tx, s.idResolver)
				if err := bhdb.DeleteAuthSecret(ctx, authSecret); err != nil {
					return err
				}
			}
		}

		result := tx.WithContext(ctx).Save(&user)

		if result.Error != nil {
			if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint \"users_principal_name_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateUserPrincipal, tx.Error)
			} else if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint \"users_email_address_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateEmail, tx.Error)
			}
		}

		return CheckError(result)
	})
}

func (s *BloodhoundDB) GetAllUsers(ctx context.Context, order string, filter model.SQLFilter) (model.Users, error) {
	var (
		users  model.Users
		result *gorm.DB
		cursor = s.preload(model.UserAssociations()).WithContext(ctx)
	)

	if order != "" {
		cursor = cursor.Order(order)
	}

	if filter.SQLString != "" {
		result = cursor.Where(filter.SQLString, filter.Params...).Find(&users)
	} else {
		result = cursor.Find(&users)
	}

	return users, CheckError(result)
}

// GetUser returns the user associated with the provided ID
// SELECT * FROM users WHERE id = ...
func (s *BloodhoundDB) GetUser(ctx context.Context, id uuid.UUID) (model.User, error) {
	var (
		user   model.User
		result = s.preload(model.UserAssociations()).WithContext(ctx).First(&user, id)
	)

	return user, CheckError(result)
}

// DeleteUser removes all roles for a given user, thereby revoking all permissions
// UPDATE users SET roles = nil WHERE user_id = ....
func (s *BloodhoundDB) DeleteUser(ctx context.Context, user model.User) error {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionDeleteUser,
		Model:  &user,
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		// Clear associations first
		if err := tx.Model(&user).WithContext(ctx).Association("Roles").Clear(); err != nil {
			return err
		} else if err := tx.Model(&model.IngestJob{}).
			Where("user_id = ?", user.ID).
			Updates(map[string]any{
				"user_email_address": user.EmailAddress,
				"user_id":            uuid.NullUUID{}}).
			Error; err != nil {
			return err
		}

		return CheckError(tx.WithContext(ctx).Delete(&user))
	})
}

// LookupUser retrieves the User row associated with the provided name. The name is matched against both the
// principal_name and email address fields of a user.
//
// SELECT * FROM users WHERE lower(principal_name) = ... or lower(email_address) = ...
func (s *BloodhoundDB) LookupUser(ctx context.Context, name string) (model.User, error) {
	var (
		user          model.User
		formattedName = strings.ToLower(name)
		result        = s.preload(model.UserAssociations()).WithContext(ctx).Where("principal_name = ? or lower(email_address) = ?", name, formattedName).First(&user)
	)

	return user, CheckError(result)
}

// CreateAuthToken creates a new AuthToken row using the provided struct
// INSERT INTO auth_tokens (...) VALUES (....)
func (s *BloodhoundDB) CreateAuthToken(ctx context.Context, authToken model.AuthToken) (model.AuthToken, error) {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionCreateAuthToken,
		Model:  &authToken,
	}
	// check for whether API keys are enabled
	apiKeysEnabled := appcfg.GetAPITokensParameter(ctx, s)
	if !apiKeysEnabled {
		return model.AuthToken{}, errors.New("API Keys are disabled")
	}

	return authToken, s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.WithContext(ctx).Create(&authToken))
	})
}

// UpdateAuthToken updates all fields in the AuthToken row as specified in the provided struct
// UPDATE auth_tokens SET key = ..., hmac_method = ..., last_access = ...
// WHERE user_id = ... AND client_id = ...
func (s *BloodhoundDB) UpdateAuthToken(ctx context.Context, authToken model.AuthToken) error {
	result := s.db.WithContext(ctx).Save(&authToken)
	return CheckError(result)
}

// GetAuthToken retrieves the AuthToken row associated with the provided ID
// SELECT * FROM auth_tokens WHERE id = ....
func (s *BloodhoundDB) GetAuthToken(ctx context.Context, id uuid.UUID) (model.AuthToken, error) {
	var (
		authToken model.AuthToken
		result    = s.db.WithContext(ctx).First(&authToken, id)
	)

	return authToken, CheckError(result)
}

func (s *BloodhoundDB) GetAllAuthTokens(ctx context.Context, order string, filter model.SQLFilter) (model.AuthTokens, error) {
	var (
		tokens model.AuthTokens
		cursor = s.db.WithContext(ctx)
	)

	if order != "" {
		cursor = cursor.Order(order)
	}

	if filter.SQLString != "" {
		cursor = cursor.Where(filter.SQLString, filter.Params...)
	}

	return tokens, CheckError(cursor.Find(&tokens))
}

func (s *BloodhoundDB) GetUserToken(ctx context.Context, userId, tokenId uuid.UUID) (model.AuthToken, error) {
	var (
		authToken model.AuthToken
		result    = s.db.WithContext(ctx).First(&authToken, "id = ? AND user_id = ?", tokenId, userId)
	)
	return authToken, CheckError(result)
}

// DeleteAuthToken deletes the provided AuthToken row
// DELETE FROM auth_tokens WHERE id = ...
func (s *BloodhoundDB) DeleteAuthToken(ctx context.Context, authToken model.AuthToken) error {
	return CheckError(s.db.WithContext(ctx).Where("id = ?", authToken.ID).Delete(&authToken))
}

// CreateAuthSecret creates a new AuthSecret row
// INSERT INTO auth_secrets (...) VALUES (....)
func (s *BloodhoundDB) CreateAuthSecret(ctx context.Context, authSecret model.AuthSecret) (model.AuthSecret, error) {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionCreateAuthSecret,
		Model:  &authSecret,
	}

	return authSecret, s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.WithContext(ctx).Create(&authSecret))
	})
}

// GetAuthSecret retrieves the AuthSecret row associated with the provided ID
// SELECT * FROM auth_secrets WHERE id = ....
func (s *BloodhoundDB) GetAuthSecret(ctx context.Context, id int32) (model.AuthSecret, error) {
	var (
		authSecret model.AuthSecret
		result     = s.db.WithContext(ctx).First(&authSecret, id)
	)

	return authSecret, CheckError(result)
}

// UpdateAuthSecret updates the auth secret with the input struct specified
// UPDATE auth_secrets SET digest = .., hmac_method = ..., expires_at = ...
// WHERE user_id = ....
func (s *BloodhoundDB) UpdateAuthSecret(ctx context.Context, authSecret model.AuthSecret) error {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionUpdateAuthSecret,
		Model:  &authSecret,
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.WithContext(ctx).Save(&authSecret))
	})
}

// DeleteAuthSecret deletes the auth secret row corresponding to the struct specified
// DELETE FROM auth_secrets WHERE user_id = ...
func (s *BloodhoundDB) DeleteAuthSecret(ctx context.Context, authSecret model.AuthSecret) error {
	auditEntry := model.AuditEntry{
		Action: model.AuditLogActionDeleteAuthSecret,
		Model:  &authSecret,
	}

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.WithContext(ctx).Delete(&authSecret))
	})
}

// CreateUserSession creates a new UserSession row
// INSERT INTO user_sessions (...) VALUES (..)
func (s *BloodhoundDB) CreateUserSession(ctx context.Context, userSession model.UserSession) (model.UserSession, error) {
	var newUserSession = userSession

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := CheckError(tx.Exec(`UPDATE users SET last_login = ? WHERE id = ?`, time.Now().UTC(), userSession.UserID)); err != nil {
			return err
		}

		return CheckError(tx.Create(&newUserSession))
	})

	return newUserSession, err
}

// EndUserSession terminates the provided session
// UPDATE user_sessions SET expires_at = <now> WHERE user_id = ...
func (s *BloodhoundDB) EndUserSession(ctx context.Context, userSession model.UserSession) {
	s.db.WithContext(ctx).Exec(`UPDATE user_sessions SET expires_at = NOW(), updated_at = NOW() WHERE user_id = ?`, userSession.UserID)
}

// corresponding retrival function is model.UserSession.GetFlag()
func (s *BloodhoundDB) SetUserSessionFlag(ctx context.Context, userSession *model.UserSession, key model.SessionFlagKey, state bool) error {
	if userSession.ID == 0 {
		return errors.New("invalid session - missing session id")
	}

	auditEntry := model.AuditEntry{}
	doAudit := false
	// only audit if the new state is true, meaning the EULA is currently being accepted
	// INFO: The FedEULA is only applicable to select enterprise installations
	if key == model.SessionFlagFedEULAAccepted && state {
		doAudit = true
		auditEntry.Action = model.AuditLogActionAcceptFedEULA
	}
	return s.MaybeAuditableTransaction(ctx, !doAudit, auditEntry, func(tx *gorm.DB) error {
		if userSession.Flags == nil {
			userSession.Flags = types.JSONBBoolObject{}
		}
		userSession.Flags[string(key)] = state
		return CheckError(tx.Model(&userSession).WithContext(ctx).Update("flags", userSession.Flags))
	})
}

func (s *BloodhoundDB) LookupActiveSessionsByUser(ctx context.Context, user model.User) ([]model.UserSession, error) {
	var userSessions []model.UserSession

	result := s.db.WithContext(ctx).Where("expires_at >= NOW() AND user_id = ?", user.ID).Find(&userSessions)
	return userSessions, CheckError(result)
}

// GetUserSession retrieves the UserSession row associated with the provided ID
// SELECT * FROM user_sessions WHERE id = ...
func (s *BloodhoundDB) GetUserSession(ctx context.Context, id int64) (model.UserSession, error) {
	var (
		userSession model.UserSession
		result      = s.preload(model.UserSessionAssociations()).WithContext(ctx).Find(&userSession, id)
	)

	return userSession, CheckError(result)
}

// SweepSessions deletes all sessions that have already expired
func (s *BloodhoundDB) SweepSessions(ctx context.Context) {
	s.db.WithContext(ctx).Where("expires_at < NOW()").Delete(&model.UserSession{})
}
