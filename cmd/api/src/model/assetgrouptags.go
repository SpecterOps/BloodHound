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

	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/dawgs/graph"
)

const (
	AssetGroupActorSystem              = "SYSTEM"
	AssetGroupTierZeroPosition         = 1
	AssetGroupTierHygienePlaceholderId = 0
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
	AssetGroupTagTypeOwned AssetGroupTagType = 3
)

type AssetGroupCertification int

const (
	AssetGroupCertificationRevoked AssetGroupCertification = -1
	AssetGroupCertificationNone    AssetGroupCertification = 0
	AssetGroupCertificationManual  AssetGroupCertification = 1
	AssetGroupCertificationAuto    AssetGroupCertification = 2
)

type AssetGroupSelectorNodeSource int

const (
	AssetGroupSelectorNodeSourceSeed   AssetGroupSelectorNodeSource = 1
	AssetGroupSelectorNodeSourceChild  AssetGroupSelectorNodeSource = 2
	AssetGroupSelectorNodeSourceParent AssetGroupSelectorNodeSource = 3
)

type AssetGroupExpansionMethod int

const (
	AssetGroupExpansionMethodNone     AssetGroupExpansionMethod = 0
	AssetGroupExpansionMethodAll      AssetGroupExpansionMethod = 1
	AssetGroupExpansionMethodChildren AssetGroupExpansionMethod = 2
	AssetGroupExpansionMethodParents  AssetGroupExpansionMethod = 3
)

type AssetGroupTag struct {
	ID              int               `json:"id"`
	Type            AssetGroupTagType `json:"type" validate:"required"`
	KindId          int               `json:"kind_id"`
	Name            string            `json:"name" validate:"required"`
	Description     string            `json:"description"`
	CreatedAt       time.Time         `json:"created_at"`
	CreatedBy       string            `json:"created_by"`
	UpdatedAt       time.Time         `json:"updated_at"`
	UpdatedBy       string            `json:"updated_by"`
	DeletedAt       null.Time         `json:"deleted_at"`
	DeletedBy       null.String       `json:"deleted_by"`
	Position        null.Int32        `json:"position"`
	RequireCertify  null.Bool         `json:"require_certify"`
	AnalysisEnabled null.Bool         `json:"analysis_enabled"`
}

type AssetGroupTags []AssetGroupTag

func (AssetGroupTag) TableName() string {
	return "asset_group_tags"
}

func (s AssetGroupTag) AuditData() AuditData {
	return AuditData{
		"id":               s.ID,
		"type":             s.Type,
		"kind_id":          s.KindId,
		"name":             s.Name,
		"description":      s.Description,
		"position":         s.Position,
		"require_certify":  s.RequireCertify,
		"analysis_enabled": s.AnalysisEnabled,
	}
}

func (s AssetGroupTag) ToKind() graph.Kind {
	return graph.StringKind(s.KindName())
}

func (s AssetGroupTag) KindName() string {
	return fmt.Sprintf("Tag_%s", strings.ReplaceAll(s.Name, " ", "_"))
}

func (s AssetGroupTag) IsStringColumn(filter string) bool {
	return filter == "name" || filter == "description"
}

func (s AssetGroupTag) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"type":             {Equals, NotEquals},
		"name":             {Equals, NotEquals, ApproximatelyEquals},
		"description":      {Equals, NotEquals, ApproximatelyEquals},
		"created_at":       {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_by":       {Equals, NotEquals},
		"updated_at":       {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_by":       {Equals, NotEquals},
		"deleted_at":       {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_by":       {Equals, NotEquals},
		"require_certify":  {Equals, NotEquals},
		"analysis_enabled": {Equals, NotEquals},
	}
}

func (s AssetGroupTag) ToType() string {
	switch s.Type {
	case AssetGroupTagTypeTier:
		return "tier"
	case AssetGroupTagTypeLabel:
		return "label"
	case AssetGroupTagTypeOwned:
		return "owned"
	default:
		return "unknown"
	}
}

func IsValidTagType(tagType AssetGroupTagType) bool {
	return tagType == AssetGroupTagTypeTier || tagType == AssetGroupTagTypeLabel || tagType == AssetGroupTagTypeOwned
}

func (s AssetGroupTag) GetExpansionMethod() AssetGroupExpansionMethod {
	switch s.Type {
	case AssetGroupTagTypeTier:
		return AssetGroupExpansionMethodAll
	case AssetGroupTagTypeLabel:
		return AssetGroupExpansionMethodChildren
	case AssetGroupTagTypeOwned:
		return AssetGroupExpansionMethodNone
	default:
		return AssetGroupExpansionMethodNone
	}
}

type SelectorSeeds []SelectorSeed

type SelectorSeed struct {
	SelectorId int          `json:"-"`
	Type       SelectorType `json:"type"`
	Value      string       `json:"value"`
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
	AutoCertify     null.Bool   `json:"auto_certify"`
	IsDefault       bool        `json:"is_default"`
	AllowDisable    bool        `json:"allow_disable"`

	Seeds []SelectorSeed `json:"seeds,omitempty" validate:"required" gorm:"-"`
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

func (s AssetGroupTagSelector) IsStringColumn(filter string) bool {
	return filter == "name" || filter == "description"
}

func (s AssetGroupTagSelector) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"auto_certify": {Equals, NotEquals},
		"created_at":   {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_by":   {Equals, NotEquals},
		"description":  {Equals, NotEquals, ApproximatelyEquals},
		"disabled_at":  {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"disabled_by":  {Equals, NotEquals},
		"is_default":   {Equals, NotEquals},
		"name":         {Equals, NotEquals, ApproximatelyEquals},
		"updated_at":   {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_by":   {Equals, NotEquals},
	}
}

type AssetGroupSelectorNodes []AssetGroupSelectorNode

type AssetGroupSelectorNode struct {
	SelectorId  int                          `json:"selector_id"`
	NodeId      graph.ID                     `json:"node_id"`
	Certified   AssetGroupCertification      `json:"certified"`
	CertifiedBy null.String                  `json:"certified_by"`
	Source      AssetGroupSelectorNodeSource `json:"source"`
	CreatedAt   time.Time                    `json:"created_at"`
	UpdatedAt   time.Time                    `json:"updated_at"`
}

func (s AssetGroupSelectorNode) TableName() string {
	return "asset_group_tag_selector_nodes"
}
