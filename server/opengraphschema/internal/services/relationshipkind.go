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
package services

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// ErrRelationshipKindNotFound indicates that no relationship kind exists for the requested id.
var ErrRelationshipKindNotFound = errors.New("relationship kind not found")

type RelationshipKind struct {
	ID            int32
	KindID        int32
	Name          string
	Description   string
	IsTraversable bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
	Info          []KindInfo
	Extension     Extension
}

func (s *Service) GetRelationshipKind(ctx context.Context, id int32) (RelationshipKind, error) {
	if relKind, err := s.db.GetRelationshipKind(ctx, id); err != nil {
		return RelationshipKind{}, err
	} else {
		if relKind.Info, err = s.db.GetKindInfos(ctx, relKind.Name); err != nil {
			return RelationshipKind{}, fmt.Errorf("fetching kind infos for relationship kind %s: %w", relKind.Name, err)
		}

		if relKind.Extension.ID != 0 {
			if relKind.Extension, err = s.db.GetExtension(ctx, relKind.Extension.ID); err != nil && !errors.Is(err, ErrExtensionNotFound) {
				return RelationshipKind{}, fmt.Errorf("fetching extension %d for relationship kind %d: %w", relKind.Extension.ID, id, err)
			}
		}

		return relKind, nil
	}
}
