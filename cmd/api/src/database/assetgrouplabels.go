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

	"github.com/specterops/bloodhound/src/model"
	"gorm.io/gorm"
)

const (
	assetGroupLabelTable        = "asset_group_labels"
	assetGroupSelectorTable     = "asset_group_label_selectors"
	assetGroupSelectorSeedTable = "asset_group_label_selector_seeds"
)

// AssetGroupLabelData defines the methods required to interact with the asset_group_labels table
type AssetGroupLabelData interface {
	CreateAssetGroupLabel(ctx context.Context, assetGroupTierId int, userId string, kindId int, name string, description string) (model.AssetGroupLabel, error)
	GetAssetGroupLabel(ctx context.Context, assetGroupLabelId int) (model.AssetGroupLabel, error)
}

// AssetGroupLabelSelectorData defines the methods required to interact with the asset_group_label_selectors and asset_group_label_selector_seeds tables
type AssetGroupLabelSelectorData interface {
	CreateAssetGroupLabelSelector(ctx context.Context, assetGroupLabelId int, userId string, name string, description string, isDefault bool, allowDisable bool, autoCertify bool, seeds []model.SelectorSeed) (model.AssetGroupLabelSelector, error)
}

func (s *BloodhoundDB) CreateAssetGroupLabelSelector(ctx context.Context, assetGroupLabelId int, userId string, name string, description string, isDefault bool, allowDisable bool, autoCertify bool, seeds []model.SelectorSeed) (model.AssetGroupLabelSelector, error) {
	var selector model.AssetGroupLabelSelector

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf("INSERT INTO %s (asset_group_label_id, created_at, created_by, updated_at, updated_by, name, description, is_default, allow_disable, auto_certify) VALUES (?, NOW(), ?, NOW(), ?, ?, ?, ?, ?, ?) RETURNING *", assetGroupSelectorTable),
			assetGroupLabelId, userId, userId, name, description, isDefault, allowDisable, autoCertify).Scan(&selector); result.Error != nil {
			return CheckError(result)
		} else {
			for _, seed := range seeds {
				if result := tx.Exec(fmt.Sprintf("INSERT INTO %s (selector_id, type, value) VALUES (?, ?, ?)", assetGroupSelectorSeedTable), selector.ID, seed.Type, seed.Value); result.Error != nil {
					return CheckError(result)
				} else {
					selector.Seeds = append(selector.Seeds, model.SelectorSeed{Type: seed.Type, Value: seed.Value})
				}
			}
		}
		return nil
	}); err != nil {
		return model.AssetGroupLabelSelector{}, err
	}
	return selector, nil
}

func (s *BloodhoundDB) GetAssetGroupLabel(ctx context.Context, assetGroupLabelId int) (model.AssetGroupLabel, error) {
	var label model.AssetGroupLabel

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf("SELECT id, asset_group_tier_id, kind_id, name, description, created_at, created_by, updated_at, updated_by, deleted_at, deleted_by FROM %s WHERE id = ?", assetGroupLabelTable), assetGroupLabelId).First(&label); result.Error != nil {
			return CheckError(result)
		}
		return nil
	}); err != nil {
		return model.AssetGroupLabel{}, err
	}
	return label, nil
}

func (s *BloodhoundDB) CreateAssetGroupLabel(ctx context.Context, assetGroupTierId int, userId string, kindId int, name string, description string) (model.AssetGroupLabel, error) {
	var label model.AssetGroupLabel

	if err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if result := tx.Raw(fmt.Sprintf("INSERT INTO %s (asset_group_tier_id, kind_id, name, description, created_at, created_by, updated_at, updated_by) VALUES (?, ?, ?, ?, NOW(), ?, NOW(), ?) RETURNING *", assetGroupLabelTable),
			assetGroupTierId, kindId, name, description, userId, userId).Scan(&label); result.Error != nil {
			return CheckError(result)
		}
		return nil
	}); err != nil {
		return model.AssetGroupLabel{}, err
	}
	return label, nil
}
