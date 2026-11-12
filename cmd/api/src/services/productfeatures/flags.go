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

// Package productfeatures defines product entitlement keys and their defaults.
// Feature evaluation and provider management belong to the OpenFeature SDK.
package productfeatures

import (
	"context"

	"github.com/open-feature/go-sdk/openfeature"
)

const (
	PZ_MULTI_TIER_ANALYSIS = "privilege_zones.multi_tier_analysis"
	PZ_TIER_LIMIT          = "privilege_zones.tier_limit"
	PZ_LABEL_LIMIT         = "privilege_zones.label_limit"
	ETAC_ENABLED           = "auth.environment_targeted_access_control"
)

// AllValues evaluates the known product entitlements for the legacy dog-tags
// endpoint. OpenFeature providers do not expose flag enumeration.
func AllValues(ctx context.Context, client *openfeature.Client) map[string]any {
	return map[string]any{
		PZ_MULTI_TIER_ANALYSIS: client.Boolean(ctx, PZ_MULTI_TIER_ANALYSIS, false, openfeature.EvaluationContext{}),
		PZ_TIER_LIMIT:          client.Int(ctx, PZ_TIER_LIMIT, 1, openfeature.EvaluationContext{}),
		PZ_LABEL_LIMIT:         client.Int(ctx, PZ_LABEL_LIMIT, 0, openfeature.EvaluationContext{}),
		ETAC_ENABLED:           client.Boolean(ctx, ETAC_ENABLED, false, openfeature.EvaluationContext{}),
	}
}
