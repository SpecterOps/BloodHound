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

	"github.com/gofrs/uuid"
)

// MaxUserPreferencesBytes bounds the opaque UI storage. Keep this in sync with
// the user_preferences_ui_storage_size database constraint and the API contract.
const MaxUserPreferencesBytes = 64 * 1024

var (
	ErrUserNotFound            = errors.New("user not found")
	ErrUserPreferencesTooLarge = errors.New("user preferences exceed 65536 bytes")
)

// GetUserPreferences returns opaque UI storage, or empty bytes for a user who
// has not saved preferences. It never creates a preference row.
func (s *Service) GetUserPreferences(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	return s.db.GetUserPreferences(ctx, userID)
}

// UpsertUserPreferences replaces the entire UI storage without decoding it.
func (s *Service) UpsertUserPreferences(ctx context.Context, userID uuid.UUID, storage []byte) error {
	if len(storage) > MaxUserPreferencesBytes {
		return ErrUserPreferencesTooLarge
	}
	return s.db.UpsertUserPreferences(ctx, userID, storage)
}
