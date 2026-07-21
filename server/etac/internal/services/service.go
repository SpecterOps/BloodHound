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

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/server/users"
)

// EnvironmentTargetedAccessControl is the domain representation of a row in the
// environment_targeted_access_control table.
type EnvironmentTargetedAccessControl struct {
	UserID        string
	EnvironmentID string
}

// AppDatabase describes the persistence capability the ETAC helpers require. It is
// the single port through which a database implementation is injected.
type AppDatabase interface {
	GetEnvironmentTargetedAccessControlForUser(ctx context.Context, userID uuid.UUID) ([]EnvironmentTargetedAccessControl, error)
}

// Service wraps a Database and dogtags.Service to provide a higher-level ETAC
// access-check suitable for consumption by BHE feature slices. Callers do not
// need to hold a reference to dogtags or to the low-level Database interface.
type Service struct {
	appdb   AppDatabase
	dogtags dogtags.Service
}

// NewService constructs an ETAC Service from the supplied Database port and
// dogtags service. The PostgreSQL implementation (Store) lives alongside in
// sql.go so the library exposes a ready-to-use service without leaking a
// storage-layer dependency.
func NewService(appdb AppDatabase, dogtagsService dogtags.Service) *Service {
	if appdb == nil {
		panic("etac: service requires a non-nil appdb")
	} else if dogtagsService == nil {
		panic("etac: service requires a non-nil dogtagsService")
	}

	return &Service{appdb: appdb, dogtags: dogtagsService}
}

// CheckUserAccess reports whether the user may access all of the supplied
// environments. Returns true without consulting the database when ETAC
// filtering is not active for this user.
func (s *Service) CheckUserAccess(ctx context.Context, user users.User, environments ...string) (bool, error) {
	if !s.ShouldFilterForETAC(user) {
		return true, nil
	}
	return s.CheckUserAccessToEnvironments(ctx, user, environments...)
}

// ShouldFilterForETAC determines whether ETAC filtering should be applied based
// on the feature flag and the user's environment access. Filtering is skipped
// when the feature is disabled or when the user has access to all environments.
func (s *Service) ShouldFilterForETAC(user users.User) bool {
	if etacEnabled := s.dogtags.GetFlagAsBool(dogtags.ETAC_ENABLED); !etacEnabled {
		return false
	} else if user.HasAllEnvironments() {
		return false
	}

	return true
}

// checkUserAccessToEnvironments reports whether the user is permitted to access
// every environment in the supplied list. Users with access to all environments
// are always permitted.
func (s *Service) CheckUserAccessToEnvironments(ctx context.Context, user users.User, environments ...string) (bool, error) {
	if user.HasAllEnvironments() {
		return true, nil
	}

	allowedList, err := s.appdb.GetEnvironmentTargetedAccessControlForUser(ctx, user.GetID())
	if err != nil {
		return false, err
	}

	allowedMap := make(map[string]struct{}, len(allowedList))
	for _, envAccess := range allowedList {
		allowedMap[envAccess.EnvironmentID] = struct{}{}
	}

	for _, env := range environments {
		if _, ok := allowedMap[env]; !ok {
			return false, nil
		}
	}

	return true, nil
}
