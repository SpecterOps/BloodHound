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
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/specterops/bloodhound/src/database/types"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

func (s *BloodhoundDB) CreateAssetGroup(ctx context.Context, name, tag string, systemGroup bool) (model.AssetGroup, error) {
	var (
		assetGroup = model.AssetGroup{
			Name:        name,
			Tag:         tag,
			SystemGroup: systemGroup,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateAssetGroup,
			Model:  &assetGroup, // Pointer is required to ensure success log contains updated fields after transaction
		}

		err error
	)

	err = s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		err := tx.Create(&assetGroup).Error
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"asset_groups_name_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateAGName, err)
			} else if strings.Contains(err.Error(), "duplicate key value violates unique constraint \"asset_groups_tag_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateAGTag, err)
			}
		}
		return err
	})

	return assetGroup, err
}

func (s *BloodhoundDB) UpdateAssetGroup(ctx context.Context, assetGroup model.AssetGroup) error {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionUpdateAssetGroup,
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
			Action: model.AuditLogActionDeleteAssetGroup,
			Model:  &assetGroup, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Delete(&assetGroup))
	})
}

func (s *BloodhoundDB) GetAssetGroup(ctx context.Context, id int32) (model.AssetGroup, error) {
	var assetGroup model.AssetGroup
	if result := s.preload(model.AssetGroupAssociations()).WithContext(ctx).First(&assetGroup, id); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return assetGroup, ErrNotFound
	} else if result.Error != nil {
		return assetGroup, result.Error
	} else if latestCollection, collectionErr := s.GetLatestAssetGroupCollection(ctx, id); errors.Is(collectionErr, ErrNotFound) {
		assetGroup.MemberCount = 0
		return assetGroup, nil
	} else if collectionErr != nil {
		return assetGroup, fmt.Errorf("error getting latest collection for asset group %s: %w", assetGroup.Name, collectionErr)
	} else {
		assetGroup.MemberCount = len(latestCollection.Entries)
		return assetGroup, nil
	}
}

func (s *BloodhoundDB) GetAllAssetGroups(ctx context.Context, order string, filter model.SQLFilter) (model.AssetGroups, error) {
	var (
		assetGroups model.AssetGroups
		result      = s.preload(model.AssetGroupAssociations()).WithContext(ctx)
	)

	if order != "" {
		result = result.Order(order)
	}

	if filter.SQLString != "" {
		result = result.Where(filter.SQLString, filter.Params...)
	}

	if result = result.Find(&assetGroups); errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return assetGroups, ErrNotFound
	} else if result.Error != nil {
		return assetGroups, result.Error
	}

	for idx := range assetGroups {
		if latestCollection, collectionErr := s.GetLatestAssetGroupCollection(ctx, assetGroups[idx].ID); errors.Is(collectionErr, ErrNotFound) {
			assetGroups[idx].MemberCount = 0
		} else if collectionErr != nil {
			return assetGroups, fmt.Errorf("error getting latest collection for asset group %s: %w", assetGroups[idx].Name, collectionErr)
		} else {
			assetGroups[idx].MemberCount = len(latestCollection.Entries)
		}
	}
	return assetGroups, nil
}

func (s *BloodhoundDB) SweepAssetGroupCollections(ctx context.Context) {
	s.db.WithContext(ctx).Where("created_at < now() - INTERVAL '30 DAYS'").Delete(&model.AssetGroupCollection{})
}

func (s *BloodhoundDB) GetAssetGroupCollections(ctx context.Context, assetGroupID int32, order string, filter model.SQLFilter) (model.AssetGroupCollections, error) {
	var (
		collections model.AssetGroupCollections
		result      *gorm.DB
	)

	if order == "" && filter.SQLString == "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).WithContext(ctx).Where("asset_group_id = ?", assetGroupID).Find(&collections)
	} else if order != "" && filter.SQLString == "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).WithContext(ctx).Order(order).Where("asset_group_id = ?", assetGroupID).Find(&collections)
	} else if order == "" && filter.SQLString != "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).WithContext(ctx).Where("asset_group_id = ?", assetGroupID).Where(filter.SQLString, filter.Params...).Find(&collections)
	} else {
		result = s.preload(model.AssetGroupCollectionAssociations()).WithContext(ctx).Order(order).Where("asset_group_id = ?", assetGroupID).Where(filter.SQLString, filter.Params...).Find(&collections)
	}
	return collections, CheckError(result)
}

func (s *BloodhoundDB) GetLatestAssetGroupCollection(ctx context.Context, assetGroupID int32) (model.AssetGroupCollection, error) {
	var (
		latestCollection model.AssetGroupCollection
		result           = s.preload(model.AssetGroupCollectionAssociations()).
					WithContext(ctx).
					Where("asset_group_id = ?", assetGroupID).
					Order("created_at DESC").
					First(&latestCollection)
	)

	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return latestCollection, ErrNotFound
	}
	return latestCollection, result.Error
}

func (s *BloodhoundDB) GetTimeRangedAssetGroupCollections(ctx context.Context, assetGroupID int32, from int64, to int64, order string) (model.AssetGroupCollections, error) {
	var (
		collections model.AssetGroupCollections
		result      *gorm.DB
	)

	if order == "" {
		result = s.preload(model.AssetGroupCollectionAssociations()).WithContext(ctx).
			Where("asset_group_id = ? AND created_at BETWEEN ? AND ?", assetGroupID, from, to).
			Find(&collections)
	} else {
		result = s.preload(model.AssetGroupCollectionAssociations()).WithContext(ctx).
			Order(order).
			Where("asset_group_id = ? AND created_at BETWEEN ? AND ?", assetGroupID, from, to).
			Find(&collections)
	}

	return collections, CheckError(result)
}

func (s *BloodhoundDB) GetAssetGroupSelector(ctx context.Context, id int32) (model.AssetGroupSelector, error) {
	var assetGroupSelector model.AssetGroupSelector
	return assetGroupSelector, CheckError(s.db.WithContext(ctx).Find(&assetGroupSelector, id))
}

func (s *BloodhoundDB) DeleteAssetGroupSelector(ctx context.Context, selector model.AssetGroupSelector) error {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionDeleteAssetGroupSelector,
			Model:  &selector, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	return s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		return CheckError(tx.Delete(&selector))
	})
}

func (s *BloodhoundDB) DeleteAssetGroupSelectorsForAssetGroups(ctx context.Context, assetGroupIds []int) error {
	return CheckError(
		s.db.WithContext(ctx).Where("asset_group_id IN ?", assetGroupIds).
			Delete(&model.AssetGroupSelector{}),
	)
}

func (s *BloodhoundDB) UpdateAssetGroupSelectors(ctx context.Context, assetGroup model.AssetGroup, selectorSpecs []model.AssetGroupSelectorSpec, systemSelector bool) (model.UpdatedAssetGroupSelectors, error) {
	var updatedSelectors = model.UpdatedAssetGroupSelectors{}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func (s *BloodhoundDB) CreateAssetGroupCollection(ctx context.Context, collection model.AssetGroupCollection, entries model.AssetGroupCollectionEntries) error {
	const CreateAssetGroupCollectionQuery = `INSERT INTO "asset_group_collection_entries"
    ("asset_group_collection_id","object_id","node_label","properties","created_at","updated_at")
	(SELECT * FROM unnest($1::bigint[], $2::text[], $3::text[], $4::jsonb[], $5::timestamp[], $5::timestamp[]));`

	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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
