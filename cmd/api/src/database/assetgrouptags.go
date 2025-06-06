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
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

type ShiftDirection int

const (
	kindTable                = "kind"
	ShiftUp   ShiftDirection = 1
	ShiftDown ShiftDirection = -1
)

// AssetGroupTagData defines the methods required to interact with the asset_group_tags table
type AssetGroupTagData interface {
	CreateAssetGroupTag(ctx context.Context, tagType model.AssetGroupTagType, user model.User, name string, description string, position null.Int32, requireCertify null.Bool) (model.AssetGroupTag, error)
	GetAssetGroupTag(ctx context.Context, assetGroupTagId int) (model.AssetGroupTag, error)
	GetAssetGroupTags(ctx context.Context, sqlFilter model.SQLFilter) (model.AssetGroupTags, error)
	GetAssetGroupTagForSelection(ctx context.Context) ([]model.AssetGroupTag, error)
	GetMaxTierPosition(ctx context.Context, tx *gorm.DB) (int32, error)
	CascadeShiftTierPositions(ctx context.Context, tx *gorm.DB, user model.User, position null.Int32, direction ShiftDirection) error
}

// AssetGroupTagSelectorData defines the methods required to interact with the asset_group_tag_selectors and asset_group_tag_selector_seeds tables
type AssetGroupTagSelectorData interface {
	CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, user model.User, name string, description string, isDefault bool, allowDisable bool, autoCertify null.Bool, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error)
	GetAssetGroupTagSelectorBySelectorId(ctx context.Context, assetGroupTagSelectorId int) (model.AssetGroupTagSelector, error)
	UpdateAssetGroupTagSelector(ctx context.Context, user model.User, selector model.AssetGroupTagSelector) (model.AssetGroupTagSelector, error)
	DeleteAssetGroupTagSelector(ctx context.Context, user model.User, selector model.AssetGroupTagSelector) error
	GetAssetGroupTagSelectorCounts(ctx context.Context, tagIds []int) (map[int]int, error)
	GetAssetGroupTagSelectorsByTagId(ctx context.Context, assetGroupTagId int, selectorSqlFilter, selectorSeedSqlFilter model.SQLFilter) (model.AssetGroupTagSelectors, error)
}

// AssetGroupTagSelectorNodeData defines the methods required to interact with the asset_group_tag_selector_nodes table
type AssetGroupTagSelectorNodeData interface {
	InsertSelectorNode(ctx context.Context, selectorId int, nodeId graph.ID, certified model.AssetGroupCertification, certifiedBy null.String, source model.AssetGroupSelectorNodeSource) error
	UpdateSelectorNodesByNodeId(ctx context.Context, selectorId int, certified model.AssetGroupCertification, certifiedBy null.String, nodeId graph.ID) error
	DeleteSelectorNodesByNodeId(ctx context.Context, selectorId int, nodeId graph.ID) error
	DeleteSelectorNodesBySelectorIds(ctx context.Context, selectorId ...int) error
	GetSelectorNodesBySelectorIds(ctx context.Context, selectorIds ...int) ([]model.AssetGroupSelectorNode, error)
	GetSelectorsByMemberId(ctx context.Context, memberId int, assetGroupTagId int) (model.AssetGroupTagSelectors, error)
}

func insertSelectorSeeds(tx *gorm.DB, selectorId int, seeds []model.SelectorSeed) ([]model.SelectorSeed, error) {
	for _, seed := range seeds {
		if result := tx.Exec(fmt.Sprintf("INSERT INTO %s (selector_id, type, value) VALUES (?, ?, ?)", seed.TableName()), selectorId, seed.Type, seed.Value); result.Error != nil {
			return nil, CheckError(result)
		}
	}
	return seeds, nil
}

func (s *BloodhoundDB) CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, user model.User, name string, description string, isDefault bool, allowDisable bool, autoCertify null.Bool, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error) {
	var (
		userIdStr = user.ID.String()
		selector  = model.AssetGroupTagSelector{
			AssetGroupTagId: assetGroupTagId,
			CreatedBy:       userIdStr,
			UpdatedBy:       userIdStr,
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

	if !autoCertify.Valid {
		return model.AssetGroupTagSelector{}, fmt.Errorf("auto_certify must be set to true or false")
	}

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)
		if result := tx.Raw(fmt.Sprintf(`
			INSERT INTO %s (asset_group_tag_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify)
			VALUES (?, NOW(), ?, NOW(), ?, ?, ?, ?, ?, ?)
			RETURNING id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify`,
			selector.TableName()),
			assetGroupTagId, userIdStr, userIdStr, name, description, isDefault, allowDisable, autoCertify).Scan(&selector); result.Error != nil {
			return CheckError(result)
		} else {
			var err error
			if selector.Seeds, err = insertSelectorSeeds(tx, selector.ID, seeds); err != nil {
				return err
			} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user, name, model.AssetGroupHistoryActionCreateSelector, assetGroupTagId, null.String{}, null.String{}); err != nil {
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

func (s *BloodhoundDB) UpdateAssetGroupTagSelector(ctx context.Context, user model.User, selector model.AssetGroupTagSelector) (model.AssetGroupTagSelector, error) {
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
			user.ID.String(), selector.Name, selector.Description, selector.DisabledAt, selector.DisabledBy, selector.AutoCertify, selector.ID); result.Error != nil {
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
			if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user, selector.Name, model.AssetGroupHistoryActionUpdateSelector, selector.AssetGroupTagId, null.String{}, null.String{}); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		return model.AssetGroupTagSelector{}, err
	}

	return selector, nil
}

func (s *BloodhoundDB) DeleteAssetGroupTagSelector(ctx context.Context, user model.User, selector model.AssetGroupTagSelector) error {
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
		} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user, selector.Name, model.AssetGroupHistoryActionDeleteSelector, selector.AssetGroupTagId, null.String{}, null.String{}); err != nil {
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

func (s *BloodhoundDB) GetAssetGroupTags(ctx context.Context, sqlFilter model.SQLFilter) (model.AssetGroupTags, error) {
	if sqlFilter.SQLString != "" {
		sqlFilter.SQLString = " AND " + sqlFilter.SQLString
	}

	var tags model.AssetGroupTags
	if result := s.db.WithContext(ctx).Raw(
		fmt.Sprintf(
			"SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify FROM %s WHERE deleted_at IS NULL%s",
			model.AssetGroupTag{}.TableName(),
			sqlFilter.SQLString,
		),
		sqlFilter.Params...,
	).Find(&tags); result.Error != nil {
		return model.AssetGroupTags{}, CheckError(result)
	}
	return tags, nil
}

func (s *BloodhoundDB) GetAssetGroupTagSelectorCounts(ctx context.Context, tagIds []int) (map[int]int, error) {
	var counts = make(map[int]int, len(tagIds))
	// initalize values to 0 for any ids that end up with no rows
	for _, i := range tagIds {
		counts[i] = 0
	}
	if rows, err := s.db.WithContext(ctx).Raw(
		fmt.Sprintf("SELECT asset_group_tag_id, COUNT(*) FROM %s WHERE asset_group_tag_id IN (?) AND disabled_at IS NULL GROUP BY asset_group_tag_id", model.AssetGroupTagSelector{}.TableName()),
		tagIds,
	).Rows(); err != nil {
		return map[int]int{}, err
	} else {
		defer rows.Close()
		var id, val int
		for rows.Next() {
			if err := rows.Scan(&id, &val); err != nil {
				return map[int]int{}, err
			}
			counts[id] = val
		}
		if err := rows.Err(); err != nil {
			return map[int]int{}, err
		}
	}
	return counts, nil
}

func (s *BloodhoundDB) CreateAssetGroupTag(ctx context.Context, tagType model.AssetGroupTagType, user model.User, name string, description string, position null.Int32, requireCertify null.Bool) (model.AssetGroupTag, error) {
	var (
		userIdStr = user.ID.String()
		tag       = model.AssetGroupTag{
			Type:           tagType,
			CreatedBy:      userIdStr,
			UpdatedBy:      userIdStr,
			Name:           name,
			Description:    description,
			Position:       position,
			RequireCertify: requireCertify,
		}

		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionCreateAssetGroupTag,
			Model:  &tag, // Pointer is required to ensure success log contains updated fields after transaction
		}

		positionProvided = tag.Position.Valid
	)

	if tag.ToType() == "unknown" {
		return model.AssetGroupTag{}, fmt.Errorf("unknown asset group tag")
	}

	if tagType != model.AssetGroupTagTypeTier && (position.Valid || requireCertify.Valid) {
		return model.AssetGroupTag{}, fmt.Errorf("position and require_certify are limited to tiers only")
	}

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		if tag.Type == model.AssetGroupTagTypeTier {
			// positionProvided := tag.Position.Valid

			if !positionProvided {
				if position, err := bhdb.GetMaxTierPosition(ctx, tx); err != nil {
					return err
				} else {
					tag.Position = null.Int32From(position + 1)
				}
			}

			if positionProvided && tag.Position == null.Int32From(1) {
				return fmt.Errorf("cannot create a new tier 0")
			}

			if positionProvided {
				if err := bhdb.CascadeShiftTierPositions(ctx, tx, user, tag.Position, ShiftUp); err != nil {
					return err
				}
			}

		}

		query := fmt.Sprintf(`
	WITH inserted_kind AS (
		INSERT INTO %s (name) VALUES (?) RETURNING id
	)
	INSERT INTO %s (type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify)
	VALUES (?, (SELECT id FROM inserted_kind), ?, ?, NOW(), ?, NOW(), ?, ?, ?)
	RETURNING id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify
`, kindTable, tag.TableName())

		if result := tx.Raw(query,
			tag.KindName(),
			tagType,
			name,
			description,
			userIdStr,
			userIdStr,
			tag.Position,
			requireCertify,
		).Scan(&tag); result.Error != nil {
			return CheckError(result)
		} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user, name, model.AssetGroupHistoryActionCreateTag, tag.ID, null.String{}, null.String{}); err != nil {
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

	sqlStr := fmt.Sprintf(`
		WITH selectors AS (
			SELECT id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify FROM %s WHERE asset_group_tag_id = ?%s
		), seeds AS (
			SELECT selector_id, type, value FROM %s %s
		)
		SELECT * FROM seeds JOIN selectors ON seeds.selector_id = selectors.id ORDER BY selectors.id`,
		model.AssetGroupTagSelector{}.TableName(), selectorSqlFilterStr, model.SelectorSeed{}.TableName(), selectorSeedSqlFilterStr)

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

func (s *BloodhoundDB) GetSelectorsByMemberId(ctx context.Context, memberId int, assetGroupTagId int) (model.AssetGroupTagSelectors, error) {
	var selectors model.AssetGroupTagSelectors

	return selectors, CheckError(s.db.WithContext(ctx).Raw(`
		SELECT s.* from asset_group_tag_selectors s
		JOIN asset_group_tag_selector_nodes n ON s.id = n.selector_id
		JOIN asset_group_tags t ON s.asset_group_tag_id = t.id
		WHERE t.id = ? AND n.node_id = ?
		`, assetGroupTagId, memberId).Find(&selectors))
}

func (s *BloodhoundDB) GetAssetGroupTagForSelection(ctx context.Context) ([]model.AssetGroupTag, error) {
	var tags []model.AssetGroupTag
	return tags, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf(`
		WITH tier AS (
			SELECT id FROM asset_group_tags WHERE type = 1 AND position = 1 AND deleted_at IS NULL LIMIT 1
		), owned AS (
			SELECT id FROM asset_group_tags WHERE type = 3 AND deleted_at IS NULL LIMIT 1
		)
		SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by FROM %s WHERE id IN ((SELECT id FROM tier), (SELECT id FROM owned))`,
		model.AssetGroupTag{}.TableName())).Find(&tags))
}

func (s *BloodhoundDB) InsertSelectorNode(ctx context.Context, selectorId int, nodeId graph.ID, certified model.AssetGroupCertification, certifiedBy null.String, source model.AssetGroupSelectorNodeSource) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("INSERT INTO %s (selector_id, node_id, certified, certified_by, source, created_at, updated_at) VALUES(?, ?, ?, ?, ?, current_timestamp, current_timestamp) ON CONFLICT DO NOTHING", model.AssetGroupSelectorNode{}.TableName()), selectorId, nodeId, certified, certifiedBy, source))
}

func (s *BloodhoundDB) UpdateSelectorNodesByNodeId(ctx context.Context, selectorId int, certified model.AssetGroupCertification, certifiedBy null.String, nodeId graph.ID) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("UPDATE %s SET certified = ?, certified_by = ?, updated_at = current_timestamp WHERE selector_id = ? AND node_id = ?", model.AssetGroupSelectorNode{}.TableName()), certified, certifiedBy, selectorId, nodeId))
}

func (s *BloodhoundDB) DeleteSelectorNodesByNodeId(ctx context.Context, selectorId int, nodeId graph.ID) error {
	return CheckError(s.db.WithContext(ctx).Exec(fmt.Sprintf("DELETE FROM %s WHERE selector_id = ? AND node_id = ?", model.AssetGroupSelectorNode{}.TableName()), selectorId, nodeId))
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
	return nodes, CheckError(s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT selector_id, node_id, certified, certified_by, source, created_at, updated_at FROM %s WHERE selector_id IN ?", model.AssetGroupSelectorNode{}.TableName()), selectorIds).Find(&nodes))
}

func (s *BloodhoundDB) GetMaxTierPosition(ctx context.Context, tx *gorm.DB) (int32, error) {
	var max null.Int32
	var tag model.AssetGroupTag

	if result := tx.WithContext(ctx).
		Raw(fmt.Sprintf("SELECT MAX(position) FROM %s", tag.TableName())).
		Scan(&max); result.Error != nil {
		return 0, CheckError(result)
	}

	return max.Int32, nil
}

func (s *BloodhoundDB) CascadeShiftTierPositions(ctx context.Context, tx *gorm.DB, user model.User, position null.Int32, direction ShiftDirection) error {
	var (
		positionOp string
	)

	switch direction {
	case ShiftUp:
		positionOp = ">="
	case ShiftDown:
		positionOp = ">"
	default:
		return fmt.Errorf("invalid shift direction")
	}

	// get affected rows
	var tags []model.AssetGroupTag
	if err := s.db.WithContext(ctx).
		Where(fmt.Sprintf("type = ? AND position %s ? AND position > 1", positionOp), model.AssetGroupTagTypeTier, position.Int32).
		Order("position ASC").
		Find(&tags).Error; err != nil {
		return fmt.Errorf("failed to fetch tags to shift: %w", err)
	}

	// update each and create history record
	for _, tag := range tags {
		var (
			auditEntry = model.AuditEntry{
				Action: model.AuditLogActionUpdateAssetGroupTag,
				Model:  &tag, // Pointer is required to ensure success log contains updated fields after transaction
			}
		)

		if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
			bhdb := NewBloodhoundDB(tx, s.idResolver)

			originalPosition := tag.Position.Int32
			if direction == ShiftUp {
				tag.Position.Int32++
			} else {
				tag.Position.Int32--
			}

			tag.UpdatedAt = time.Now()
			tag.UpdatedBy = user.ID.String()

			if err := s.db.WithContext(ctx).Save(&tag).Error; err != nil {
				return fmt.Errorf("failed to update tag position: %w", err)
			}

			if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user, tag.Name, model.AssetGroupHistoryActionUpdateTag, tag.ID, null.String{}, null.StringFrom(fmt.Sprintf("original position %d, updated positon %d", originalPosition, tag.Position.Int32))); err != nil {
				return fmt.Errorf("failed to create history record: %w", err)
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}
