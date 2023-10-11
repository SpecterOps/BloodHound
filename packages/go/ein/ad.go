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

package ein

import (
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
)

func ConvertSessionObject(session Session) IngestibleSession {
	return IngestibleSession{
		Source:    session.ComputerSID,
		Target:    session.UserSID,
		LogonType: session.LogonType,
	}
}

func ConvertObjectToNode(item IngestBase, itemType graph.Kind) IngestibleNode {
	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: item.Properties,
		Label:       itemType,
	}
}

func ParseObjectContainer(item IngestBase, itemType graph.Kind) IngestibleRelationship {
	containingPrincipal := item.ContainedBy
	if containingPrincipal.ObjectIdentifier != "" {
		return IngestibleRelationship{
			Source:     containingPrincipal.ObjectIdentifier,
			SourceType: containingPrincipal.Kind(),
			TargetType: itemType,
			Target:     item.ObjectIdentifier,
			RelProps:   map[string]any{"isacl": false},
			RelType:    ad.Contains,
		}
	}

	return IngestibleRelationship{}
}

func ParsePrimaryGroup(item IngestBase, itemType graph.Kind, primaryGroupSid string) IngestibleRelationship {
	if primaryGroupSid != "" {
		return IngestibleRelationship{
			Source:     item.ObjectIdentifier,
			SourceType: itemType,
			TargetType: ad.Group,
			Target:     primaryGroupSid,
			RelProps:   map[string]any{"isacl": false, "isprimarygroup": true},
			RelType:    ad.MemberOf,
		}
	}

	return IngestibleRelationship{}
}

func ParseGroupMembershipData(group Group) ParsedGroupMembershipData {
	result := ParsedGroupMembershipData{}
	for _, member := range group.Members {
		if strings.HasPrefix(member.ObjectIdentifier, "DN=") {
			result.DistinguishedNameMembers = append(result.DistinguishedNameMembers, IngestibleRelationship{
				Source:     member.ObjectIdentifier,
				SourceType: member.Kind(),
				Target:     group.ObjectIdentifier,
				TargetType: ad.Group,
				RelProps:   map[string]any{"isacl": false, "isprimarygroup": false},
				RelType:    ad.MemberOf})
		} else {
			result.RegularMembers = append(result.RegularMembers, IngestibleRelationship{
				Source:     member.ObjectIdentifier,
				SourceType: member.Kind(),
				Target:     group.ObjectIdentifier,
				TargetType: ad.Group,
				RelProps:   map[string]any{"isacl": false, "isprimarygroup": false},
				RelType:    ad.MemberOf})
		}
	}

	return result
}

func ParseACEData(aces []ACE, targetID string, targetType graph.Kind) []IngestibleRelationship {
	converted := make([]IngestibleRelationship, 0)

	for _, ace := range aces {
		if ace.PrincipalSID == targetID {
			continue
		}

		if rightKind, err := analysis.ParseKind(ace.RightName); err != nil {
			log.Errorf("error during ParseACEData: %v", err)
			continue
		} else if !ad.IsACLKind(rightKind) {
			log.Errorf("non-ace edge type given to process aces: %s", ace.RightName)
			continue
		} else {
			converted = append(converted, IngestibleRelationship{
				Source:     ace.PrincipalSID,
				SourceType: ace.Kind(),
				Target:     targetID,
				TargetType: targetType,
				RelProps:   map[string]any{"isacl": true, "isinherited": ace.IsInherited},
				RelType:    rightKind,
			})
		}
	}

	return converted
}

func convertSPNData(spns []SPNTarget, sourceID string) []IngestibleRelationship {
	converted := make([]IngestibleRelationship, len(spns))

	for i, s := range spns {
		if kind, err := analysis.ParseKind(s.Service); err != nil {
			log.Errorf("error during processSPNTargets: %v", err)
		} else {
			converted[i] = IngestibleRelationship{
				Source:     sourceID,
				Target:     s.ComputerSID,
				SourceType: ad.User,
				TargetType: ad.Computer,
				RelProps:   map[string]any{"isacl": true, "port": s.Port},
				RelType:    kind,
			}
		}
	}

	return converted
}

func ParseUserMiscData(user User) []IngestibleRelationship {
	data := make([]IngestibleRelationship, 0)

	data = append(data, convertSPNData(user.SPNTargets, user.ObjectIdentifier)...)
	if rel := ParsePrimaryGroup(user.IngestBase, ad.User, user.PrimaryGroupSID); rel.IsValid() {
		data = append(data, rel)
	}

	for _, target := range user.AllowedToDelegate {
		data = append(data, IngestibleRelationship{
			Source:     user.ObjectIdentifier,
			SourceType: ad.User,
			Target:     target.ObjectIdentifier,
			TargetType: target.Kind(),
			RelType:    ad.AllowedToDelegate,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	for _, target := range user.HasSIDHistory {
		data = append(data, IngestibleRelationship{
			Source:     user.ObjectIdentifier,
			SourceType: ad.User,
			Target:     target.ObjectIdentifier,
			TargetType: target.Kind(),
			RelType:    ad.HasSIDHistory,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	return data
}

func ParseChildObjects(data []TypedPrincipal, containerId string, containerType graph.Kind) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, len(data))
	for i, childObject := range data {
		relationships[i] = IngestibleRelationship{
			Source:     containerId,
			SourceType: containerType,
			TargetType: childObject.Kind(),
			Target:     childObject.ObjectIdentifier,
			RelProps:   map[string]any{"isacl": false},
			RelType:    ad.Contains,
		}
	}

	return relationships
}
func ParseGpLinks(links []GPLink, itemIdentifier string, itemType graph.Kind) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, len(links))
	for i, gpLink := range links {
		relationships[i] = IngestibleRelationship{
			Source:     gpLink.Guid,
			SourceType: ad.GPO,
			Target:     itemIdentifier,
			TargetType: itemType,
			RelProps:   map[string]any{"isacl": false, "enforced": gpLink.IsEnforced},
			RelType:    ad.GPLink,
		}
	}

	return relationships
}

func ParseDomainTrusts(domain Domain) ParsedDomainTrustData {
	parsedData := ParsedDomainTrustData{}
	for _, trust := range domain.Trusts {
		parsedData.ExtraNodeProps = append(parsedData.ExtraNodeProps, IngestibleNode{
			PropertyMap: map[string]any{"name": trust.TargetDomainName},
			ObjectID:    trust.TargetDomainSid,
			Label:       ad.Domain,
		})

		var dir = trust.TrustDirection
		if dir == TrustDirectionInbound || dir == TrustDirectionBidirectional {
			parsedData.TrustRelationships = append(parsedData.TrustRelationships, IngestibleRelationship{
				Source:     domain.ObjectIdentifier,
				SourceType: ad.Domain,
				Target:     trust.TargetDomainSid,
				TargetType: ad.Domain,
				RelProps: map[string]any{
					"isacl":        false,
					"sidfiltering": trust.SidFilteringEnabled,
					"trusttype":    trust.TrustType,
					"transitive":   trust.IsTransitive},
				RelType: ad.TrustedBy,
			})
		}

		if dir == TrustDirectionOutbound || dir == TrustDirectionBidirectional {
			parsedData.TrustRelationships = append(parsedData.TrustRelationships, IngestibleRelationship{
				Source:     trust.TargetDomainSid,
				SourceType: ad.Domain,
				Target:     domain.ObjectIdentifier,
				TargetType: ad.Domain,
				RelProps: map[string]any{
					"isacl":        false,
					"sidfiltering": trust.SidFilteringEnabled,
					"trusttype":    trust.TrustType,
					"transitive":   trust.IsTransitive},
				RelType: ad.TrustedBy,
			})
		}
	}

	return parsedData
}

// ParseComputerMiscData parses AllowedToDelegate, AllowedToAct, HasSIDHistory,DumpSMSAPassword and Sessions
func ParseComputerMiscData(computer Computer) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, target := range computer.AllowedToDelegate {
		relationships = append(relationships, IngestibleRelationship{
			Source:     computer.ObjectIdentifier,
			SourceType: ad.Computer,
			Target:     target.ObjectIdentifier,
			TargetType: target.Kind(),
			RelType:    ad.AllowedToDelegate,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	for _, actor := range computer.AllowedToAct {
		relationships = append(relationships, IngestibleRelationship{
			Source:     actor.ObjectIdentifier,
			SourceType: actor.Kind(),
			Target:     computer.ObjectIdentifier,
			TargetType: ad.Computer,
			RelType:    ad.AllowedToAct,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	for _, target := range computer.DumpSMSAPassword {
		relationships = append(relationships, IngestibleRelationship{
			Source:     computer.ObjectIdentifier,
			SourceType: ad.Computer,
			Target:     target.ObjectIdentifier,
			TargetType: target.Kind(),
			RelType:    ad.DumpSMSAPassword,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	for _, target := range computer.HasSIDHistory {
		relationships = append(relationships, IngestibleRelationship{
			Source:     computer.ObjectIdentifier,
			SourceType: ad.Computer,
			Target:     target.ObjectIdentifier,
			TargetType: target.Kind(),
			RelType:    ad.HasSIDHistory,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	if computer.Sessions.Collected {
		for _, session := range computer.Sessions.Results {
			relationships = append(relationships, IngestibleRelationship{
				Source:     session.ComputerSID,
				SourceType: ad.Computer,
				Target:     session.UserSID,
				TargetType: ad.User,
				RelType:    ad.HasSession,
				RelProps:   map[string]any{"isacl": false},
			})
		}
	}

	if computer.PrivilegedSessions.Collected {
		for _, session := range computer.PrivilegedSessions.Results {
			relationships = append(relationships, IngestibleRelationship{
				Source:     session.ComputerSID,
				SourceType: ad.Computer,
				Target:     session.UserSID,
				TargetType: ad.User,
				RelType:    ad.HasSession,
				RelProps:   map[string]any{"isacl": false},
			})
		}
	}

	if computer.RegistrySessions.Collected {
		for _, session := range computer.RegistrySessions.Results {
			relationships = append(relationships, IngestibleRelationship{
				Source:     session.ComputerSID,
				SourceType: ad.Computer,
				Target:     session.UserSID,
				TargetType: ad.User,
				RelType:    ad.HasSession,
				RelProps:   map[string]any{"isacl": false},
			})
		}
	}

	return relationships
}

func ConvertLocalGroup(localGroup LocalGroupAPIResult, computer Computer) ParsedLocalGroupData {
	parsedData := ParsedLocalGroupData{}
	if localGroup.Name != IgnoredName {
		parsedData.Nodes = append(parsedData.Nodes, IngestibleNode{
			ObjectID: localGroup.ObjectIdentifier,
			PropertyMap: map[string]any{
				"name": localGroup.Name,
			},
			Label: ad.LocalGroup,
		})
	}

	parsedData.Relationships = append(parsedData.Relationships, IngestibleRelationship{
		Source:     localGroup.ObjectIdentifier,
		SourceType: ad.LocalGroup,
		TargetType: ad.Computer,
		Target:     computer.ObjectIdentifier,
		RelProps:   map[string]any{"isacl": false},
		RelType:    ad.LocalToComputer,
	})

	for _, member := range localGroup.Results {
		parsedData.Relationships = append(parsedData.Relationships, IngestibleRelationship{
			Source:     member.ObjectIdentifier,
			SourceType: member.Kind(),
			TargetType: ad.LocalGroup,
			Target:     localGroup.ObjectIdentifier,
			RelProps:   map[string]any{"isacl": false},
			RelType:    ad.MemberOfLocalGroup,
		})
	}

	for _, name := range localGroup.LocalNames {
		parsedData.Nodes = append(parsedData.Nodes, IngestibleNode{
			ObjectID: name.ObjectIdentifier,
			PropertyMap: map[string]any{
				"name": name.PrincipalName,
			},
			Label: ad.Entity,
		})
	}

	return parsedData
}

func ParseUserRightData(userRight UserRightsAssignmentAPIResult, computer Computer, right graph.Kind) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, grant := range userRight.Results {
		relationships = append(relationships, IngestibleRelationship{
			Source:     grant.ObjectIdentifier,
			SourceType: grant.Kind(),
			TargetType: ad.Computer,
			Target:     computer.ObjectIdentifier,
			RelProps:   map[string]any{"isacl": false},
			RelType:    right,
		})
	}

	return relationships
}

func ParseEnterpriseCAMiscData(enterpriseCA EnterpriseCA) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	enabledCertTemplates := make([]string, 0)

	for _, actor := range enterpriseCA.EnabledCertTemplates {
		enabledCertTemplates = append(enabledCertTemplates, actor.ObjectIdentifier)
		relationships = append(relationships, IngestibleRelationship{
			Source:     actor.ObjectIdentifier,
			SourceType: ad.CertTemplate,
			Target:     enterpriseCA.ObjectIdentifier,
			TargetType: ad.EnterpriseCA,
			RelType:    ad.PublishedTo,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	if enterpriseCA.HostingComputer != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     enterpriseCA.HostingComputer,
			SourceType: ad.Computer,
			Target:     enterpriseCA.ObjectIdentifier,
			TargetType: ad.EnterpriseCA,
			RelType:    ad.HostsCAService,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	// if enterpriseCA.CARegistryData != "" {
	// 	//TODO: Handle CASecurity

	// 	if enterpriseCA.CARegistryData.EnrollmentAgentRestrictionsCollected {
	// 		for _, restiction := range enterpriseCA.CARegistryData.EnrollmentAgentRestrictions {
	// 			if restiction.AccessType == "AccessAllowedCallback" {
	// 				templates := make([]string, 0)
	// 				if restiction.AllTemplates {
	// 					templates = enabledCertTemplates
	// 				}
	// 				else {
	// 					templates = append(templates, restiction.Template.ObjectIdentifier)
	// 				}

	// 				// TODO: Handle Targets

	// 				for _, template := range templates {
	// 					relationships = append(relationships, IngestibleRelationship{
	// 						Source:     restiction.Agent.ObjectIdentifier,
	// 						SourceType: restiction.Agent.Kind(),
	// 						Target:     template,
	// 						TargetType: ad.CertTemplate,
	// 						RelType:    ad.DelegatedEnrollmentAgent,
	// 						RelProps:   map[string]any{"isacl": false},
	// 					})

	// 				}
	// 			}
	// 		}
	// 	}
	// }

	return relationships
}

func ParseRootCAMiscData(rootCA RootCA) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	// if domainID, err := domain.Properties.Get(common.domainID.String()).String(); err != nil {
	// 	return nil, err
	// } else if domainID == objectID {

	// if rootCA.Properties.domainsid != "" {
	// 	relationships = append(relationships, IngestibleRelationship{
	// 		Source:     rootCA.ObjectIdentifier,
	// 		SourceType: ad.RootCA,
	// 		Target:     rootCA.Properties.domainsid,
	// 		TargetType: ad.Domain,
	// 		RelType:    ad.RootCAFor,
	// 		RelProps:   map[string]any{"isacl": false},
	// 	})
	// }

	return relationships
}

type CertificateMappingMethod int

const (
	CertificateMappingManytoMany                  CertificateMappingMethod = 1
	CertificateMappingOneToOne                    CertificateMappingMethod = 1 << 1
	CertificateMappingUserPrincipalName           CertificateMappingMethod = 1 << 2
	CertificateMappingKerberosCertificate         CertificateMappingMethod = 1 << 3
	CertificateMappingKerberosExplicitCertificate CertificateMappingMethod = 1 << 4
)

func ParseDCRegistryData(computer Computer) IngestibleNode {
	var (
		prettyCertificateMappingMethodMappings map[string]string = map[string]string{
			"01": "0x01: Many-to-one (issuer certificate)",
			"02": "0x02: One-to-one (subject/issuer)",
			"04": "0x04: User principal name (UPN/SAN)",
			"08": "0x08: Kerberos service-for-user (S4U) certificate",
			"10": "0x10: Kerberos service-for-user (S4U) explicit certificate",
		}
		prettyStrongCertificateBindingEnforcementMappings []string = []string{
			"0: Disabled",
			"1: Compatibility mode",
			"2: Full enforcement mode",
		}
	)
	propMap := make(map[string]any)

	if computer.DCRegistryData.CertificateMappingMethods.Collected && computer.DCRegistryData.CertificateMappingMethods.Value >= 0 {
		propMap["CertificateMappingMethodsCollected"] = true
		propMap["CertificateMappingMethodsHex"] = fmt.Sprintf("0x%02x", computer.DCRegistryData.CertificateMappingMethods.Value)

		var prettyMappings []string

		if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingManytoMany) != 0 {
			prettyMappings = append(prettyMappings, prettyCertificateMappingMethodMappings["01"])
		}
		if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingOneToOne) != 0 {
			prettyMappings = append(prettyMappings, prettyCertificateMappingMethodMappings["02"])
		}
		if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingUserPrincipalName) != 0 {
			prettyMappings = append(prettyMappings, prettyCertificateMappingMethodMappings["04"])
		}
		if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingKerberosCertificate) != 0 {
			prettyMappings = append(prettyMappings, prettyCertificateMappingMethodMappings["08"])
		}
		if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingKerberosExplicitCertificate) != 0 {
			prettyMappings = append(prettyMappings, prettyCertificateMappingMethodMappings["10"])
		}

		propMap["CertificateMappingMethodsPretty"] = prettyMappings
	}

	if computer.DCRegistryData.StrongCertificateBindingEnforcement.Collected {
		propMap["StrongCertificateBindingEnforcementCollected"] = true
		propMap["StrongCertificateBindingEnforcementInt"] = computer.DCRegistryData.StrongCertificateBindingEnforcement.Value
		propMap["StrongCertificateBindingEnforcementPretty"] = prettyStrongCertificateBindingEnforcementMappings[computer.DCRegistryData.StrongCertificateBindingEnforcement.Value]
	}

	return IngestibleNode{
		ObjectID:    computer.ObjectIdentifier,
		PropertyMap: propMap,
		Label:       ad.Computer,
	}
}
