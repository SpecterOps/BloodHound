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
	"strconv"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (s *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, req v2.GraphSchemaExtension) error {
	var (
		environments = make([]database.EnvironmentInput, len(req.Environments))
	)

	for i, environment := range req.Environments {
		environments[i] = database.EnvironmentInput{
			EnvironmentKindName: environment.EnvironmentKind,
			SourceKindName:      environment.SourceKind,
			PrincipalKinds:      environment.PrincipalKinds,
		}
	}

	// TODO: Temporary hardcoded value but needs to be updated to pass in the extension ID
	err := s.openGraphSchemaRepository.UpsertGraphSchemaExtension(ctx, 1, environments)
	if err != nil {
		return fmt.Errorf("error upserting graph extension: %w", err)
	}

	return nil
}

func (s *OpenGraphSchemaService) GetExtensions(ctx context.Context) ([]v2.ExtensionInfo, error) {
	// Sort results by display name
	extensions, count, err := s.openGraphSchemaRepository.GetGraphSchemaExtensions(ctx, model.Filters{}, model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}}, 0, 0)
	if err != nil {
		return []v2.ExtensionInfo{}, fmt.Errorf("error retrieving graph extensions: %w", err)
	}

	apiExtensions := make([]v2.ExtensionInfo, count)

	for i, extension := range extensions {
		apiExtensions[i] = v2.ExtensionInfo{
			Id:      strconv.Itoa(int(extension.ID)),
			Name:    extension.DisplayName,
			Version: extension.Version,
		}
	}

	return apiExtensions, nil
}
