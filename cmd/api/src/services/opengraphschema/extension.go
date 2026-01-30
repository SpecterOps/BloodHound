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
	"errors"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// UpsertOpenGraphExtension - validates the incoming graph schema, passes it to the DB layer for upserting and if successful
// updates the in memory kinds map.
func (o *OpenGraphSchemaService) UpsertOpenGraphExtension(ctx context.Context, openGraphExtension model.GraphExtensionInput) (bool, error) {
	var (
		err          error
		schemaExists bool
	)

	if err = validateGraphExtension(openGraphExtension); err != nil {
		return schemaExists, fmt.Errorf("%w: %w", model.ErrGraphExtensionValidation, err)
	} else if schemaExists, err = o.openGraphSchemaRepository.UpsertOpenGraphExtension(ctx, openGraphExtension); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return schemaExists, fmt.Errorf("%w: %w", model.ErrGraphExtensionValidation, err)
		}
		return schemaExists, fmt.Errorf("graph schema upsert error: %w", err)
	} else if err = o.graphDBKindRepository.RefreshKinds(ctx); err != nil {
		return schemaExists, fmt.Errorf("%w: %w", model.ErrGraphDBRefreshKinds, err)
	}
	return schemaExists, nil
}

// validateGraphExtension - Ensures the incoming model.GraphExtensionInput has an extension name, node kinds exist, and
// there are no duplicate kinds. Also ensures node and edge kinds are prepended with the extension namespace.
func validateGraphExtension(graphExtension model.GraphExtensionInput) error {
	var (
		kinds      = make(map[string]any, 0)
		properties = make(map[string]any, 0)
	)
	if graphExtension.ExtensionInput.Name == "" {
		return errors.New("graph schema extension name is required")
	} else if graphExtension.ExtensionInput.Version == "" {
		return errors.New("graph schema extension version is required")
	} else if graphExtension.ExtensionInput.Namespace == "" {
		return errors.New("graph schema extension namespace is required")
	} else if len(graphExtension.NodeKindsInput) == 0 {
		return errors.New("graph schema node kinds are required")
	}
	for _, kind := range graphExtension.NodeKindsInput {
		if !strings.HasPrefix(kind.Name, fmt.Sprintf("%s_", graphExtension.ExtensionInput.Namespace)) {
			return fmt.Errorf("graph schema kind %s is missing extension namespace prefix", kind.Name)
		}
		if _, ok := kinds[kind.Name]; ok {
			return fmt.Errorf("duplicate graph kinds: %s", kind.Name)
		}
		kinds[kind.Name] = struct{}{}
	}
	for _, kind := range graphExtension.RelationshipKindsInput {
		if !strings.HasPrefix(kind.Name, fmt.Sprintf("%s_", graphExtension.ExtensionInput.Namespace)) {
			return fmt.Errorf("graph schema edge kind %s is missing extension namespace prefix", kind.Name)
		}
		if _, ok := kinds[kind.Name]; ok {
			return fmt.Errorf("duplicate graph kinds: %s", kind.Name)
		}
		kinds[kind.Name] = struct{}{}
	}
	for _, property := range graphExtension.PropertiesInput {
		if _, ok := properties[property.Name]; ok {
			return fmt.Errorf("duplicate graph properties: %s", property.Name)
		}
		properties[property.Name] = struct{}{}
	}
	return nil
}

func (o *OpenGraphSchemaService) ListExtensions(ctx context.Context) (model.GraphSchemaExtensions, error) {
	// Sort results by display name
	extensions, _, err := o.openGraphSchemaRepository.GetGraphSchemaExtensions(ctx, model.Filters{}, model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}}, 0, 0)
	if err != nil {
		return model.GraphSchemaExtensions{}, fmt.Errorf("error retrieving graph extensions: %w", err)
	}

	return extensions, nil
}
