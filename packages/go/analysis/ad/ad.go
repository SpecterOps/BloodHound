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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/traversal"

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

func FetchWellKnownTierZeroEntities(ctx context.Context, db graph.Database, domainSID string) (graph.NodeSet, error) {
	defer log.Measure(log.LevelInfo, "FetchWellKnownTierZeroEntities")()

	nodes := graph.NewNodeSet()

	return nodes, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
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
				return err
			}
		}

		return nil
	})
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

func CalculateCrossProductNodeSets(groupExpansions impact.PathAggregator, nodeSets ...[]*graph.Node) cardinality.Duplex[uint32] {
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
	unrollSet.Each(func(id uint32) bool {
		//If group expansions contains this ID and its cardinality is > 0, it's a group/localgroup
		idCardinality := groupExpansions.Cardinality(id).Cardinality()
		if idCardinality > 0 {
			tempMap[id] = idCardinality
		}

		return true
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

	unrollSet.Each(func(remainder uint32) bool {
		if checkSet.Contains(remainder) {
			resultEntities.Add(remainder)
		}

		return true
	})

	return resultEntities
}

func CalculateCrossProductBitmaps(groupExpansions impact.PathAggregator, nodeSets ...cardinality.Duplex[uint32]) cardinality.Duplex[uint32] {
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
	nodeSets[1].Each(func(id uint32) bool {
		checkSet.Add(id)
		idCardinality := groupExpansions.Cardinality(id)
		idCardinalityCount := getCardinalityCount(id, idCardinality, cardinalityCache)
		cardinalityCache[id] = idCardinalityCount
		if idCardinalityCount > 0 {
			checkSet.Or(idCardinality.(cardinality.Duplex[uint32]))
		}

		return true
	})

	//If we have more than 2 bitmaps, we need to AND everything together
	if len(nodeSets) > 2 {
		for i := 2; i < len(nodeSets); i++ {
			tempSet := cardinality.NewBitmap32()
			nodeSets[i].Each(func(id uint32) bool {
				tempSet.Add(id)
				idCardinality := groupExpansions.Cardinality(id)
				idCardinalityCount := getCardinalityCount(id, idCardinality, cardinalityCache)
				cardinalityCache[id] = idCardinalityCount
				if idCardinalityCount > 0 {
					tempSet.Or(idCardinality.(cardinality.Duplex[uint32]))
				}

				return true
			})

			checkSet.And(tempSet)
		}
	}

	//checkSet should have all the valid principals from all other sets at this point
	//Check first degree principals in our reference set first
	nodeSets[0].Each(func(id uint32) bool {
		if checkSet.Contains(id) {
			resultEntities.Add(id)
		} else {
			idCardinality := groupExpansions.Cardinality(id)
			idCardinalityCount := getCardinalityCount(id, idCardinality, cardinalityCache)
			cardinalityCache[id] = idCardinalityCount
			if idCardinalityCount > 0 {
				unrollSet.Or(idCardinality.(cardinality.Duplex[uint32]))
			}
		}

		return true
	})

	tempMap := map[uint32]uint64{}
	//Find all the groups in our secondary targets and map them to their cardinality in our expansions
	//Saving off to a map to prevent multiple lookups on the expansions
	//Unhandled error here is irrelevant, we can never return an error
	unrollSet.Each(func(id uint32) bool {
		//If group expansions contains this ID and its cardinality is > 0, it's a group/localgroup
		idCardinality := groupExpansions.Cardinality(id).Cardinality()
		if idCardinality > 0 {
			tempMap[id] = idCardinality
		}

		return true
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

	unrollSet.Each(func(remainder uint32) bool {
		if checkSet.Contains(remainder) {
			resultEntities.Add(remainder)
		}

		return true
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
			if results, err := GetADCSESC1EdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC3 {
			if results, err := GetADCSESC3EdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC6a {
			if results, err := GetADCSESC6aEdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC9a {
			if results, err := GetADCSESC9aEdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		} else if edge.Kind == ad.ADCSESC10a {
			if results, err := GetADCSESC10aEdgeComposition(ctx, db, edge); err != nil {
				return err
			} else {
				pathSet = results
			}
		}
		return nil
	})
}

func ADCSESC3Path1Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.And(
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC3Path2Pattern(domainId graph.ID, enterpriseCAs, candidateTemplates cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.KindIn(query.End(), ad.CertTemplate),
			query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
			query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
			query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(candidateTemplates)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...))).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC3Path3Pattern() traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.KindIn(query.Relationship(), ad.Enroll),
		))
}

func ADCSESC6aPath1Pattern() traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.Equals(query.EndProperty(ad.IsUserSpecifiesSanEnabled.String()), true),
			query.KindIn(query.Relationship(), ad.Enroll),
		))
}

func ADCSESC6aPath2Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.And(
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Equals(query.EndProperty(ad.NoSecurityExtension.String()), true),
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC6aPath3Pattern(domainId graph.ID, enterpriseCAs, candidateTemplates cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.KindIn(query.End(), ad.CertTemplate),
			query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(candidateTemplates)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...))).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainId),
		))
}

func ADCSESC6aPath4Pattern(domainId graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		Outbound(
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.Enroll),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.InIDs(query.End(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.CanAbuseWeakCertBinding),
			query.KindIn(query.End(), ad.Computer),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.DCFor, ad.TrustedBy),
			query.Equals(query.EndID(), domainId),
		))
}

func GetADCSESC6aEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		closureErr           error
		startNode            *graph.Node
		traversalInst        = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		lock                 = &sync.Mutex{}
		paths                = graph.PathSet{}
		certTemplateSegments = map[graph.ID][]*graph.PathSegment{}
		enterpriseCASegments = map[graph.ID][]*graph.PathSegment{}
		certTemplates        = cardinality.NewBitmap32()
		enterpriseCAs        = cardinality.NewBitmap32()
		path1EnterpriseCAs   = cardinality.NewBitmap32()
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else {
			startNode = node
			return nil
		}
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath1Pattern().Do(func(terminal *graph.PathSegment) error {
			enterpriseCA := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			path1EnterpriseCAs.Add(enterpriseCA.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath2Pattern(edge.EndID, path1EnterpriseCAs).Do(func(terminal *graph.PathSegment) error {
			certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
			certTemplates.Add(certTemplate.ID.Uint32())
			lock.Unlock()

			return nil
		})}); err != nil {
		return nil, err
	}
	log.Infof("certtemplates %v", certTemplates.Slice())

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath3Pattern(edge.EndID, path1EnterpriseCAs, certTemplates).Do(func(terminal *graph.PathSegment) error {
			certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			certTemplateSegments[certTemplate.ID] = append(certTemplateSegments[certTemplate.ID], terminal)
			certTemplates.Add(certTemplate.ID.Uint32())
			lock.Unlock()

			return nil
		})}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC6aPath4Pattern(edge.EndID, path1EnterpriseCAs).Do(func(terminal *graph.PathSegment) error {
			enterpriseCA := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			paths.AddPath(terminal.Path())
			enterpriseCASegments[enterpriseCA.ID] = append(enterpriseCASegments[enterpriseCA.ID], terminal)
			enterpriseCAs.Add(enterpriseCA.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	email, err := startNode.Properties.Get(common.Email.String()).String()
	if err != nil {
		log.Warnf("unable to access property %s for node with id %d: %v", common.Email.String(), startNode.ID, err)
	}

	certTemplates.Each(func(value uint32) bool {
		var certTemplate *graph.Node

		if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
			if node, err := ops.FetchNode(tx, graph.ID(value)); err != nil {
				return err
			} else {
				certTemplate = node
				return nil
			}
		}); err != nil {
			closureErr = fmt.Errorf("could not fetch cert template node: %w", err)
			return false
		}

		schemaVersion, err := certTemplate.Properties.Get(ad.SchemaVersion.String()).Float64()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SchemaVersion.String(), certTemplate.ID, err)
		}
		subjectAltRequireEmail, err := certTemplate.Properties.Get(ad.SubjectAltRequireEmail.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectAltRequireEmail.String(), certTemplate.ID, err)
		}
		subjectRequireEmail, err := certTemplate.Properties.Get(ad.SubjectRequireEmail.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectRequireEmail.String(), certTemplate.ID, err)
		}
		subjectAltRequireDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectAltRequireDNS.String(), certTemplate.ID, err)
		}
		subjectAltRequireDomainDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDomainDNS.String()).Bool()
		if err != nil {
			log.Warnf("unable to access property %s for certTemplate with id %d: %v", ad.SubjectAltRequireDomainDNS.String(), certTemplate.ID, err)
		}

		for _, segment := range certTemplateSegments[graph.ID(value)] {
			if startNode.Kinds.ContainsOneOf(ad.User) {
				if subjectAltRequireDNS || subjectAltRequireDomainDNS {
					continue
				} else if email == "" && !((!subjectAltRequireEmail && !subjectRequireEmail) || schemaVersion == 1) {
					continue
				} else {
					log.Infof("Found ESC6a Path: %s", graph.FormatPathSegment(segment))
					paths.AddPath(segment.Path())
				}
			} else if startNode.Kinds.ContainsOneOf(ad.Computer) {
				if email == "" && !((!subjectAltRequireEmail && !subjectRequireEmail) || schemaVersion == 1) {
					continue
				} else {
					log.Infof("Found ESC6a Path: %s", graph.FormatPathSegment(segment))
					paths.AddPath(segment.Path())
				}
			} else {
				log.Infof("Found ESC6a Path: %s", graph.FormatPathSegment(segment))
				paths.AddPath(segment.Path())
			}

		}

		return true
	})

	if closureErr != nil {
		return paths, closureErr
	}

	if paths.Len() > 0 {
		enterpriseCAs.Each(func(value uint32) bool {
			for _, segment := range enterpriseCASegments[graph.ID(value)] {
				paths.AddPath(segment.Path())
			}
			return true
		})
	}

	return paths, nil
}

func GetADCSESC3EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode *graph.Node

		traversalInst           = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                   = graph.PathSet{}
		path1CandidateSegments  = map[graph.ID][]*graph.PathSegment{}
		path2CandidateSegments  = map[graph.ID][]*graph.PathSegment{}
		lock                    = &sync.Mutex{}
		path1CertTemplates      = cardinality.NewBitmap32()
		path2CertTemplates      = cardinality.NewBitmap32()
		enterpriseCANodes       = cardinality.NewBitmap32()
		enterpriseCASegments    = map[graph.ID][]*graph.PathSegment{}
		path2CandidateTemplates = cardinality.NewBitmap32()
		enrollOnBehalfOfPaths   graph.PathSet
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else {
			startNode = node
			return nil
		}
	}); err != nil {
		return nil, err
	}

	//Start by fetching all EnterpriseCA nodes that our user has Enroll rights on via group membership or directly (P4/P5)
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC3Path3Pattern().Do(func(terminal *graph.PathSegment) error {
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			enterpriseCASegments[enterpriseCANode.ID] = append(enterpriseCASegments[enterpriseCANode.ID], terminal)
			enterpriseCANodes.Add(enterpriseCANode.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	//Use the enterprise CA nodes we gathered to filter the first set of paths for P1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC3Path1Pattern(edge.EndID, enterpriseCANodes).Do(func(terminal *graph.PathSegment) error {
			certTemplateNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			path1CandidateSegments[certTemplateNode.ID] = append(path1CandidateSegments[certTemplateNode.ID], terminal)
			path1CertTemplates.Add(certTemplateNode.ID.Uint32())
			lock.Unlock()

			return nil
		})}); err != nil {
		return nil, err
	}

	//Find all cert templates we have EnrollOnBehalfOf from our first group of templates to prefilter again
	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if p, err := ops.FetchPathSet(tx.Relationships().Filter(
			query.And(
				query.InIDs(query.StartID(), cardinality.DuplexToGraphIDs(path1CertTemplates)...),
				query.KindIn(query.Relationship(), ad.EnrollOnBehalfOf),
				query.KindIn(query.End(), ad.CertTemplate)),
		)); err != nil {
			return err
		} else {
			enrollOnBehalfOfPaths = p
			return nil
		}
	}); err != nil {
		return nil, err
	}

	for _, path := range enrollOnBehalfOfPaths {
		path2CandidateTemplates.Add(path.Terminal().ID.Uint32())
	}

	//Use our enterprise ca + candidate templates as filters for the third query (P2)
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC3Path2Pattern(edge.EndID, enterpriseCANodes, path2CandidateTemplates).Do(func(terminal *graph.PathSegment) error {
			certTemplateNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
			})

			lock.Lock()
			path2CandidateSegments[certTemplateNode.ID] = append(path2CandidateSegments[certTemplateNode.ID], terminal)
			path2CertTemplates.Add(certTemplateNode.ID.Uint32())
			lock.Unlock()

			return nil
		})}); err != nil {
		return nil, err
	}

	//EnrollOnBehalfOf is used to join P1 and P2, so we'll use it as the key
	for _, p3 := range enrollOnBehalfOfPaths {
		ct1 := p3.Root()
		ct2 := p3.Terminal()

		if !path1CertTemplates.Contains(ct1.ID.Uint32()) {
			continue
		}

		if !path2CertTemplates.Contains(ct2.ID.Uint32()) {
			continue
		}

		p1paths := path1CandidateSegments[ct1.ID]
		p2paths := path2CandidateSegments[ct2.ID]

		for _, p1 := range p1paths {
			eca1 := p1.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) && enterpriseCANodes.Contains(nextSegment.Node.ID.Uint32())
			})

			for _, p2 := range p2paths {
				eca2 := p2.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA) && enterpriseCANodes.Contains(nextSegment.Node.ID.Uint32())
				})

				for _, p4 := range enterpriseCASegments[eca1.ID] {
					paths.AddPath(p4.Path())
				}

				for _, p5 := range enterpriseCASegments[eca2.ID] {
					paths.AddPath(p5.Path())
				}

				paths.AddPath(p3)
				paths.AddPath(p1.Path())
				paths.AddPath(p2.Path())

				if collected, err := eca2.Properties.Get(ad.EnrollmentAgentRestrictionsCollected.String()).Bool(); err != nil {
					log.Errorf("error getting enrollmentagentcollected for eca2 %d: %v", eca2.ID, err)
				} else if hasRestrictions, err := eca2.Properties.Get(ad.HasEnrollmentAgentRestrictions.String()).Bool(); err != nil {
					log.Errorf("error getting hasenrollmentagentrestrictions for ca %d: %v", eca2.ID, err)
				} else if collected && hasRestrictions {
					if p6, err := getDelegatedEnrollmentAgentPath(ctx, startNode, ct2, db); err != nil {
						log.Infof("Error getting p6 for composition: %v", err)
					} else {
						paths.AddPathSet(p6)
					}
				}
			}
		}
	}

	return paths, nil
}

func getDelegatedEnrollmentAgentPath(ctx context.Context, startNode, certTemplate2 *graph.Node, db graph.Database) (graph.PathSet, error) {
	var pathSet graph.PathSet

	return pathSet, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if paths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.InIDs(query.StartID(), startNode.ID),
			query.InIDs(query.EndID(), certTemplate2.ID),
			query.KindIn(query.Relationship(), ad.DelegatedEnrollmentAgent),
		))); err != nil {
			return err
		} else {
			pathSet = paths
			return nil
		}
	})
}

func ADCSESC1Path1Pattern(domainID graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
			query.Kind(query.End(), ad.CertTemplate),
			query.Or(
				query.And(
					query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
					query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
					query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), true),
				),
				query.And(
					query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
					query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), true),
				),
			),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo, ad.IssuedSignedBy),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainID),
		))
}

func ADCSESC1Path2Pattern(domainID graph.ID, enterpriseCAs cardinality.Duplex[uint32]) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.Enroll),
			query.InIDs(query.EndID(), cardinality.DuplexToGraphIDs(enterpriseCAs)...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainID),
		))
}

func GetADCSESC1EdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		startNode *graph.Node

		traversalInst      = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths              = graph.PathSet{}
		candidateSegments  = map[graph.ID][]*graph.PathSegment{}
		path1EnterpriseCAs = cardinality.NewBitmap32()
		path2EnterpriseCAs = cardinality.NewBitmap32()
		lock               = &sync.Mutex{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if node, err := ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else {
			startNode = node
			return nil
		}
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC1Path1Pattern(edge.EndID).Do(func(terminal *graph.PathSegment) error {
			// Find the CA and track it before stuffing this path into the candidates
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			candidateSegments[enterpriseCANode.ID] = append(candidateSegments[enterpriseCANode.ID], terminal)
			path1EnterpriseCAs.Add(enterpriseCANode.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: ADCSESC1Path2Pattern(edge.EndID, path1EnterpriseCAs).Do(func(terminal *graph.PathSegment) error {
			// Find the CA and track it before stuffing this path into the candidates
			enterpriseCANode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			candidateSegments[enterpriseCANode.ID] = append(candidateSegments[enterpriseCANode.ID], terminal)
			path2EnterpriseCAs.Add(enterpriseCANode.ID.Uint32())
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	// Intersect the CAs and take only those seen in both paths
	path1EnterpriseCAs.And(path2EnterpriseCAs)

	// Render paths from the segments
	path1EnterpriseCAs.Each(func(value uint32) bool {
		for _, segment := range candidateSegments[graph.ID(value)] {
			paths.AddPath(segment.Path())
		}

		return true
	})

	return paths, nil
}

func getGoldenCertEdgeComposition(tx graph.Transaction, edge *graph.Relationship) (graph.PathSet, error) {
	finalPaths := graph.NewPathSet()
	//Grab the start node (computer) as well as the target domain node
	if startNode, targetDomainNode, err := ops.FetchRelationshipNodes(tx, edge); err != nil {
		return finalPaths, err
	} else {
		//Find hosted enterprise CA
		if ecaPaths, err := ops.FetchPathSet(tx.Relationships().Filter(query.And(
			query.Equals(query.StartID(), startNode.ID),
			query.KindIn(query.End(), ad.EnterpriseCA),
			query.KindIn(query.Relationship(), ad.HostsCAService),
		))); err != nil {
			log.Errorf("error getting hostscaservice edge to enterprise ca for computer %d : %v", startNode.ID, err)
		} else {
			for _, ecaPath := range ecaPaths {
				eca := ecaPath.Terminal()
				if chainToRootCAPaths, err := FetchEnterpriseCAsCertChainPathToDomain(tx, eca, targetDomainNode); err != nil {
					log.Errorf("error getting eca %d path to domain %d: %v", eca.ID, targetDomainNode.ID, err)
				} else if chainToRootCAPaths.Len() == 0 {
					continue
				} else if trustedForAuthPaths, err := FetchEnterpriseCAsTrustedForAuthPathToDomain(tx, eca, targetDomainNode); err != nil {
					log.Errorf("error getting eca %d path to domain %d via trusted for auth: %v", eca.ID, targetDomainNode.ID, err)
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

func adcsESC9aPath1Pattern(domainID graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(
			1, 1,
			query.And(
				query.KindIn(query.Relationship(), ad.GenericWrite, ad.GenericAll, ad.Owns, ad.WriteOwner, ad.WriteDACL),
				query.KindIn(query.End(), ad.Computer, ad.User),
			),
		).
		OutboundWithDepth(
			0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			),
		).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
				query.Kind(query.End(), ad.CertTemplate),
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Equals(query.EndProperty(ad.NoSecurityExtension.String()), true),
				query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), false),
				query.Or(
					query.Equals(query.EndProperty(ad.SubjectAltRequireUPN.String()), true),
					query.Equals(query.EndProperty(ad.SubjectAltRequireSPN.String()), true),
				),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo, ad.IssuedSignedBy),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainID),
		))
}

func adcsESC9APath2Pattern(caNodes []graph.ID, domainId graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(0, 0, query.And(
			query.Kind(query.Relationship(), ad.MemberOf),
			query.Kind(query.End(), ad.Group),
		)).
		Outbound(query.And(
			query.Kind(query.Relationship(), ad.Enroll),
			query.InIDs(query.End(), caNodes...),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.TrustedForNTAuth),
			query.Kind(query.End(), ad.NTAuthStore),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.NTAuthStoreFor),
			query.Equals(query.EndID(), domainId),
		))
}

func adcsESC9APath3Pattern(caIDs []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		Inbound(
			query.KindIn(query.Relationship(), ad.DCFor, ad.TrustedBy),
		).
		Inbound(query.And(
			query.Kind(query.Relationship(), ad.CanAbuseWeakCertBinding),
			query.InIDs(query.StartID(), caIDs...),
		))
}

func GetADCSESC9aEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*
		MATCH (n {objectid:'S-1-5-21-3933516454-2894985453-2515407000-500'})-[:ADCSESC9a]->(d:Domain {objectid:'S-1-5-21-3933516454-2894985453-2515407000'})
		OPTIONAL MATCH p1 = (n)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(m)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
		WHERE ct.requiresmanagerapproval = false
		AND ct.authenticationenabled = true
		AND ct.nosecurityextension = true
		AND ct.enrolleesuppliessubject = false
		AND (ct.subjectaltrequireupn = true OR ct.subjectaltrequirespn = true)
		AND (
		(ct.schemaversion > 1 AND ct.authorizedsignatures = 0)
		OR ct.schemaversion = 1
		)
		AND (
		m:Computer
		OR (m:User AND ct.subjectaltrequiredns = false AND ct.subjectaltrequiredomaindns = false)
		)
		OPTIONAL MATCH p2 = (m)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
		OPTIONAL MATCH p3 = (ca)-[:CanAbuseWeakCertBinding|DCFor|TrustedBy*1..]->(d)
		RETURN p1,p2,p3
	*/

	var (
		startNode *graph.Node
		endNode   *graph.Node

		traversalInst          = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                  = graph.PathSet{}
		path1CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		victimCANodes          = map[graph.ID][]graph.ID{}
		path2CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path3CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		p2canodes              = make([]graph.ID, 0)
		nodeMap                = map[graph.ID]*graph.Node{}
		lock                   = &sync.Mutex{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	//Fully manifest p1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC9aPath1Pattern(edge.EndID).Do(func(terminal *graph.PathSegment) error {
			victimNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Depth() == 1
			})

			if victimNode.Kinds.ContainsOneOf(ad.User) {
				certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				if !certTemplateValidForUserVictim(certTemplate) {
					return nil
				}
			}

			caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			path1CandidateSegments[victimNode.ID] = append(path1CandidateSegments[victimNode.ID], terminal)
			nodeMap[victimNode.ID] = victimNode
			victimCANodes[victimNode.ID] = append(victimCANodes[victimNode.ID], caNode.ID)
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	for victim, p1CANodes := range victimCANodes {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: nodeMap[victim],
			Driver: adcsESC9APath2Pattern(p1CANodes, edge.EndID).Do(func(terminal *graph.PathSegment) error {
				caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				path2CandidateSegments[caNode.ID] = append(path2CandidateSegments[caNode.ID], terminal)
				p2canodes = append(p2canodes, caNode.ID)
				lock.Unlock()

				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	if len(p2canodes) > 0 {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: endNode,
			Driver: adcsESC9APath3Pattern(p2canodes).Do(func(terminal *graph.PathSegment) error {
				caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				path3CandidateSegments[caNode.ID] = append(path3CandidateSegments[caNode.ID], terminal)
				lock.Unlock()
				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	for _, p1paths := range path1CandidateSegments {
		for _, p1path := range p1paths {
			caNode := p1path.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			if p2segments, ok := path2CandidateSegments[caNode.ID]; !ok {
				continue
			} else if p3segments, ok := path3CandidateSegments[caNode.ID]; !ok {
				continue
			} else {
				paths.AddPath(p1path.Path())
				for _, p2 := range p2segments {
					paths.AddPath(p2.Path())
				}

				for _, p3 := range p3segments {
					paths.AddPath(p3.Path())
				}
			}
		}
	}

	return paths, nil
}

func certTemplateValidForUserVictim(certTemplate *graph.Node) bool {
	if subjectAltRequireDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDNS.String()).Bool(); err != nil {
		return false
	} else if subjectAltRequireDNS {
		return false
	} else if subjectAltRequireDomainDNS, err := certTemplate.Properties.Get(ad.SubjectAltRequireDomainDNS.String()).Bool(); err != nil {
		return false
	} else if subjectAltRequireDomainDNS {
		return false
	} else {
		return true
	}
}

func adcsESC10aPath1Pattern(domainID graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		OutboundWithDepth(
			1, 1,
			query.And(
				query.KindIn(query.Relationship(), ad.GenericWrite, ad.GenericAll, ad.Owns, ad.WriteOwner, ad.WriteDACL),
				query.KindIn(query.End(), ad.Computer, ad.User),
			),
		).
		OutboundWithDepth(
			0, 0,
			query.And(
				query.Kind(query.Relationship(), ad.MemberOf),
				query.Kind(query.End(), ad.Group),
			),
		).
		Outbound(
			query.And(
				query.KindIn(query.Relationship(), ad.GenericAll, ad.Enroll, ad.AllExtendedRights),
				query.Kind(query.End(), ad.CertTemplate),
				query.Equals(query.EndProperty(ad.RequiresManagerApproval.String()), false),
				query.Equals(query.EndProperty(ad.AuthenticationEnabled.String()), true),
				query.Equals(query.EndProperty(ad.EnrolleeSuppliesSubject.String()), false),
				query.Or(
					query.Equals(query.EndProperty(ad.SubjectAltRequireUPN.String()), true),
					query.Equals(query.EndProperty(ad.SubjectAltRequireSPN.String()), true),
				),
				query.Or(
					query.Equals(query.EndProperty(ad.SchemaVersion.String()), 1),
					query.And(
						query.GreaterThan(query.EndProperty(ad.SchemaVersion.String()), 1),
						query.Equals(query.EndProperty(ad.AuthorizedSignatures.String()), 0),
					),
				),
			),
		).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.PublishedTo, ad.IssuedSignedBy),
			query.Kind(query.End(), ad.EnterpriseCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.IssuedSignedBy, ad.EnterpriseCAFor),
			query.Kind(query.End(), ad.RootCA),
		)).
		Outbound(query.And(
			query.KindIn(query.Relationship(), ad.RootCAFor),
			query.Equals(query.EndID(), domainID),
		))
}

func adcsESC10APath3Pattern(caIDs []graph.ID) traversal.PatternContinuation {
	return traversal.NewPattern().
		Inbound(
			query.KindIn(query.Relationship(), ad.DCFor, ad.TrustedBy),
		).
		Inbound(query.And(
			query.Kind(query.Relationship(), ad.CanAbuseUPNCertMapping),
			query.InIDs(query.StartID(), caIDs...),
		))
}

func GetADCSESC10aEdgeComposition(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	/*MATCH (n {objectid:'S-1-5-21-3933516454-2894985453-2515407000-500'})-[:ADCSESC10a]->(d:Domain {objectid:'S-1-5-21-3933516454-2894985453-2515407000'})
	OPTIONAL MATCH p1 = (n)-[:GenericAll|GenericWrite|Owns|WriteOwner|WriteDacl]->(m)-[:MemberOf*0..]->()-[:GenericAll|Enroll|AllExtendedRights]->(ct)-[:PublishedTo]->(ca)-[:IssuedSignedBy|EnterpriseCAFor|RootCAFor*1..]->(d)
	WHERE ct.requiresmanagerapproval = false
	  AND ct.authenticationenabled = true
	  AND ct.enrolleesuppliessubject = false
	  AND (ct.subjectaltrequireupn = true OR ct.subjectaltrequirespn = true)
	  AND (
	    (ct.schemaversion > 1 AND ct.authorizedsignatures = 0)
	    OR ct.schemaversion = 1
	  )
	  AND (
	    m:Computer
	    OR (m:User AND ct.subjectaltrequiredns = false AND ct.subjectaltrequiredomaindns = false)
	  )
	OPTIONAL MATCH p2 = (m)-[:MemberOf*0..]->()-[:Enroll]->(ca)-[:TrustedForNTAuth]->(nt)-[:NTAuthStoreFor]->(d)
	OPTIONAL MATCH p3 = (ca)-[:CanAbuseUPNCertMapping|DCFor|TrustedBy*1..]->(d)
	RETURN p1,p2,p3*/
	var (
		startNode *graph.Node
		endNode   *graph.Node

		traversalInst          = traversal.New(db, analysis.MaximumDatabaseParallelWorkers)
		paths                  = graph.PathSet{}
		path1CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		victimCANodes          = map[graph.ID][]graph.ID{}
		path2CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		path3CandidateSegments = map[graph.ID][]*graph.PathSegment{}
		p2canodes              = make([]graph.ID, 0)
		nodeMap                = map[graph.ID]*graph.Node{}
		lock                   = &sync.Mutex{}
	)

	if err := db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		var err error
		if startNode, err = ops.FetchNode(tx, edge.StartID); err != nil {
			return err
		} else if endNode, err = ops.FetchNode(tx, edge.EndID); err != nil {
			return err
		} else {
			return nil
		}
	}); err != nil {
		return nil, err
	}

	//Fully manifest p1
	if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
		Root: startNode,
		Driver: adcsESC10aPath1Pattern(edge.EndID).Do(func(terminal *graph.PathSegment) error {
			victimNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Depth() == 1
			})

			if victimNode.Kinds.ContainsOneOf(ad.User) {
				certTemplate := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.CertTemplate)
				})

				if !certTemplateValidForUserVictim(certTemplate) {
					return nil
				}
			}

			caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			lock.Lock()
			path1CandidateSegments[victimNode.ID] = append(path1CandidateSegments[victimNode.ID], terminal)
			nodeMap[victimNode.ID] = victimNode
			victimCANodes[victimNode.ID] = append(victimCANodes[victimNode.ID], caNode.ID)
			lock.Unlock()

			return nil
		}),
	}); err != nil {
		return nil, err
	}

	//We can re-use p2 from ESC9a, since they're the same
	for victim, p1CANodes := range victimCANodes {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: nodeMap[victim],
			Driver: adcsESC9APath2Pattern(p1CANodes, edge.EndID).Do(func(terminal *graph.PathSegment) error {
				caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				path2CandidateSegments[caNode.ID] = append(path2CandidateSegments[caNode.ID], terminal)
				p2canodes = append(p2canodes, caNode.ID)
				lock.Unlock()

				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	if len(p2canodes) > 0 {
		if err := traversalInst.BreadthFirst(ctx, traversal.Plan{
			Root: endNode,
			Driver: adcsESC10APath3Pattern(p2canodes).Do(func(terminal *graph.PathSegment) error {
				caNode := terminal.Search(func(nextSegment *graph.PathSegment) bool {
					return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
				})

				lock.Lock()
				path3CandidateSegments[caNode.ID] = append(path3CandidateSegments[caNode.ID], terminal)
				lock.Unlock()
				return nil
			}),
		}); err != nil {
			return nil, err
		}
	}

	for _, p1paths := range path1CandidateSegments {
		for _, p1path := range p1paths {
			caNode := p1path.Search(func(nextSegment *graph.PathSegment) bool {
				return nextSegment.Node.Kinds.ContainsOneOf(ad.EnterpriseCA)
			})

			if p2segments, ok := path2CandidateSegments[caNode.ID]; !ok {
				continue
			} else if p3segments, ok := path3CandidateSegments[caNode.ID]; !ok {
				continue
			} else {
				paths.AddPath(p1path.Path())
				for _, p2 := range p2segments {
					paths.AddPath(p2.Path())
				}

				for _, p3 := range p3segments {
					paths.AddPath(p3.Path())
				}
			}
		}
	}

	return paths, nil
}
