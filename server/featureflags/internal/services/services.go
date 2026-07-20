// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package services

//go:generate go tool mockery

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// Feature-flag keys referenced by BHE feature slices. These mirror the keys
// defined in cmd/api/src/model/appcfg but are redeclared here so consumers do
// not need to import the appcfg package.
const (
	FeatureOpenHoundSupport = "openhound_support"
	FeatureAlerts           = "alerts"
)

// ErrNotFound indicates that no feature flag exists for the requested key.
var (
	ErrNotFound         = errors.New("feature flag not found")
	ErrNotUserUpdatable = errors.New("feature flag is not user updatable")
)

// FeatureFlag is the domain representation of a row in the feature_flags table.
type FeatureFlag struct {
	ID            int32
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Key           string
	Name          string
	Description   string
	Enabled       bool
	UserUpdatable bool
	DeletedAt     sql.NullTime
}

// TODO: once the audit log middleware is instantuated, this should be removed
func (s FeatureFlag) AuditData() model.AuditData {
	return model.AuditData{
		"name":    s.Name,
		"key":     s.Key,
		"enabled": s.Enabled,
	}
}

// Database describes the persistence capabilities the feature-flag Service
// requires. Implementations translate driver-level not-found errors into
// ErrNotFound so the Service can reason about them in domain terms. It is the
// single port through which a database implementation is injected.
type Database interface {
	GetFlagByKey(ctx context.Context, key string) (FeatureFlag, error)
	GetFlagByID(ctx context.Context, id int32) (FeatureFlag, error)
	GetAllFlags(ctx context.Context) ([]FeatureFlag, error)
	SetFlag(ctx context.Context, featureFlag FeatureFlag) error
}

// Service implements feature-flag use cases on top of a Database implementation.
type Service struct {
	db Database
}

// NewService constructs a Service from the supplied Database port. The
// PostgreSQL implementation (Store) lives alongside in sql.go so callers obtain
// a ready-to-use service without taking on a storage-layer dependency directly.
func NewService(db Database) *Service {
	if db == nil {
		panic("feature-flag: service requires a non-nil Database")
	}
	return &Service{db: db}
}

// GetFlagByKey returns the feature flag for the supplied key, or ErrNotFound
// when no flag exists.
func (s *Service) GetFlagByKey(ctx context.Context, key string) (FeatureFlag, error) {
	return s.db.GetFlagByKey(ctx, key)
}

// IsEnabled reports whether the feature flag for the supplied key is enabled.
func (s *Service) IsEnabled(ctx context.Context, key string) (bool, error) {
	flag, err := s.db.GetFlagByKey(ctx, key)
	if err != nil {
		return false, err
	}
	return flag.Enabled, nil
}

// GetAllFlags returns all the feature flags
func (s *Service) GetAllFlags(ctx context.Context) ([]FeatureFlag, error) {
	return s.db.GetAllFlags(ctx)
}

// ToggleFlag enables/disables the feature flag by the feature flag id
func (s *Service) ToggleFlag(ctx context.Context, id int32) (FeatureFlag, error) {
	flag, err := s.db.GetFlagByID(ctx, id)
	if err != nil {
		return flag, err
	}

	if !flag.UserUpdatable {
		return flag, ErrNotUserUpdatable
	}

	flag.Enabled = !flag.Enabled

	if err := s.db.SetFlag(ctx, flag); err != nil {
		return flag, err
	}

	return flag, nil
}
