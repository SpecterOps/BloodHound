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

package database

import (
	"context"
	"fmt"

	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

const (
	kindTable = "kind"
)

// AssetGroupTagData defines the methods required to interact with the asset_group_tags table
type AssetGroupTagData interface {
	CreateAssetGroupTag(ctx context.Context, tagType model.AssetGroupTagType, userId string, name string, description string, position null.Int32, requireCertify null.Bool) (model.AssetGroupTag, error)
	GetAssetGroupTag(ctx context.Context, assetGroupTagId int) (model.AssetGroupTag, error)
}

// AssetGroupTagSelectorData defines the methods required to interact with the asset_group_tag_selectors and asset_group_tag_selector_seeds tables
type AssetGroupTagSelectorData interface {
	CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, userId string, name string, description string, isDefault bool, allowDisable bool, autoCertify bool, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error)
	GetAssetGroupTagSelectorsByTagId(ctx context.Context, assetGroupLabelID int, sqlFilter model.SQLFilter) ([]model.AssetGroupTagSelector, error)
}

func (s *BloodhoundDB) CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, userId string, name string, description string, isDefault bool, allowDisable bool, autoCertify bool, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error) {
	var (
		selector = model.AssetGroupTagSelector{
			AssetGroupTagId: assetGroupTagId,
			CreatedBy:       userId,
			UpdatedBy:       userId,
			Name:            name,
			Description:     description,
			IsDefault:       isDefault,
			AllowDisable:    allowDisable,
			AutoCertify:     autoCertify,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateAssetGroupTagSelector,
			Model:  &selector, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)
		if result := tx.Raw(fmt.Sprintf(`
			INSERT INTO %s (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify) 
			VALUES (?, NOW(), ?, NOW(), ?, ?, ?, ?, ?, ?) 
			RETURNING id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify`,
			selector.TableName()),
			assetGroupTagId, userId, userId, name, description, isDefault, allowDisable, autoCertify).Scan(&selector); result.Error != nil {
			return CheckError(result)
		} else {
			for _, seed := range seeds {
				if result := tx.Exec(fmt.Sprintf("INSERT INTO %s (selector_id, type, value) VALUES (?, ?, ?)", seed.TableName()), selector.ID, seed.Type, seed.Value); result.Error != nil {
					return CheckError(result)
				} else {
					selector.Seeds = append(selector.Seeds, model.SelectorSeed{Type: seed.Type, Value: seed.Value})
				}
			}
			if err := bhdb.CreateAssetGroupHistoryRecord(ctx, userId, name, model.AssetGroupHistoryActionCreateSelector, assetGroupTagId, null.String{}, null.String{}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return model.AssetGroupTagSelector{}, err
	}

	return selector, nil
}

func (s *BloodhoundDB) GetAssetGroupTag(ctx context.Context, assetGroupTagId int) (model.AssetGroupTag, error) {
	var tag model.AssetGroupTag
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify FROM %s WHERE id = ? AND deleted_at IS NULL", tag.TableName()), assetGroupTagId).First(&tag); result.Error != nil {
		return model.AssetGroupTag{}, CheckError(result)
	} else {
		return tag, nil
	}
}

func (s *BloodhoundDB) CreateAssetGroupTag(ctx context.Context, tagType model.AssetGroupTagType, userId string, name string, description string, position null.Int32, requireCertify null.Bool) (model.AssetGroupTag, error) {
	var (
		tag = model.AssetGroupTag{
			Type:           tagType,
			CreatedBy:      userId,
			UpdatedBy:      userId,
			Name:           name,
			Description:    description,
			Position:       position,
			RequireCertify: requireCertify,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateAssetGroupTag,
			Model:  &tag, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if tagType != model.AssetGroupTagTypeTier && (position.Valid || requireCertify.Valid) {
		return model.AssetGroupTag{}, fmt.Errorf("position and require_certify are limited to tiers only")
	}

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		var kindId int
		if result := tx.Raw(fmt.Sprintf("INSERT INTO %s (name) VALUES (?) RETURNING id", kindTable), tag.ToKind()).Scan(&kindId); result.Error != nil {
			return CheckError(result)
		} else if result := tx.Raw(fmt.Sprintf(`
			INSERT INTO %s (type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify) 
			VALUES (?, ?, ?, ?, NOW(), ?, NOW(), ?, ?, ?) 
			RETURNING id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify`,
			tag.TableName()),
			tagType, kindId, name, description, userId, userId, position, requireCertify).Scan(&tag); result.Error != nil {
			return CheckError(result)
		} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, userId, name, model.AssetGroupHistoryActionCreateTag, tag.ID, null.String{}, null.String{}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.AssetGroupTag{}, err
	}
	return tag, nil
}

func (s *BloodhoundDB) GetAssetGroupTagSelectorsByTagId(ctx context.Context, assetGroupTagId int, sqlFilter model.SQLFilter) ([]model.AssetGroupTagSelector, error) {
	var selectors []model.AssetGroupTagSelector
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify FROM %s WHERE asset_group_tag_id = ?", model.AssetGroupTagSelector{}.TableName()), assetGroupTagId).Find(&selectors); result.Error != nil {
		return []model.AssetGroupTagSelector{}, CheckError(result)
	} else {
		for index, selector := range selectors {
			if sqlFilter.SQLString != "" {
				result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT selector_id, type, value FROM %s WHERE selector_id = ? AND type = ?", model.SelectorSeed{}.TableName()), selector.ID, sqlFilter.Params).Scan(&selectors[index].Seeds)
				if result.Error != nil {
					return []model.AssetGroupTagSelector{}, CheckError(result)
				} else {
					result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT selector_id, type, value FROM %s WHERE selector_id = ?", model.SelectorSeed{}.TableName()), selector.ID).Scan(&selectors[index].Seeds)
					if result.Error != nil {
						return []model.AssetGroupTagSelector{}, CheckError(result)

					}
				}
			}
		}
	}
	return selectors, nil
}
