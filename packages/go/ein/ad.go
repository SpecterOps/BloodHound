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
	"github.com/specterops/bloodhound/slices"
)

var unenforcedDomainProperties = make(map[string]map[string]any) // key: objectid, value: {propertyKey: propertyValue}
var enforcedDomainProperties = make(map[string][]string)         // key: objectid, value: [propertyKey]

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

// ParseDomainMiscData parses misc data applied at the domain level
func ParseDomainMiscData(domains []Domain) []IngestibleNode {
	unenforcedDomainProperties = make(map[string]map[string]any)
	enforcedDomainProperties = make(map[string][]string)

	ingestibleNodes := make([]IngestibleNode, 0)
	for _, domain := range domains {

		// GPOChanges
		domainProperties := map[string]map[string]any{"unenforced": {}, "enforced": {}}
		linkTypes := [2]LinkType{domain.GPOChanges.Unenforced, domain.GPOChanges.Enforced}

		for index, linkType := range [2]string{"unenforced", "enforced"} {
			// Password policies
			if val, err := linkTypes[index].PasswordPolicies["MinimumPasswordAge"]; err != false {
				domainProperties[linkType]["minimumPasswordAge"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["MaximumPasswordAge"]; err != false {
				domainProperties[linkType]["maximumPasswordAge"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["MinimumPasswordLength"]; err != false {
				domainProperties[linkType]["minimumPasswordLength"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["PasswordComplexity"]; err != false {
				domainProperties[linkType]["passwordComplexity"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["PasswordHistorySize"]; err != false {
				domainProperties[linkType]["passwordHistorySize"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["ClearTextPassword"]; err != false {
				domainProperties[linkType]["clearTextPassword"] = val
			}

			// Lockout policies
			if val, err := linkTypes[index].LockoutPolicies["LockoutDuration"]; err != false {
				domainProperties[linkType]["lockoutDuration"] = val
			}
			if val, err := linkTypes[index].LockoutPolicies["LockoutBadCount"]; err != false {
				domainProperties[linkType]["lockoutBadCount"] = val
			}
			if val, err := linkTypes[index].LockoutPolicies["ResetLockoutCount"]; err != false {
				domainProperties[linkType]["resetLockoutCount"] = val
			}
			if val, err := linkTypes[index].LockoutPolicies["ForceLogoffWhenHourExpire"]; err != false {
				domainProperties[linkType]["forceLogoffWhenHourExpire"] = val
			}

			// SMB signings
			if val, err := linkTypes[index].SMBSigning["RequiresServerSMBSigning"]; err != false {
				domainProperties[linkType]["requiresServerSMBSigning"] = val
			}
			if val, err := linkTypes[index].SMBSigning["RequiresClientSMBSigning"]; err != false {
				domainProperties[linkType]["requiresClientSMBSigning"] = val
			}
			if val, err := linkTypes[index].SMBSigning["EnablesServerSMBSigning"]; err != false {
				domainProperties[linkType]["enablesServerSMBSigning"] = val
			}
			if val, err := linkTypes[index].SMBSigning["EnablesClientSMBSigning"]; err != false {
				domainProperties[linkType]["enablesClientSMBSigning"] = val
			}

			// LDAP signings
			if val, err := linkTypes[index].LDAPSigning["RequiresLDAPClientSigning"]; err != false {
				domainProperties[linkType]["requiresLDAPClientSigning"] = val
			}
			if val, err := linkTypes[index].LDAPSigning["LDAPEnforceChannelBinding"]; err != false {
				domainProperties[linkType]["LDAPEnforceChannelBinding"] = val
			}

			// LM authentication level
			if val, err := linkTypes[index].LMAuthenticationLevel["LmCompatibilityLevel"]; err != false {
				domainProperties[linkType]["lmCompatibilityLevel"] = val
			}

			// MSCache
			if val, err := linkTypes[index].MSCache["CachedLogonsCount"]; err != false {
				domainProperties[linkType]["cachedLogonsCount"] = val
			}
		}

		// add properties for each affected computer
		for _, computer := range domain.GPOChanges.AffectedComputers {
			computerId := computer.ObjectIdentifier

			// store unenforced properties to manage precedence with OU properties
			unenforcedDomainProperties[computerId] = domainProperties["unenforced"]

			// add enforced properties
			ingestibleNodes = append(ingestibleNodes, IngestibleNode{
				PropertyMap: domainProperties["enforced"],
				ObjectID:    computerId,
				Label:       ad.Computer,
			})

			// store defined property keys
			for key, value := range domainProperties["enforced"] {
				if value != nil {
					enforcedDomainProperties[computerId] = append(enforcedDomainProperties[computerId], key)
				}
			}
		}

		//DNS Property
		zones := domain.DNSProperty
		dNSProps := make(map[string]any)
		for zone, dNSProperties := range zones {
			for dNSProperty, value := range dNSProperties {

				switch dNSProperty {

				// allow_update
				case "allowUpdate":
					switch int(value.(float64)) {
					case 2:
						dNSProps[zone+" "+dNSProperty] = "secure"
						break
					default:
						dNSProps[zone+" "+dNSProperty] = "unsecure"
						break
					}
					break
				}
			}
		}
		ingestibleNodes = append(ingestibleNodes, IngestibleNode{
			PropertyMap: dNSProps,
			ObjectID:    domain.ObjectIdentifier,
			Label:       ad.Domain,
		})
	}
	return ingestibleNodes

}

// ParseOUMiscData parses GPOChanges data and manages prioritization of the policies
func ParseOUMiscData(ous []OU) []IngestibleNode {
	ingestibleNodes := make([]IngestibleNode, 0)
	var blockInheritanceComputers []string
	var propertiesComputers []map[string]map[string]any // key: objectid, value: { propKey: propValue }
	foundInCurrentProperties := false

	for _, ou := range ous {
		// add GPOChanges properties to the affected computers
		blockInheritance := ou.GPOChanges.BlockInheritance
		ouProperties := map[string]map[string]any{"unenforced": {}, "enforced": {}}
		linkTypes := [2]LinkType{ou.GPOChanges.Unenforced, ou.GPOChanges.Enforced}
		foundInEnforcedDomainProperties := false

		for index, linkType := range [2]string{"unenforced", "enforced"} {
			// Password policies
			if val, err := linkTypes[index].PasswordPolicies["MinimumPasswordAge"]; err != false {
				ouProperties[linkType]["minimumPasswordAge"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["MaximumPasswordAge"]; err != false {
				ouProperties[linkType]["maximumPasswordAge"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["MinimumPasswordLength"]; err != false {
				ouProperties[linkType]["minimumPasswordLength"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["PasswordComplexity"]; err != false {
				ouProperties[linkType]["passwordComplexity"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["PasswordHistorySize"]; err != false {
				ouProperties[linkType]["passwordHistorySize"] = val
			}
			if val, err := linkTypes[index].PasswordPolicies["ClearTextPassword"]; err != false {
				ouProperties[linkType]["clearTextPassword"] = val
			}

			// Lockout policies
			if val, err := linkTypes[index].LockoutPolicies["LockoutDuration"]; err != false {
				ouProperties[linkType]["lockoutDuration"] = val
			}
			if val, err := linkTypes[index].LockoutPolicies["LockoutBadCount"]; err != false {
				ouProperties[linkType]["lockoutBadCount"] = val
			}
			if val, err := linkTypes[index].LockoutPolicies["ResetLockoutCount"]; err != false {
				ouProperties[linkType]["resetLockoutCount"] = val
			}
			if val, err := linkTypes[index].LockoutPolicies["ForceLogoffWhenHourExpire"]; err != false {
				ouProperties[linkType]["forceLogoffWhenHourExpire"] = val
			}

			// SMB signings
			if val, err := linkTypes[index].SMBSigning["RequiresServerSMBSigning"]; err != false {
				ouProperties[linkType]["requiresServerSMBSigning"] = val
			}
			if val, err := linkTypes[index].SMBSigning["RequiresClientSMBSigning"]; err != false {
				ouProperties[linkType]["requiresClientSMBSigning"] = val
			}
			if val, err := linkTypes[index].SMBSigning["EnablesServerSMBSigning"]; err != false {
				ouProperties[linkType]["enablesServerSMBSigning"] = val
			}
			if val, err := linkTypes[index].SMBSigning["EnablesClientSMBSigning"]; err != false {
				ouProperties[linkType]["enablesClientSMBSigning"] = val
			}

			// LDAP signings
			if val, err := linkTypes[index].LDAPSigning["RequiresLDAPClientSigning"]; err != false {
				ouProperties[linkType]["requiresLDAPClientSigning"] = val
			}
			if val, err := linkTypes[index].LDAPSigning["LDAPEnforceChannelBinding"]; err != false {
				ouProperties[linkType]["LDAPEnforceChannelBinding"] = val
			}

			// LM authentication level
			if val, err := linkTypes[index].LMAuthenticationLevel["LmCompatibilityLevel"]; err != false {
				ouProperties[linkType]["lmCompatibilityLevel"] = val
			}

			// MSCache
			if val, err := linkTypes[index].MSCache["CachedLogonsCount"]; err != false {
				ouProperties[linkType]["cachedLogonsCount"] = val
			}
		}

		for _, computer := range ou.GPOChanges.AffectedComputers {

			computerIdentifier := computer.ObjectIdentifier
			foundInPropertiesComputers := false

			// remove unenforced OU properties overlapping with domain enforced properties
			for computerId, enforcedDomainProps := range enforcedDomainProperties {
				if computerId == computerIdentifier {
					for _, enforcedDomainProp := range enforcedDomainProps {
						delete(ouProperties["unenforced"], enforcedDomainProp)
					}
				}
			}

			// unenforced properties
			if !slices.Contains(blockInheritanceComputers, computerIdentifier) {
				// add, if the computer is already affected by properties
				for _, propertiesComputer := range propertiesComputers {
					for objectid := range propertiesComputer {
						if objectid == computerIdentifier {
							// add properties which do not exist yet and are not enforced at the domain level
							for unenforcedPropKey, unenforcedPropValue := range ouProperties["unenforced"] {
								foundInEnforcedDomainProperties = false
								for enforcedDomainPropertiesComputerIdentifier, enforcedDomainPropertyKeys := range enforcedDomainProperties {
									if enforcedDomainPropertiesComputerIdentifier == computerIdentifier && slices.Contains(enforcedDomainPropertyKeys, unenforcedPropKey) {
										foundInEnforcedDomainProperties = true
										break
									}
								}
								if !foundInEnforcedDomainProperties && propertiesComputer[computerIdentifier][unenforcedPropKey] == nil {
									propertiesComputer[computerIdentifier][unenforcedPropKey] = unenforcedPropValue
								}
							}
							foundInPropertiesComputers = true
						}
					}
				}

				// create, if the computer has not already been affected by properties
				if !foundInPropertiesComputers {
					propertiesComputers = append(propertiesComputers, map[string]map[string]any{computerIdentifier: copyMap(ouProperties["unenforced"])})
				}

				if blockInheritance {
					// add the computer to the block inheritance array
					blockInheritanceComputers = append(blockInheritanceComputers, computerIdentifier)
					// remove unenforced properties set by domains
					delete(unenforcedDomainProperties, computerIdentifier)
				}
			}

			// enforced properties
			// add, if this computer is already affected by properties
			foundInPropertiesComputers = false
			for _, propertiesComputer := range propertiesComputers {
				for objectid := range propertiesComputer {
					if objectid == computerIdentifier {
						for enforcedPropKey, enforcedPropValue := range ouProperties["enforced"] {
							// override the property if it is empty or not enforced at domain level
							foundInEnforcedDomainProperties = false
							for enforcedDomainPropertiesComputerIdentifier, enforcedDomainPropertyKeys := range enforcedDomainProperties {
								if enforcedDomainPropertiesComputerIdentifier == objectid && slices.Contains(enforcedDomainPropertyKeys, enforcedPropKey) {
									foundInEnforcedDomainProperties = true
									break
								}
							}
							if !foundInEnforcedDomainProperties {
								propertiesComputer[computerIdentifier][enforcedPropKey] = enforcedPropValue
							}
						}
						foundInPropertiesComputers = true
					}
				}
			}

			// create if the computer has not already been affected by properties
			if !foundInPropertiesComputers {
				propertiesComputers = append(propertiesComputers, map[string]map[string]any{computerIdentifier: copyMap(ouProperties["enforced"])})
			}
		}
	}

	// finally add domain unenforced properties, without overriding existing properties
	for domainComputerId, unenforcedDomainProps := range unenforcedDomainProperties {
		for _, propertiesComputer := range propertiesComputers {
			for computerId, properties := range propertiesComputer {
				if domainComputerId == computerId {
					for unenforcedDomainKey, unenforcedDomainValue := range unenforcedDomainProps {
						foundInCurrentProperties = false
						for propKey := range properties {
							if propKey == unenforcedDomainKey {
								foundInCurrentProperties = true
								break
							}
						}
						if !foundInCurrentProperties && !slices.Contains(enforcedDomainProperties[computerId], unenforcedDomainKey) {
							properties[unenforcedDomainKey] = unenforcedDomainValue
						}
					}
				}
			}
		}
	}

	for _, propertiesComputer := range propertiesComputers {
		for objectid, properties := range propertiesComputer {
			ingestibleNodes = append(ingestibleNodes, IngestibleNode{
				PropertyMap: properties,
				ObjectID:    objectid,
				Label:       ad.Computer,
			})
		}
	}

	return ingestibleNodes
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
