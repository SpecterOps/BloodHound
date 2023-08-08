// Copyright 2023 Specter Ops, Inc.
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

	"github.com/specterops/bloodhound/src/database/types"
)

type AssetGroupSelector struct {
	AssetGroupID   int32  `json:"asset_group_id"`
	Name           string `json:"name"`
	Selector       string `json:"selector"`
	SystemSelector bool   `json:"system_selector"`

	Serial
}

func (s AssetGroupSelector) AuditData() AuditData {
	return AuditData{
		"name":     s.Name,
		"selector": s.Selector,
	}
}

type AssetGroupSelectors []AssetGroupSelector

func (s AssetGroupSelectors) Strings() []string {
	selectorStrings := make([]string, len(s))

	for idx := 0; idx < len(s); idx++ {
		selectorStrings[idx] = s[idx].Selector
	}

	return selectorStrings
}

// AssetGroupAssociations returns a list of AssetGroup model associations to load eagerly by default with GORM
// Preload(...). Note: this does not include the "Collections" association on-purpose since this collection grows
// over time and may require additional parameters for fetching.
func AssetGroupAssociations() []string {
	return []string{
		"Selectors",
	}
}

type AssetGroup struct {
	Name        string                `json:"name"`
	Tag         string                `json:"tag"`
	SystemGroup bool                  `json:"system_group"`
	Selectors   AssetGroupSelectors   `gorm:"constraint:OnDelete:CASCADE;"`
	Collections AssetGroupCollections `json:"-" gorm:"constraint:OnDelete:CASCADE;"`
	MemberCount int                   `json:"member_count" gorm:"-"`

	Serial
}

func (s AssetGroup) AuditData() AuditData {
	return AuditData{
		"asset_group_id":   s.ID,
		"asset_group_name": s.Name,
		"asset_group_tag":  s.Tag,
	}
}

type AssetGroups []AssetGroup

func (s AssetGroups) IsSortable(column string) bool {
	switch column {
	case "name",
		"tag",
		"member_count":
		return true
	default:
		return false
	}
}

func (s AssetGroups) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"name":         {Equals, NotEquals},
		"tag":          {Equals, NotEquals},
		"system_group": {Equals, NotEquals},
		"member_count": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"id":           {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at":   {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at":   {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_at":   {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s AssetGroups) IsString(column string) bool {
	switch column {
	case "name",
		"tag",
		"system_group":
		return true
	default:
		return false
	}
}

func (s AssetGroups) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s AssetGroups) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

func (s AssetGroups) FindByName(name string) (AssetGroup, bool) {
	for _, group := range s {
		if group.Name == name {
			return group, true
		}
	}

	return AssetGroup{}, false
}

// AssetGroupCollectionAssociations returns a list of AssetGroupCollection model associations to eagerly by default
// with GORM Preload(...).
func AssetGroupCollectionAssociations() []string {
	return []string{"Entries"}
}

type AssetGroupCollection struct {
	AssetGroupID int32                       `json:"-"`
	Entries      AssetGroupCollectionEntries `json:"entries" gorm:"constraint:OnDelete:CASCADE;"`

	BigSerial
}

type AssetGroupCollections []AssetGroupCollection

func (s AssetGroupCollections) IsSortable(column string) bool {
	switch column {
	case "id",
		"created_at",
		"updated_at",
		"deleted_at":
		return true
	default:
		return false
	}
}

func (s AssetGroupCollections) ValidFilters() map[string][]FilterOperator {
	return map[string][]FilterOperator{
		"id":         {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"created_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"updated_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
		"deleted_at": {Equals, GreaterThan, GreaterThanOrEquals, LessThan, LessThanOrEquals, NotEquals},
	}
}

func (s AssetGroupCollections) GetFilterableColumns() []string {
	var columns = make([]string, 0)
	for column := range s.ValidFilters() {
		columns = append(columns, column)
	}
	return columns
}

func (s AssetGroupCollections) GetValidFilterPredicatesAsStrings(column string) ([]string, error) {
	if predicates, validColumn := s.ValidFilters()[column]; !validColumn {
		return []string{}, fmt.Errorf("the specified column cannot be filtered")
	} else {
		var stringPredicates = make([]string, 0)
		for _, predicate := range predicates {
			stringPredicates = append(stringPredicates, string(predicate))
		}
		return stringPredicates, nil
	}
}

type AssetGroupCollectionEntry struct {
	AssetGroupCollectionID int64                   `json:"-"`
	ObjectID               string                  `json:"object_id"`
	NodeLabel              string                  `json:"node_label"`
	Properties             types.JSONUntypedObject `json:"properties"`

	BigSerial
}

type AssetGroupCollectionEntries []AssetGroupCollectionEntry

type AssetGroupSelectorSpec struct {
	SelectorName   string `json:"selector_name"`
	EntityObjectID string `json:"sid"`
	Action         string `json:"action"`
}

const (
	SelectorSpecActionAdd    = "add"
	SelectorSpecActionRemove = "remove"
)

func (s AssetGroupSelectorSpec) Validate() error {
	return nil
}

func (s AssetGroupSelectorSpec) AuditData() AuditData {
	return AuditData{
		"selector_name":             s.SelectorName,
		"selector_entity_object_id": s.EntityObjectID,
	}
}
