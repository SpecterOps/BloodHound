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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . AgiData
package agi

import (
	"context"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/model"
)

type AgiData interface {
	GetAllAssetGroups(ctx context.Context, order string, filter model.SQLFilter) (model.AssetGroups, error)
	GetAssetGroup(ctx context.Context, id int32) (model.AssetGroup, error)
	CreateAssetGroupCollection(ctx context.Context, collection model.AssetGroupCollection, entries model.AssetGroupCollectionEntries) error
}

func FetchAssetGroupNodes(tx graph.Transaction, assetGroupTag string, isSystemGroup bool) (graph.NodeSet, error) {
	var (
		assetGroupNodes graph.NodeSet
		tagPropertyStr  = common.SystemTags.String()
		err             error
	)

	if !isSystemGroup {
		tagPropertyStr = common.UserTags.String()
	}

	if assetGroupNodes, err = ops.FetchNodeSet(tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.KindIn(query.Node(), ad.Entity, azure.Entity),
			query.StringContains(query.NodeProperty(tagPropertyStr), assetGroupTag),
		)
	})); err != nil {
		return graph.NodeSet{}, err
	} else {
		// tags are space seperated, so we have to loop and remove any that are not exact matches
		for _, node := range assetGroupNodes {
			tags, _ := node.Properties.Get(tagPropertyStr).String()
			if !slices.Contains(strings.Split(tags, " "), assetGroupTag) {
				assetGroupNodes.Remove(node.ID)
			}
		}
	}

	return assetGroupNodes, err
}

func RunAssetGroupIsolationCollections(ctx context.Context, db AgiData, graphDB graph.Database) error {
	defer log.Measure(log.LevelInfo, "Asset Group Isolation Collections")()

	if assetGroups, err := db.GetAllAssetGroups(ctx, "", model.SQLFilter{}); err != nil {
		return err
	} else {
		return graphDB.WriteTransaction(ctx, func(tx graph.Transaction) error {
			for _, assetGroup := range assetGroups {
				if assetGroupNodes, err := FetchAssetGroupNodes(tx, assetGroup.Tag, assetGroup.SystemGroup); err != nil {
					return err
				} else {
					var (
						entries    = make(model.AssetGroupCollectionEntries, len(assetGroupNodes))
						collection = model.AssetGroupCollection{
							AssetGroupID: assetGroup.ID,
						}
					)

					idx := 0
					for _, node := range assetGroupNodes {
						if objectID, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
							log.Errorf("Node %d that does not have valid %s property", node.ID, common.ObjectID)
						} else {
							entries[idx] = model.AssetGroupCollectionEntry{
								ObjectID:   objectID,
								NodeLabel:  analysis.GetNodeKindDisplayLabel(node),
								Properties: node.Properties.Map,
							}
						}
						idx++
					}

					// Enter a collection, even if it's empty to signal that we did do a tagging/collection run
					if err := db.CreateAssetGroupCollection(ctx, collection, entries); err != nil {
						return err
					}
				}
			}

			return nil
		})
	}
}

func UpdateAssetGroupIsolationTags(ctx context.Context, db AgiData, graphDb graph.Database) error {
	if assetGroups, err := db.GetAllAssetGroups(ctx, "", model.SQLFilter{}); err != nil {
		return err
	} else {
		return graphDb.WriteTransaction(ctx, func(tx graph.Transaction) error {
			for _, assetGroup := range assetGroups {
				if assetGroupNodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
					tagPropertyStr := common.SystemTags.String()

					if !assetGroup.SystemGroup {
						tagPropertyStr = common.UserTags.String()
					}

					return query.And(
						query.KindIn(query.Node(), ad.Entity, azure.Entity),
						query.In(query.NodeProperty(common.ObjectID.String()), assetGroup.Selectors.Strings()),
						query.Not(query.StringContains(query.NodeProperty(tagPropertyStr), assetGroup.Tag)),
					)
				})); err != nil {
					return err
				} else {
					for _, node := range assetGroupNodes {
						tagPropertyStr := common.SystemTags.String()

						if !assetGroup.SystemGroup {
							tagPropertyStr = common.UserTags.String()
						}

						if tags, err := node.Properties.Get(tagPropertyStr).String(); err != nil {
							if graph.IsErrPropertyNotFound(err) {
								node.Properties.Set(tagPropertyStr, assetGroup.Tag)
							} else {
								return err
							}
						} else {
							node.Properties.Set(tagPropertyStr, tags+" "+assetGroup.Tag)
						}

						if err := tx.UpdateNode(node); err != nil {
							return err
						}
					}
				}
			}

			return nil
		})
	}
}
