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
)

func (o *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, req v2.GraphSchemaExtension) error {
	for _, env := range req.Environments {
		// TODO: Update temporary hardcoded extensionID once extension work is complete
		if err := o.openGraphSchemaRepository.UpsertSchemaEnvironmentWithPrincipalKinds(ctx, 1, env.EnvironmentKind, env.SourceKind, env.PrincipalKinds); err != nil {
			return fmt.Errorf("failed to upload environments with principal kinds: %w", err)
		}
	}

	return nil
}
