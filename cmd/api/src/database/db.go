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
	"fmt"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/auth"
	"github.com/specterops/bloodhound/src/ctx"
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

	GetConfigurationParameter(parameter string) (appcfg.Parameter, error)
	SetConfigurationParameter(appConfig appcfg.Parameter) error
	GetAllConfigurationParameters() (appcfg.Parameters, error)
	CreateIngestTask(ingestTask model.IngestTask) (model.IngestTask, error)
	GetAllIngestTasks() (model.IngestTasks, error)
	DeleteIngestTask(ingestTask model.IngestTask) error
	GetIngestTasksForJob(jobID int64) (model.IngestTasks, error)
	GetUnfinishedIngestIDs() ([]int64, error)
	CreateAssetGroup(name, tag string, systemGroup bool) (model.AssetGroup, error)
	UpdateAssetGroup(assetGroup model.AssetGroup) error
	DeleteAssetGroup(assetGroup model.AssetGroup) error
	GetAssetGroup(id int32) (model.AssetGroup, error)
	GetAllAssetGroups(order string, filter model.SQLFilter) (model.AssetGroups, error)
	SweepAssetGroupCollections()
	GetAssetGroupCollections(assetGroupID int32, order string, filter model.SQLFilter) (model.AssetGroupCollections, error)
	GetLatestAssetGroupCollection(assetGroupID int32) (model.AssetGroupCollection, error)
	GetTimeRangedAssetGroupCollections(assetGroupID int32, from int64, to int64, order string) (model.AssetGroupCollections, error)
	GetAllAssetGroupCollections() (model.AssetGroupCollections, error)
	GetAssetGroupSelector(id int32) (model.AssetGroupSelector, error)
	UpdateAssetGroupSelector(selector model.AssetGroupSelector) error
	DeleteAssetGroupSelector(selector model.AssetGroupSelector) error
	RemoveAssetGroupSelector(selector model.AssetGroupSelector) error
	CreateRawAssetGroupSelector(assetGroup model.AssetGroup, name, selector string) (model.AssetGroupSelector, error)
	CreateAssetGroupSelector(assetGroup model.AssetGroup, spec model.AssetGroupSelectorSpec, systemSelector bool) (model.AssetGroupSelector, error)
	UpdateAssetGroupSelectors(ctx ctx.Context, assetGroup model.AssetGroup, selectorSpecs []model.AssetGroupSelectorSpec, systemSelector bool) (map[string]model.AssetGroupSelectors, error)
	GetAllAssetGroupSelectors() (model.AssetGroupSelectors, error)
	CreateAssetGroupCollection(collection model.AssetGroupCollection, entries model.AssetGroupCollectionEntries) error
	RawFirst(value any) error
	Wipe() error
	MigrateModels([]any) error
	AppendAuditLog(ctx ctx.Context, action string, data model.Auditable) error
	GetAuditLogs(offset, limit int) (model.AuditLogs, error)
	GetAuditLogsBetween(before, after time.Time, offset, limit int, order string, filter model.SQLFilter) (model.AuditLogs, error)
	GetAuditLogsCount() (int, error)
	CreateRole(role model.Role) (model.Role, error)
	UpdateRole(role model.Role) error
	GetAllRoles(order string, filter model.SQLFilter) (model.Roles, error)
	GetRoles(ids []int32) (model.Roles, error)
	GetRolesByName(names []string) (model.Roles, error)
	GetRole(id int32) (model.Role, error)
	LookupRoleByName(name string) (model.Role, error)
	GetAllPermissions(order string, filter model.SQLFilter) (model.Permissions, error)
	GetPermission(id int) (model.Permission, error)
	CreatePermission(permission model.Permission) (model.Permission, error)
	InitializeSAMLAuth(adminUser model.User, samlProvider model.SAMLProvider) (model.SAMLProvider, model.Installation, error)
	InitializeSecretAuth(adminUser model.User, authSecret model.AuthSecret) (model.Installation, error)
	CreateInstallation() (model.Installation, error)
	GetInstallation() (model.Installation, error)
	HasInstallation() (bool, error)
	CreateUser(user model.User) (model.User, error)
	UpdateUser(user model.User) error
	GetAllUsers(order string, filter model.SQLFilter) (model.Users, error)
	GetUser(id uuid.UUID) (model.User, error)
	DeleteUser(user model.User) error
	LookupUser(principalName string) (model.User, error)
	CreateAuthToken(authToken model.AuthToken) (model.AuthToken, error)
	UpdateAuthToken(authToken model.AuthToken) error
	GetAllAuthTokens(order string, filter model.SQLFilter) (model.AuthTokens, error)
	GetAuthToken(id uuid.UUID) (model.AuthToken, error)
	ListUserTokens(userID uuid.UUID, order string, filter model.SQLFilter) (model.AuthTokens, error)
	GetUserToken(userId, tokenId uuid.UUID) (model.AuthToken, error)
	DeleteAuthToken(authToken model.AuthToken) error
	CreateAuthSecret(authSecret model.AuthSecret) (model.AuthSecret, error)
	GetAuthSecret(id int32) (model.AuthSecret, error)
	UpdateAuthSecret(authSecret model.AuthSecret) error
	DeleteAuthSecret(authSecret model.AuthSecret) error
	CreateSAMLIdentityProvider(samlProvider model.SAMLProvider) (model.SAMLProvider, error)
	UpdateSAMLIdentityProvider(samlProvider model.SAMLProvider) error
	LookupSAMLProviderByName(name string) (model.SAMLProvider, error)
	GetAllSAMLProviders() (model.SAMLProviders, error)
	GetSAMLProvider(id int32) (model.SAMLProvider, error)
	GetSAMLProviderUsers(id int32) (model.Users, error)
	DeleteSAMLProvider(samlProvider model.SAMLProvider) error
	CreateUserSession(userSession model.UserSession) (model.UserSession, error)
	LookupActiveSessionsByUser(user model.User) ([]model.UserSession, error)
	EndUserSession(userSession model.UserSession)
	GetUserSession(id int64) (model.UserSession, error)
	SweepSessions()
	CreateADDataQualityStats(stats model.ADDataQualityStats) (model.ADDataQualityStats, error)
	GetADDataQualityStats(domainSid string, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.ADDataQualityStats, int, error)
	CreateADDataQualityAggregation(aggregation model.ADDataQualityAggregation) (model.ADDataQualityAggregation, error)
	GetADDataQualityAggregations(start time.Time, end time.Time, sort_by string, limit int, skip int) (model.ADDataQualityAggregations, int, error)
	CreateAzureDataQualityStats(stats model.AzureDataQualityStats) (model.AzureDataQualityStats, error)
	GetAzureDataQualityStats(tenantId string, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.AzureDataQualityStats, int, error)
	CreateAzureDataQualityAggregation(aggregation model.AzureDataQualityAggregation) (model.AzureDataQualityAggregation, error)
	GetAzureDataQualityAggregations(start time.Time, end time.Time, sort_by string, limit int, skip int) (model.AzureDataQualityAggregations, int, error)
	CreateFileUploadJob(job model.FileUploadJob) (model.FileUploadJob, error)
	UpdateFileUploadJob(job model.FileUploadJob) error
	GetFileUploadJob(id int64) (model.FileUploadJob, error)
	GetAllFileUploadJobs(skip int, limit int, order string, filter model.SQLFilter) ([]model.FileUploadJob, int, error)
	GetFileUploadJobsWithStatus(status model.JobStatus) ([]model.FileUploadJob, error)
	ListSavedQueries(userID uuid.UUID, order string, filter model.SQLFilter, skip, limit int) (model.SavedQueries, int, error)
	CreateSavedQuery(userID uuid.UUID, name string, query string) (model.SavedQuery, error)
}

type BloodhoundDB struct {
	db         *gorm.DB
	idResolver auth.IdentityResolver // TODO: this really needs to be elsewhere. something something separation of concerns
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

func (s *BloodhoundDB) RawFirst(value any) error {
	return CheckError(s.db.Model(value).First(value))
}

func (s *BloodhoundDB) RawDelete(value any) error {
	return CheckError(s.db.Delete(value))
}

func (s *BloodhoundDB) Wipe() error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		var tables []string

		if result := tx.Raw("select table_name from information_schema.tables where table_schema = current_schema()").Scan(&tables); result.Error != nil {
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

func (s *BloodhoundDB) MigrateModels(models []any) error {
	// Run the migrator
	if err := migration.NewMigrator(s.db).Migrate(models); err != nil {
		log.Errorf("Error during database migration phase: %v", err)
		return err
	}

	return nil
}
