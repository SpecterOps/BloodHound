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
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/database/migration"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/agi"
	"github.com/specterops/bloodhound/cmd/api/src/services/dataquality"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var (
	ErrNotFound = errors.New("entity not found")
)

var (
	ErrDuplicateAGName                           = errors.New("duplicate asset group name")
	ErrDuplicateAGTag                            = errors.New("duplicate asset group tag")
	ErrDuplicateAGTagSelectorName                = errors.New("duplicate asset group tag selector name")
	ErrDuplicateSSOProviderName                  = errors.New("duplicate sso provider name")
	ErrDuplicateUserPrincipal                    = errors.New("duplicate user principal name")
	ErrDuplicateEmail                            = errors.New("duplicate user email address")
	ErrDuplicateCustomNodeKindName               = errors.New("duplicate custom node kind name")
	ErrDuplicateKindName                         = errors.New("duplicate kind name")
	ErrDuplicateGlyph                            = errors.New("duplicate glyph")
	ErrPositionOutOfRange                        = errors.New("position out of range")
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
	upload.UploadData
	GetAllIngestTasks(ctx context.Context) (model.IngestTasks, error)
	CountAllIngestTasks(ctx context.Context) (int64, error)
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
	DeleteAssetGroupSelectorsForAssetGroups(ctx context.Context, assetGroupIds []int) error

	Wipe(ctx context.Context) error
	Migrate(ctx context.Context) error
	PopulateExtensionData(ctx context.Context) error
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

	// SSO
	SSOProviderData
	OIDCProviderData
	SAMLProviderData

	// Sessions
	CreateUserSession(ctx context.Context, userSession model.UserSession) (model.UserSession, error)
	SetUserSessionFlag(ctx context.Context, userSession *model.UserSession, key model.SessionFlagKey, state bool) error
	LookupActiveSessionsByUser(ctx context.Context, user model.User) ([]model.UserSession, error)
	EndUserSession(ctx context.Context, userSession model.UserSession)
	GetUserSession(ctx context.Context, id int64) (model.UserSession, error)
	SweepSessions(ctx context.Context)

	// Data Quality
	dataquality.DataQualityData
	GetADDataQualityStats(ctx context.Context, domainSid string, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.ADDataQualityStats, int, error)
	GetAggregateADDataQualityStats(ctx context.Context, domainSIDs []string, start time.Time, end time.Time) (model.ADDataQualityStats, error)
	GetADDataQualityAggregations(ctx context.Context, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.ADDataQualityAggregations, int, error)
	GetAzureDataQualityStats(ctx context.Context, tenantId string, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.AzureDataQualityStats, int, error)
	GetAzureDataQualityAggregations(ctx context.Context, start time.Time, end time.Time, sort_by string, limit int, skip int) (model.AzureDataQualityAggregations, int, error)
	DeleteAllDataQuality(ctx context.Context) error

	// Saved Queries
	SavedQueriesData

	// Saved Queries Permissions
	SavedQueriesPermissionsData

	// Analysis Request
	AnalysisRequestData

	// Datapipe Status
	DatapipeStatusData

	// Asset Group Tags
	AssetGroupHistoryData
	AssetGroupTagData
	AssetGroupTagSelectorData
	AssetGroupTagSelectorNodeData

	// Custom Node Kinds
	CustomNodeKindData

	// Source Kinds
	SourceKindsData

	// Environment Targeted Access Control
	EnvironmentTargetedAccessControlData

	// OpenGraph Schema
	OpenGraphSchema

	// Kind
	Kind
}

type BloodhoundDB struct {
	db         *gorm.DB
	idResolver auth.IdentityResolver // TODO: this really needs to be elsewhere. something something separation of concerns
}

func (s *BloodhoundDB) Close(ctx context.Context) {
	if sqlDBRef, err := s.db.WithContext(ctx).DB(); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed to fetch SQL DB reference from GORM: %v", err))
	} else if err := sqlDBRef.Close(); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Failed closing database: %v", err))
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

// Transaction executes the given function within a database transaction.
// The function receives a new BloodhoundDB instance backed by the transaction,
// allowing all existing methods to participate in the transaction.
// If the function returns an error, the transaction is rolled back.
// If the function returns nil, the transaction is committed.
// Optional sql.TxOptions can be provided to configure isolation level and read-only mode.
func (s *BloodhoundDB) Transaction(ctx context.Context, fn func(tx *BloodhoundDB) error, opts ...*sql.TxOptions) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewBloodhoundDB(tx, s.idResolver))
	}, opts...)
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

		if result := tx.Raw("SELECT table_name FROM information_schema.tables WHERE table_schema = CURRENT_SCHEMA() AND NOT table_name ILIKE '%pg_stat%'").Scan(&tables); result.Error != nil {
			return result.Error
		}

		for _, table := range tables {
			stmt := fmt.Sprintf(`DROP TABLE IF EXISTS "%s" CASCADE`, table)

			if err := tx.Exec(stmt).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *BloodhoundDB) Migrate(ctx context.Context) error {
	// Run the migrator
	if err := migration.NewMigrator(s.db.WithContext(ctx)).ExecuteStepwiseMigrations(); err != nil {
		slog.ErrorContext(ctx, "Error during SQL database migration phase", attr.Error(err))
		return err
	}

	return nil
}

func (s *BloodhoundDB) PopulateExtensionData(ctx context.Context) error {
	if err := migration.NewMigrator(s.db.WithContext(ctx)).ExecuteExtensionDataPopulation(); err != nil {
		slog.ErrorContext(ctx, "Error during extensions data population phase", attr.Error(err))
		return err
	}

	return nil
}
