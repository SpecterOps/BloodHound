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

package services

import (
	"context"
	"errors"
	"strconv"
)

// TierHygienePlaceholderID is the sentinel tag id used to request tiering-agnostic
// (hygiene) findings when no concrete asset group tag is supplied.
const TierHygienePlaceholderID = 0

// AssetGroupTag is the domain representation of an asset group tag. Only the fields
// consumed by callers of this slice are surfaced; the persistence layer maps its rows
// into this type at the store boundary.
type AssetGroupTag struct {
	ID int
}

// ErrAssetGroupTagNotFound indicates that no asset group tag exists for the requested id.
var ErrAssetGroupTagNotFound = errors.New("asset group tag not found")

// ErrTierZeroTagNotFound indicates that no tier zero asset group tag could be resolved.
var ErrTierZeroTagNotFound = errors.New("tier zero asset group tag not found")

// GetTierZeroTag returns the tier zero asset group tag, or ErrTierZeroTagNotFound when
// none is configured.
func (s *Service) GetTierZeroTag(ctx context.Context) (AssetGroupTag, error) {
	return s.db.GetTierZeroTag(ctx)
}

// ResolveTagIDsWithFallback resolves the asset group tag ids to query for. When
// maybeAssetGroupTagID is supplied it is validated against the tag table (except for the
// hygiene placeholder id, which is passed through). When it is empty the tier zero tag and
// the hygiene placeholder are returned as the default set.
func (s *Service) ResolveTagIDsWithFallback(ctx context.Context, maybeAssetGroupTagID string) ([]int, error) {
	var tagIDs []int

	if maybeAssetGroupTagID != "" {
		if tagID, err := strconv.Atoi(maybeAssetGroupTagID); err != nil {
			return tagIDs, err
		} else if tagID == TierHygienePlaceholderID {
			// This is a workaround to supply tiering agnostic findings
			tagIDs = append(tagIDs, TierHygienePlaceholderID)
			return tagIDs, nil
		} else if _, err = s.db.GetAssetGroupTagByID(ctx, tagID); err != nil {
			return tagIDs, err
		} else {
			tagIDs = append(tagIDs, tagID)
			return tagIDs, nil
		}
	}

	// Fallback to tier zero and hygiene if not supplied
	if tierZeroTag, err := s.db.GetTierZeroTag(ctx); err != nil {
		return tagIDs, err
	} else {
		// We need both the hygiene placeholder and the tier zero asset group tag id
		tagIDs = append(tagIDs, TierHygienePlaceholderID, tierZeroTag.ID)
		return tagIDs, nil
	}
}
