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

package graphify

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
)

func convertGenericNode(entity ein.GenericNode, converted *ConvertedData) error {
	objectID := strings.ToUpper(entity.ID) // BloodHound convention: object IDs are uppercased

	node := ein.IngestibleNode{
		ObjectID:    objectID,
		PropertyMap: entity.Properties,
		Labels:      graph.StringsToKinds(entity.Kinds),
	}

	if node.PropertyMap == nil {
		node.PropertyMap = make(map[string]any)
	}

	// If a "objectid" is present in the property map, verify it matches the top-level ID
	if rawID, ok := node.PropertyMap["objectid"]; ok {
		if propertyID, ok := rawID.(string); ok && !strings.EqualFold(propertyID, objectID) {
			slog.Warn("objectid in property map does not match top-level id of node; skipping.",
				slog.Any("properties.objectid", propertyID),
				slog.String("expected objectid", objectID),
			)
			return fmt.Errorf("skipping invalid node. objectid: %s", objectID)
		}
	}

	// the first element in node.Labels determines which icon the UI renders for the node.
	// it is critical to specify this information because a node can have up to 3 kinds.
	node.PropertyMap[common.PrimaryKind.String()] = node.Labels[0]

	converted.NodeProps = append(converted.NodeProps, node)
	return nil
}

func convertGenericEdge(entity ein.GenericEdge, converted *ConvertedData) error {
	ingestibleRel := ein.NewIngestibleRelationship(
		ein.IngestibleEndpoint{
			Value:   strings.ToUpper(entity.Start.Value),
			MatchBy: ein.IngestMatchStrategy(entity.Start.MatchBy),
			Kind:    graph.StringKind(entity.Start.Kind),
		},
		ein.IngestibleEndpoint{
			Value:   strings.ToUpper(entity.End.Value),
			MatchBy: ein.IngestMatchStrategy(entity.End.MatchBy),
			Kind:    graph.StringKind(entity.End.Kind),
		},
		ein.IngestibleRel{
			RelProps: entity.Properties,
			RelType:  graph.StringKind(entity.Kind),
		},
	)

	converted.RelProps = append(converted.RelProps, ingestibleRel)
	return nil
}

func convertComputerData(computer ein.Computer, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertComputerToNode(computer, ingestTime)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, computer.Aces, computer.ObjectIdentifier, ad.Computer)...)
	if primaryGroupRel := ein.ParsePrimaryGroup(computer.IngestBase, ad.Computer, computer.PrimaryGroupSID); primaryGroupRel.IsValid() {
		converted.RelProps = append(converted.RelProps, primaryGroupRel)
	}
	if containsIdentityRel := ein.ParseDomainForIdentity(computer.IngestBase, ad.Computer, computer.DomainSID); containsIdentityRel.IsValid() {
		converted.RelProps = append(converted.RelProps, containsIdentityRel)
	}

	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(computer.IngestBase, ad.Computer, baseNodeProp)...)
	converted.RelProps = append(converted.RelProps, ein.ParseComputerMiscData(computer)...)

	for _, localGroup := range computer.LocalGroups {
		// throw away groups with 0 members unless the group name begins with "DIRECT ACCESS USERS"
		if !localGroup.Collected || (len(localGroup.Results) == 0 && !strings.HasPrefix(strings.ToUpper(localGroup.Name), "DIRECT ACCESS USERS")) {
			continue
		}

		parsedLocalGroupData := ein.ConvertLocalGroup(localGroup, computer)
		converted.RelProps = append(converted.RelProps, parsedLocalGroupData.Relationships...)
		converted.NodeProps = append(converted.NodeProps, parsedLocalGroupData.Nodes...)
	}

	for _, userRight := range computer.UserRights {
		if !userRight.Collected || len(userRight.Results) == 0 {
			continue
		}

		if userRight.Privilege == ein.UserRightRemoteInteractiveLogon {
			converted.RelProps = append(converted.RelProps, ein.ParseUserRightData(userRight, computer, ad.RemoteInteractiveLogonRight)...)
			baseNodeProp.PropertyMap[ad.HasURA.String()] = true
		}
	}

	converted.NodeProps = append(converted.NodeProps, ein.ParseDCRegistryData(computer))
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)

}

func convertUserData(user ein.User, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(user.IngestBase, ad.User, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, user.Aces, user.ObjectIdentifier, ad.User)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(user.IngestBase, ad.User, baseNodeProp)...)
	converted.RelProps = append(converted.RelProps, ein.ParseUserMiscData(user)...)
	if containsIdentityRel := ein.ParseDomainForIdentity(user.IngestBase, ad.User, user.DomainSID); containsIdentityRel.IsValid() {
		converted.RelProps = append(converted.RelProps, containsIdentityRel)
	}
}

func convertGroupData(group ein.Group, converted *ConvertedGroupData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(group.IngestBase, ad.Group, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(group.IngestBase, ad.Group, baseNodeProp)...)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, group.Aces, group.ObjectIdentifier, ad.Group)...)

	groupMembershipData := ein.ParseGroupMembershipData(group)
	converted.RelProps = append(converted.RelProps, groupMembershipData.RegularMembers...)
	converted.DistinguishedNameProps = append(converted.DistinguishedNameProps, groupMembershipData.DistinguishedNameMembers...)
}

func convertDomainData(domain ein.Domain, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(domain.IngestBase, ad.Domain, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, domain.Aces, domain.ObjectIdentifier, ad.Domain)...)
	if len(domain.ChildObjects) > 0 {
		converted.RelProps = append(converted.RelProps, ein.ParseChildObjects(domain.ChildObjects, domain.ObjectIdentifier, ad.Domain)...)
	}

	if len(domain.Links) > 0 {
		converted.RelProps = append(converted.RelProps, ein.ParseGpLinks(domain.Links, domain.ObjectIdentifier, ad.Domain)...)
	}

	domainTrustData := ein.ParseDomainTrusts(domain)
	converted.RelProps = append(converted.RelProps, domainTrustData.TrustRelationships...)
	converted.NodeProps = append(converted.NodeProps, domainTrustData.ExtraNodeProps...)

	parsedLocalGroupData := ein.ParseGPOChanges(domain.GPOChanges)
	converted.RelProps = append(converted.RelProps, parsedLocalGroupData.Relationships...)
	converted.NodeProps = append(converted.NodeProps, parsedLocalGroupData.Nodes...)
}

func convertGPOData(gpo ein.GPO, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(ein.IngestBase(gpo), ad.GPO, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, gpo.Aces, gpo.ObjectIdentifier, ad.GPO)...)
}

func convertOUData(ou ein.OU, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(ou.IngestBase, ad.OU, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, ou.Aces, ou.ObjectIdentifier, ad.OU)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(ou.IngestBase, ad.OU, baseNodeProp)...)

	converted.RelProps = append(converted.RelProps, ein.ParseGpLinks(ou.Links, ou.ObjectIdentifier, ad.OU)...)
	if len(ou.ChildObjects) > 0 {
		converted.RelProps = append(converted.RelProps, ein.ParseChildObjects(ou.ChildObjects, ou.ObjectIdentifier, ad.OU)...)
	}

	parsedLocalGroupData := ein.ParseGPOChanges(ou.GPOChanges)
	converted.RelProps = append(converted.RelProps, parsedLocalGroupData.Relationships...)
	converted.NodeProps = append(converted.NodeProps, parsedLocalGroupData.Nodes...)
}

func convertSessionData(session ein.Session, converted *ConvertedSessionData) {
	converted.SessionProps = append(converted.SessionProps, ein.ConvertSessionObject(session))
}

func CreateConvertedSessionData(count int) ConvertedSessionData {
	converted := ConvertedSessionData{}
	converted.SessionProps = make([]ein.IngestibleSession, 0, count)
	return converted
}

func convertContainerData(container ein.Container, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(container.IngestBase, ad.Container, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, container.Aces, container.ObjectIdentifier, ad.Container)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(container.IngestBase, ad.Container, baseNodeProp)...)

	if len(container.ChildObjects) > 0 {
		converted.RelProps = append(converted.RelProps, ein.ParseChildObjects(container.ChildObjects, container.ObjectIdentifier, ad.Container)...)
	}
}

func convertAIACAData(aiaca ein.AIACA, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(ein.IngestBase(aiaca), ad.AIACA, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, aiaca.Aces, aiaca.ObjectIdentifier, ad.AIACA)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(ein.IngestBase(aiaca), ad.AIACA, baseNodeProp)...)
}

func convertRootCAData(rootca ein.RootCA, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(rootca.IngestBase, ad.RootCA, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, rootca.Aces, rootca.ObjectIdentifier, ad.RootCA)...)
	converted.RelProps = append(converted.RelProps, ein.ParseRootCAMiscData(rootca)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(rootca.IngestBase, ad.RootCA, baseNodeProp)...)
}

func convertEnterpriseCAData(enterpriseca ein.EnterpriseCA, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProps := ein.ConvertEnterpriseCAToNode(enterpriseca, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProps)
	converted.NodeProps = append(converted.NodeProps, ein.ParseCARegistryProperties(enterpriseca))
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProps, enterpriseca.Aces, enterpriseca.ObjectIdentifier, ad.EnterpriseCA)...)
	converted.RelProps = append(converted.RelProps, ein.ParseEnterpriseCAMiscData(enterpriseca)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(enterpriseca.IngestBase, ad.EnterpriseCA, baseNodeProps)...)
}

func convertNTAuthStoreData(ntauthstore ein.NTAuthStore, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(ntauthstore.IngestBase, ad.NTAuthStore, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseNTAuthStoreData(ntauthstore)...)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, ntauthstore.Aces, ntauthstore.ObjectIdentifier, ad.NTAuthStore)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(ntauthstore.IngestBase, ad.NTAuthStore, baseNodeProp)...)
}

func convertCertTemplateData(certtemplate ein.CertTemplate, converted *ConvertedData, ingestTime time.Time) {
	baseNodeProp := ein.ConvertObjectToNode(ein.IngestBase(certtemplate), ad.CertTemplate, ingestTime)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, certtemplate.Aces, certtemplate.ObjectIdentifier, ad.CertTemplate)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(ein.IngestBase(certtemplate), ad.CertTemplate, baseNodeProp)...)
}

func convertIssuancePolicy(issuancePolicy ein.IssuancePolicy, converted *ConvertedData, ingestTime time.Time) {
	props := ein.ConvertObjectToNode(issuancePolicy.IngestBase, ad.IssuancePolicy, ingestTime)
	if issuancePolicy.GroupLink.ObjectIdentifier != "" {
		converted.RelProps = append(converted.RelProps, ein.NewIngestibleRelationship(
			ein.IngestibleEndpoint{
				Value: issuancePolicy.ObjectIdentifier,
				Kind:  ad.IssuancePolicy,
			},
			ein.IngestibleEndpoint{
				Value: issuancePolicy.GroupLink.ObjectIdentifier,
				Kind:  issuancePolicy.GroupLink.Kind(),
			},
			ein.IngestibleRel{
				RelProps: map[string]any{"isacl": false},
				RelType:  ad.OIDGroupLink,
			},
		))
		props.PropertyMap[ad.GroupLinkID.String()] = issuancePolicy.GroupLink.ObjectIdentifier
	}

	converted.NodeProps = append(converted.NodeProps, props)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(props, issuancePolicy.Aces, issuancePolicy.ObjectIdentifier, ad.IssuancePolicy)...)
	converted.RelProps = append(converted.RelProps, ein.ParseObjectContainer(issuancePolicy.IngestBase, ad.IssuancePolicies, props)...)
}
