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

	"github.com/specterops/bloodhound/server/featureflags/internal/services"
)

type FeatureFlagView struct {
	ID            int32        `json:"id"`
	CreatedAt     time.Time    `json:"created_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	DeletedAt     sql.NullTime `json:"deleted_at"`
	Key           string       `json:"key"`
	Name          string       `json:"name"`
	Description   string       `json:"description"`
	Enabled       bool         `json:"enabled"`
	UserUpdatable bool         `json:"user_updatable"`
}

func BuildFeatureFlagView(tf services.FeatureFlag) FeatureFlagView {
	return FeatureFlagView{
		ID:            tf.ID,
		CreatedAt:     tf.CreatedAt,
		UpdatedAt:     tf.UpdatedAt,
		Key:           tf.Key,
		Name:          tf.Name,
		Description:   tf.Description,
		Enabled:       tf.Enabled,
		UserUpdatable: tf.UserUpdatable,
	}
}

func (s FeatureFlagView) JSONView() ([]byte, error) { return json.Marshal(s) }

type FeatureFlagsView []FeatureFlagView

func BuildFeatureFlagsView(flags []services.FeatureFlag) FeatureFlagsView {
	views := make(FeatureFlagsView, 0, len(flags))
	for _, flag := range flags {
		views = append(views, BuildFeatureFlagView(flag))
	}
	return views
}

func (s FeatureFlagsView) JSONView() ([]byte, error) { return json.Marshal(s) }
