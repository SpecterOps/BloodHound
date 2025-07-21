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
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/dawgs/graph"
	"gorm.io/gorm"
)

const (
	kindTable = "kind"
)

// AssetGroupTagData defines the methods required to interact with the asset_group_tags table
type AssetGroupTagData interface {
	CreateAssetGroupTag(ctx context.Context, tagType model.AssetGroupTagType, user model.User, name string, description string, position null.Int32, requireCertify null.Bool) (model.AssetGroupTag, error)
	UpdateAssetGroupTag(ctx context.Context, user model.User, tag model.AssetGroupTag) (model.AssetGroupTag, error)
	DeleteAssetGroupTag(ctx context.Context, user model.User, assetGroupTag model.AssetGroupTag) error
	GetAssetGroupTag(ctx context.Context, assetGroupTagId int) (model.AssetGroupTag, error)
	GetAssetGroupTags(ctx context.Context, sqlFilter model.SQLFilter) (model.AssetGroupTags, error)
	GetOrderedAssetGroupTagTiers(ctx context.Context) ([]model.AssetGroupTag, error)
	GetAssetGroupTagForSelection(ctx context.Context) ([]model.AssetGroupTag, error)
}

// AssetGroupTagSelectorData defines the methods required to interact with the asset_group_tag_selectors and asset_group_tag_selector_seeds tables
type AssetGroupTagSelectorData interface {
	CreateAssetGroupTagSelector(ctx context.Context, assetGroupTagId int, user model.User, name string, description string, isDefault bool, allowDisable bool, autoCertify null.Bool, seeds []model.SelectorSeed) (model.AssetGroupTagSelector, error)
	GetAssetGroupTagSelectorBySelectorId(ctx context.Context, assetGroupTagSelectorId int) (model.AssetGroupTagSelector, error)
	UpdateAssetGroupTagSelector(ctx context.Context, actorId, email string, selector model.AssetGroupTagSelector) (model.AssetGroupTagSelector, error)
	DeleteAssetGroupTagSelector(ctx context.Context, user model.User, selector model.AssetGroupTagSelector) error
	GetAssetGroupTagSelectorCounts(ctx context.Context, tagIds []int) (map[int]int, error)
	GetAssetGroupTagSelectorsByTagId(ctx context.Context, assetGroupTagId int, selectorSqlFilter, selectorSeedSqlFilter model.SQLFilter, skip, limit int) (model.AssetGroupTagSelectors, int, error)
	GetCustomAssetGroupTagSelectorsToMigrate(ctx context.Context) (model.AssetGroupTagSelectors, error)
	GetAssetGroupTagSelectors(ctx context.Context, sqlFilter model.SQLFilter, limit int) (model.AssetGroupTagSelectors, error)
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
			} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user.ID.String(), user.EmailAddress.ValueOrZero(), name, model.AssetGroupHistoryActionCreateSelector, assetGroupTagId, null.String{}, null.String{}); err != nil {
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

func (s *BloodhoundDB) UpdateAssetGroupTagSelector(ctx context.Context, actorId, emailAddress string, selector model.AssetGroupTagSelector) (model.AssetGroupTagSelector, error) {
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
			actorId, selector.Name, selector.Description, selector.DisabledAt, selector.DisabledBy, selector.AutoCertify, selector.ID); result.Error != nil {
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
			if err := bhdb.CreateAssetGroupHistoryRecord(ctx, actorId, emailAddress, selector.Name, model.AssetGroupHistoryActionUpdateSelector, selector.AssetGroupTagId, null.String{}, null.String{}); err != nil {
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
		} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user.ID.String(), user.EmailAddress.ValueOrZero(), selector.Name, model.AssetGroupHistoryActionDeleteSelector, selector.AssetGroupTagId, null.String{}, null.String{}); err != nil {
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
	if result := s.db.WithContext(ctx).Raw(fmt.Sprintf("SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify, analysis_enabled FROM %s WHERE id = ? AND deleted_at IS NULL", tag.TableName()), assetGroupTagId).First(&tag); result.Error != nil {
		return model.AssetGroupTag{}, CheckError(result)
	} else {
		return tag, nil
	}
}

func (s *BloodhoundDB) GetOrderedAssetGroupTagTiers(ctx context.Context) ([]model.AssetGroupTag, error) {
	var tags model.AssetGroupTags
	if result := s.db.WithContext(ctx).Raw(
		fmt.Sprintf(
			"SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify, analysis_enabled FROM %s WHERE type = ? AND deleted_at IS NULL ORDER BY position ASC",
			model.AssetGroupTag{}.TableName(),
		), model.AssetGroupTagTypeTier,
	).Find(&tags); result.Error != nil {
		return model.AssetGroupTags{}, CheckError(result)
	}
	return tags, nil
}

func (s *BloodhoundDB) GetAssetGroupTags(ctx context.Context, sqlFilter model.SQLFilter) (model.AssetGroupTags, error) {
	if sqlFilter.SQLString != "" {
		sqlFilter.SQLString = " AND " + sqlFilter.SQLString
	}

	var tags model.AssetGroupTags

	if result := s.db.WithContext(ctx).Raw(
		fmt.Sprintf(
			"SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify, analysis_enabled FROM %s WHERE deleted_at IS NULL%s ORDER BY name ASC",
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
	)

	if tag.ToType() == "unknown" {
		return model.AssetGroupTag{}, fmt.Errorf("unknown asset group tag")
	} else if tagType != model.AssetGroupTagTypeTier && (position.Valid || requireCertify.Valid) {
		return model.AssetGroupTag{}, fmt.Errorf("position and require_certify are limited to tiers only")
	} else if tagType == model.AssetGroupTagTypeTier {
		tag.AnalysisEnabled = null.BoolFrom(false)
	}

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		if tag.Type == model.AssetGroupTagTypeTier {

			orderedTags, err := bhdb.GetOrderedAssetGroupTagTiers(ctx)
			if err != nil {
				return err
			}

			if !tag.Position.Valid {
				tag.Position.SetValid(int32(len(orderedTags) + 1))
			}
			pos := int(tag.Position.ValueOrZero())
			if pos <= 1 || pos > len(orderedTags)+1 {
				return ErrPositionOutOfRange
			}

			orderedTags = slices.Insert(orderedTags, pos-1, tag)

			if err := bhdb.UpdateTierPositions(ctx, user, orderedTags, tag.ID); err != nil {
				return err
			}

		}

		query := fmt.Sprintf(`
			WITH inserted_kind AS (
				INSERT INTO %s (name) VALUES (?) RETURNING id
			)
			INSERT INTO %s (type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify, analysis_enabled)
			VALUES (?, (SELECT id FROM inserted_kind), ?, ?, NOW(), ?, NOW(), ?, ?, ?, ?)
			RETURNING id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify, analysis_enabled
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
			tag.AnalysisEnabled,
		).Scan(&tag); result.Error != nil {
			if strings.Contains(result.Error.Error(), "duplicate key value violates unique constraint \"kind_name_key\"") {
				return fmt.Errorf("%w: %v", ErrDuplicateKindName, result.Error)
			}
			return CheckError(result)
		} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user.ID.String(), user.EmailAddress.ValueOrZero(), name, model.AssetGroupHistoryActionCreateTag, tag.ID, null.String{}, null.String{}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return model.AssetGroupTag{}, err
	}

	return tag, nil
}

func (s *BloodhoundDB) UpdateAssetGroupTag(ctx context.Context, user model.User, tag model.AssetGroupTag) (model.AssetGroupTag, error) {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionUpdateAssetGroupTag,
			Model:  &tag, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if tag.Type == model.AssetGroupTagTypeTier {
		if !tag.Position.Valid {
			return model.AssetGroupTag{}, fmt.Errorf("position is required for an existing tier")
		}
	} else if tag.Position.Valid || tag.RequireCertify.Valid || tag.AnalysisEnabled.Valid {
		return model.AssetGroupTag{}, fmt.Errorf("position, require_certify, and analysis_enabled are limited to tiers only")
	}

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		var newPosition = tag.Position // only set for tiers

		origTag, err := bhdb.GetAssetGroupTag(ctx, tag.ID)
		if err != nil {
			return err
		}

		if tag.Type == model.AssetGroupTagTypeTier {
			if !origTag.Position.Equal(tag.Position) {
				origPosInt := int(origTag.Position.ValueOrZero())
				newPosInt := int(tag.Position.ValueOrZero())
				orderedTags, err := bhdb.GetOrderedAssetGroupTagTiers(ctx)
				if err != nil {
					return err
				}

				if newPosInt <= 1 || newPosInt > len(orderedTags) {
					return ErrPositionOutOfRange
				}

				orderedTags = slices.Delete(orderedTags, origPosInt-1, origPosInt)
				orderedTags = slices.Insert(orderedTags, newPosInt-1, tag)

				if err := bhdb.UpdateTierPositions(ctx, user, orderedTags, tag.ID); err != nil {
					return err
				}
			}
		}

		if result := tx.Exec(
			fmt.Sprintf(
				`UPDATE %s
				SET
					name = ?,
					description = ?,
					position = ?,
					require_certify = ?,
					analysis_enabled = ?,
					updated_at = NOW(),
					updated_by = ?
				WHERE id = ?`,
				tag.TableName(),
			),
			tag.Name,
			tag.Description,
			newPosition, // this is the same as tag.Position for non-tiers
			tag.RequireCertify,
			tag.AnalysisEnabled,
			user.ID.String(),
			tag.ID,
		); result.Error != nil {
			var pgErr *pgconn.PgError
			if errors.As(result.Error, &pgErr) &&
				pgErr.Code == "23505" && // unique_violation
				pgErr.ConstraintName == "agl_name_unique_index" {
				return fmt.Errorf("tag name must be unique: %w: %v", ErrDuplicateAGName, result.Error)
			}
			return CheckError(result)
		} else {
			if origTag.Name != tag.Name {
				if result := tx.Exec(
					fmt.Sprintf(`UPDATE %s SET name = ? WHERE id = ?`, kindTable),
					tag.KindName(),
					tag.KindId,
				); result.Error != nil {
					return CheckError(result)
				}
			}

			if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user.ID.String(), user.EmailAddress.ValueOrZero(), tag.Name, model.AssetGroupHistoryActionUpdateTag, tag.ID, null.String{}, null.String{}); err != nil {
				return err
			}

			// Analysis was updated, create an additional separate history record
			if !origTag.AnalysisEnabled.Equal(tag.AnalysisEnabled) {
				action := model.AssetGroupHistoryActionAnalysisDisabledTag
				if tag.AnalysisEnabled.ValueOrZero() {
					action = model.AssetGroupHistoryActionAnalysisEnabledTag
				}
				if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user.ID.String(), user.EmailAddress.ValueOrZero(), tag.Name, action, tag.ID, null.String{}, null.String{}); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		return model.AssetGroupTag{}, err
	}

	return tag, nil
}

func (s *BloodhoundDB) DeleteAssetGroupTag(ctx context.Context, user model.User, assetGroupTag model.AssetGroupTag) error {
	var (
		auditEntry = model.AuditEntry{
			Action: model.AuditLogActionDeleteAssetGroupTag,
			Model:  &assetGroupTag, // Pointer is required to ensure success log contains updated fields after transaction
		}
	)

	if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
		bhdb := NewBloodhoundDB(tx, s.idResolver)

		if selectors, _, err := bhdb.GetAssetGroupTagSelectorsByTagId(ctx, assetGroupTag.ID, model.SQLFilter{}, model.SQLFilter{}, 0, 0); err != nil {
			return err
		} else {
			for _, selector := range selectors {
				if err := bhdb.DeleteAssetGroupTagSelector(ctx, user, selector); err != nil {
					return err
				}
			}
		}

		if assetGroupTag.Type == model.AssetGroupTagTypeTier && assetGroupTag.Position.Valid && assetGroupTag.Position.Int32 == 1 {
			return fmt.Errorf("you cannot delete a tier in the 1st position")
		} else if assetGroupTag.Type == model.AssetGroupTagTypeOwned {
			return fmt.Errorf("you cannot delete the owned tag")
		}

		if result := tx.Exec(fmt.Sprintf(`
			UPDATE %s SET kind_id = null, updated_at = NOW(), updated_by = ?, deleted_at = NOW(), deleted_by = ?, position = null
			WHERE id = ?`,
			assetGroupTag.TableName()),
			user.ID.String(), user.ID.String(), assetGroupTag.ID); result.Error != nil {
			return CheckError(result)
		} else if result := tx.Exec("DELETE FROM kind WHERE id = ?", assetGroupTag.KindId); result.Error != nil {
			return CheckError(result)
		} else if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user.ID.String(), user.EmailAddress.ValueOrZero(), assetGroupTag.Name, model.AssetGroupHistoryActionDeleteTag, assetGroupTag.ID, null.String{}, null.String{}); err != nil {
			return err
		}

		if assetGroupTag.Type == model.AssetGroupTagTypeTier && assetGroupTag.Position.Valid && assetGroupTag.Position.Int32 > 1 {
			if orderedTags, err := bhdb.GetOrderedAssetGroupTagTiers(ctx); err != nil {
				return err
			} else {
				if err := bhdb.UpdateTierPositions(ctx, user, orderedTags); err != nil {
					return err
				}
			}
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
}

func (s *BloodhoundDB) GetAssetGroupTagSelectorsByTagId(ctx context.Context, assetGroupTagId int, selectorSqlFilter, selectorSeedSqlFilter model.SQLFilter, skip, limit int) (model.AssetGroupTagSelectors, int, error) {
	var (
		results         = model.AssetGroupTagSelectors{}
		skipLimitString string
		totalRowCount   int
	)

	var selectorSqlFilterStr string
	if selectorSqlFilter.SQLString != "" {
		selectorSqlFilterStr = " AND " + selectorSqlFilter.SQLString
	}

	var selectorSeedSqlFilterStr string
	if selectorSeedSqlFilter.SQLString != "" {
		selectorSeedSqlFilterStr = " WHERE " + selectorSeedSqlFilter.SQLString
	}

	if limit > 0 {
		skipLimitString += fmt.Sprintf(" LIMIT %d", limit)
	}

	if skip > 0 {
		skipLimitString += fmt.Sprintf(" OFFSET %d", skip)
	}

	baseSqlStr := fmt.Sprintf(`
		WITH selectors AS (
			SELECT id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify FROM %s WHERE asset_group_tag_id = ?%s
		), seeds AS (
			SELECT selector_id, type, value FROM %s %s
		)`,
		model.AssetGroupTagSelector{}.TableName(), selectorSqlFilterStr, model.SelectorSeed{}.TableName(), selectorSeedSqlFilterStr)

	sqlStr := fmt.Sprintf("%s SELECT * FROM seeds JOIN selectors ON seeds.selector_id = selectors.id ORDER BY selectors.id %s", baseSqlStr, skipLimitString)

	if rows, err := s.db.WithContext(ctx).Raw(sqlStr, append(append([]any{assetGroupTagId}, selectorSqlFilter.Params...), selectorSeedSqlFilter.Params...)...).Rows(); err != nil {
		return model.AssetGroupTagSelectors{}, 0, err
	} else {
		defer rows.Close()

		index := -1
		for rows.Next() {
			var (
				selector model.AssetGroupTagSelector
				seed     model.SelectorSeed
			)

			if err := s.db.ScanRows(rows, &seed); err != nil {
				return model.AssetGroupTagSelectors{}, 0, err
			}

			if index < 0 || seed.SelectorId != results[index].ID {
				if err := s.db.ScanRows(rows, &selector); err != nil {
					return model.AssetGroupTagSelectors{}, 0, err
				}
				results = append(results, selector)
				index++
			}

			results[index].Seeds = append(results[index].Seeds, seed)
		}

		// we need an overall count of the rows if pagination is supplied
		if limit > 0 || skip > 0 {
			countSqlStr := baseSqlStr + " SELECT COUNT(*) FROM seeds JOIN selectors ON seeds.selector_id = selectors.id"
			if err := s.db.WithContext(ctx).Raw(countSqlStr, append(append([]any{assetGroupTagId}, selectorSqlFilter.Params...), selectorSeedSqlFilter.Params...)...).Scan(&totalRowCount).Error; err != nil {
				return model.AssetGroupTagSelectors{}, 0, err
			}
		} else {
			totalRowCount = len(results)
		}
	}

	return results, totalRowCount, nil
}

func (s *BloodhoundDB) GetCustomAssetGroupTagSelectorsToMigrate(ctx context.Context) (model.AssetGroupTagSelectors, error) {
	var results = model.AssetGroupTagSelectors{}

	sqlStr := fmt.Sprintf(`
		WITH selectors AS (
			SELECT id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify FROM %s WHERE created_at = updated_at AND created_at < '2025-05-28' AND is_default = false
		), seeds AS (
			SELECT selector_id, type, value FROM %s WHERE type = 1
		)
		SELECT * FROM seeds JOIN selectors ON seeds.selector_id = selectors.id WHERE value = name ORDER BY selectors.id`,
		model.AssetGroupTagSelector{}.TableName(), model.SelectorSeed{}.TableName())

	if rows, err := s.db.WithContext(ctx).Raw(sqlStr).Rows(); err != nil {
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
		SELECT id, type, kind_id, name, description, created_at, created_by, updated_at, updated_by, position, require_certify, analysis_enabled FROM %s WHERE id IN ((SELECT id FROM tier), (SELECT id FROM owned))`,
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

func (s *BloodhoundDB) UpdateTierPositions(ctx context.Context, user model.User, orderedTags model.AssetGroupTags, ignoredTagIds ...int) error {
	for newPos, tag := range orderedTags {
		newPos++ // position is 1 based not zero

		if slices.Contains(ignoredTagIds, tag.ID) || tag.Position.ValueOrZero() == int32(newPos) {
			continue
		}

		var (
			auditEntry = model.AuditEntry{
				Action: model.AuditLogActionUpdateAssetGroupTag,
				Model:  &tag, // Pointer is required to ensure success log contains updated fields after transaction
			}
		)

		if err := s.AuditableTransaction(ctx, auditEntry, func(tx *gorm.DB) error {
			bhdb := NewBloodhoundDB(tx, s.idResolver)

			tag.UpdatedAt = time.Now()
			tag.UpdatedBy = user.ID.String()
			tag.Position.SetValid(int32(newPos))

			if err := tx.WithContext(ctx).Save(&tag).Error; err != nil {
				return fmt.Errorf("failed to update tag position: %w", err)
			}

			if err := bhdb.CreateAssetGroupHistoryRecord(ctx, user.ID.String(), user.EmailAddress.ValueOrZero(), tag.Name, model.AssetGroupHistoryActionUpdateTag, tag.ID, null.String{}, null.String{}); err != nil {
				return fmt.Errorf("failed to create history record: %w", err)
			}

			return nil
		}); err != nil {
			return err
		}
	}

	return nil
}

func (s *BloodhoundDB) GetAssetGroupTagSelectors(ctx context.Context, sqlFilter model.SQLFilter, limit int) (model.AssetGroupTagSelectors, error) {
	var selectors = model.AssetGroupTagSelectors{}

	if sqlFilter.SQLString != "" {
		sqlFilter.SQLString = "WHERE " + sqlFilter.SQLString
	}

	limitStr := ""
	if limit > 0 {
		limitStr = "LIMIT " + strconv.Itoa(limit)
	}

	if result := s.db.WithContext(ctx).Raw(
		fmt.Sprintf(
			"SELECT id, asset_group_tag_id, created_at, created_by, updated_at, updated_by, disabled_at, disabled_by, name, description, is_default, allow_disable, auto_certify FROM %s %s ORDER BY name ASC, asset_group_tag_id ASC, id ASC %s",
			model.AssetGroupTagSelector{}.TableName(),
			sqlFilter.SQLString,
			limitStr,
		),
		sqlFilter.Params...,
	).Scan(&selectors); result.Error != nil {
		return model.AssetGroupTagSelectors{}, CheckError(result)
	}

	return selectors, nil
}
