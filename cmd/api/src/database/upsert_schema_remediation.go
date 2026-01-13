// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
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
)

// UpsertRemediation validates and upserts a remediation.
// If the remediation exists for the finding ID, it is updated. If it doesn't already exist, it is created.
// Findings information must be inserted first before inserting remediation information.
func (s *BloodhoundDB) UpsertRemediation(ctx context.Context, findingId int32, shortDescription, longDescription, shortRemediation, longRemediation string) error {
	if _, err := s.GetRemediationByFindingId(ctx, findingId); err != nil && !errors.Is(err, ErrNotFound) {
		return fmt.Errorf("error retrieving remediation by finding id '%d': %w", findingId, err)
	} else if err == nil {
		// Remediation exists - update it
		if _, err := s.UpdateRemediation(ctx, findingId, shortDescription, longDescription, shortRemediation, longRemediation); err != nil {
			return fmt.Errorf("error updating remediation by finding id '%d': %w", findingId, err)
		}
	} else {
		if _, err := s.CreateRemediation(ctx, findingId, shortDescription, longDescription, shortRemediation, longRemediation); err != nil {
			return fmt.Errorf("error creating remediation by finding id '%d': %w", findingId, err)
		}
	}
	return nil
}
