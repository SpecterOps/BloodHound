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

package routes

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/api/router"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/dataquality/internal/handlers"
)

type FeatureFlags interface {
	IsEnabled(ctx context.Context, key string) (bool, error)
}

func Register(routerInst *router.Router, handlerSet *handlers.Handlers, featureFlags FeatureFlags) {
	routerInst.GET("/api/v2/data-quality-stats-aggregations", handlerSet.ListAggregations).
		RequirePermissions(auth.Permissions().GraphDBRead).
		CheckFeatureFlag(featureFlags, appcfg.FeatureOpenGraphExtensionManagement).
		WithSort(handlers.AggregationsView{}).
		WithPaging(params.PagingConfig{DefaultLimit: 1000})
}
