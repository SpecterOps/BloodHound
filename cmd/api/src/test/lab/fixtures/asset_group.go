// Copyright 2024 Specter Ops, Inc.
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

package fixtures

import (
	"context"
	"fmt"
	"log"

	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
)

func NewAssetGroupFixture(db *lab.Fixture[*database.BloodhoundDB], name string, tag string, systemGroup bool) *lab.Fixture[*model.AssetGroup] {
	fixture := lab.NewFixture(
		func(h *lab.Harness) (*model.AssetGroup, error) {
			if tx, ok := lab.Unpack(h, db); !ok {
				return nil, fmt.Errorf("unable to unpack BloodhoundDB")
			} else if assetGroup, err := tx.CreateAssetGroup(context.Background(), name, tag, systemGroup); err != nil {
				return nil, fmt.Errorf("unable to create asset group: %w", err)
			} else {
				return &assetGroup, nil
			}
		},
		func(h *lab.Harness, assetGroup *model.AssetGroup) error {
			if tx, ok := lab.Unpack(h, PostgresFixture); !ok {
				return fmt.Errorf("unable to unpack BloodhoundDB")
			} else {
				if err := tx.DeleteAssetGroup(context.Background(), *assetGroup); err != nil {
					return fmt.Errorf("failure deleting asset group: %w", err)
				}
			}
			return nil
		},
	)

	if err := lab.SetDependency(fixture, db); err != nil {
		log.Fatalf("AssetGroupFixture dependency error: %v", err)
	} else if err := lab.SetDependency(fixture, BHAdminApiClientFixture); err != nil {
		log.Fatalf("AssetGroupFixture dependency error: %v", err)
	}

	return fixture
}

func NewAssetGroupSelectorFixture(db *lab.Fixture[*database.BloodhoundDB], assetGroupFixture *lab.Fixture[*model.AssetGroup], selectorName, objectId string) *lab.Fixture[*model.AssetGroupSelector] {
	fixture := lab.NewFixture(
		func(h *lab.Harness) (*model.AssetGroupSelector, error) {
			if assetGroup, ok := lab.Unpack(h, assetGroupFixture); !ok {
				return nil, fmt.Errorf("unable to unpack AssetGroupFixture")
			} else if db, ok := lab.Unpack(h, db); !ok {
				return nil, fmt.Errorf("unable to unpack BloodhoundDB")
			} else if selector, err := db.CreateAssetGroupSelector(
				*assetGroup,
				model.AssetGroupSelectorSpec{
					SelectorName:   selectorName,
					EntityObjectID: objectId,
				},
				false,
			); err != nil {
				return nil, fmt.Errorf("failure creating asset group selector: %v", err)
			} else {
				return &selector, nil
			}
		},
		func(harness *lab.Harness, assetGroupSelector *model.AssetGroupSelector) error {
			if tx, ok := lab.Unpack(harness, PostgresFixture); !ok {
				return fmt.Errorf("unable to unpack PostgresFixture")
			} else {
				if err := tx.DeleteAssetGroupSelector(context.Background(), *assetGroupSelector); err != nil {
					return fmt.Errorf("failure deleting asset group selector")
				}
			}

			return nil
		},
	)

	if err := lab.SetDependency(fixture, assetGroupFixture); err != nil {
		log.Fatalf("AssetGroupSelectorFixture dependency error: %v", err)
	} else if err := lab.SetDependency(fixture, db); err != nil {
		log.Fatalf("AssetGroupSelectorFixture dependency error: %v", err)
	} else if err := lab.SetDependency(fixture, BHAdminApiClientFixture); err != nil {
		log.Fatalf("AssetGroupSelectorFixture dependency error: %v", err)
	}

	return fixture
}
