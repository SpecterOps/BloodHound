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
	"fmt"
)

type EnvironmentInput struct {
	EnvironmentKindName string
	SourceKindName      string
	PrincipalKinds      []string
}

type FindingInput struct {
	Name                 string
	DisplayName          string
	RelationshipKindName string
	EnvironmentKindName  string
	SourceKindName       string
	RemediationInput     RemediationInput
}

type RemediationInput struct {
	ShortDescription string
	LongDescription  string
	ShortRemediation string
	LongRemediation  string
}

func (s *BloodhoundDB) UpsertGraphSchemaExtension(ctx context.Context, extensionID int32, environments []EnvironmentInput, findings []FindingInput) error {
	return s.Transaction(ctx, func(tx *BloodhoundDB) error {
		for _, env := range environments {
			if err := tx.UpsertSchemaEnvironmentWithPrincipalKinds(ctx, extensionID, env.EnvironmentKindName, env.SourceKindName, env.PrincipalKinds); err != nil {
				return fmt.Errorf("failed to upload environment with principal kinds: %w", err)
			}
		}

		for _, finding := range findings {
			if schemaFinding, err := tx.UpsertFinding(ctx, extensionID, finding.SourceKindName, finding.RelationshipKindName, finding.EnvironmentKindName, finding.Name, finding.DisplayName); err != nil {
				return fmt.Errorf("failed to upsert finding: %w", err)
			} else {
				if err := tx.UpsertRemediation(ctx, schemaFinding.ID, finding.RemediationInput.ShortDescription, finding.RemediationInput.LongDescription, finding.RemediationInput.ShortRemediation, finding.RemediationInput.LongRemediation); err != nil {
					return fmt.Errorf("failed to upsert remediation: %w", err)
				}
			}
		}

		return nil
	})
}
