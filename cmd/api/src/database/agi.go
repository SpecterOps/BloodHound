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

package database

import (
	"context"
	"time"

	"gorm.io/gorm"

	"github.com/specterops/bloodhound/src/ctx"
	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
)

func (s *BloodhoundDB) CreateAssetGroup(ctx context.Context, name, tag string, systemGroup bool) (model.AssetGroup, error) {
	var (
		assetGroup = model.AssetGroup{
			Name:        name,
			Tag:         tag,
			SystemGroup: systemGroup,
		}

		auditEntry = model.AuditEntry{
			Action: "CreateAssetGroup",
			Model:  &assetGroup, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return assetGroup, s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Create(&assetGroup))
	})
}

func (s *BloodhoundDB) UpdateAssetGroup(ctx context.Context, assetGroup model.AssetGroup) error {
	var (
		auditEntry = model.AuditEntry{
			Action: "UpdateAssetGroup",
			Model:  &assetGroup, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Save(&assetGroup))
	})
}

func (s *BloodhoundDB) DeleteAssetGroup(ctx context.Context, assetGroup model.AssetGroup) error {
	var (
		auditEntry = model.AuditEntry{
			Action: "DeleteAssetGroup",
			Model:  &assetGroup, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Delete(&assetGroup))
	})
}

func (s *BloodhoundDB) GetAssetGroup(id int32) (model.AssetGroup, error) {
	var (
		assetGroup model.AssetGroup
		result     = s.preload(model.AssetGroupAssociations()).First(&assetGroup, id)
	)
	return assetGroup, CheckError(result)
}

func (s *BloodhoundDB) GetAllAssetGroups(order string, filter model.SQLFilter) (model.AssetGroups, error) {
	var (
		assetGroups model.AssetGroups
		result      *gorm.DB
	)

	if order != "" && filter.SQLString == "" {
		result = s.preload(model.AssetGroupAssociations()).Order(order).Find(&assetGroups)
	} else if order != "" && filter.SQLString != "" {
		result = s.preload(model.AssetGroupAssociations()).Where(filter.SQLString, filter.Params).Order(order).Find(&assetGroups)
	} else if order == "" && filter.SQLString != "" {
		result = s.preload(model.AssetGroupAssociations()).Where(filter.SQLString, filter.Params).Find(&assetGroups)
	} else {
		result = s.preload(model.AssetGroupAssociations()).Find(&assetGroups)
	}

	if result.Error != nil {
		return assetGroups, CheckError(result)
	}

	for idx, assetGroup := range assetGroups {
		if latestCollection, err := s.GetLatestAssetGroupCollection(assetGroup.ID); err != nil {
			if err == ErrNotFound {
				assetGroup.MemberCount = 0
			} else {
				return assetGroups, err
			}
		} else {
			assetGroups[idx].MemberCount = len(latestCollection.Entries)
		}
	}

	return assetGroups, nil
}

func (s *BloodhoundDB) SweepAssetGroupCollections() {
	s.db.Where("created_at < now() - INTERVAL '30 DAYS'").Delete(&model.AssetGroupCollection{})
}

func (s *BloodhoundDB) GetAssetGroupCollections(assetGroupID int32, order string, filter model.SQLFilter) (model.AssetGroupCollections, error) {
	var (
		collections model.AssetGroupCollections
		result      *gorm.DB
	)

	if order == "" && filter.SQLString == "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).Where("asset_group_id = ?", assetGroupID).Find(&collections)
	} else if order != "" && filter.SQLString == "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).Order(order).Where("asset_group_id = ?", assetGroupID).Find(&collections)
	} else if order == "" && filter.SQLString != "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).Where("asset_group_id = ?", assetGroupID).Where(filter.SQLString, filter.Params).Find(&collections)
	} else {
		result = s.preload(model.AssetGroupCollectionAssociations()).Order(order).Where("asset_group_id = ?", assetGroupID).Where(filter.SQLString, filter.Params).Find(&collections)
	}
	return collections, CheckError(result)
}

// GetLatestAssetGroupCollection has been DEPRECATED as part of V1 and will be deleted. Use GetAllAssetGroupCollections with filters and limits instead
func (s *BloodhoundDB) GetLatestAssetGroupCollection(assetGroupID int32) (model.AssetGroupCollection, error) {
	var collection model.AssetGroupCollection

	result := s.preload(model.AssetGroupCollectionAssociations()).Where("asset_group_id = ?", assetGroupID).Last(&collection)
	return collection, CheckError(result)
}

// GetTimeRangedAssetGroupCollections has been DEPRECATED as part of V1 and will be deleted. Use GetAllAssetGroupCollections with filters instead
func (s *BloodhoundDB) GetTimeRangedAssetGroupCollections(assetGroupID int32, from int64, to int64, order string) (model.AssetGroupCollections, error) {
	var (
		collections model.AssetGroupCollections
		result      *gorm.DB
	)

	if order == "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).
			Where("asset_group_id = ? AND created_at BETWEEN ? AND ?", assetGroupID, from, to).
			Find(&collections)
	} else {
		result = s.preload(model.AssetGroupCollectionAssociations()).
			Order(order).
			Where("asset_group_id = ? AND created_at BETWEEN ? AND ?", assetGroupID, from, to).
			Find(&collections)
	}

	return collections, CheckError(result)
}

func (s *BloodhoundDB) GetAllAssetGroupCollections() (model.AssetGroupCollections, error) {
	var collections model.AssetGroupCollections

	result := s.preload(model.AssetGroupCollectionAssociations()).Find(&collections)
	return collections, CheckError(result)
}

func (s *BloodhoundDB) GetAssetGroupSelector(id int32) (model.AssetGroupSelector, error) {
	var assetGroupSelector model.AssetGroupSelector
	return assetGroupSelector, CheckError(s.db.Find(&assetGroupSelector, id))
}

func (s *BloodhoundDB) UpdateAssetGroupSelector(ctx context.Context, selector model.AssetGroupSelector) error {
	var (
		auditEntry = model.AuditEntry{
			Action: "UpdateAssetGroupSelector",
			Model:  &selector, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Save(&selector))
	})
}

func (s *BloodhoundDB) DeleteAssetGroupSelector(ctx context.Context, selector model.AssetGroupSelector) error {
	var (
		auditEntry = model.AuditEntry{
			Action: "DeleteAssetGroupSelector",
			Model:  &selector, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Delete(&selector))
	})
}

func (s *BloodhoundDB) CreateRawAssetGroupSelector(assetGroup model.AssetGroup, name, selector string) (model.AssetGroupSelector, error) {
	assetGroupSelector := model.AssetGroupSelector{
		AssetGroupID: assetGroup.ID,
		Name:         name,
		Selector:     selector,
	}

	return assetGroupSelector, CheckError(s.db.Create(&assetGroupSelector))
}

func (s *BloodhoundDB) CreateAssetGroupSelector(assetGroup model.AssetGroup, spec model.AssetGroupSelectorSpec, systemSelector bool) (model.AssetGroupSelector, error) {
	assetGroupSelector := model.AssetGroupSelector{
		AssetGroupID:   assetGroup.ID,
		Name:           spec.SelectorName,
		Selector:       spec.EntityObjectID,
		SystemSelector: systemSelector,
	}

	return assetGroupSelector, CheckError(s.db.Create(&assetGroupSelector))
}

func (s *BloodhoundDB) UpdateAssetGroupSelectors(ctx ctx.Context, assetGroup model.AssetGroup, selectorSpecs []model.AssetGroupSelectorSpec, systemSelector bool) (model.UpdatedAssetGroupSelectors, error) {
	var updatedSelectors = model.UpdatedAssetGroupSelectors{}

	err := s.db.Transaction(func(tx *gorm.DB) error {
		for _, selectorSpec := range selectorSpecs {
			switch selectorSpec.Action {
			case model.SelectorSpecActionAdd:
				assetGroupSelector := model.AssetGroupSelector{
					AssetGroupID:   assetGroup.ID,
					Name:           selectorSpec.SelectorName,
					Selector:       selectorSpec.EntityObjectID,
					SystemSelector: systemSelector,
				}

				if selectorsMatched := tx.Where("asset_group_id=? AND name=?", assetGroup.ID, selectorSpec.SelectorName).Find(&model.AssetGroupSelector{}).RowsAffected; selectorsMatched == 0 {
					// create a new db entry only if it doesn't exist, otherwise continue execution
					if result := tx.Create(&assetGroupSelector); result.Error != nil {
						return CheckError(result)
					}
				}

				updatedSelectors.Added = append(updatedSelectors.Added, assetGroupSelector)

			case model.SelectorSpecActionRemove:
				if result := tx.Where("asset_group_id=? AND name=?", assetGroup.ID, selectorSpec.SelectorName).Delete(&model.AssetGroupSelector{}); result.Error != nil {
					return CheckError(result)
				} else {
					updatedSelectors.Removed = append(updatedSelectors.Removed, model.AssetGroupSelector{
						AssetGroupID: assetGroup.ID,
						Name:         selectorSpec.SelectorName,
						Selector:     selectorSpec.EntityObjectID,
					})
				}
			}
		}

		return nil
	})

	return updatedSelectors, err
}

func (s *BloodhoundDB) GetAllAssetGroupSelectors() (model.AssetGroupSelectors, error) {
	var assetGroupSelectors model.AssetGroupSelectors

	return assetGroupSelectors, CheckError(s.db.Find(&assetGroupSelectors))
}

func (s *BloodhoundDB) CreateAssetGroupCollection(collection model.AssetGroupCollection, entries model.AssetGroupCollectionEntries) error {
	const CreateAssetGroupCollectionQuery = `INSERT INTO "asset_group_collection_entries"
    ("asset_group_collection_id","object_id","node_label","properties","created_at","updated_at")
	(SELECT * FROM unnest($1::bigint[], $2::text[], $3::text[], $4::jsonb[], $5::timestamp[], $5::timestamp[]));`

	return s.db.Transaction(func(tx *gorm.DB) error {
		var newCollection = collection

		if result := tx.Create(&newCollection); result.Error != nil {
			return CheckError(result)
		}

		// GORM will fail on an attempt to insert a nil slice, so we have to guard against empty entry arrays here
		if len(entries) > 0 {
			var (
				agIds      = make([]int64, len(entries))
				objectIds  = make([]string, len(entries))
				labels     = make([]string, len(entries))
				properties = make([]types.JSONUntypedObject, len(entries))
				timestamps = make([]time.Time, len(entries))
				now        = time.Now()
			)

			for idx := range entries {
				agIds[idx] = newCollection.ID
				objectIds[idx] = entries[idx].ObjectID
				labels[idx] = entries[idx].NodeLabel
				properties[idx] = entries[idx].Properties
				timestamps[idx] = now
			}

			return CheckError(tx.Exec(CreateAssetGroupCollectionQuery, agIds, objectIds, labels, properties, timestamps))
		}

		return nil
	})
}
