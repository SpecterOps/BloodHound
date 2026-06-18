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
// Package featureflags is a self-contained feature-flag library. It owns the
// feature-flag domain (the FeatureFlag type, the Database port and the Service
// in service.go), the PostgreSQL adapter (Store in sql.go) and the Register
// entry point that wires them together so callers obtain a ready-to-use service
// without reaching into the storage layer.
package featureflags

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/server/featureflags/internal/appdb"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

const (
	FeatureOpenHoundSupport = services.FeatureOpenHoundSupport
	FeatureAlerts           = services.FeatureAlerts
)

type Service interface {
	IsEnabled(ctx context.Context, key string) (bool, error)
}

// Register wires the feature-flag service to its PostgreSQL store and returns
// the constructed service for use by BHE feature slices.
func Register(pool *pgxpool.Pool) Service {
	return services.NewService(appdb.NewStore(pool))
}
