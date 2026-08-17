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
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/featureflags/internal/appdb"
	"github.com/specterops/bloodhound/server/featureflags/internal/handlers"
	"github.com/specterops/bloodhound/server/featureflags/internal/routes"
	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

const (
	FeatureOpenHoundSupport = services.FeatureOpenHoundSupport
	FeatureAlerts           = services.FeatureAlerts
)

type FeatureFlagRequestAdapter interface {
	IsEnabled(ctx context.Context, key string) (bool, error)
}

func NewFeatureFlagRequestAdapter(pool *pgxpool.Pool) FeatureFlagRequestAdapter {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	return handlerSet
}

func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store      = appdb.NewStore(pool)
		svc        = services.NewService(store)
		handlerSet = handlers.NewHandlersContainer(svc)
	)

	routes.Register(routerInst, handlerSet)
}
