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

package handlers

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/specterops/bloodhound/server/savedqueries/internal/services"
)

type SavedQueryView struct {
	UserID            string       `json:"user_id"`
	Name              string       `json:"name"`
	Query             string       `json:"query"`
	Description       string       `json:"description"`
	Category          string       `json:"category,omitempty"`
	SchemaExtensionID *int32       `json:"extension_id,omitempty"`
	ID                int64        `json:"id"`
	CreatedAt         time.Time    `json:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at"`
	DeletedAt         sql.NullTime `json:"deleted_at"`
}

func BuildSavedQueryView(savedQuery services.SavedQuery) SavedQueryView {
	return SavedQueryView{
		UserID:            savedQuery.UserID,
		Name:              savedQuery.Name,
		Query:             savedQuery.Query,
		Description:       savedQuery.Description,
		Category:          savedQuery.Category,
		SchemaExtensionID: savedQuery.SchemaExtensionID,
		ID:                savedQuery.ID,
		CreatedAt:         savedQuery.CreatedAt,
		UpdatedAt:         savedQuery.UpdatedAt,
		DeletedAt:         savedQuery.DeletedAt,
	}
}

func (s SavedQueryView) JSONView() ([]byte, error) {
	return json.Marshal(s)
}
