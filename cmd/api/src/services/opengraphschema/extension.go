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

package opengraphschema

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
)

// UpsertOpenGraphExtension - validates the incoming graph schema, passes it to the DB layer for upserting and if successful
// updates the in memory kinds map.
func (s *OpenGraphSchemaService) UpsertOpenGraphExtension(ctx context.Context, openGraphExtension model.GraphExtensionInput) (bool, error) {
	var (
		err          error
		schemaExists bool
	)

	if err = openGraphExtension.Validate(); err != nil {
		return schemaExists, fmt.Errorf("%w: %w", model.ErrGraphExtensionValidation, err)
	} else if schemaExists, err = s.openGraphSchemaRepository.UpsertOpenGraphExtension(ctx, openGraphExtension); err != nil {
		if model.ErrIsGraphSchemaDuplicateError(err) {
			return schemaExists, fmt.Errorf("%w: %w", model.ErrGraphExtensionValidation, err)
		}
		return schemaExists, fmt.Errorf("graph schema upsert error: %w", err)
	} else if err = s.graphDBKindRepository.RefreshKinds(ctx); err != nil {
		return schemaExists, fmt.Errorf("%w: %w", model.ErrGraphDBRefreshKinds, err)
	}
	return schemaExists, nil
}

// GetGraphSchemaExtensions retrieves extensions from the repository with filtering, sorting, and pagination
func (s *OpenGraphSchemaService) GetGraphSchemaExtensions(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error) {
	return s.openGraphSchemaRepository.GetGraphSchemaExtensions(ctx, filters, sort, skip, limit)
}

// ListExtensions returns all extensions sorted by display name
// This is a convenience wrapper around GetGraphSchemaExtensions for the list endpoint
func (s *OpenGraphSchemaService) ListExtensions(ctx context.Context) (model.GraphSchemaExtensions, error) {
	var (
		sort = model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}}
	)
	if extensions, _, err := s.GetGraphSchemaExtensions(ctx, model.Filters{}, sort, 0, 0); err != nil {
		return model.GraphSchemaExtensions{}, fmt.Errorf("error retrieving graph extensions: %w", err)
	} else {
		return extensions, nil
	}
}

func (s *OpenGraphSchemaService) DeleteExtension(ctx context.Context, extensionID int32) error {
	if err := s.openGraphSchemaRepository.DeleteGraphSchemaExtension(ctx, extensionID); err != nil {
		return fmt.Errorf("error deleting graph extension: %w", err)
	} else if err := s.graphDBKindRepository.RefreshKinds(ctx); err != nil {
		return fmt.Errorf("%w: %w", model.ErrGraphDBRefreshKinds, err)
	}

	return nil
}

// GetEnvironmentKindsAndSchemaEnvironmentData - returns all environment kinds as graph.Kinds and a map of
// their schema environments. If the findings feature flag is not enabled, it will only return builtin environment kinds.
// TODO: Remove the onlyBuiltin parameter once the appcfg.FeatureOpenGraphFindings feature flag is removed.
func (s *OpenGraphSchemaService) GetEnvironmentKindsAndSchemaEnvironmentData(ctx context.Context, onlyBuiltin bool) (graph.Kinds, model.EnvironmentKindsToEnvironment, error) {
	var filters = make(model.Filters)
	if onlyBuiltin {
		filters = model.Filters{"is_builtin": []model.Filter{{Operator: model.Equals, Value: "true", SetOperator: model.FilterAnd}}}
	}
	if environments, err := s.openGraphSchemaRepository.GetEnvironmentsFiltered(ctx, filters); err != nil {
		return nil, nil, err
	} else {
		// Build environment kind mappings
		environmentKinds := make([]graph.Kind, 0)
		envKindToEnvironment := make(model.EnvironmentKindsToEnvironment, len(environments))
		for _, env := range environments {
			environmentKinds = append(environmentKinds, graph.StringKind(env.EnvironmentKindName))
			envKindToEnvironment[env.EnvironmentKindName] = env
		}
		return environmentKinds, envKindToEnvironment, nil
	}
}

// GetSchemaFindings - returns all schema findings filtered and sorted by the given criteria.
func (s *OpenGraphSchemaService) GetSchemaFindings(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) ([]model.SchemaFinding, int, error) {
	return s.openGraphSchemaRepository.GetSchemaFindings(ctx, filters, sort, skip, limit)
}
