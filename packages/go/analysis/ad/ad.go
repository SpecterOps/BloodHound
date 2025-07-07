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
	"log/slog"
	"sort"
	"strings"
	"time"

	"github.com/specterops/bloodhound/analysis/ad/internal/nodeprops"
	"github.com/specterops/bloodhound/analysis/ad/wellknown"
	"github.com/specterops/bloodhound/analysis/impact"
	"github.com/specterops/bloodhound/bhlog/measure"
	"github.com/specterops/bloodhound/dawgs/cardinality"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

const (
	AdminSDHolderDNPrefix = "CN=ADMINSDHOLDER,CN=SYSTEM,"
)

func TierZeroWellKnownSIDSuffixes() []string {
	return []string{
		wellknown.EnterpriseDomainControllersGroupSIDSuffix.String(),
		wellknown.AdministratorAccountSIDSuffix.String(),
		wellknown.DomainAdminsGroupSIDSuffix.String(),
		wellknown.DomainControllersGroupSIDSuffix.String(),
		wellknown.SchemaAdminsGroupSIDSuffix.String(),
		wellknown.EnterpriseAdminsGroupSIDSuffix.String(),
		wellknown.KeyAdminsGroupSIDSuffix.String(),
		wellknown.EnterpriseKeyAdminsGroupSIDSuffix.String(),
		wellknown.BackupOperatorsGroupSIDSuffix.String(),
		wellknown.AdministratorsSIDSuffix.String(),
	}
}

func FetchWellKnownTierZeroEntities(ctx context.Context, db graph.Database, domainSID string) (graph.NodeSet, error) {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "FetchWellKnownTierZeroEntities")()

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

		// AdminSDHolder
		if err := tx.Nodes().Filterf(func() graph.Criteria {
			return query.And(
				query.KindIn(query.Node(), ad.Container),
				query.StringStartsWith(query.NodeProperty(ad.DistinguishedName.String()), AdminSDHolderDNPrefix),
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

		return nil
	})
}

func FixWellKnownNodeTypes(ctx context.Context, db graph.Database) error {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "Fix well known node types")()

	groupSuffixes := []string{
		wellknown.EnterpriseKeyAdminsGroupSIDSuffix.String(),
		wellknown.KeyAdminsGroupSIDSuffix.String(),
		wellknown.EnterpriseDomainControllersGroupSIDSuffix.String(),
		wellknown.DomainAdminsGroupSIDSuffix.String(),
		wellknown.DomainControllersGroupSIDSuffix.String(),
		wellknown.SchemaAdminsGroupSIDSuffix.String(),
		wellknown.EnterpriseAdminsGroupSIDSuffix.String(),
		wellknown.AdministratorsSIDSuffix.String(),
		wellknown.BackupOperatorsGroupSIDSuffix.String(),
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
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "Domain Associations")()

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

		// TODO: Reimplement unassociated node pruning if we decide that FOSS needs unassociated node pruning
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
				slog.Error(fmt.Sprintf("Domain node %d does not have a valid object ID", node.ID))
			} else if domainName, err := node.Properties.Get(common.Name.String()).String(); err != nil {
				slog.Error(fmt.Sprintf("Domain node %d does not have a valid name", node.ID))
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

func LinkWellKnownNodes(ctx context.Context, db graph.Database) error {
	defer measure.ContextMeasure(ctx, slog.LevelInfo, "Link well-known nodes")()

	var (
		errors        = util.NewErrorCollector()
		newProperties = graph.NewProperties()
	)

	domains, err := GetCollectedDomains(ctx, db)
	if err != nil {
		return err
	}

	newProperties.Set(common.LastSeen.String(), time.Now().UTC())

	for _, domain := range domains {
		if err := linkWellKnownNodesForDomain(ctx, db, domain, domains.Slice(), newProperties); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf(
				"Error linking well-known nodes for domain %d: %v",
				domain.ID,
				err,
			))
			errors.Add(fmt.Errorf("failed linking well-known nodes for domain %d: %w", domain.ID, err))
		}
	}

	return errors.Combined()
}

func linkWellKnownNodesForDomain(ctx context.Context, db graph.Database, domain *graph.Node, allDomains []*graph.Node, newProperties *graph.Properties) error {
	domainSid, domainName, err := nodeprops.ReadDomainIDandNameAsString(domain)
	if err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("Error getting domain sid or name for domain %d: %v", domain.ID, err))
		return err
	}

	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		wellKnownNodes, err := createWellKnownNodesForDomain(tx, domain.ID, domainSid, domainName)
		if err != nil {
			return err
		}

		if err := createWellKnownLinksWithinDomain(tx, wellKnownNodes, newProperties); err != nil {
			return err
		}

		if err := handleGuestNodeSpecialCase(tx, domainSid, wellKnownNodes.Everyone, newProperties); err != nil {
			return err
		}

		return handleTrustRelationships(ctx, tx, domain, allDomains, wellKnownNodes.Everyone, wellKnownNodes.AuthUsers, newProperties)
	})
}

type wellKnownNodesSet struct {
	Everyone                                *graph.Node
	AuthUsers                               *graph.Node
	DomainUsers                             *graph.Node
	DomainComputers                         *graph.Node
	Network                                 *graph.Node
	ThisOrganization                        *graph.Node
	ThisOrganizationCertificate             *graph.Node
	AuthenticationAuthorityAssertedIdentity *graph.Node
	KeyTrust                                *graph.Node
	MFAKeyProperty                          *graph.Node
	NTLMAuthentication                      *graph.Node
	SChannelAuthentication                  *graph.Node
}

func createWellKnownNodesForDomain(tx graph.Transaction, domainID graph.ID, domainSid, domainName string) (*wellKnownNodesSet, error) {
	nodes := &wellKnownNodesSet{}
	var err error

	// Create/get primary nodes first
	nodes.Everyone, err = getOrCreateWellKnownGroup(
		tx,
		wellknown.EveryoneSIDSuffix,
		domainSid,
		domainName,
		wellknown.DefineNodeName(wellknown.EveryoneNodeNamePrefix, domainName),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting everyone for domain %d: %w", domainID, err)
	}

	nodes.AuthUsers, err = getOrCreateWellKnownGroup(
		tx,
		wellknown.AuthenticatedUsersSIDSuffix,
		domainSid,
		domainName,
		wellknown.DefineNodeName(wellknown.AuthenticatedUsersNodeNamePrefix, domainName),
	)
	if err != nil {
		return nil, fmt.Errorf("error getting auth users node for domain %d: %w", domainID, err)
	}

	// Create/get all other nodes
	nodeDefinitions := []struct {
		target     **graph.Node
		sidSuffix  wellknown.SIDSuffix
		namePrefix wellknown.NodeNamePrefix
	}{
		{&nodes.DomainUsers, wellknown.DomainUsersSIDSuffix, wellknown.DomainUsersNodeNamePrefix},
		{&nodes.DomainComputers, wellknown.DomainComputersSIDSuffix, wellknown.DomainComputerNodeNamePrefix},
		{&nodes.Network, wellknown.NetworkSIDSuffix, wellknown.NetworkNodeNamePrefix},
		{&nodes.ThisOrganization, wellknown.ThisOrganizationSIDSuffix, wellknown.ThisOrganizationNodeNamePrefix},
		{&nodes.ThisOrganizationCertificate, wellknown.ThisOrganizationCertificateSIDSuffix, wellknown.ThisOrganizationCertificateNodeNamePrefix},
		{&nodes.AuthenticationAuthorityAssertedIdentity, wellknown.AuthenticationAuthorityAssertedIdentitySIDSuffix, wellknown.AuthenticationAuthorityAssertedIdentityNodeNamePrefix},
		{&nodes.KeyTrust, wellknown.KeyTrustSIDSuffix, wellknown.KeyTrustNodeNamePrefix},
		{&nodes.MFAKeyProperty, wellknown.MFAKeyPropertySIDSuffix, wellknown.MFAKeyPropertyNodeNamePrefix},
		{&nodes.NTLMAuthentication, wellknown.NTLMAuthenticationSIDSuffix, wellknown.NTLMAuthenticationNodeNamePrefix},
		{&nodes.SChannelAuthentication, wellknown.SchannelAuthenticationSIDSuffix, wellknown.SchannelAuthenticationNodeNamePrefix},
	}

	for _, def := range nodeDefinitions {
		*def.target, err = getOrCreateWellKnownGroup(
			tx,
			def.sidSuffix,
			domainSid,
			domainName,
			wellknown.DefineNodeName(def.namePrefix, domainName),
		)
		if err != nil {
			return nil, fmt.Errorf("error getting %s node for domain %d: %w", def.namePrefix, domainID, err)
		}
	}

	return nodes, nil
}

func createWellKnownLinksWithinDomain(tx graph.Transaction, nodes *wellKnownNodesSet, newProperties *graph.Properties) error {
	// Define all the links to create
	memberOfLinks := []struct {
		from *graph.Node
		to   *graph.Node
	}{
		{nodes.DomainUsers, nodes.AuthUsers},
		{nodes.DomainComputers, nodes.AuthUsers},
		{nodes.AuthUsers, nodes.Everyone},
	}

	claimSpecialIdentityLinks := []struct {
		from *graph.Node
		to   *graph.Node
	}{
		{nodes.Everyone, nodes.Network},
		{nodes.Everyone, nodes.ThisOrganization},
		{nodes.Everyone, nodes.ThisOrganizationCertificate},
		{nodes.Everyone, nodes.AuthenticationAuthorityAssertedIdentity},
		{nodes.Everyone, nodes.KeyTrust},
		{nodes.Everyone, nodes.MFAKeyProperty},
		{nodes.Everyone, nodes.NTLMAuthentication},
		{nodes.Everyone, nodes.SChannelAuthentication},
	}

	// Create MemberOf links
	for _, link := range memberOfLinks {
		if err := createOrUpdateWellKnownLink(tx, link.from, link.to, newProperties, ad.MemberOf); err != nil {
			return err
		}
	}

	// Create ClaimSpecialIdentity links
	for _, link := range claimSpecialIdentityLinks {
		if err := createOrUpdateWellKnownLink(tx, link.from, link.to, newProperties, ad.ClaimSpecialIdentity); err != nil {
			return err
		}
	}

	return nil
}

func handleGuestNodeSpecialCase(tx graph.Transaction, domainSid string, everyoneNode *graph.Node, newProperties *graph.Properties) error {
	guestNode, err := tx.Nodes().Filterf(func() graph.Criteria {
		// Don't create Guest if it does not exist - not a special identity but a real object
		return query.Equals(query.NodeProperty(common.ObjectID.String()), wellknown.DefineSID(
			domainSid,
			wellknown.GuestSIDSuffix))
	}).First()

	if err != nil {
		// Guest node doesn't exist, which is fine
		return nil
	}

	// Guest - MemberOf -> Everyone
	if err := createOrUpdateWellKnownLink(tx, guestNode, everyoneNode, newProperties, ad.MemberOf); err != nil {
		return err
	}

	// If guest is enabled, create Everyone - ClaimSpecialIdentity -> Guest
	if guestEnabled, err := guestNode.Properties.Get(common.Enabled.String()).Bool(); err == nil && guestEnabled {
		return createOrUpdateWellKnownLink(tx, everyoneNode, guestNode, newProperties, ad.ClaimSpecialIdentity)
	}

	return nil
}

func handleTrustRelationships(ctx context.Context, tx graph.Transaction, domain *graph.Node, allDomains []*graph.Node, everyoneNode, authUsersNode *graph.Node, newProperties *graph.Properties) error {
	trustingDomains, err := fetchFirstDegreeNodes(tx, domain, ad.SameForestTrust, ad.CrossForestTrust)
	if err != nil {
		return fmt.Errorf("error getting trusting domains for domain %d: %w", domain.ID, err)
	}

	for _, trustingDomain := range trustingDomains {
		if !domainsContain(allDomains, trustingDomain) {
			continue // Make sure it is collected
		}

		trustingDomainSid, trustingDomainName, err := nodeprops.ReadDomainIDandNameAsString(trustingDomain)
		if err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error getting domain sid or name for domain %d: %v", trustingDomain.ID, err))
			continue
		}

		// Get/create special identity nodes in trusting domain
		everyoneNodeTrustingDomain, err := getOrCreateWellKnownGroup(
			tx,
			wellknown.EveryoneSIDSuffix,
			trustingDomainSid,
			trustingDomainName,
			wellknown.DefineNodeName(wellknown.EveryoneNodeNamePrefix, trustingDomainName),
		)
		if err != nil {
			return fmt.Errorf("error getting everyone for domain %d: %w", trustingDomain.ID, err)
		}

		authUsersNodeTrustingDomain, err := getOrCreateWellKnownGroup(
			tx,
			wellknown.AuthenticatedUsersSIDSuffix,
			trustingDomainSid,
			trustingDomainName,
			wellknown.DefineNodeName(wellknown.AuthenticatedUsersNodeNamePrefix, trustingDomainName),
		)
		if err != nil {
			return fmt.Errorf("error getting auth users node for domain %d: %w", trustingDomain.ID, err)
		}

		// Create cross-domain trust links
		if err := createOrUpdateWellKnownLink(tx, authUsersNode, authUsersNodeTrustingDomain, newProperties, ad.MemberOf); err != nil {
			return err
		}

		if err := createOrUpdateWellKnownLink(tx, everyoneNode, everyoneNodeTrustingDomain, newProperties, ad.MemberOf); err != nil {
			return err
		}
	}

	return nil
}

func domainsContain(domains []*graph.Node, target *graph.Node) bool {
	for _, domain := range domains {
		if domain.ID == target.ID {
			return true
		}
	}
	return false
}

func getOrCreateWellKnownGroup(
	tx graph.Transaction,
	wellKnownSid wellknown.SIDSuffix,
	domainSid, domainName, nodeName string,
) (
	*graph.Node,
	error,
) {
	var objectId string
	if wellKnownSid == wellknown.DomainUsersSIDSuffix || wellKnownSid == wellknown.DomainComputersSIDSuffix {
		objectId = wellknown.DefineSID(domainSid, wellKnownSid)
	} else {
		objectId = wellknown.DefineSID(domainName, wellKnownSid)
	}

	// Only filter by ObjectID, not by kind
	if wellKnownNode, err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.Equals(query.NodeProperty(common.ObjectID.String()), objectId)
	}).First(); err != nil && !graph.IsErrNotFound(err) {
		return nil, err
	} else if graph.IsErrNotFound(err) {
		// Only create a new node if no node with this ObjectID exists
		properties := graph.AsProperties(graph.PropertyMap{
			common.Name:     nodeName,
			common.ObjectID: objectId,
			ad.DomainSID:    domainSid,
			common.LastSeen: time.Now().UTC(),
			ad.DomainFQDN:   domainName,
		})
		return tx.CreateNode(properties, ad.Entity, ad.Group)
	} else {
		// If a node with this ObjectID exists (regardless of its kind), return it
		// Optionally, we could add the ad.Group kind if it's missing
		if !wellKnownNode.Kinds.ContainsOneOf(ad.Group) {
			// Add the ad.Group kind if it's missing
			wellKnownNode.AddKinds(ad.Group)
			if err := tx.UpdateNode(wellKnownNode); err != nil {
				return nil, fmt.Errorf("failed to update node with Group kind: %w", err)
			}
		}
		return wellKnownNode, nil
	}
}

func createOrUpdateWellKnownLink(
	tx graph.Transaction,
	startNode *graph.Node,
	endNode *graph.Node,
	props *graph.Properties,
	edgeType graph.Kind,
) error {
	if rel, err := tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Equals(query.StartID(), startNode.ID),
			query.Equals(query.EndID(), endNode.ID),
			query.Kind(query.Relationship(), edgeType),
		)
	}).First(); err != nil && !graph.IsErrNotFound(err) {
		return err
	} else if graph.IsErrNotFound(err) {
		if _, err := tx.CreateRelationshipByIDs(
			startNode.ID,
			endNode.ID,
			edgeType,
			props,
		); err != nil {
			return err
		} else {
			return nil
		}
	} else {
		rel.Properties.Set(common.LastSeen.String(), time.Now().UTC())
		return tx.UpdateRelationship(rel)
	}
}

// CalculateCrossProductNodeSets finds the intersection of the given sets of nodes.
// See CalculateCrossProductNodeSetsDoc.md for explaination of the specialGroups (Authenticated Users and Everyone) and why we treat them the way we do
func CalculateCrossProductNodeSets(tx graph.Transaction, groupExpansions impact.PathAggregator, nodeSlices ...[]*graph.Node) cardinality.Duplex[uint64] {
	if len(nodeSlices) < 2 {
		slog.Error("Cross products require at least 2 nodesets")
		return cardinality.NewBitmap64()
	}

	// The intention is that the node sets being passed into this function contain all the first degree principals for control
	var (
		// Temporary storage for first degree and unrolled sets without auth users/everyone
		firstDegreeSets []cardinality.Duplex[uint64]
		unrolledSets    []cardinality.Duplex[uint64]

		// This is the set we use as a reference set to check against checkset
		unrolledRefSet = cardinality.NewBitmap64()

		// This is the set we use to aggregate multiple sets together it should have all the valid principals from all other sets at this point
		checkSet = cardinality.NewBitmap64()

		// This is our set of entities that have the complete cross product of permissions
		resultEntities = cardinality.NewBitmap64()
	)

	// Get the IDs of the Auth. Users and Everyone groups
	specialGroups, err := FetchAuthUsersAndEveryoneGroups(tx)
	if err != nil {
		slog.Error(fmt.Sprintf("Could not fetch groups: %s", err.Error()))
	}

	// Unroll all nodesets
	for _, nodeSlice := range nodeSlices {
		var (
			firstDegreeSet = cardinality.NewBitmap64()
			unrolledSet    = cardinality.NewBitmap64()
		)

		for _, entity := range nodeSlice {
			entityID := entity.ID.Uint64()

			firstDegreeSet.Add(entityID)
			unrolledSet.Add(entityID)

			if entity.Kinds.ContainsOneOf(ad.Group, ad.LocalGroup) {
				unrolledSet.Or(groupExpansions.Cardinality(entity.ID.Uint64()))
			}
		}

		// Skip sets containing Auth. Users or Everyone
		hasSpecialGroup := false

		for _, specialGroup := range specialGroups {
			if unrolledSet.Contains(specialGroup.ID.Uint64()) {
				hasSpecialGroup = true
				break
			}
		}

		if !hasSpecialGroup {
			unrolledSets = append(unrolledSets, unrolledSet)
			firstDegreeSets = append(firstDegreeSets, firstDegreeSet)
		}
	}

	// If every nodeset (unrolled) includes Auth. Users/Everyone then return all nodesets (first degree)
	if len(firstDegreeSets) == 0 {
		for _, nodeSet := range nodeSlices {
			for _, entity := range nodeSet {
				resultEntities.Add(entity.ID.Uint64())
			}
		}

		return resultEntities
	} else if len(firstDegreeSets) == 1 { // If every nodeset (unrolled) except one includes Auth. Users/Everyone then return that one nodeset (first degree)
		return firstDegreeSets[0]
	} else {
		// This means that len(firstDegreeSets) must be greater than or equal to 2 i.e. we have at least two nodesets (unrolled) without Auth. Users/Everyone
		checkSet.Or(unrolledSets[1])

		for _, unrolledSet := range unrolledSets[2:] {
			checkSet.And(unrolledSet)
		}
	}

	// Check first degree principals in our reference set (firstDegreeSets[0]) first
	firstDegreeSets[0].Each(func(id uint64) bool {
		if checkSet.Contains(id) {
			resultEntities.Add(id)
		} else {
			unrolledRefSet.Or(groupExpansions.Cardinality(id))
		}

		return true
	})

	// Find all the groups in our secondary targets and map them to their cardinality in our expansions
	// Saving off to a map to prevent multiple lookups on the expansions
	tempMap := map[uint64]uint64{}
	unrolledRefSet.Each(func(id uint64) bool {
		// If group expansions contains this ID and its cardinality is > 0, it's a group/localgroup
		idCardinality := groupExpansions.Cardinality(id).Cardinality()
		if idCardinality > 0 {
			tempMap[id] = idCardinality
		}

		return true
	})

	// Save the map keys to a new slice, this represents our list of groups in the expansion
	keys := make([]uint64, 0, len(tempMap))

	for key := range tempMap {
		keys = append(keys, key)
	}

	// Sort by cardinality we saved in the map, which will give us all the groups sorted by their number of members
	sort.Slice(keys, func(i, j int) bool {
		return tempMap[keys[i]] < tempMap[keys[j]]
	})

	for _, groupId := range keys {
		// If the set doesn't contain our key, it means that we've already encapsulated this group in a previous shortcut so skip it
		if !unrolledRefSet.Contains(groupId) {
			continue
		}

		if checkSet.Contains(groupId) {
			// If this entity is a cross product, add it to result entities, remove the group id from the second set and xor the group's membership with the result set
			resultEntities.Add(groupId)

			unrolledRefSet.Remove(groupId)
			unrolledRefSet.Xor(groupExpansions.Cardinality(groupId))
		} else {
			// If this isn't a match, remove it from the second set to ensure we don't check it again, but leave its membership
			unrolledRefSet.Remove(groupId)
		}
	}

	unrolledRefSet.Each(func(remainder uint64) bool {
		if checkSet.Contains(remainder) {
			resultEntities.Add(remainder)
		}

		return true
	})

	return resultEntities
}

func GetEdgeCompositionPath(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.PathSet, error) {
	var (
		err     error
		pathSet = graph.NewPathSet()
	)

	if err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		switch edge.Kind {
		case ad.GoldenCert:
			pathSet, err = getGoldenCertEdgeComposition(tx, edge)
		case ad.ADCSESC1:
			pathSet, err = GetADCSESC1EdgeComposition(ctx, db, edge)
		case ad.ADCSESC3:
			pathSet, err = GetADCSESC3EdgeComposition(ctx, db, edge)
		case ad.ADCSESC4:
			pathSet, err = GetADCSESC4EdgeComposition(ctx, db, edge)
		case ad.ADCSESC6a, ad.ADCSESC6b:
			pathSet, err = GetADCSESC6EdgeComposition(ctx, db, edge)
		case ad.ADCSESC9a:
			pathSet, err = GetADCSESC9aEdgeComposition(ctx, db, edge)
		case ad.ADCSESC9b:
			pathSet, err = GetADCSESC9bEdgeComposition(ctx, db, edge)
		case ad.ADCSESC10a, ad.ADCSESC10b:
			pathSet, err = GetADCSESC10EdgeComposition(ctx, db, edge)
		case ad.ADCSESC13:
			pathSet, err = GetADCSESC13EdgeComposition(ctx, db, edge)
		case ad.CoerceAndRelayNTLMToADCS:
			pathSet, err = GetCoerceAndRelayNTLMtoADCSEdgeComposition(ctx, db, edge)
		case ad.CoerceAndRelayNTLMToSMB:
			pathSet, err = GetCoerceAndRelayNTLMtoSMBEdgeComposition(ctx, db, edge)
		case ad.GPOAppliesTo:
			pathSet, err = GetGPOAppliesToComposition(ctx, db, edge)
		case ad.CanApplyGPO:
			pathSet, err = GetCanApplyGPOComposition(ctx, db, edge)

		}
		return err
	}); err != nil {
		return graph.NewPathSet(), err
	}
	return pathSet, nil
}

func GetRelayTargets(ctx context.Context, db graph.Database, edge *graph.Relationship) (graph.NodeSet, error) {
	var (
		err     error
		nodeSet = graph.NewNodeSet()
	)

	if err = db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		switch edge.Kind {
		case ad.CoerceAndRelayNTLMToLDAP:
			nodeSet, err = GetVulnerableDomainControllersForRelayNTLMtoLDAP(ctx, db, edge)
		case ad.CoerceAndRelayNTLMToLDAPS:
			nodeSet, err = GetVulnerableDomainControllersForRelayNTLMtoLDAPS(ctx, db, edge)
		case ad.CoerceAndRelayNTLMToADCS:
			nodeSet, err = GetVulnerableEnterpriseCAsForRelayNTLMtoADCS(ctx, db, edge)
		case ad.CoerceAndRelayNTLMToSMB:
			nodeSet, err = GetCoercionTargetsForCoerceAndRelayNTLMtoSMB(ctx, db, edge)
		}
		return err
	}); err != nil {
		return graph.NewNodeSet(), err
	}
	return nodeSet, nil
}
