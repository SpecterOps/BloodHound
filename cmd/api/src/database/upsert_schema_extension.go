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

func (s *BloodhoundDB) UpsertGraphSchemaExtension(ctx context.Context, extensionID int32, environments []EnvironmentInput) error {
	return s.Transaction(ctx, func(tx *BloodhoundDB) error {
		for _, env := range environments {
			if err := tx.UpsertSchemaEnvironmentWithPrincipalKinds(ctx, extensionID, env.EnvironmentKindName, env.SourceKindName, env.PrincipalKinds); err != nil {
				return fmt.Errorf("failed to upsert environment with principal kinds: %w", err)
			}
		}

		return nil
	})
}
