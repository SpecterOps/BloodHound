// Copyright 2025 Specter Ops, Inc.
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
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (o *OpenGraphSchemaService) UpsertSchemaEnvironment(ctx context.Context, graphSchema model.SchemaEnvironment) error {
	if err := o.validate(ctx, graphSchema); err != nil {
		return fmt.Errorf("error validating schema environment: %w", err)
	} else if environment, err := o.openGraphSchemaRepository.GetSchemaEnvironmentById(ctx, graphSchema.ID); err != nil && !errors.Is(err, database.ErrNotFound) {
		return fmt.Errorf("error retrieving schema environment id %d: %w", graphSchema.ID, err)
	} else if !errors.Is(err, database.ErrNotFound) {
		// Environment exists - delete and recreate
		if err := o.openGraphSchemaRepository.DeleteSchemaEnvironment(ctx, environment.ID); err != nil {
			return fmt.Errorf("error deleting schema environment %d: %w", environment.ID, err)
		}
		if _, err := o.openGraphSchemaRepository.CreateSchemaEnvironment(ctx, graphSchema.SchemaExtensionId, graphSchema.EnvironmentKindId, graphSchema.SourceKindId); err != nil {
			return fmt.Errorf("error creating schema environment %d: %w", environment.ID, err)
		}
	} else {
		// Environment not found - just create
		if _, err := o.openGraphSchemaRepository.CreateSchemaEnvironment(ctx, graphSchema.SchemaExtensionId, graphSchema.EnvironmentKindId, graphSchema.SourceKindId); err != nil {
			return fmt.Errorf("error creating schema environment: %w", err)
		}
	}
	return nil
}

func (o *OpenGraphSchemaService) validate(ctx context.Context, graphSchema model.SchemaEnvironment) error {
	kinds, err := o.openGraphSchemaRepository.GetSourceKinds(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving source kinds: %w", err)
	}

	var (
		foundEnvironmentKind = false
		foundSourceKind      = false
	)

	for _, v := range kinds {
		if graphSchema.EnvironmentKindId == int32(v.ID) {
			foundEnvironmentKind = true
		}
		if graphSchema.SourceKindId == int32(v.ID) {
			foundSourceKind = true
		}
		// Early exit if both found
		if foundEnvironmentKind && foundSourceKind {
			break
		}
	}

	if !foundEnvironmentKind {
		return fmt.Errorf("invalid environment kind id %d", graphSchema.EnvironmentKindId)
	}
	if !foundSourceKind {
		return fmt.Errorf("invalid source kind id %d", graphSchema.SourceKindId)
	}

	return nil
}
