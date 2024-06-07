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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../LICENSE.header -destination=./mocks/db.go -package=mocks . Database

import (
	"context"
	"fmt"
	"github.com/specterops/bloodhound/src/services/agi"
	"github.com/specterops/bloodhound/src/services/dataquality"
	"github.com/specterops/bloodhound/src/services/fileupload"
	"github.com/specterops/bloodhound/src/services/ingest"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/database/migration"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	ErrNotFound = errors.Error("entity not found")
)

func IsUnexpectedDatabaseError(err error) bool {
	return err != nil && err != ErrNotFound
}

// Database describes the old interface for communicating with the application database
//
// Deprecated: When writing code in the new structure, do not pass this interface. Instead, create an interface containing
// the methods you wish to use in your service implementation:
// https://specterops.atlassian.net/wiki/spaces/BE/pages/194412923/Restructure+API+Endpoints+Guide+RFC?atlOrigin=eyJpIjoiZjhkOGI0ZDFlMjEzNDkwMDlkMzRhM2QxYTRjMzlmODYiLCJwIjoiY29uZmx1ZW5jZS1jaGF0cy1pbnQifQ
type Database interface {
	appcfg.ParameterService
	appcfg.FeatureFlagService

	Close(ctx context.Context)

	// Ingest
	ingest.IngestData
	GetAllIngestTasks(ctx context.Context) (model.IngestTasks, error)
	DeleteIngestTask(ctx context.Context, ingestTask model.IngestTask) error
	GetIngestTasksForJob(ctx context.Context, jobID int64) (model.IngestTasks, error)

	// Asset Groups
	agi.AgiData
	CreateAssetGroup(ctx context.Context, name, tag string, systemGroup bool) (model.AssetGroup, error)
	UpdateAssetGroup(ctx context.Context, assetGroup model.AssetGroup) error
	DeleteAssetGroup(ctx context.Context, assetGroup model.AssetGroup) error
	SweepAssetGroupCollections(ctx context.Context)
	GetAssetGroupCollections(ctx context.Context, assetGroupID int32, order string, filter model.SQLFilter) (model.AssetGroupCollections, error)
	GetLatestAssetGroupCollection(ctx context.Context, assetGroupID int32) (model.AssetGroupCollection, error)
	GetTimeRangedAssetGroupCollections(ctx context.Context, assetGroupID int32, from int64, to int64, order string) (model.AssetGroupCollections, error)
	GetAssetGroupSelector(ctx context.Context, id int32) (model.AssetGroupSelector, error)
	DeleteAssetGroupSelector(ctx context.Context, selector model.AssetGroupSelector) error
	UpdateAssetGroupSelectors(ctx context.Context, assetGroup model.AssetGroup, selectorSpecs []model.AssetGroupSelectorSpec, systemSelector bool) (model.UpdatedAssetGroupSelectors, error)

	Wipe(ctx context.Context) error
	Migrate(ctx context.Context) error
	RequiresMigration(ctx context.Context) (bool, error)
	CreateInstallation(ctx context.Context) (model.Installation, error)
	GetInstallation(ctx context.Context) (model.Installation, error)
	HasInstallation(ctx context.Context) (bool, error)

	// Audit Logs
	CreateAuditLog(ctx context.Context, auditLog model.AuditLog) error
	AppendAuditLog(ctx context.Context, entry model.AuditEntry) error
	ListAuditLogs(ctx context.Context, before, after time.Time, offset, limit int, order string, filter model.SQLFilter) (model.AuditLogs, int, error)

	// Roles
	GetAllRoles(ctx context.Context, order string, filter model.SQLFilter) (model.Roles, error)
	GetRoles(ctx context.Context, ids []int32) (model.Roles, error)
	GetRole(ctx context.Context, id int32) (model.Role, error)

	// Permissions
	GetAllPermissions(ctx context.Context, order string, filter model.SQLFilter) (model.Permissions, error)
	GetPermission(ctx context.Context, id int) (model.Permission, error)

	// Users
	CreateUser(ctx context.Context, user model.User) (model.User, error)
	UpdateUser(ctx context.Context, user model.User) error
	GetAllUsers(ctx context.Context, order string, filter model.SQLFilter) (model.Users, error)
	GetUser(ctx context.Context, id uuid.UUID) (model.User, error)
	DeleteUser(ctx context.Context, user model.User) error
	LookupUser(ctx context.Context, principalName string) (model.User, error)

	// Auth
	CreateAuthToken(ctx context.Context, authToken model.AuthToken) (model.AuthToken, error)
	UpdateAuthToken(ctx context.Context, authToken model.AuthToken) error
	GetAllAuthTokens(ctx context.Context, order string, filter model.SQLFilter) (model.AuthTokens, error)
	GetAuthToken(ctx context.Context, id uuid.UUID) (model.AuthToken, error)
	GetUserToken(ctx context.Context, userId, tokenId uuid.UUID) (model.AuthToken, error)
	DeleteAuthToken(ctx context.Context, authToken model.AuthToken) error
	CreateAuthSecret(ctx context.Context, authSecret model.AuthSecret) (model.AuthSecret, error)
	GetAuthSecret(ctx context.Context, id int32) (model.AuthSecret, error)
	UpdateAuthSecret(ctx context.Context, authSecret model.AuthSecret) error
	DeleteAuthSecret(ctx context.Context, authSecret model.AuthSecret) error
	InitializeSecretAuth(ctx context.Context, adminUser model.User, authSecret model.AuthSecret) (model.Installation, error)

	// SAML
	CreateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider) (model.SAMLProvider, error)
	UpdateSAMLIdentityProvider(ctx context.Context, samlProvider model.SAMLProvider) error
	LookupSAMLProviderByName(ctx context.Context, name string) (model.SAMLProvider, error)
	GetAllSAMLProviders(ctx context.Context) (model.SAMLProviders, error)
	GetSAMLProvider(ctx context.Context, id int32) (model.SAMLProvider, error)
	GetSAMLProviderUsers(ctx context.Context, id int32) (model.Users, error)
	DeleteSAMLProvider(ctx context.Context, samlProvider model.SAMLProvider) error

	// Sessions
	CreateUserSession(ctx context.Context, userSession model.UserSession) (model.UserSession, error)
	LookupActiveSessionsByUser(ctx context.Context, user model.User) ([]model.UserSession, error)
	EndUserSession(ctx context.Context, userSession model.UserSession)
	GetUserSession(ctx context.Context, id int64) (model.UserSession, error)
	SweepSessions(ctx context.Context)

	// Data Quality
	dataquality.DataQualityData
	GetADDataQualityStats(ctx context.Context, domainSid string, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.ADDataQualityStats, int, error)
	GetADDataQualityAggregations(ctx context.Context, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.ADDataQualityAggregations, int, error)
	GetAzureDataQualityStats(ctx context.Context, tenantId string, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.AzureDataQualityStats, int, error)
	GetAzureDataQualityAggregations(ctx context.Context, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.AzureDataQualityAggregations, int, error)
	DeleteAllDataQuality(ctx context.Context) error

	// File Upload
	fileupload.FileUploadData

	// Saved Queries
	ListSavedQueries(ctx context.Context, userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) (model.SavedQueries, int, error)
	CreateSavedQuery(ctx context.Context, userID uuid.UUID, name string, query string) (model.SavedQuery, error)
	DeleteSavedQuery(ctx context.Context, id int) error
	SavedQueryBelongsToUser(ctx context.Context, userID uuid.UUID, savedQueryID int) (bool, error)
	DeleteAssetGroupSelectorsForAssetGroups(ctx context.Context, assetGroupIds []int) error

	// Analysis Request
	AnalysisRequestData
}

type BloodhoundDB struct {
	db         *gorm.DB
	idResolver auth.IdentityResolver // TODO: this really needs to be elsewhere. something something separation of concerns
}

func (s *BloodhoundDB) Close(ctx context.Context) {
	if sqlDBRef, err := s.db.WithContext(ctx).DB(); err != nil {
		log.Errorf("Failed to fetch SQL DB reference from GORM: %v", err)
	} else if err := sqlDBRef.Close(); err != nil {
		log.Errorf("Failed closing database: %v", err)
	}
}

func (s *BloodhoundDB) preload(associations []string) *gorm.DB {
	cursor := s.db
	for _, association := range associations {
		cursor = cursor.Preload(association)
	}

	return cursor
}

func (s *BloodhoundDB) Scope(scopeFuncs ...ScopeFunc) *gorm.DB {
	scopes := make([]func(*gorm.DB) *gorm.DB, len(scopeFuncs))
	for idx, scopeFunc := range scopeFuncs {
		scopes[idx] = scopeFunc
	}

	return s.db.Scopes(scopes...)
}

func NewBloodhoundDB(db *gorm.DB, idResolver auth.IdentityResolver) *BloodhoundDB {
	return &BloodhoundDB{db: db, idResolver: idResolver}
}

func OpenDatabase(connection string) (*gorm.DB, error) {
	gormConfig := &gorm.Config{
		Logger: &GormLogAdapter{
			SlowQueryErrorThreshold: time.Second * 10,
			SlowQueryWarnThreshold:  time.Second * 1,
		},
	}

	if db, err := gorm.Open(postgres.Open(connection), gormConfig); err != nil {
		return nil, err
	} else {
		return db, nil
	}
}

func (s *BloodhoundDB) RawDelete(value any) error {
	return CheckError(s.db.Delete(value))
}

func (s *BloodhoundDB) Wipe(ctx context.Context) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var tables []string

		if result := tx.Raw("select table_name from information_schema.tables where table_schema = current_schema() and not table_name ilike '%pg_stat%'").Scan(&tables); result.Error != nil {
			return result.Error
		}

		for _, table := range tables {
			stmt := fmt.Sprintf(`drop table if exists "%s" cascade`, table)

			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *BloodhoundDB) RequiresMigration(ctx context.Context) (bool, error) {
	return migration.NewMigrator(s.db.WithContext(ctx)).RequiresMigration()
}

func (s *BloodhoundDB) Migrate(ctx context.Context) error {
	// Run the migrator
	if err := migration.NewMigrator(s.db.WithContext(ctx)).Migrate(); err != nil {
		log.Errorf("Error during SQL database migration phase: %v", err)
		return err
	}

	return nil
}
