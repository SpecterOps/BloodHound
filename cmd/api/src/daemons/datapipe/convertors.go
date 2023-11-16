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
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
)

func convertComputerData(data []ein.Computer, adcsEnabled bool) ConvertedData {
	converted := ConvertedData{}

	for _, computer := range data {
		baseNodeProp := ein.ConvertObjectToNode(computer.IngestBase, ad.Computer)
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(computer.Aces, computer.ObjectIdentifier, ad.Computer)...)
		if primaryGroupRel := ein.ParsePrimaryGroup(computer.IngestBase, ad.Computer, computer.PrimaryGroupSID); primaryGroupRel.IsValid() {
			converted.RelProps = append(converted.RelProps, primaryGroupRel)
		}

		if containingPrincipalRel := ein.ParseObjectContainer(computer.IngestBase, ad.Computer); containingPrincipalRel.IsValid() {
			converted.RelProps = append(converted.RelProps, containingPrincipalRel)
		}

		converted.RelProps = append(converted.RelProps, ein.ParseComputerMiscData(computer)...)
		for _, localGroup := range computer.LocalGroups {
			if !localGroup.Collected || len(localGroup.Results) == 0 {
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
				converted.RelProps = append(converted.RelProps, ein.ParseUserRightData(userRight, computer, ad.RemoteInteractiveLogonPrivilege)...)
				baseNodeProp.PropertyMap[ad.HasURA.String()] = true
			}
		}

		if adcsEnabled {
			converted.NodeProps = append(converted.NodeProps, ein.ParseDCRegistryData(computer))
		}

		converted.NodeProps = append(converted.NodeProps, baseNodeProp)
	}

	return converted
}

func convertUserData(data []ein.User) ConvertedData {
	converted := ConvertedData{}

	for _, user := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(user.IngestBase, ad.User))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(user.Aces, user.ObjectIdentifier, ad.User)...)
		if rel := ein.ParseObjectContainer(user.IngestBase, ad.User); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}
		converted.RelProps = append(converted.RelProps, ein.ParseUserMiscData(user)...)
	}

	return converted
}

func convertGroupData(data []ein.Group) ConvertedGroupData {
	converted := ConvertedGroupData{}

	for _, group := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(group.IngestBase, ad.Group))

		if rel := ein.ParseObjectContainer(group.IngestBase, ad.Group); rel.Source != "" && rel.Target != "" {
			converted.RelProps = append(converted.RelProps, rel)
		}
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(group.Aces, group.ObjectIdentifier, ad.Group)...)

		groupMembershipData := ein.ParseGroupMembershipData(group)
		converted.RelProps = append(converted.RelProps, groupMembershipData.RegularMembers...)
		converted.DistinguishedNameProps = append(converted.DistinguishedNameProps, groupMembershipData.DistinguishedNameMembers...)
	}

	return converted
}

func convertDomainData(data []ein.Domain) ConvertedData {
	converted := ConvertedData{}

	for _, domain := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(domain.IngestBase, ad.Domain))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(domain.Aces, domain.ObjectIdentifier, ad.Domain)...)
		if len(domain.ChildObjects) > 0 {
			converted.RelProps = append(converted.RelProps, ein.ParseChildObjects(domain.ChildObjects, domain.ObjectIdentifier, ad.Domain)...)
		}

		if len(domain.Links) > 0 {
			converted.RelProps = append(converted.RelProps, ein.ParseGpLinks(domain.Links, domain.ObjectIdentifier, ad.Domain)...)
		}

		domainTrustData := ein.ParseDomainTrusts(domain)
		converted.RelProps = append(converted.RelProps, domainTrustData.TrustRelationships...)
		converted.NodeProps = append(converted.NodeProps, domainTrustData.ExtraNodeProps...)
	}

	return converted
}

func convertGPOData(data []ein.GPO) ConvertedData {
	converted := ConvertedData{}

	for _, gpo := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(ein.IngestBase(gpo), ad.GPO))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(gpo.Aces, gpo.ObjectIdentifier, ad.GPO)...)
	}

	return converted
}

func convertOUData(data []ein.OU) ConvertedData {
	converted := ConvertedData{}
	for _, ou := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(ou.IngestBase, ad.OU))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(ou.Aces, ou.ObjectIdentifier, ad.OU)...)
		if container := ein.ParseObjectContainer(ou.IngestBase, ad.OU); container.IsValid() {
			converted.RelProps = append(converted.RelProps, container)
		}

		converted.RelProps = append(converted.RelProps, ein.ParseGpLinks(ou.Links, ou.ObjectIdentifier, ad.OU)...)
		if len(ou.ChildObjects) > 0 {
			converted.RelProps = append(converted.RelProps, ein.ParseChildObjects(ou.ChildObjects, ou.ObjectIdentifier, ad.OU)...)
		}
	}

	return converted
}

func convertSessionData(data []ein.Session) ConvertedSessionData {
	converted := CreateConvertedSessionData(len(data))
	for i, s := range data {
		converted.SessionProps[i] = ein.ConvertSessionObject(s)
	}

	return converted
}

func CreateConvertedSessionData(count int) ConvertedSessionData {
	converted := ConvertedSessionData{}
	converted.SessionProps = make([]ein.IngestibleSession, count)
	return converted
}

func convertContainerData(data []ein.Container) ConvertedData {
	converted := ConvertedData{}
	for _, container := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(container.IngestBase, ad.Container))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(container.Aces, container.ObjectIdentifier, ad.Container)...)

		if rel := ein.ParseObjectContainer(container.IngestBase, ad.Container); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}

		if len(container.ChildObjects) > 0 {
			converted.RelProps = append(converted.RelProps, ein.ParseChildObjects(container.ChildObjects, container.ObjectIdentifier, ad.Container)...)
		}
	}

	return converted
}

func convertAIACAData(data []ein.AIACA) ConvertedData {
	converted := ConvertedData{}

	for _, aiaca := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(ein.IngestBase(aiaca), ad.AIACA))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(aiaca.Aces, aiaca.ObjectIdentifier, ad.AIACA)...)

		if rel := ein.ParseObjectContainer(ein.IngestBase(aiaca), ad.AIACA); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}
	}

	return converted
}

func convertRootCAData(data []ein.RootCA) ConvertedData {
	converted := ConvertedData{}

	for _, rootca := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(rootca.IngestBase, ad.RootCA))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(rootca.Aces, rootca.ObjectIdentifier, ad.RootCA)...)
		converted.RelProps = append(converted.RelProps, ein.ParseRootCAMiscData(rootca)...)

		if rel := ein.ParseObjectContainer(rootca.IngestBase, ad.RootCA); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}
	}

	return converted
}

func convertEnterpriseCAData(data []ein.EnterpriseCA) ConvertedData {
	converted := ConvertedData{}

	for _, enterpriseca := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(enterpriseca.IngestBase, ad.EnterpriseCA))
		converted.NodeProps = append(converted.NodeProps, ein.ParseCARegistryProperties(enterpriseca))
		converted.RelProps = append(converted.RelProps, ein.ParseEnterpriseCAMiscData(enterpriseca)...)

		if rel := ein.ParseObjectContainer(enterpriseca.IngestBase, ad.EnterpriseCA); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}
	}

	return converted
}

func convertNTAuthStoreData(data []ein.NTAuthStore) ConvertedData {
	converted := ConvertedData{}

	for _, ntauthstore := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(ntauthstore.IngestBase, ad.NTAuthStore))
		converted.RelProps = append(converted.RelProps, ein.ParseNTAuthStoreData(ntauthstore)...)
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(ntauthstore.Aces, ntauthstore.ObjectIdentifier, ad.NTAuthStore)...)

		if rel := ein.ParseObjectContainer(ntauthstore.IngestBase, ad.NTAuthStore); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}
	}

	return converted
}

func convertCertTemplateData(data []ein.CertTemplate) ConvertedData {
	converted := ConvertedData{}

	for _, certtemplate := range data {
		converted.NodeProps = append(converted.NodeProps, ein.ConvertObjectToNode(ein.IngestBase(certtemplate), ad.CertTemplate))
		converted.RelProps = append(converted.RelProps, ein.ParseACEData(certtemplate.Aces, certtemplate.ObjectIdentifier, ad.CertTemplate)...)

		if rel := ein.ParseObjectContainer(ein.IngestBase(certtemplate), ad.CertTemplate); rel.IsValid() {
			converted.RelProps = append(converted.RelProps, rel)
		}
	}

	return converted
}
