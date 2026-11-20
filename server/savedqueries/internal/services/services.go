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

//go:generate go tool mockery

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var ErrSavedQueryNotFound = errors.New("saved query not found")

type Actor struct {
	UserID          string
	IsAdministrator bool
}

type SavedQuery struct {
	ID                int64
	UserID            string
	Name              string
	Query             string
	Description       string
	Category          string
	SchemaExtensionID *int32
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         sql.NullTime
}

type SavedQueryAccess struct {
	PublicOrShared bool
}

type Database interface {
	GetSavedQuery(ctx context.Context, savedQueryID int64, actorUserID string) (SavedQuery, SavedQueryAccess, error)
}

type Service struct {
	database Database
}

func NewService(databaseInterface Database) *Service {
	return &Service{database: databaseInterface}
}

func (s *Service) GetSavedQuery(ctx context.Context, savedQueryID int64, actor Actor) (SavedQuery, error) {
	savedQuery, access, err := s.database.GetSavedQuery(ctx, savedQueryID, actor.UserID)
	if err != nil {
		return SavedQuery{}, err
	}

	if savedQuery.UserID != actor.UserID && !actor.IsAdministrator && !access.PublicOrShared {
		return SavedQuery{}, ErrSavedQueryNotFound
	}

	return savedQuery, nil
}
