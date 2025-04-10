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

package datapipe

import (
	"strings"

	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func convertComputerData(computer ein.Computer, converted *ConvertedData) {
	baseNodeProp := ein.ConvertComputerToNode(computer)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, computer.Aces, computer.ObjectIdentifier, ad.Computer)...)
	if primaryGroupRel := ein.ParsePrimaryGroup(computer.IngestBase, ad.Computer, computer.PrimaryGroupSID); primaryGroupRel.IsValid() {
		converted.RelProps = append(converted.RelProps, primaryGroupRel)
	}

	if containingPrincipalRel := ein.ParseObjectContainer(computer.IngestBase, ad.Computer); containingPrincipalRel.IsValid() {
		converted.RelProps = append(converted.RelProps, containingPrincipalRel)
	}

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

func convertUserData(user ein.User, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(user.IngestBase, ad.User)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, user.Aces, user.ObjectIdentifier, ad.User)...)
	if rel := ein.ParseObjectContainer(user.IngestBase, ad.User); rel.IsValid() {
		converted.RelProps = append(converted.RelProps, rel)
	}
	converted.RelProps = append(converted.RelProps, ein.ParseUserMiscData(user)...)
}

func convertGroupData(group ein.Group, converted *ConvertedGroupData) {
	baseNodeProp := ein.ConvertObjectToNode(group.IngestBase, ad.Group)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)

	if rel := ein.ParseObjectContainer(group.IngestBase, ad.Group); rel.Source != "" && rel.Target != "" {
		converted.RelProps = append(converted.RelProps, rel)
	}
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, group.Aces, group.ObjectIdentifier, ad.Group)...)

	groupMembershipData := ein.ParseGroupMembershipData(group)
	converted.RelProps = append(converted.RelProps, groupMembershipData.RegularMembers...)
	converted.DistinguishedNameProps = append(converted.DistinguishedNameProps, groupMembershipData.DistinguishedNameMembers...)
}

func convertDomainData(domain ein.Domain, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(domain.IngestBase, ad.Domain)
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

func convertGPOData(gpo ein.GPO, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(ein.IngestBase(gpo), ad.GPO)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, gpo.Aces, gpo.ObjectIdentifier, ad.GPO)...)
}

func convertOUData(ou ein.OU, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(ou.IngestBase, ad.OU)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, ou.Aces, ou.ObjectIdentifier, ad.OU)...)
	if container := ein.ParseObjectContainer(ou.IngestBase, ad.OU); container.IsValid() {
		converted.RelProps = append(converted.RelProps, container)
	}

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
	converted.SessionProps = make([]ein.IngestibleSession, count)
	return converted
}

func convertContainerData(container ein.Container, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(container.IngestBase, ad.Container)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, container.Aces, container.ObjectIdentifier, ad.Container)...)

	if rel := ein.ParseObjectContainer(container.IngestBase, ad.Container); rel.IsValid() {
		converted.RelProps = append(converted.RelProps, rel)
	}

	if len(container.ChildObjects) > 0 {
		converted.RelProps = append(converted.RelProps, ein.ParseChildObjects(container.ChildObjects, container.ObjectIdentifier, ad.Container)...)
	}
}

func convertAIACAData(aiaca ein.AIACA, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(ein.IngestBase(aiaca), ad.AIACA)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, aiaca.Aces, aiaca.ObjectIdentifier, ad.AIACA)...)

	if rel := ein.ParseObjectContainer(ein.IngestBase(aiaca), ad.AIACA); rel.IsValid() {
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertRootCAData(rootca ein.RootCA, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(rootca.IngestBase, ad.RootCA)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, rootca.Aces, rootca.ObjectIdentifier, ad.RootCA)...)
	converted.RelProps = append(converted.RelProps, ein.ParseRootCAMiscData(rootca)...)

	if rel := ein.ParseObjectContainer(rootca.IngestBase, ad.RootCA); rel.IsValid() {
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertEnterpriseCAData(enterpriseca ein.EnterpriseCA, converted *ConvertedData) {
	baseNodeProp := ein.ConvertEnterpriseCAToNode(enterpriseca)
	converted.NodeProps = append(converted.NodeProps, ein.ConvertEnterpriseCAToNode(enterpriseca))
	converted.NodeProps = append(converted.NodeProps, ein.ParseCARegistryProperties(enterpriseca))
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, enterpriseca.Aces, enterpriseca.ObjectIdentifier, ad.EnterpriseCA)...)
	converted.RelProps = append(converted.RelProps, ein.ParseEnterpriseCAMiscData(enterpriseca)...)

	if rel := ein.ParseObjectContainer(enterpriseca.IngestBase, ad.EnterpriseCA); rel.IsValid() {
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertNTAuthStoreData(ntauthstore ein.NTAuthStore, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(ntauthstore.IngestBase, ad.NTAuthStore)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseNTAuthStoreData(ntauthstore)...)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, ntauthstore.Aces, ntauthstore.ObjectIdentifier, ad.NTAuthStore)...)

	if rel := ein.ParseObjectContainer(ntauthstore.IngestBase, ad.NTAuthStore); rel.IsValid() {
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertCertTemplateData(certtemplate ein.CertTemplate, converted *ConvertedData) {
	baseNodeProp := ein.ConvertObjectToNode(ein.IngestBase(certtemplate), ad.CertTemplate)
	converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	converted.RelProps = append(converted.RelProps, ein.ParseACEData(baseNodeProp, certtemplate.Aces, certtemplate.ObjectIdentifier, ad.CertTemplate)...)

	if rel := ein.ParseObjectContainer(ein.IngestBase(certtemplate), ad.CertTemplate); rel.IsValid() {
		converted.RelProps = append(converted.RelProps, rel)
	}
}

func convertIssuancePolicy(issuancePolicy ein.IssuancePolicy, converted *ConvertedData) {
	props := ein.ConvertObjectToNode(issuancePolicy.IngestBase, ad.IssuancePolicy)
	if issuancePolicy.GroupLink.ObjectIdentifier != "" {
		converted.RelProps = append(converted.RelProps, ein.NewIngestibleRelationship(
			ein.IngestibleSource{
				Source:     issuancePolicy.ObjectIdentifier,
				SourceType: ad.IssuancePolicy,
			},
			ein.IngestibleTarget{
				Target:     issuancePolicy.GroupLink.ObjectIdentifier,
				TargetType: issuancePolicy.GroupLink.Kind(),
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

	if container := ein.ParseObjectContainer(issuancePolicy.IngestBase, ad.IssuancePolicy); container.IsValid() {
		converted.RelProps = append(converted.RelProps, container)
	}

}
