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

// Package opengraphschema is the wireup module for the open graph schema feature.
// It composes the store and service so other feature slices can obtain a
// ready-to-use adapter without reaching into the persistence layer.
package opengraphschema

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/appdb"
	"github.com/specterops/bloodhound/server/opengraphschema/internal/services"
	"github.com/specterops/dawgs/graph"
)

// OpenGraphSchemaRequestAdapter is the exported open graph schema capability
// other feature slices depend on. It is satisfied by the internal service
// constructed by NewOpenGraphSchemaRequestAdapter.
type OpenGraphSchemaRequestAdapter interface {
	GetEnvironmentKindsAndSchemaEnvironmentData(ctx context.Context, onlyBuiltin bool) (graph.Kinds, model.EnvironmentKindsToEnvironment, error)
}

// NewOpenGraphSchemaRequestAdapter builds a ready-to-use open graph schema
// adapter backed by the pgx pool, wiring the store and service together so
// callers obtain the service without reaching into the persistence layer.
func NewOpenGraphSchemaRequestAdapter(pool *pgxpool.Pool) OpenGraphSchemaRequestAdapter {
	return services.NewService(appdb.NewStore(pool))
}
