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

// validateGraphExtension validates the incoming model.GraphExtensionInput for the
// following fields: name, version, namespace.
// It also ensures node kinds exist and there are no duplicate kinds. Additionally, it
// validates that node and edge kinds must be prepended with the extension namespace.
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
	for _, environment := range graphExtension.EnvironmentsInput {
		if !strings.HasPrefix(environment.EnvironmentKindName, fmt.Sprintf("%s_", graphExtension.ExtensionInput.Namespace)) {
			return fmt.Errorf("graph schema environment kind %s is missing extension namespace prefix", environment.EnvironmentKindName)
		}
		if _, ok := kinds[environment.EnvironmentKindName]; !ok {
			return fmt.Errorf("graph schema environment %s not declared as a node kind", environment.EnvironmentKindName)
		}
		if environment.SourceKindName == "" {
			return fmt.Errorf("graph schema environment source kind cannot be empty")
		}
		if _, ok := kinds[environment.SourceKindName]; ok {
			return fmt.Errorf("graph schema environment source kind %s should not be declared as a node or relationship kind", environment.SourceKindName)
		}
		for _, principalKind := range environment.PrincipalKinds {
			if !strings.HasPrefix(principalKind, fmt.Sprintf("%s_", graphExtension.ExtensionInput.Namespace)) {
				return fmt.Errorf("graph schema environment principal kind %s is missing extension namespace prefix", principalKind)
			}
		}
	}
	for _, finding := range graphExtension.FindingsInput {
		if !strings.HasPrefix(finding.Name, fmt.Sprintf("%s_", graphExtension.ExtensionInput.Namespace)) {
			return fmt.Errorf("graph schema finding %s is missing extension namespace prefix", finding.Name)
		}
		if !strings.HasPrefix(finding.EnvironmentKindName, fmt.Sprintf("%s_", graphExtension.ExtensionInput.Namespace)) {
			return fmt.Errorf("graph schema finding environment kind %s is missing extension namespace prefix", finding.EnvironmentKindName)
		}
		if !strings.HasPrefix(finding.RelationshipKindName, fmt.Sprintf("%s_", graphExtension.ExtensionInput.Namespace)) {
			return fmt.Errorf("graph schema finding relationship kind %s is missing extension namespace prefix", finding.RelationshipKindName)
		}
		if _, ok := kinds[finding.EnvironmentKindName]; !ok {
			return fmt.Errorf("graph schema finding environment kind %s not declared as a node kind", finding.EnvironmentKindName)
		}
		if _, ok := kinds[finding.RelationshipKindName]; !ok {
			return fmt.Errorf("graph schema finding relationship kind %s not declared as a relationship kind", finding.RelationshipKindName)
		}
		if finding.SourceKindName == "" {
			return fmt.Errorf("graph schema finding source kind cannot be empty")
		}
		if _, ok := kinds[finding.SourceKindName]; ok {
			return fmt.Errorf("graph schema finding source kind %s should not be declared as a node or relationship kind", finding.SourceKindName)
		}
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
