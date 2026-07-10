// Copyright 2026 Specter Ops, Inc.
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
// Package etac is a self-contained Environment Targeted Access Control (ETAC)
// library. It owns the ETAC domain (the User and EnvironmentTargetedAccessControl
// types, the Database port and the Service in service.go), the PostgreSQL
// adapter (Store in sql.go) and the Register entry point that wires them
// together so callers obtain a ready-to-use service without reaching into the
// storage layer.
package etac

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/bloodhound/server/etac/internal/appdb"
	"github.com/specterops/bloodhound/server/etac/internal/services"
	"github.com/specterops/bloodhound/server/users"
)

type Service interface {
	CheckUserAccess(ctx context.Context, user users.User, environments ...string) (bool, error)
}

// Register wires the ETAC service to its PostgreSQL store and the supplied
// dogtags service, returning the constructed service for use by BHE feature
// slices.
func Register(pool *pgxpool.Pool, dogtagsService dogtags.Service) Service {
	return services.NewService(appdb.NewStore(pool), dogtagsService)
}
