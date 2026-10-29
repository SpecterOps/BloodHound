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

package dataquality

import (
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/server/dataquality/internal/appdb"
	"github.com/specterops/bloodhound/server/dataquality/internal/handlers"
	"github.com/specterops/bloodhound/server/dataquality/internal/routes"
	"github.com/specterops/bloodhound/server/dataquality/internal/services"
	"github.com/specterops/bloodhound/server/featureflags"
	"github.com/specterops/bloodhound/server/opengraphschema"
)

func Register(routerInst *router.Router, pool *pgxpool.Pool) {
	var (
		store      = appdb.NewStore(pool)
		service    = services.NewService(store, opengraphschema.NewOpenGraphSchemaRequestAdapter(pool))
		handlerSet = handlers.NewHandlersContainer(service)
	)
	routes.Register(routerInst, handlerSet, featureflags.NewFeatureFlagRequestAdapter(pool))
}
