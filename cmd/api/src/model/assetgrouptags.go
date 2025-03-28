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
	"fmt"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/database/types/null"
)

const (
	AssetGroupActorSystem = "SYSTEM"
)

type SelectorType int

const (
	SelectorTypeObjectId SelectorType = 1
	SelectorTypeCypher   SelectorType = 2
)

type AssetGroupTagType int

const (
	AssetGroupTagTypeTier  AssetGroupTagType = 1
	AssetGroupTagTypeLabel AssetGroupTagType = 2
)

type AssetGroupTag struct {
	ID             int               `json:"id"`
	Type           AssetGroupTagType `json:"type"`
	KindId         int               `json:"kind_id"`
	Name           string            `json:"name"`
	Description    string            `json:"description"`
	CreatedAt      time.Time         `json:"created_at"`
	CreatedBy      string            `json:"created_by"`
	UpdatedAt      time.Time         `json:"updated_at"`
	UpdatedBy      string            `json:"updated_by"`
	DeletedAt      null.Time         `json:"deleted_at"`
	DeletedBy      null.String       `json:"deleted_by"`
	Position       null.Int32        `json:"position"`
	RequireCertify null.Bool         `json:"require_certify"`
}

func (AssetGroupTag) TableName() string {
	return "asset_group_tags"
}

func (s AssetGroupTag) AuditData() AuditData {
	return AuditData{
		"id":              s.ID,
		"type":            s.Type,
		"kind_id":         s.KindId,
		"name":            s.Name,
		"description":     s.Description,
		"position":        s.Position,
		"require_certify": s.RequireCertify,
	}
}

func (s AssetGroupTag) ToKind() graph.Kind {
	return graph.StringKind(fmt.Sprintf("Tag_%s", strings.ReplaceAll(s.Name, " ", "_")))
}

type SelectorSeed struct {
	Type  SelectorType `json:"type"`
	Value string       `json:"value"`
}

func (SelectorSeed) TableName() string {
	return "asset_group_tag_selector_seeds"
}

func (s SelectorSeed) AuditData() AuditData {
	return AuditData{
		"type":  s.Type,
		"value": s.Value,
	}
}

func (s SelectorSeed) IsString(column string) bool {
	switch column {
	case "type":
		return true
	default:
		return false
	}
}

func (s SelectorSeed) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{"type": {Equals, NotEquals}}
}

type AssetGroupTagSelectors []AssetGroupTagSelector

type AssetGroupTagSelector struct {
	ID              int         `json:"id"`
	AssetGroupTagId int         `json:"asset_group_tag_id"`
	CreatedAt       time.Time   `json:"created_at"`
	CreatedBy       string      `json:"created_by"`
	UpdatedAt       time.Time   `json:"updated_at"`
	UpdatedBy       string      `json:"updated_by"`
	DisabledAt      null.Time   `json:"disabled_at"`
	DisabledBy      null.String `json:"disabled_by"`
	Name            string      `json:"name" validate:"required"`
	Description     string      `json:"description"`
	AutoCertify     bool        `json:"auto_certify"`
	IsDefault       bool        `json:"is_default"`
	AllowDisable    bool        `json:"allow_disable"`

	Seeds []SelectorSeed `json:"seeds" validate:"required" gorm:"-"`
}

func (AssetGroupTagSelector) TableName() string {
	return "asset_group_tag_selectors"
}

func (s AssetGroupTagSelector) AuditData() AuditData {
	return AuditData{
		"id":                 s.ID,
		"asset_group_tag_id": s.AssetGroupTagId,
		"name":               s.Name,
		"description":        s.Description,
		"auto_certify":       s.AutoCertify,
		"is_default":         s.IsDefault,
	}
}

func (s AssetGroupTagSelector) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"disabled_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s AssetGroupTagSelector) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

type ListSelectorsResponse struct {
	Selectors []AssetGroupTagSelector `json:"selectors"`
}
