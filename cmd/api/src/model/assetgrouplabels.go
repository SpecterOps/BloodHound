// Copyright 2025 Specter Ops, Inc.
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

package model

import (
	"time"

	"github.com/specterops/bloodhound/src/database/types/null"
)

type SelectorType int

const (
	SelectorTypeObjectId SelectorType = 1
	SelectorTypeCypher   SelectorType = 2
)

type AssetGroupLabel struct {
	ID               int         `json:"id"`
	AssetGroupTierId null.Int32  `json:"asset_group_tier_id"`
	KindId           int         `json:"kind_id"`
	Name             string      `json:"name"`
	Description      string      `json:"description"`
	CreatedAt        time.Time   `json:"created_at"`
	CreatedBy        string      `json:"created_by"`
	UpdatedAt        time.Time   `json:"updated_at"`
	UpdatedBy        string      `json:"updated_by"`
	DeletedAt        null.Time   `json:"deleted_at"`
	DeletedBy        null.String `json:"deleted_by"`
}

type SelectorSeed struct {
	Type  SelectorType `json:"type"`
	Value string       `json:"value"`
}

type AssetGroupLabelSelector struct {
	ID                int       `json:"id"`
	AssetGroupLabelId int       `json:"asset_group_label_id"`
	CreatedAt         time.Time `json:"created_at"`
	CreatedBy         string    `json:"created_by"`
	UpdatedAt         time.Time `json:"updated_at"`
	UpdatedBy         string    `json:"updated_by"`
	DisabledAt        null.Time `json:"disabled_at"`
	DisabledBy        string    `json:"disabled_by"`
	Name              string    `json:"name" validate:"required"`
	Description       string    `json:"description"`
	AutoCertify       bool      `json:"auto_certify"`
	IsDefault         bool      `json:"is_default"`

	Seeds []SelectorSeed `json:"seeds" validate:"required" gorm:"-"`
}
