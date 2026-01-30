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
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (s *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, req v2.GraphSchemaExtension) error {
	var (
		environments = make([]model.EnvironmentInput, len(req.Environments))
		findings     = make([]model.FindingInput, len(req.Findings))
	)

	for i, environment := range req.Environments {
		environments[i] = model.EnvironmentInput{
			EnvironmentKindName: environment.EnvironmentKind,
			SourceKindName:      environment.SourceKind,
			PrincipalKinds:      environment.PrincipalKinds,
		}
	}

	for i, finding := range req.Findings {
		findings[i] = model.FindingInput{
			Name:                 finding.Name,
			DisplayName:          finding.DisplayName,
			SourceKindName:       finding.SourceKind,
			RelationshipKindName: finding.RelationshipKind,
			EnvironmentKindName:  finding.EnvironmentKind,
			RemediationInput: model.RemediationInput{
				ShortDescription: finding.Remediation.ShortDescription,
				LongDescription:  finding.Remediation.LongDescription,
				ShortRemediation: finding.Remediation.ShortRemediation,
				LongRemediation:  finding.Remediation.LongRemediation,
			},
		}
	}

	_, err := s.openGraphSchemaRepository.UpsertOpenGraphExtension(ctx, model.GraphExtensionInput{
		EnvironmentsInput: environments,
		FindingsInput:     findings,
	})
	if err != nil {
		return fmt.Errorf("error upserting graph extension: %w", err)
	}

	return nil
}

func (s *OpenGraphSchemaService) ListExtensions(ctx context.Context) ([]v2.ExtensionInfo, error) {
	// Sort results by display name
	extensions, _, err := s.openGraphSchemaRepository.GetGraphSchemaExtensions(ctx, model.Filters{}, model.Sort{{Column: "display_name", Direction: model.AscendingSortDirection}}, 0, 0)
	if err != nil {
		return []v2.ExtensionInfo{}, fmt.Errorf("error retrieving graph extensions: %w", err)
	}

	apiExtensions := make([]v2.ExtensionInfo, len(extensions))

	for i, extension := range extensions {
		apiExtensions[i] = v2.ExtensionInfo{
			ID:      strconv.Itoa(int(extension.ID)),
			Name:    extension.DisplayName,
			Version: extension.Version,
		}
	}

	return apiExtensions, nil
}
