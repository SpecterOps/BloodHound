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
	"strings"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/slicesext"
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

// ParseComputerMiscData parses AllowedToDelegate, AllowedToAct, HasSIDHistory,DumpSMSAPassword,DCFor and Sessions
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

	if computer.IsDC && computer.DomainSID != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     computer.ObjectIdentifier,
			SourceType: ad.Computer,
			TargetType: ad.Domain,
			Target:     computer.DomainSID,
			RelProps:   map[string]any{"isacl": false},
			RelType:    ad.DCFor,
		})
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

func ParseCARegistryProperties(enterpriseCA EnterpriseCA) IngestibleNode {
	propMap := make(map[string]any)

	// HasEnrollmentAgentRestrictions
	if enterpriseCA.CARegistryData.EnrollmentAgentRestrictions.Collected {

		if len(enterpriseCA.CARegistryData.EnrollmentAgentRestrictions.Restrictions) > 0 {
			propMap[ad.HasEnrollmentAgentRestrictions.String()] = true
		} else {
			propMap[ad.HasEnrollmentAgentRestrictions.String()] = false
		}
	}

	// IsUserSpecifiesSanEnabled
	if enterpriseCA.CARegistryData.IsUserSpecifiesSanEnabled.Collected {
		propMap[ad.IsUserSpecifiesSanEnabled.String()] = enterpriseCA.CARegistryData.IsUserSpecifiesSanEnabled.Value
	}

	return IngestibleNode{
		ObjectID:    enterpriseCA.ObjectIdentifier,
		PropertyMap: propMap,
		Label:       ad.EnterpriseCA,
	}
}

func ParseEnterpriseCAMiscData(enterpriseCA EnterpriseCA) []IngestibleRelationship {
	var (
		relationships        = make([]IngestibleRelationship, 0)
		enabledCertTemplates = make([]string, 0)
	)

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

	relationships = handleEnterpriseCAEnrollmentAgentRestrictions(enterpriseCA, relationships, enabledCertTemplates)
	relationships = handleEnterpriseCASecurity(enterpriseCA, relationships)

	return relationships
}

func handleEnterpriseCAEnrollmentAgentRestrictions(enterpriseCA EnterpriseCA, relationships []IngestibleRelationship, enabledCertTemplates []string) []IngestibleRelationship {

	if enterpriseCA.CARegistryData.EnrollmentAgentRestrictions.Collected {
		for _, restriction := range enterpriseCA.CARegistryData.EnrollmentAgentRestrictions.Restrictions {
			if restriction.AccessType == AccessAllowedCallback {
				templates := make([]string, 0)
				if restriction.AllTemplates {
					templates = enabledCertTemplates
				} else {
					templates = append(templates, restriction.Template.ObjectIdentifier)
				}

				for _, template := range templates {
					relationships = append(relationships, IngestibleRelationship{
						Source:     restriction.Agent.ObjectIdentifier,
						SourceType: restriction.Agent.Kind(),
						Target:     template,
						TargetType: ad.CertTemplate,
						RelType:    ad.DelegatedEnrollmentAgent,
						RelProps:   map[string]any{"isacl": false},
					})

				}
			}
		}
	}

	return relationships
}

func handleEnterpriseCASecurity(enterpriseCA EnterpriseCA, relationships []IngestibleRelationship) []IngestibleRelationship {
	if enterpriseCA.CARegistryData.CASecurity.Collected {
		caSecurityData := slicesext.Filter(enterpriseCA.CARegistryData.CASecurity.Data, func(s ACE) bool {
			if s.PrincipalType == ad.LocalGroup.String() {
				return false
			}
			if s.RightName == ad.Owns.String() {
				return false
			} else {
				return true
			}
		})

		filteredACES := slicesext.Filter(enterpriseCA.Aces, func(s ACE) bool {
			if s.PrincipalSID == enterpriseCA.HostingComputer {
				return true
			} else {
				if s.RightName == ad.ManageCA.String() || s.RightName == ad.ManageCertificates.String() || s.RightName == ad.Enroll.String() {
					return false
				} else {
					return true
				}
			}
		})

		combinedData := append(caSecurityData, filteredACES...)
		relationships = append(relationships, ParseACEData(combinedData, enterpriseCA.ObjectIdentifier, ad.EnterpriseCA)...)

	} else {
		relationships = append(relationships, ParseACEData(enterpriseCA.Aces, enterpriseCA.ObjectIdentifier, ad.EnterpriseCA)...)
	}

	return relationships
}

func ParseRootCAMiscData(rootCA RootCA) []IngestibleRelationship {
	var (
		relationships = make([]IngestibleRelationship, 0)
		domainsid     = rootCA.DomainSID
	)

	if domainsid != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     rootCA.ObjectIdentifier,
			SourceType: ad.RootCA,
			Target:     domainsid,
			TargetType: ad.Domain,
			RelType:    ad.RootCAFor,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	return relationships
}

func ParseNTAuthStoreData(ntAuthStore NTAuthStore) []IngestibleRelationship {
	var (
		relationships = make([]IngestibleRelationship, 0)
		domainsid     = ntAuthStore.DomainSID
	)

	if domainsid != "" {
		relationships = append(relationships, IngestibleRelationship{
			Source:     ntAuthStore.ObjectIdentifier,
			SourceType: ad.NTAuthStore,
			Target:     domainsid,
			TargetType: ad.Domain,
			RelType:    ad.NTAuthStoreFor,
			RelProps:   map[string]any{"isacl": false},
		})
	}

	return relationships
}

type CertificateMappingMethod int

const (
	CertificateMappingManytoMany                     CertificateMappingMethod = 1
	CertificateMappingOneToOne                       CertificateMappingMethod = 1 << 1
	CertificateMappingUserPrincipalName              CertificateMappingMethod = 1 << 2
	CertificateMappingKerberosS4UCertificate         CertificateMappingMethod = 1 << 3
	CertificateMappingKerberosS4UExplicitCertificate CertificateMappingMethod = 1 << 4
)

// Prettified definitions for DCRegistryData
const (
	RegValNotExisting = "Registry value does not exist"

	PrettyCertMappingManyToOne                      = "0x01: Many-to-one (issuer certificate)"
	PrettyCertMappingOneToOne                       = "0x02: One-to-one (subject/issuer)"
	PrettyCertMappingUserPrincipalName              = "0x04: User principal name (UPN/SAN)"
	PrettyCertMappingKerberosS4UCertificate         = "0x08: Kerberos service-for-user (S4U) certificate"
	PrettyCertMappingKerberosS4UExplicitCertificate = "0x10: Kerberos service-for-user (S4U) explicit certificate"

	PrettyStrongCertBindingEnforcementDisabled      = "Disabled"
	PrettyStrongCertBindingEnforcementCompatibility = "Compatibility mode"
	PrettyStrongCertBindingEnforcementFull          = "Full enforcement mode"
)

func ParseDCRegistryData(computer Computer) IngestibleNode {
	var ()
	propMap := make(map[string]any)

	if computer.DCRegistryData.CertificateMappingMethods.Collected {
		propMap[ad.CertificateMappingMethodsRaw.String()] = computer.DCRegistryData.CertificateMappingMethods.Value
		var prettyMappings []string

		if computer.DCRegistryData.CertificateMappingMethods.Value == -1 {
			prettyMappings = append(prettyMappings, RegValNotExisting)
		} else {
			if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingManytoMany) != 0 {
				prettyMappings = append(prettyMappings, PrettyCertMappingManyToOne)
			}
			if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingOneToOne) != 0 {
				prettyMappings = append(prettyMappings, PrettyCertMappingOneToOne)
			}
			if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingUserPrincipalName) != 0 {
				prettyMappings = append(prettyMappings, PrettyCertMappingUserPrincipalName)
			}
			if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingKerberosS4UCertificate) != 0 {
				prettyMappings = append(prettyMappings, PrettyCertMappingKerberosS4UCertificate)
			}
			if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingKerberosS4UExplicitCertificate) != 0 {
				prettyMappings = append(prettyMappings, PrettyCertMappingKerberosS4UExplicitCertificate)
			}
		}

		propMap[ad.CertificateMappingMethods.String()] = prettyMappings
	}

	if computer.DCRegistryData.StrongCertificateBindingEnforcement.Collected {
		propMap[ad.StrongCertificateBindingEnforcementRaw.String()] = computer.DCRegistryData.StrongCertificateBindingEnforcement.Value

		switch computer.DCRegistryData.StrongCertificateBindingEnforcement.Value {
		case -1:
			propMap[ad.StrongCertificateBindingEnforcement.String()] = RegValNotExisting
		case 0:
			propMap[ad.StrongCertificateBindingEnforcement.String()] = PrettyStrongCertBindingEnforcementDisabled
		case 1:
			propMap[ad.StrongCertificateBindingEnforcement.String()] = PrettyStrongCertBindingEnforcementCompatibility
		case 2:
			propMap[ad.StrongCertificateBindingEnforcement.String()] = PrettyStrongCertBindingEnforcementFull
		}
	}

	return IngestibleNode{
		ObjectID:    computer.ObjectIdentifier,
		PropertyMap: propMap,
		Label:       ad.Computer,
	}
}
