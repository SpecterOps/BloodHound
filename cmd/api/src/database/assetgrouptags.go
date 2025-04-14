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

	"github.com/specterops/bloodhound/dawgs/graph"
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
	CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, userId string, name string, description string, isDefault bool, allowDisable bool, autoCertify null.Bool, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error)
	GetAssetGroupTagSelectorBySelectorId(ctx context.Context, assetGroupTagSelectorId int) (model.AssetGroupTagSelector, error)
	UpdateAssetGroupTagSelector(ctx context.Context, userId string, selector model.AssetGroupTagSelector) (model.AssetGroupTagSelector, error)
	DeleteAssetGroupTagSelector(ctx context.Context, userId string, selector model.AssetGroupTagSelector) error
	GetAssetGroupTagSelectorsByTagId(ctx context.Context, assetGroupTagId int, selectorSqlFilter, selectorSeedSqlFilter model.SQLFilter) (model.AssetGroupTagSelectors, error)
	GetAssetGroupTagForSelection(ctx context.Context) ([]model.AssetGroupTag, error)
}

// AssetGroupTagSelectorNodeData defines the methods required to interact with the asset_group_tag_selector_nodes table
type AssetGroupTagSelectorNodeData interface {
	InsertSelectorNode(ctx context.Context, selectorId int, nodeId graph.ID, certified model.AssetGroupCertification, certifiedBy null.String, source model.AssetGroupSelectorNodeSource) error
	UpdateSelectorNodesByNodeId(ctx context.Context, selectorId int, certified model.AssetGroupCertification, certifiedBy null.String, nodeIds []graph.ID) error
	DeleteSelectorNodesByNodeId(ctx context.Context, selectorId int, nodeIds []graph.ID) error
	DeleteSelectorNodesBySelectorIds(ctx context.Context, selectorId ...int) error
	GetSelectorNodesBySelectorIds(ctx context.Context, selectorIds ...int) ([]model.AssetGroupSelectorNode, error)
}

func insertSelectorSeeds(tx *gorm.DB, selectorId int, seeds []model.SelectorSeed) ([]model.SelectorSeed, error) {
	for _, seed := range seeds {
		if result := tx.Exec(fmt.Sprintf("INSERT INTO %s (selector_id, type, value) VALUES (?, ?, ?)", seed.TableName()), selectorId, seed.Type, seed.Value); result.Error != nil {
			return nil, CheckError(result)
		}
	}
	return seeds, nil
}

func (s *BloodhoundDB) CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, userId string, name string, description string, isDefault bool, allowDisable bool, autoCertify null.Bool, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error) {
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
			var err error
			if selector.Seeds, err = insertSelectorSeeds(tx, selector.ID, seeds); err != nil {
				return err
			} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, userId, name, model.AssetGroupHistoryActionCreateSelector, assetGroupTagId, null.String{}, null.String{}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return model.AssetGroupTagSelector{}, err
	}

	return selector, nil
}

func (s *BloodhoundDB) GetAssetGroupTagSelectorBySelectorId(ctx context.Context, assetGroupTagSelectorId int) (model.AssetGroupTagSelector, error) {
	var (
		selector = model.AssetGroupTagSelector{
			ID: assetGroupTagSelectorId,
		}
	)

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf(`
			SELECT id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify
			FROM %s WHERE id = ?`,
			selector.TableName()),
			assetGroupTagSelectorId).First(&selector); result.Error != nil {
			return CheckError(result)
		} else if result := tx.Raw(fmt.Sprintf("SELECT selector_id, type, value FROM %s WHERE selector_id = ?", (model.SelectorSeed{}).TableName()), selector.ID).Find(&selector.Seeds); result.Error != nil {
			return CheckError(result)
		}
		return nil
	}); err != nil {
		return model.AssetGroupTagSelector{}, err
	}

	return selector, nil
}

func (s *BloodhoundDB) UpdateAssetGroupTagSelector(ctx context.Context, userId string, selector model.AssetGroupTagSelector) (model.AssetGroupTagSelector, error) {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionUpdateAssetGroupTagSelector,
			Model:  &selector, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)
		if result := tx.Exec(fmt.Sprintf(`
			UPDATE %s SET updated_at = NOW(), updated_by = ?, name = ?, description = ?, disabled_at = ?, disabled_by = ?, auto_certify = ?
			WHERE id = ?`,
			selector.TableName()),
			userId, selector.Name, selector.Description, selector.DisabledAt, selector.DisabledBy, selector.AutoCertify, selector.ID); result.Error != nil {
			return CheckError(result)
		} else {
			if selector.Seeds != nil {
				// delete old seeds and re-insert the new ones
				if result := tx.Exec(fmt.Sprintf("DELETE FROM %s WHERE selector_id = ?", model.SelectorSeed{}.TableName()), selector.ID); result.Error != nil {
					return CheckError(result)
				} else if _, err := insertSelectorSeeds(tx, selector.ID, selector.Seeds); err != nil {
					return err
				}
			}
			if err := bhdb.CreateAssetGroupHistoryRecord(ctx, userId, selector.Name, model.AssetGroupHistoryActionUpdateSelector, selector.AssetGroupTagId, null.String{}, null.String{}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return model.AssetGroupTagSelector{}, err
	}

	return selector, nil
}

func (s *BloodhoundDB) DeleteAssetGroupTagSelector(ctx context.Context, userId string, selector model.AssetGroupTagSelector) error {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionDeleteAssetGroupTagSelector,
			Model:  &selector, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)
		if result := tx.Exec(fmt.Sprintf(`DELETE FROM %s WHERE id = ?`, selector.TableName()), selector.ID); result.Error != nil {
			return CheckError(result)
		} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, userId, selector.Name, model.AssetGroupHistoryActionDeleteSelector, selector.AssetGroupTagId, null.String{}, null.String{}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	return nil
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

	if tag.ToType() == "unknown" {
		return model.AssetGroupTag{}, fmt.Errorf("unknown asset group tag")
	}

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

func (s *BloodhoundDB) GetAssetGroupTagSelectorsByTagId(ctx context.Context, assetGroupTagId int, selectorSqlFilter, selectorSeedSqlFilter model.SQLFilter) (model.AssetGroupTagSelectors, error) {
	var results = model.AssetGroupTagSelectors{}

	var selectorSqlFilterStr string
	if selectorSqlFilter.SQLString != "" {
		selectorSqlFilterStr = " AND " + selectorSqlFilter.SQLString
	}

	var selectorSeedSqlFilterStr string
	if selectorSeedSqlFilter.SQLString != "" {
		selectorSeedSqlFilterStr = " WHERE " + selectorSeedSqlFilter.SQLString
	}

	sqlStr := fmt.Sprintf(`WITH selectors
  AS (SELECT  id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify FROM %s WHERE asset_group_tag_id = ?%s),
seeds
  AS (SELECT selector_id, type, value FROM %s %s)
SELECT * FROM seeds JOIN selectors ON seeds.selector_id = selectors.id ORDER BY selectors.id`, model.AssetGroupTagSelector{}.TableName(), selectorSqlFilterStr, model.SelectorSeed{}.TableName(), selectorSeedSqlFilterStr)

	if rows, err := s.db.WithContext(ctx).Raw(sqlStr, append(append([]any{assetGroupTagId}, selectorSqlFilter.Params...), selectorSeedSqlFilter.Params...)...).Rows(); err != nil {
		return model.AssetGroupTagSelectors{}, err
	} else {
		defer rows.Close()

		index := -1
		for rows.Next() {
			var (
				selector model.AssetGroupTagSelector
				seed     model.SelectorSeed
			)

			if err := s.db.ScanRows(rows, &seed); err != nil {
				return model.AssetGroupTagSelectors{}, err
			}

			if index < 0 || seed.SelectorId != results[index].ID {
				if err := s.db.ScanRows(rows, &selector); err != nil {
					return model.AssetGroupTagSelectors{}, err
				}
				results = append(results, selector)
				index++
			}

			results[index].Seeds = append(results[index].Seeds, seed)
		}
	}

	return results, nil
}

func (s *BloodhoundDB) GetAssetGroupTagForSelection(ctx context.Context) ([]model.AssetGroupTag, error) {
	var tags []model.AssetGroupTag
	return tags, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf(`WITH
tier AS (SELECT id FROM asset_group_tags WHERE type = 1 AND position = 1 AND deleted_at IS NULL LIMIT 1),
owned AS (SELECT id FROM asset_group_tags WHERE type = 3 AND deleted_at IS NULL LIMIT 1)
SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by FROM %s WHERE id IN ((SELECT id FROM tier), (SELECT id FROM owned))`, model.AssetGroupTag{}.TableName())).Find(&tags))
}

func (s *BloodhoundDB) InsertSelectorNode(ctx context.Context, selectorId int, nodeId graph.ID, certified model.AssetGroupCertification, certifiedBy null.String, source model.AssetGroupSelectorNodeSource) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("INSERT INTO %s (selector_id, node_id, certified, certified_by, source) VALUES(?, ?, ?, ?, ?) ON CONFLICT DO NOTHING", model.AssetGroupSelectorNode{}.TableName()), selectorId, nodeId, certified, certifiedBy, source))
}

func (s *BloodhoundDB) UpdateSelectorNodesByNodeId(ctx context.Context, selectorId int, certified model.AssetGroupCertification, certifiedBy null.String, nodeIds []graph.ID) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("UPDATE %s SET certified = ?, certified_by = ? WHERE selector_id = ? AND node_id = IN ?", model.AssetGroupSelectorNode{}.TableName()), certified, certifiedBy, selectorId, nodeIds))
}

func (s *BloodhoundDB) DeleteSelectorNodesByNodeId(ctx context.Context, selectorId int, nodeIds []graph.ID) error {
	if len(nodeIds) == 0 {
		return nil
	}
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE selector_id = ? AND node_id IN ?", model.AssetGroupSelectorNode{}.TableName()), selectorId, nodeIds))
}

func (s *BloodhoundDB) DeleteSelectorNodesBySelectorIds(ctx context.Context, selectorIds ...int) error {
	if len(selectorIds) == 0 {
		return nil
	}
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE selector_id IN ?", model.AssetGroupSelectorNode{}.TableName()), selectorIds))
}

func (s *BloodhoundDB) GetSelectorNodesBySelectorIds(ctx context.Context, selectorIds ...int) ([]model.AssetGroupSelectorNode, error) {
	var nodes []model.AssetGroupSelectorNode
	if len(selectorIds) == 0 {
		return nodes, nil
	}
	return nodes, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT selector_id, node_id, certified, certified_by, source FROM %s WHERE selector_id IN ?", model.AssetGroupSelectorNode{}.TableName()), selectorIds).Find(&nodes))
}
