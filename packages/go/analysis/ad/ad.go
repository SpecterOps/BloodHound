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

package ad

import (
	"context"
	"fmt"
	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/traversal"
	"sort"
	"strings"
	"time"

	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/dawgs/cardinality"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

var (
	ErrNoSuchGroup   = errors.New("no group found")
	AdminGroupSuffix = "-544"
	RDPGroupSuffix   = "-555"
)

const (
	EnterpriseDomainControllersGroupSIDSuffix = "1-5-9"
	AdministratorAccountSIDSuffix             = "-500"
	DomainAdminsGroupSIDSuffix                = "-512"
	DomainControllersGroupSIDSuffix           = "-516"
	SchemaAdminsGroupSIDSuffix                = "-518"
	EnterpriseAdminsGroupSIDSuffix            = "-519"
	KeyAdminsGroupSIDSuffix                   = "-526"
	EnterpriseKeyAdminsGroupSIDSuffix         = "-527"
	AdministratorsGroupSIDSuffix              = "-544"
	BackupOperatorsGroupSIDSuffix             = "-551"
	DomainUsersSuffix                         = "-513"
	AuthenticatedUsersSuffix                  = "-S-1-5-11"
	EveryoneSuffix                            = "-S-1-1-0"
	DomainComputersSuffix                     = "-515"
)

func TierZeroWellKnownSIDSuffixes() []string {
	return []string{
		EnterpriseDomainControllersGroupSIDSuffix,
		AdministratorAccountSIDSuffix,
		DomainAdminsGroupSIDSuffix,
		DomainControllersGroupSIDSuffix,
		SchemaAdminsGroupSIDSuffix,
		EnterpriseAdminsGroupSIDSuffix,
		KeyAdminsGroupSIDSuffix,
		EnterpriseKeyAdminsGroupSIDSuffix,
		BackupOperatorsGroupSIDSuffix,
		AdministratorsGroupSIDSuffix,
	}
}
func FetchWellKnownTierZeroEntities(tx graph.Transaction, domainSID string) (graph.NodeSet, error) {
	nodes := graph.NewNodeSet()

	for _, wellKnownSIDSuffix := range TierZeroWellKnownSIDSuffixes() {
		if err := tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				// Make sure we have the Group or User label. This should cover the case for URA as well as filter out all the other localgroups
				query.KindIn(query.Node(), ad.Group, ad.User),
				query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), wellKnownSIDSuffix),
				query.Equals(query.NodeProperty(ad.DomainSID.String()), domainSID),
			)
		}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				nodes.Add(node)
			}

			return cursor.Error()
		}); err != nil {
			return nil, err
		}
	}

	return nodes, nil
}

func FixWellKnownNodeTypes(ctx context.Context, db graph.Database) error {
	defer log.Measure(log.LevelInfo, "Fix well known node types")()
	groupSuffixes := []string{EnterpriseKeyAdminsGroupSIDSuffix,
		KeyAdminsGroupSIDSuffix,
		EnterpriseDomainControllersGroupSIDSuffix,
		DomainAdminsGroupSIDSuffix,
		DomainControllersGroupSIDSuffix,
		SchemaAdminsGroupSIDSuffix,
		EnterpriseAdminsGroupSIDSuffix,
		AdministratorsGroupSIDSuffix,
		BackupOperatorsGroupSIDSuffix,
	}

	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for _, suffix := range groupSuffixes {
			if nodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
				return query.And(
					query.StringEndsWith(query.NodeProperty(common.ObjectID.String()), suffix),
					query.Not(query.KindIn(query.Node(), ad.Group, ad.LocalGroup)),
				)
			})); err != nil && !graph.IsErrNotFound(err) {
				return err
			} else if graph.IsErrNotFound(err) {
				continue
			} else {
				for _, node := range nodes {
					node.AddKinds(ad.Group)
					if err := tx.UpdateNode(node); err != nil {
						return err
					}
				}
			}
		}

		return nil
	})
}

func RunDomainAssociations(ctx context.Context, db graph.Database) error {
	defer log.Measure(log.LevelInfo, "Domain Associations")()

	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		if domainNamesByObjectID, err := grabDomainInformation(tx); err != nil {
			return fmt.Errorf("error grabbing domain information for association: %w", err)
		} else if unnamedNodes, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
			return query.Not(query.Exists(query.NodeProperty(common.Name.String())))
		})); err != nil {
			return fmt.Errorf("error grabbing unnnamed nodes for association: %w", err)
		} else {
			for _, unnamedNode := range unnamedNodes {
				if nodeObjectID, err := unnamedNode.Properties.Get(common.ObjectID.String()).String(); err == nil {
					if objectIDSuffixIdx := strings.LastIndex(nodeObjectID, "-"); objectIDSuffixIdx >= 0 {
						nodeDomainSID := nodeObjectID[:objectIDSuffixIdx]

						if domainName, hasDomain := domainNamesByObjectID[nodeDomainSID]; hasDomain {
							unnamedNode.Properties.Set(common.Name.String(), fmt.Sprintf("(%s) %s", domainName, nodeObjectID))
							unnamedNode.Properties.Set(ad.DomainSID.String(), nodeDomainSID)

							if err := tx.UpdateNode(unnamedNode); err != nil {
								return fmt.Errorf("error renaming nodes during association: %w", err)
							}
						}
					}
				}
			}
		}

		//TODO: Reimplement unassociated node pruning if we decide that FOSS needs unassociated node pruning
		return nil
	})

}

func grabDomainInformation(tx graph.Transaction) (map[string]string, error) {
	domainNamesByObjectID := make(map[string]string)

	if err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.Kind(query.Node(), ad.Domain)
	}).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			if domainObjectID, err := node.Properties.Get(common.ObjectID.String()).String(); err != nil {
				log.Errorf("Domain node %d does not have a valid object ID", node.ID)
			} else if domainName, err := node.Properties.Get(common.Name.String()).String(); err != nil {
				log.Errorf("Domain node %d does not have a valid name", node.ID)
			} else {
				domainNamesByObjectID[domainObjectID] = domainName
			}
		}

		return cursor.Error()
	}); err != nil {
		return nil, err
	} else {
		return domainNamesByObjectID, nil
	}
}

func LinkWellKnownGroups(ctx context.Context, db graph.Database) error {
	defer log.Measure(log.LevelInfo, "Link well known groups")()

	var (
		errors        = util.NewErrorCollector()
		newProperties = graph.NewProperties()
	)

	if domains, err := GetCollectedDomains(ctx, db); err != nil {
		return err
	} else {
		newProperties.Set(common.LastSeen.String(), time.Now().UTC())

		for _, domain := range domains {
			if domainSid, err := domain.Properties.Get(ad.DomainSID.String()).String(); err != nil {
				log.Errorf("Error getting domain sid for domain %d: %v", domain.ID, err)
			} else if domainName, err := domain.Properties.Get(common.Name.String()).String(); err != nil {
				log.Errorf("Error getting domain name for domain %d: %v", domain.ID, err)
			} else {
				var (
					domainId         = domain.ID
					domainUsersId    = fmt.Sprintf("%s%s", domainSid, DomainUsersSuffix)
					authUsersId      = fmt.Sprintf("%s%s", domainName, AuthenticatedUsersSuffix)
					everyoneId       = fmt.Sprintf("%s%s", domainName, EveryoneSuffix)
					domainComputerId = fmt.Sprintf("%s%s", domainSid, DomainComputersSuffix)
				)

				if err := db.WriteTransaction(ctx, func(tx graph.Transaction) error {
					if domainUserNode, err := getOrCreateWellKnownGroup(tx, domainUsersId, domainSid, domainName, fmt.Sprintf("DOMAIN USERS@%s", domainName)); err != nil {
						return fmt.Errorf("error getting domain users node for domain %d: %w", domainId, err)
					} else if authUsersNode, err := getOrCreateWellKnownGroup(tx, authUsersId, domainSid, domainName, fmt.Sprintf("AUTHENTICATED USERS@%s", domainName)); err != nil {
						return fmt.Errorf("error getting auth users node for domain %d: %w", domainId, err)
					} else if everyoneNode, err := getOrCreateWellKnownGroup(tx, everyoneId, domainSid, domainName, fmt.Sprintf("EVERYONE@%s", domainName)); err != nil {
						return fmt.Errorf("error getting everyone for domain %d: %w", domainId, err)
					} else if domainComputerNode, err := getOrCreateWellKnownGroup(tx, domainComputerId, domainSid, domainName, fmt.Sprintf("DOMAIN COMPUTERS@%s", domainName)); err != nil {
						return fmt.Errorf("error getting domain computers node for domain %d: %w", domainId, err)
					} else if err := createOrUpdateWellKnownLink(tx, domainUserNode, authUsersNode, newProperties); err != nil {
						return err
					} else if err := createOrUpdateWellKnownLink(tx, domainComputerNode, authUsersNode, newProperties); err != nil {
						return err
					} else if err := createOrUpdateWellKnownLink(tx, authUsersNode, everyoneNode, newProperties); err != nil {
						return err
					} else {
						return nil
					}
				}); err != nil {
					log.Errorf("Error linking well known groups for domain %d: %v", domain.ID, err)
					errors.Add(fmt.Errorf("failed linking well known groups for domain %d: %w", domain.ID, err))
				}
			}
		}

		return errors.Combined()
	}
}

func getOrCreateWellKnownGroup(tx graph.Transaction, wellKnownSid string, domainSid, domainName, nodeName string) (*graph.Node, error) {
	if wellKnownNode, err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.NodeProperty(common.ObjectID.String()), wellKnownSid),
			query.Kind(query.Node(), ad.Group),
		)
	}).First(); err != nil && !graph.IsErrNotFound(err) {
		return nil, err
	} else if graph.IsErrNotFound(err) {
		properties := graph.AsProperties(graph.PropertyMap{
			common.Name:     nodeName,
			common.ObjectID: wellKnownSid,
			ad.DomainSID:    domainSid,
			common.LastSeen: time.Now().UTC(),
			ad.DomainFQDN:   domainName,
		})
		return tx.CreateNode(properties, ad.Entity, ad.Group)
	} else {
		return wellKnownNode, nil
	}
}

func createOrUpdateWellKnownLink(tx graph.Transaction, startNode *graph.Node, endNode *graph.Node, props *graph.Properties) error {
	if rel, err := tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), startNode.ID),
			query.Equals(query.EndID(), endNode.ID),
			query.Kind(query.Relationship(), ad.MemberOf),
		)
	}).First(); err != nil && !graph.IsErrNotFound(err) {
		return err
	} else if graph.IsErrNotFound(err) {
		if _, err := tx.CreateRelationshipByIDs(startNode.ID, endNode.ID, ad.MemberOf, props); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		rel.Properties.Set(common.LastSeen.String(), time.Now().UTC())
		return tx.UpdateRelationship(rel)
	}
}

func CalculateCrossProductNodeSetsNew(groupExpansions impact.PathAggregator, nodeSets ...[]*graph.Node) cardinality.Duplex[uint32] {
	if len(nodeSets) < 2 {
		log.Errorf("cross products require at least 2 nodesets")
		return cardinality.NewBitmap32()
	}

	//The intention is that the node sets being passed into this function contain all the first degree principals for control
	var (
		resultEntities = cardinality.NewBitmap32()
		unrollSet      = cardinality.NewBitmap32()
		checkSet       = cardinality.NewBitmap32()
	)

	//We need to fully unroll node sets 1-X into a single bitmap which we will check against
	for _, entity := range nodeSets[1] {
		checkSet.Add(entity.ID.Uint32())
		if entity.Kinds.ContainsOneOf(ad.Group) {
			checkSet.Or(groupExpansions.Cardinality(entity.ID.Uint32()).(cardinality.Duplex[uint32]))
		}
	}

	if len(nodeSets) > 2 {
		for i := 2; i < len(nodeSets); i++ {
			tempSet := cardinality.NewBitmap32()
			for _, entity := range nodeSets[i] {
				tempSet.Add(entity.ID.Uint32())
				if entity.Kinds.ContainsOneOf(ad.Group) {
					tempSet.Or(groupExpansions.Cardinality(entity.ID.Uint32()).(cardinality.Duplex[uint32]))
				}
			}
			checkSet.And(tempSet)
		}
	}

	//checkSet should have all the valid principals from all other sets at this point
	//Check first degree principals in our reference set first
	for _, entity := range nodeSets[0] {
		if checkSet.Contains(entity.ID.Uint32()) {
			resultEntities.Add(entity.ID.Uint32())
		} else if entity.Kinds.ContainsOneOf(ad.Group, ad.LocalGroup) {
			unrollSet.Or(groupExpansions.Cardinality(entity.ID.Uint32()).(cardinality.Duplex[uint32]))
		}
	}

	tempMap := map[uint32]uint64{}
	//Find all the groups in our secondary targets and map them to their cardinality in our expansions
	//Saving off to a map to prevent multiple lookups on the expansions
	//Unhandled error here is irrelevant, we can never return an error
	unrollSet.Each(func(id uint32) (bool, error) {
		//If group expansions contains this ID and its cardinality is > 0, it's a group/localgroup
		idCardinality := groupExpansions.Cardinality(id).Cardinality()
		if idCardinality > 0 {
			tempMap[id] = idCardinality
		}

		return true, nil
	})

	//Save the map keys to a new slice, this represents our list of groups in the expansion
	keys := make([]uint32, len(tempMap))
	i := 0
	for key := range tempMap {
		keys[i] = key
		i++
	}

	//Sort by cardinality we saved in the map, which will give us all the groups sorted by their number of members
	sort.Slice(keys, func(i, j int) bool {
		return tempMap[keys[i]] < tempMap[keys[j]]
	})

	for _, groupId := range keys {
		//If the set doesn't contain our key, it means that we've already encapsulated this group in a previous shortcut so skip it
		if !unrollSet.Contains(groupId) {
			continue
		}
		if checkSet.Contains(groupId) {
			//If this entity is a cross product, add it to result entities, remove the group id from the second set and xor the group's membership with the result set
			resultEntities.Add(groupId)
			unrollSet.Remove(groupId)
			unrollSet.Xor(groupExpansions.Cardinality(groupId).(cardinality.Duplex[uint32]))
		} else {
			//If this isn't a match, remove it from the second set to ensure we don't check it again, but leave its membership
			unrollSet.Remove(groupId)
		}
	}

	unrollSet.Each(func(remainder uint32) (bool, error) {
		if checkSet.Contains(remainder) {
			resultEntities.Add(remainder)
		}

		return true, nil
	})

	return resultEntities
}

func CalculateCrossProductBitmapsNew(groupExpansions impact.PathAggregator, nodeSets ...cardinality.Duplex[uint32]) cardinality.Duplex[uint32] {
	if len(nodeSets) < 2 {
		log.Errorf("cross products require at least 2 nodesets")
		return cardinality.NewBitmap32()
	}

	//The intention is that the node sets being passed into this function contain all the first degree principals for control
	var (
		resultEntities   = cardinality.NewBitmap32()
		unrollSet        = cardinality.NewBitmap32()
		cardinalityCache = map[uint32]uint64{}
		checkSet         = cardinality.NewBitmap32()
	)

	//Take the second of our node sets and unroll it all into a single bitmap
	nodeSets[1].Each(func(id uint32) (bool, error) {
		checkSet.Add(id)
		idCardinality := groupExpansions.Cardinality(id)
		idCardinalityCount := getCardinalityCount(id, idCardinality, cardinalityCache)
		cardinalityCache[id] = idCardinalityCount
		if idCardinalityCount > 0 {
			cardinalityCache[id] = idCardinalityCount
			checkSet.Or(idCardinality.(cardinality.Duplex[uint32]))
		}

		return true, nil
	})

	//If we have more than 2 bitmaps, we need to AND everything together
	if len(nodeSets) > 2 {
		for i := 2; i < len(nodeSets); i++ {
			tempSet := cardinality.NewBitmap32()
			nodeSets[i].Each(func(id uint32) (bool, error) {
				tempSet.Add(id)
				idCardinality := groupExpansions.Cardinality(id)
				idCardinalityCount := getCardinalityCount(id, idCardinality, cardinalityCache)
				cardinalityCache[id] = idCardinalityCount
				if idCardinalityCount > 0 {
					cardinalityCache[id] = idCardinalityCount
					tempSet.Or(idCardinality.(cardinality.Duplex[uint32]))
				}

				return true, nil
			})

			checkSet.And(tempSet)
		}
	}

	//checkSet should have all the valid principals from all other sets at this point
	//Check first degree principals in our reference set first
	nodeSets[0].Each(func(id uint32) (bool, error) {
		if checkSet.Contains(id) {
			resultEntities.Add(id)
		} else {
			idCardinality := groupExpansions.Cardinality(id)
			idCardinalityCount := getCardinalityCount(id, idCardinality, cardinalityCache)
			if idCardinalityCount > 0 {
				cardinalityCache[id] = idCardinalityCount
				unrollSet.Or(idCardinality.(cardinality.Duplex[uint32]))
			}
		}

		return true, nil
	})

	tempMap := map[uint32]uint64{}
	//Find all the groups in our secondary targets and map them to their cardinality in our expansions
	//Saving off to a map to prevent multiple lookups on the expansions
	//Unhandled error here is irrelevant, we can never return an error
	unrollSet.Each(func(id uint32) (bool, error) {
		//If group expansions contains this ID and its cardinality is > 0, it's a group/localgroup
		idCardinality := groupExpansions.Cardinality(id).Cardinality()
		if idCardinality > 0 {
			tempMap[id] = idCardinality
		}

		return true, nil
	})

	//Save the map keys to a new slice, this represents our list of groups in the expansion
	keys := make([]uint32, len(tempMap))
	i := 0
	for key := range tempMap {
		keys[i] = key
		i++
	}

	//Sort by cardinality we saved in the map, which will give us all the groups sorted by their number of members
	sort.Slice(keys, func(i, j int) bool {
		return tempMap[keys[i]] < tempMap[keys[j]]
	})

	for _, groupId := range keys {
		//If the set doesn't contain our key, it means that we've already encapsulated this group in a previous shortcut so skip it
		if !unrollSet.Contains(groupId) {
			continue
		}
		if checkSet.Contains(groupId) {
			//If this entity is a cross product, add it to result entities, remove the group id from the second set and xor the group's membership with the result set
			resultEntities.Add(groupId)
			unrollSet.Remove(groupId)
			unrollSet.Xor(groupExpansions.Cardinality(groupId).(cardinality.Duplex[uint32]))
		} else {
			//If this isn't a match, remove it from the second set to ensure we don't check it again, but leave its membership
			unrollSet.Remove(groupId)
		}
	}

	unrollSet.Each(func(remainder uint32) (bool, error) {
		if checkSet.Contains(remainder) {
			resultEntities.Add(remainder)
		}

		return true, nil
	})

	return resultEntities
}

func getCardinalityCount(id uint32, expansions cardinality.Provider[uint32], cardinalityCache map[uint32]uint64) uint64 {
	if idCardinality, ok := cardinalityCache[id]; ok {
		return idCardinality
	} else {
		idCardinality = expansions.Cardinality()
		return idCardinality
	}
}

func CalculateCrossProductBitmaps(firstNodeSet, secondNodeSet cardinality.Duplex[uint32], groupExpansions impact.PathAggregator) cardinality.Duplex[uint32] {
	//The intention is that the node sets being passed into this function contain all the first degree principals for control
	var (
		resultEntities   = cardinality.NewBitmap32()
		firstSetUnroll   = cardinality.NewBitmap32()
		secondSetUnroll  = cardinality.NewBitmap32()
		cardinalityCache = map[uint32]uint64{}
	)

	//Unroll all groups in our first nodeset into a bitmap
	firstNodeSet.Each(func(id uint32) (bool, error) {
		firstSetUnroll.Add(id)
		idCardinality := groupExpansions.Cardinality(id)
		idCardinalityCount := idCardinality.Cardinality()
		if idCardinalityCount > 0 {
			cardinalityCache[id] = idCardinalityCount
			firstSetUnroll.Or(idCardinality.(cardinality.Duplex[uint32]))
		}

		return true, nil
	})

	secondNodeSet.Each(func(id uint32) (bool, error) {
		if firstSetUnroll.Contains(id) {
			resultEntities.Add(id)
		} else {
			idCardinality := groupExpansions.Cardinality(id)
			idCardinalityCount := idCardinality.Cardinality()
			if idCardinalityCount > 0 {
				cardinalityCache[id] = idCardinalityCount
				secondSetUnroll.Or(idCardinality.(cardinality.Duplex[uint32]))
			}
		}
		return true, nil
	})

	tempMap := map[uint32]uint64{}
	//Find all the groups in our secondary targets and map them to their cardinality in our expansions
	//Saving off to a map to prevent multiple lookups on the expansions
	//Unhandled error here is irrelevant, we can never return an error
	secondSetUnroll.Each(func(id uint32) (bool, error) {
		//If group expansions contains this ID and its cardinality is > 0, it's a group/localgroup
		idCardinality := groupExpansions.Cardinality(id).Cardinality()
		if idCardinality > 0 {
			tempMap[id] = idCardinality
		}

		return true, nil
	})

	//Save the map keys to a new slice, this represents our list of groups in the expansion
	keys := make([]uint32, len(tempMap))
	i := 0
	for key := range tempMap {
		keys[i] = key
		i++
	}

	//Sort by cardinality we saved in the map, which will give us all the groups sorted by their number of members
	sort.Slice(keys, func(i, j int) bool {
		return tempMap[keys[i]] < tempMap[keys[j]]
	})

	for _, groupId := range keys {
		//If the set doesn't contain our key, it means that we've already encapsulated this group in a previous shortcut so skip it
		if !secondSetUnroll.Contains(groupId) {
			continue
		}
		if firstSetUnroll.Contains(groupId) {
			//If this entity is a cross product, add it to result entities, remove the group id from the second set and xor the group's membership with the result set
			resultEntities.Add(groupId)
			secondSetUnroll.Remove(groupId)
			secondSetUnroll.Xor(groupExpansions.Cardinality(groupId).(cardinality.Duplex[uint32]))
		} else {
			//If this isn't a match, remove it from the second set to ensure we don't check it again, but leave its membership
			secondSetUnroll.Remove(groupId)
		}
	}

	secondSetUnroll.Each(func(remainder uint32) (bool, error) {
		if firstSetUnroll.Contains(remainder) {
			resultEntities.Add(remainder)
		}

		return true, nil
	})

	return resultEntities
}

func CalculateCrossProductNodeSets(firstNodeSet, secondNodeSet []*graph.Node, groupExpansions impact.PathAggregator) cardinality.Duplex[uint32] {
	//The intention is that the node sets being passed into this function contain all the first degree principals for control
	var (
		resultEntities  = cardinality.NewBitmap32()
		firstSetUnroll  = cardinality.NewBitmap32()
		secondSetUnroll = cardinality.NewBitmap32()
	)

	//Unroll all groups in our first nodeset into a bitmap
	for _, entity := range firstNodeSet {
		firstSetUnroll.Add(entity.ID.Uint32())
		if entity.Kinds.ContainsOneOf(ad.Group) {
			firstSetUnroll.Or(groupExpansions.Cardinality(entity.ID.Uint32()).(cardinality.Duplex[uint32]))
		}
	}

	//Check all our first degree entities directly for matches
	for _, entity := range secondNodeSet {
		if firstSetUnroll.Contains(entity.ID.Uint32()) {
			resultEntities.Add(entity.ID.Uint32())
		} else if entity.Kinds.ContainsOneOf(ad.Group, ad.LocalGroup) {
			//If our entity doesn't match directly, but is a group or local group we want to expand its membership into a map
			secondSetUnroll.Or(groupExpansions.Cardinality(entity.ID.Uint32()).(cardinality.Duplex[uint32]))
		}
	}

	tempMap := map[uint32]uint64{}
	//Find all the groups in our secondary targets and map them to their cardinality in our expansions
	//Saving off to a map to prevent multiple lookups on the expansions
	//Unhandled error here is irrelevant, we can never return an error
	secondSetUnroll.Each(func(id uint32) (bool, error) {
		//If group expansions contains this ID and its cardinality is > 0, it's a group/localgroup
		idCardinality := groupExpansions.Cardinality(id).Cardinality()
		if idCardinality > 0 {
			tempMap[id] = idCardinality
		}

		return true, nil
	})

	//Save the map keys to a new slice, this represents our list of groups in the expansion
	keys := make([]uint32, len(tempMap))
	i := 0
	for key := range tempMap {
		keys[i] = key
		i++
	}

	//Sort by cardinality we saved in the map, which will give us all the groups sorted by their number of members
	sort.Slice(keys, func(i, j int) bool {
		return tempMap[keys[i]] < tempMap[keys[j]]
	})

	for _, groupId := range keys {
		//If the set doesn't contain our key, it means that we've already encapsulated this group in a previous shortcut so skip it
		if !secondSetUnroll.Contains(groupId) {
			continue
		}
		if firstSetUnroll.Contains(groupId) {
			//If this entity is a cross product, add it to result entities, remove the group id from the second set and xor the group's membership with the result set
			resultEntities.Add(groupId)
			secondSetUnroll.Remove(groupId)
			secondSetUnroll.Xor(groupExpansions.Cardinality(groupId).(cardinality.Duplex[uint32]))
		} else {
			//If this isn't a match, remove it from the second set to ensure we don't check it again, but leave its membership
			secondSetUnroll.Remove(groupId)
		}
	}

	secondSetUnroll.Each(func(remainder uint32) (bool, error) {
		if firstSetUnroll.Contains(remainder) {
			resultEntities.Add(remainder)
		}

		return true, nil
	})

	return resultEntities
}

func GetEdgeCompositionPath(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	pathSet := graph.NewPathSet()
	return pathSet, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if edge.Kind == ad.GoldenCert {
			if results, err := getGoldenCertEdgeComposition(tx, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC1 {
			if results, err := getADCSESC1EdgeComposition(tx, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		}
		return nil
	})
}

var (
	esc3EdgeCompSegment1Kinds      = graph.Kinds{ad.GenericAll, ad.Enroll, ad.AllExtendedRights, ad.MemberOf}
	esc3EdgeCompPath1Segment2Kinds = graph.Kinds{ad.PublishedTo, ad.IssuedSignedBy, ad.EnterpriseCAFor, ad.RootCAFor}
	esc3EdgeCompPath2Segment2Kinds = graph.Kinds{ad.PublishedTo, ad.TrustedForNTAuth, ad.NTAuthStoreFor}
)

func GetADCSESC3EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH p1 = (x)-[:MemberOf|GenericAll|Enroll|AllExtendedRights]->(ct1:CertTemplate)-[PublishedTo]->(eca1:EnterpriseCA)-[:IssuedSignedBy|EnterpriseCAFor*1..]->(rca:RootCA)-[:RootCAFor]->(d:Domain)
		WHERE x.objectid = "<principal objectid>"
		WHERE d.objectid = "<domain objectid>">
		AND ct1.requiresmanagerapproval = false
		AND ct1.schemaversion = 1 OR (ct1.schemaversion > 1 AND ct1.authorizedsignatures = 0)
		MATCH p2 = (x)-[:GenericAll|Enroll|AllExtendedRights]->(ct2:CertTemplate)-[PublishedTo]->(eca2:EnterpriseCA)-[:TrustedForNTAuth]->(:NTAuthStore)-[:NTAuthStoreFor]->(d)
		WHERE ct2.authenticationenabled = true
		AND ct2.requiresmanagerapproval = false
		MATCH p3 = (ct1)-[:EnrollOnBehalfOf]->(ct2)
		MATCH p4 = (x)-[:Enroll|MemberOf*..]->(eca1)
		MATCH p5 = (x)-[:Enroll|MemberOf*..]->(eca2)
		RETURN p1,p2,p3,p4,p5


		Additionally, if eca2.enrollmentagentrestrictionscollected == true AND eca2.hasenrollmentagentrestrictions == true then we should include an additional path:
		MATCH p6 = (x)-[:DelegatedEnrollmentAgent]->(ct2)
	*/

	var (
		paths           graph.PathSet
		traversalInst   = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		traversalBitmap = cardinality.NewBitmap32()
		ct1Nodes        = cardinality.NewBitmap32()
		ct2Nodes        = cardinality.NewBitmap32()
		ecaNodes        = cardinality.NewBitmap32()
		p6CTNodes       = cardinality.NewBitmap32()
	)

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: graph.NewNode(edge.StartID, graph.NewProperties(), ad.Entity),
		Driver: func(ctx context.Context, tx graph.Transaction, segment *graph.PathSegment) ([]*graph.PathSegment, error) {
			var (
				criteria = []graph.Criteria{
					query.Equals(query.StartID(), segment.Node.ID),
				}
				sawControlRel = false
			)

			// Don't revisit nodes
			if !traversalBitmap.CheckedAdd(segment.Node.ID.Uint32()) {
				return nil, nil
			}

			var (
				path = segment.Path()
			)

			// Walk the path to check for a control relationship
			path.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
				sawControlRel = relationship.Kind.Is(ad.ACLRelationships()...)
				return !sawControlRel
			})

			// Swap query criteria depending on the of the traversal
			if sawControlRel {
				// Don't expand this path any further
				if segment.Edge.Kind.Is(ad.MemberOf) {
					return nil, nil
				}

				// Switch to expanding ADCS kinds
				criteria = append(criteria, query.KindIn(query.Relationship(), esc3EdgeCompPath1Segment2Kinds...))
			} else {
				// Expand membership and ACL kinds
				criteria = append(criteria, query.KindIn(query.Relationship(), esc3EdgeCompSegment1Kinds...))
			}

			// Is this node the terminal node?
			if segment.Node.ID == edge.EndID {
				// Add the path
				paths.AddPath(path)

				path.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
					// Capture the cert template nodes
					if end.Kinds.ContainsOneOf(ad.CertTemplate) {
						ct1Nodes.Add(end.ID.Uint32())
					}

					// Capture the enterprise CA nodes
					if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
						ecaNodes.Add(end.ID.Uint32())
					}

					return true
				})

				return nil, nil
			}

			// Is this node a cert template?
			if segment.Node.Kinds.ContainsOneOf(ad.CertTemplate) && !isStartCertTemplateValidESC3(segment.Node) {
				return nil, nil
			}

			var (
				nextSegments []*graph.PathSegment
				nextQuery    = tx.Relationships().Filter(query.And(criteria...))
			)

			// Order by relationship ID so that skip and limit behave somewhat predictably - cost of this is pretty
			// small even for large result sets
			nextQuery.OrderBy(query.Order(query.Identity(query.Relationship()), query.Ascending()))

			return nextSegments, nextQuery.FetchDirection(graph.DirectionInbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
				for next := range cursor.Chan() {
					nextSegments = append(nextSegments, segment.Descend(next.Node, next.Relationship))
				}

				return cursor.Error()
			})
		},
	}); err != nil {
		return nil, err
	}

	traversalBitmap.Clear()

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: graph.NewNode(edge.StartID, graph.NewProperties(), ad.Entity),
		Driver: func(ctx context.Context, tx graph.Transaction, segment *graph.PathSegment) ([]*graph.PathSegment, error) {
			var (
				criteria = []graph.Criteria{
					query.Equals(query.StartID(), segment.Node.ID),
				}
				sawControlRel = false
			)

			// Don't revisit nodes
			if !traversalBitmap.CheckedAdd(segment.Node.ID.Uint32()) {
				return nil, nil
			}

			var (
				path = segment.Path()
			)

			// Walk the path to check for a control relationship
			path.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
				sawControlRel = relationship.Kind.Is(ad.ACLRelationships()...)
				return !sawControlRel
			})

			// Swap query criteria depending on the of the traversal
			if sawControlRel {
				// Don't expand this path any further
				if segment.Edge.Kind.Is(ad.MemberOf) {
					return nil, nil
				}

				// Switch to expanding ADCS kinds
				criteria = append(criteria, query.KindIn(query.Relationship(), esc3EdgeCompPath2Segment2Kinds...))
			} else {
				// Expand membership and ACL kinds
				criteria = append(criteria, query.KindIn(query.Relationship(), esc3EdgeCompSegment1Kinds...))
			}

			// Is this node the terminal node?
			if segment.Node.ID == edge.EndID {
				// Add the path
				paths.AddPath(path)

				path.Walk(func(start, end *graph.Node, relationship *graph.Relationship) bool {
					// Capture the cert template nodes
					if end.Kinds.ContainsOneOf(ad.CertTemplate) {
						ct2Nodes.Add(end.ID.Uint32())
					}

					// Capture the enterprise CA nodes
					if end.Kinds.ContainsOneOf(ad.EnterpriseCA) {
						ecaNodes.Add(end.ID.Uint32())

						// In addition, to account for P6 we must track CertTemplates that are published to ECA nodes with certain conditions
						if collected, err := end.Properties.Get(ad.EnrollmentAgentRestrictionsCollected.String()).Bool(); err != nil {
							log.Errorf("error getting enrollmentagentcollected for eca2 %d: %w", end.ID, err)
						} else if hasRestrictions, err := end.Properties.Get(ad.HasEnrollmentAgentRestrictions.String()).Bool(); err != nil {
							log.Errorf("error getting hasenrollmentagentrestrictions for ca %d: %w", end.ID, err)
						} else if collected && hasRestrictions {
							// The start node here MUST be a CertTemplate as it is the only node kind that can traverse PublishedTo to an EnterpriseCA
							p6CTNodes.Add(start.ID.Uint32())
						}
					}

					return true
				})

				return nil, nil
			}

			// Is this node a cert template?
			if segment.Node.Kinds.ContainsOneOf(ad.CertTemplate) && !isEndCertTemplateValidESC3(segment.Node) {
				return nil, nil
			}

			var (
				nextSegments []*graph.PathSegment
				nextQuery    = tx.Relationships().Filter(query.And(criteria...))
			)

			// Order by relationship ID so that skip and limit behave somewhat predictably - cost of this is pretty
			// small even for large result sets
			nextQuery.OrderBy(query.Order(query.Identity(query.Relationship()), query.Ascending()))

			return nextSegments, nextQuery.FetchDirection(graph.DirectionInbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
				for next := range cursor.Chan() {
					nextSegments = append(nextSegments, segment.Descend(next.Node, next.Relationship))
				}

				return cursor.Error()
			})
		},
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: graph.NewNode(edge.StartID, graph.NewProperties(), ad.Entity),
		Driver: func(ctx context.Context, tx graph.Transaction, segment *graph.PathSegment) ([]*graph.PathSegment, error) {
			if ecaNodes.Contains(segment.Node.ID.Uint32()) {
				// We reached one of the ECA nodes
				paths.AddPath(segment.Path())
				return nil, nil
			}

			var (
				nextSegments []*graph.PathSegment
				nextQuery    = tx.Relationships().Filter(query.And(
					query.Equals(query.StartID(), segment.Node.ID),
					query.KindIn(query.Relationship(), ad.Enroll, ad.MemberOf),
				))
			)

			// Order by relationship ID so that skip and limit behave somewhat predictably - cost of this is pretty
			// small even for large result sets
			nextQuery.OrderBy(query.Order(query.Identity(query.Relationship()), query.Ascending()))

			return nextSegments, nextQuery.FetchDirection(graph.DirectionInbound, func(cursor graph.Cursor[graph.DirectionalResult]) error {
				for next := range cursor.Chan() {
					nextSegments = append(nextSegments, segment.Descend(next.Node, next.Relationship))
				}

				return cursor.Error()
			})
		},
	}); err != nil {
		return nil, err
	}

	return paths, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if templateEscalationPaths, err := ops.FetchPathSet(tx, tx.Relationships().Filter(query.And(
			query.InIDs(query.StartID(), cardinality.DuplexToGraphIDs(ct1Nodes)...),
			query.Kind(query.Relationship(), ad.EnrollOnBehalfOf),
			query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(ct2Nodes)...),
		))); err != nil {
			return err
		} else {
			paths.AddPathSet(templateEscalationPaths)
		}

		if p6CTNodes.Cardinality() > 0 {
			if delegatedEnrollmentEscalationPaths, err := ops.FetchPathSet(tx, tx.Relationships().Filter(query.And(
				query.Equals(query.StartID(), edge.StartID),
				query.Kind(query.Relationship(), ad.DelegatedEnrollmentAgent),
				query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(p6CTNodes)...),
			))); err != nil {
				return err
			} else {
				paths.AddPathSet(delegatedEnrollmentEscalationPaths)
			}
		}

		return nil
	})
}

func getADCSESC1EdgeComposition(tx graph.Transaction, edge *graph.Relationship) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()
	//Grab the start node as well as the target domain node
	if startNode, targetDomainNode, err := ops.FetchRelationshipNodes(tx, edge); err != nil {
		return finalPaths, err
	} else {
		//Find cert templates that we have control over using enroll/acls
		if pathsToTemplates, err := ops.TraversePaths(tx, ops.TraversalPlan{
			Root:      startNode,
			Direction: graph.DirectionOutbound,
			BranchQuery: func() graph.Criteria {
				return query.KindIn(query.Relationship(), ad.Enroll, ad.GenericAll, ad.AllExtendedRights, ad.MemberOf)
			},
			DescentFilter: OutboundControlDescentFilter,
			PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
				node := segment.Node
				if !node.Kinds.ContainsOneOf(ad.CertTemplate) {
					return false
				} else if props, err := getValidatePublishedCertTemplateForEsc1PropertyValues(node); err != nil {
					log.Errorf("Error getting props for certtemplate %d: %v", node.ID, err)
					return false
				} else if !validatePublishedCertTemplateForEsc1(props) {
					return false
				}

				return true
			},
		}); err != nil {
			log.Errorf("Error getting paths from start node %d to templates: %v", startNode.ID, err)
			return finalPaths, err
		} else {
			for _, path := range pathsToTemplates {
				certTemplate := path.Terminal()
				if ecaPaths, err := ops.FetchPathSet(tx, tx.Relationships().Filter(FilterPublishedCAs(certTemplate))); err != nil {
					log.Errorf("error getting eca published to for cert template %d : %v", certTemplate.ID, err)
				} else {
					for _, ecaPath := range ecaPaths {
						eca := ecaPath.Terminal()
						if domainPaths, err := FetchEnterpriseCAsCertChainPathToDomain(tx, eca, targetDomainNode); err != nil {
							log.Errorf("error getting eca %d path to domain %d: %v", eca.ID, targetDomainNode.ID, err)
						} else if domainPaths.Len() == 0 {
							continue
						} else if trustedForAuthPaths, err := FetchEnterpriseCAsTrustedForAuthPathToDomain(tx, eca, targetDomainNode); err != nil {
							log.Errorf("error getting eca %d path to domain %d via trusted for auth: %v", eca.ID, targetDomainNode.ID, err)
						} else if trustedForAuthPaths.Len() == 0 {
							continue
						} else if userPathsToCa, err := ops.TraversePaths(tx, ops.TraversalPlan{
							Root:      startNode,
							Direction: graph.DirectionOutbound,
							BranchQuery: func() graph.Criteria {
								return query.KindIn(query.Relationship(), ad.Enroll, ad.MemberOf)
							},
							DescentFilter: OutboundControlDescentFilter,
							PathFilter: func(ctx *ops.TraversalContext, segment *graph.PathSegment) bool {
								return segment.Node.ID == eca.ID
							},
						}); err != nil {
							log.Errorf("Error getting paths from start node %d to enterprise ca: %v", startNode.ID, err)
						} else if userPathsToCa.Len() == 0 {
							continue
						} else {
							finalPaths.AddPath(path)
							finalPaths.AddPath(ecaPath)
							finalPaths.AddPathSet(domainPaths)
							finalPaths.AddPathSet(trustedForAuthPaths)
							finalPaths.AddPathSet(userPathsToCa)
						}
					}
				}
			}
		}

		return finalPaths, nil
	}
}

func getGoldenCertEdgeComposition(tx graph.Transaction, edge *graph.Relationship) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()
	//Grab the start node (computer) as well as the target domain node
	if startNode, targetDomainNode, err := ops.FetchRelationshipNodes(tx, edge); err != nil {
		return finalPaths, err
	} else {
		//Find hosted enterprise CA
		if ecaPaths, err := ops.FetchPathSet(tx, tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), startNode.ID),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.KindIn(query.Relationship(), ad.HostsCAService),
		))); err != nil {
			log.Errorf("error getting hostscaservice edge to enterprise ca for computer %d : %w", startNode.ID, err)
		} else {
			for _, ecaPath := range ecaPaths {
				eca := ecaPath.Terminal()
				if chainToRootCAPaths, err := FetchEnterpriseCAsCertChainPathToDomain(tx, eca, targetDomainNode); err != nil {
					log.Errorf("error getting eca %d path to domain %d: %w", eca.ID, targetDomainNode.ID, err)
				} else if chainToRootCAPaths.Len() == 0 {
					continue
				} else if trustedForAuthPaths, err := FetchEnterpriseCAsTrustedForAuthPathToDomain(tx, eca, targetDomainNode); err != nil {
					log.Errorf("error getting eca %d path to domain %d via trusted for auth: %w", eca.ID, targetDomainNode.ID, err)
				} else if trustedForAuthPaths.Len() == 0 {
					continue
				} else {
					finalPaths.AddPath(ecaPath)
					finalPaths.AddPathSet(chainToRootCAPaths)
					finalPaths.AddPathSet(trustedForAuthPaths)
				}
			}
		}

		return finalPaths, nil
	}
}
