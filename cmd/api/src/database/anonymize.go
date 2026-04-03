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
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (s *BloodhoundDB) GetAnonymizeTranslationEntries(ctx context.Context) ([]model.AnonymizeTranslationEntry, error) {
	var entries []model.AnonymizeTranslationEntry
	result := s.db.WithContext(ctx).Find(&entries)
	return entries, result.Error
}

func (s *BloodhoundDB) SaveAnonymizeTranslationEntries(ctx context.Context, entries []model.AnonymizeTranslationEntry) error {
	now := time.Now()
	for i := range entries {
		entries[i].CreatedAt = now
	}
	return s.db.WithContext(ctx).CreateInBatches(entries, 500).Error
}

func (s *BloodhoundDB) DeleteAnonymizeTranslationEntries(ctx context.Context) error {
	return s.db.WithContext(ctx).Where("1=1").Delete(&model.AnonymizeTranslationEntry{}).Error
}

func (s *BloodhoundDB) SearchAnonymizeTranslationEntries(ctx context.Context, query string) ([]model.AnonymizeTranslationEntry, error) {
	var entries []model.AnonymizeTranslationEntry
	likeQuery := "%" + query + "%"
	result := s.db.WithContext(ctx).
		Where("original_value ILIKE ? OR anonymized_value ILIKE ?", likeQuery, likeQuery).
		Limit(100).
		Find(&entries)
	return entries, result.Error
}
