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

package database

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// UpsertOpenGraphExtension - compares then upserts the incoming model.GraphExtension with the one stored in
// the BloodHoundDB.
//
// During development, it was decided to push the upsert logic down to the database layer due to difficulties of
// decoupling the database and service layers while still providing transactional guarantees. The following
// functions use models intended for the service layer and call the database public methods directly, rather
// than using an interface.
func (s *BloodhoundDB) UpsertOpenGraphExtension(ctx context.Context, graphExtension model.GraphExtension) (bool, error) {
	var (
		err                    error
		schemaExists           bool
		existingGraphExtension model.GraphExtension

		tx                      = s.db.WithContext(ctx).Begin()
		bloodhoundDBTransaction = BloodhoundDB{db: tx}
	)
	// Check for an immediate error after beginning the transaction
	if err = tx.Error; err != nil {
		return false, err
	}

	defer func() {
		tx.Rollback() // rollback is a no-op if the tx has already been committed
	}()

	if existingGraphExtension, err = s.GetGraphExtensionByName(ctx, graphExtension.GraphSchemaExtension.Name); err != nil {
		if !errors.Is(err, ErrNotFound) {
			return schemaExists, err
		} else {
			// extension does not exist so create extension
			if graphExtension.GraphSchemaExtension, err = s.CreateGraphSchemaExtension(ctx, graphExtension.GraphSchemaExtension.Name,
				graphExtension.GraphSchemaExtension.DisplayName, graphExtension.GraphSchemaExtension.Version, graphExtension.GraphSchemaExtension.Namespace); err != nil {
				return schemaExists, err
			}
		}
	} else {
		// extension exists, transfer model.Serial and update
		schemaExists = true
		if existingGraphExtension.GraphSchemaExtension.IsBuiltin {
			return schemaExists, model.GraphExtensionBuiltInError
		}
		graphExtension.GraphSchemaExtension.Serial = existingGraphExtension.GraphSchemaExtension.Serial
		if graphExtension.GraphSchemaExtension, err = s.UpdateGraphSchemaExtension(ctx, graphExtension.GraphSchemaExtension); err != nil {
			return schemaExists, err
		}
	}

	// Ensure existing environments and findings are deleted prior to upsert
	if existingGraphExtension.GraphEnvironments != nil && len(existingGraphExtension.GraphEnvironments) > 0 {
		for _, env := range existingGraphExtension.GraphEnvironments {
			if err = bloodhoundDBTransaction.DeleteEnvironment(ctx, env.ID); err != nil {
				return schemaExists, err
			}
		}
	}
	if existingGraphExtension.GraphFindings != nil && len(existingGraphExtension.GraphFindings) > 0 {
		for _, finding := range existingGraphExtension.GraphFindings {
			if err = bloodhoundDBTransaction.DeleteSchemaRelationshipFinding(ctx, finding.ID); err != nil {
				return schemaExists, err
			}
		}
	}

	// explicitly assign nodes, properties and edges their extension id
	for idx := range graphExtension.GraphSchemaNodeKinds {
		graphExtension.GraphSchemaNodeKinds[idx].SchemaExtensionId = graphExtension.GraphSchemaExtension.ID
	}
	for idx := range graphExtension.GraphSchemaEdgeKinds {
		graphExtension.GraphSchemaEdgeKinds[idx].SchemaExtensionId = graphExtension.GraphSchemaExtension.ID
	}
	for idx := range graphExtension.GraphSchemaProperties {
		graphExtension.GraphSchemaProperties[idx].SchemaExtensionId = graphExtension.GraphSchemaExtension.ID
	}

	if graphExtension.GraphSchemaNodeKinds, err = bloodhoundDBTransaction.UpsertGraphSchemaNodeKinds(ctx,
		graphExtension.GraphSchemaNodeKinds, existingGraphExtension.GraphSchemaNodeKinds); err != nil {
		return false, fmt.Errorf("failed to upsert node kinds: %w", err)
	} else if graphExtension.GraphSchemaEdgeKinds, err = bloodhoundDBTransaction.UpsertGraphSchemaEdgeKinds(ctx,
		graphExtension.GraphSchemaEdgeKinds, existingGraphExtension.GraphSchemaEdgeKinds); err != nil {
		return false, fmt.Errorf("failed to upsert edge kinds: %w", err)
	} else if graphExtension.GraphSchemaProperties, err = bloodhoundDBTransaction.UpsertGraphSchemaProperties(ctx,
		graphExtension.GraphSchemaProperties, existingGraphExtension.GraphSchemaProperties); err != nil {
		return false, fmt.Errorf("failed to upsert properties: %w", err)
	} else if err = bloodhoundDBTransaction.upsertGraphEnvironments(ctx, graphExtension.GraphSchemaExtension.ID,
		graphExtension.GraphEnvironments); err != nil {
		return false, err
	} else if err = bloodhoundDBTransaction.upsertFindingsAndRemediations(ctx, graphExtension.GraphSchemaExtension.ID,
		graphExtension.GraphFindings); err != nil {
		return false, err
	}

	if err = tx.Commit().Error; err != nil {
		return false, err
	} else {
		return schemaExists, nil
	}
}

// GetGraphExtensionByName - returns the service layer model.GraphExtension for the provided graph extension name.
func (s *BloodhoundDB) GetGraphExtensionByName(ctx context.Context, graphExtensionName string) (model.GraphExtension, error) {
	var (
		graphExtension = model.GraphExtension{
			GraphSchemaExtension:  model.GraphSchemaExtension{},
			GraphSchemaProperties: make(model.GraphSchemaProperties, 0),
			GraphSchemaEdgeKinds:  make(model.GraphSchemaEdgeKinds, 0),
			GraphSchemaNodeKinds:  make(model.GraphSchemaNodeKinds, 0),
			GraphEnvironments:     make(model.GraphEnvironments, 0),
		}
	)

	if extensions, totalRecords, err := s.GetGraphSchemaExtensions(ctx,
		model.Filters{"name": []model.Filter{{ // check to see if extension exists
			Operator:    model.Equals,
			Value:       graphExtensionName,
			SetOperator: model.FilterAnd,
		}}}, model.Sort{}, 0, 1); err != nil && !errors.Is(err, ErrNotFound) {
		return graphExtension, err
	} else if totalRecords == 0 || errors.Is(err, ErrNotFound) {
		return graphExtension, ErrNotFound
	} else {
		graphExtension.GraphSchemaExtension = extensions[0]
		var (
			schemaIdFilter = model.Filter{
				Operator:    model.Equals,
				Value:       strconv.FormatInt(int64(graphExtension.GraphSchemaExtension.ID), 10),
				SetOperator: model.FilterAnd,
			}
		)
		if graphExtension.GraphSchemaNodeKinds, _, err = s.GetGraphSchemaNodeKinds(ctx,
			model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, ErrNotFound) {
			return graphExtension, err
		} else if graphExtension.GraphSchemaEdgeKinds, _, err = s.GetGraphSchemaEdgeKinds(ctx,
			model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, ErrNotFound) {
			return graphExtension, err
		} else if graphExtension.GraphSchemaProperties, _, err = s.GetGraphSchemaProperties(ctx,
			model.Filters{"schema_extension_id": []model.Filter{schemaIdFilter}}, model.Sort{}, 0, 0); err != nil && !errors.Is(err, ErrNotFound) {
			return graphExtension, err
		} else if graphExtension.GraphEnvironments, err = s.GetGraphEnvironmentsByGraphSchemaExtensionId(ctx,
			graphExtension.GraphSchemaExtension.ID); err != nil && !errors.Is(err, ErrNotFound) {
			return graphExtension, err
		} else if graphExtension.GraphFindings, err = s.GetGraphFindingsBySchemaExtensionId(ctx, graphExtension.GraphSchemaExtension.ID); err != nil && !errors.Is(err, ErrNotFound) {
			return graphExtension, err
		}
	}

	return graphExtension, nil
}
