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

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/database"
)

func (s *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, req v2.GraphSchemaExtension) error {
	var (
		environments = make([]database.EnvironmentInput, len(req.Environments))
		findings     = make([]database.FindingInput, len(req.Findings))
	)

	for i, environment := range req.Environments {
		environments[i] = database.EnvironmentInput{
			EnvironmentKindName: environment.EnvironmentKind,
			SourceKindName:      environment.SourceKind,
			PrincipalKinds:      environment.PrincipalKinds,
		}
	}

	for i, finding := range req.Findings {
		findings[i] = database.FindingInput{
			Name:                 finding.Name,
			DisplayName:          finding.DisplayName,
			SourceKindName:       finding.SourceKind,
			RelationshipKindName: finding.RelationshipKind,
			EnvironmentKindName:  finding.EnvironmentKind,
			RemediationInput: database.RemediationInput{
				ShortDescription: finding.Remediation.ShortDescription,
				LongDescription:  finding.Remediation.LongDescription,
				ShortRemediation: finding.Remediation.ShortRemediation,
				LongRemediation:  finding.Remediation.LongRemediation,
			},
		}
	}

	// TODO: Temporary hardcoded value but needs to be updated to pass in the extension ID
	err := s.openGraphSchemaRepository.UpsertGraphSchemaExtension(ctx, 1, environments, findings)
	if err != nil {
		return fmt.Errorf("error upserting graph extension: %w", err)
	}

	return nil
}
