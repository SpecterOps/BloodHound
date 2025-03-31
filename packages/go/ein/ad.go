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
	"log/slog"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/analysis"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/common"
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
	itemProps := item.Properties
	if itemProps == nil {
		itemProps = make(map[string]any)
	}

	if itemType == ad.Domain {
		convertInvalidDomainProperties(itemProps)
	}

	convertOwnsEdgeToProperty(item, itemProps)

	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: itemProps,
		Label:       itemType,
	}
}

func ConvertComputerToNode(item Computer) IngestibleNode {
	itemProps := item.Properties
	if itemProps == nil {
		itemProps = make(map[string]any)
	}

	convertOwnsEdgeToProperty(item.IngestBase, itemProps)

	if item.IsWebClientRunning.Collected {
		itemProps[ad.WebClientRunning.String()] = item.IsWebClientRunning.Result
	}

	if item.SmbInfo.Collected {
		itemProps[ad.SMBSigning.String()] = item.SmbInfo.SigningEnabled
	}

	if item.NTLMRegistryData.Collected {
		/*
			RestrictSendingNtlmTraffic is sent to us as an uint
			The possible values are
				0: Allow All
				1: Audit All
				2: Deny All
		*/
		if item.NTLMRegistryData.RestrictSendingNtlmTraffic == 0 {
			itemProps[ad.RestrictOutboundNTLM.String()] = false
		} else {
			itemProps[ad.RestrictOutboundNTLM.String()] = true
		}
	}

	if ldapEnabled, ok := itemProps["ldapenabled"]; ok {
		delete(itemProps, "ldapenabled")
		itemProps[ad.LDAPAvailable.String()] = ldapEnabled
	}

	if ldapsEnabled, ok := itemProps["ldapsenabled"]; ok {
		delete(itemProps, "ldapsenabled")
		itemProps[ad.LDAPSAvailable.String()] = ldapsEnabled
	}

	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: itemProps,
		Label:       ad.Computer,
	}
}

func ConvertEnterpriseCAToNode(item EnterpriseCA) IngestibleNode {
	itemProps := item.Properties
	if itemProps == nil {
		itemProps = make(map[string]any)
	}

	convertOwnsEdgeToProperty(item.IngestBase, itemProps)

	var (
		httpEndpoints    = make([]string, 0)
		httpsEndpoints   = make([]string, 0)
		hasCollectedData bool
	)

	for _, endpoint := range item.HttpEnrollmentEndpoints {
		if !endpoint.Collected {
			continue
		}

		hasCollectedData = true

		if endpoint.Result.ADCSWebEnrollmentHTTP {
			httpEndpoints = append(httpEndpoints, endpoint.Result.Url)
		}

		if endpoint.Result.ADCSWebEnrollmentHTTPS && !endpoint.Result.ADCSWebEnrollmentEPA {
			httpsEndpoints = append(httpsEndpoints, endpoint.Result.Url)
		}
	}

	if len(httpEndpoints) > 0 && len(httpsEndpoints) > 0 {
		itemProps[ad.HTTPEnrollmentEndpoints.String()] = httpEndpoints
		itemProps[ad.HTTPSEnrollmentEndpoints.String()] = httpsEndpoints
		itemProps[ad.HasVulnerableEndpoint.String()] = true
	} else if len(httpsEndpoints) > 0 {
		itemProps[ad.HTTPSEnrollmentEndpoints.String()] = httpsEndpoints
		itemProps[ad.HasVulnerableEndpoint.String()] = true
	} else if len(httpEndpoints) > 0 {
		itemProps[ad.HTTPSEnrollmentEndpoints.String()] = httpsEndpoints
		itemProps[ad.HasVulnerableEndpoint.String()] = true
	} else if hasCollectedData {
		//If we have collected data but no endpoints, we can mark this enterprise CA as not having a vulnerable endpoint
		itemProps[ad.HasVulnerableEndpoint.String()] = false
	}

	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: itemProps,
		Label:       ad.EnterpriseCA,
	}
}

// This function is to support our new method of doing Owns edges and makes older data sets backwards compatible
func convertOwnsEdgeToProperty(item IngestBase, itemProps map[string]any) {
	for _, ace := range item.Aces {
		if rightName, err := analysis.ParseKind(ace.RightName); err != nil {
			continue
		} else if rightName.Is(ad.Owns) || rightName.Is(ad.OwnsRaw) {
			itemProps[ad.OwnerSid.String()] = ace.PrincipalSID
			return
		}
	}
}

func convertInvalidDomainProperties(itemProps map[string]any) {
	convertProperty(itemProps, "machineaccountquota", stringToInt)
	convertProperty(itemProps, "minpwdlength", stringToInt)
	convertProperty(itemProps, "pwdproperties", stringToInt)
	convertProperty(itemProps, "pwdhistorylength", stringToInt)
	convertProperty(itemProps, "lockoutthreshold", stringToInt)
	convertProperty(itemProps, "expirepasswordsonsmartcardonlyaccounts", stringToBool)
}

func convertProperty(itemProps map[string]any, keyName string, conversionFunction func(map[string]any, string)) {
	conversionFunction(itemProps, keyName)
}

func stringToBool(itemProps map[string]any, keyName string) {
	if rawProperty, ok := itemProps[keyName]; ok {
		switch converted := rawProperty.(type) {
		case string:
			if final, err := strconv.ParseBool(converted); err != nil {
				delete(itemProps, keyName)
			} else {
				itemProps[keyName] = final
			}
		case bool:
		// pass
		default:
			slog.Debug(fmt.Sprintf("Removing %s with type %T", converted, converted))
			delete(itemProps, keyName)
		}
	}
}

func stringToInt(itemProps map[string]any, keyName string) {
	if rawProperty, ok := itemProps[keyName]; ok {
		switch converted := rawProperty.(type) {
		case string:
			if final, err := strconv.Atoi(converted); err != nil {
				delete(itemProps, keyName)
			} else {
				itemProps[keyName] = final
			}
		case int:
		// pass
		default:
			slog.Debug(fmt.Sprintf("Removing %s with type %T", keyName, converted))
			delete(itemProps, keyName)
		}
	}
}

func ParseObjectContainer(item IngestBase, itemType graph.Kind) IngestibleRelationship {
	containingPrincipal := item.ContainedBy
	if containingPrincipal.ObjectIdentifier != "" {
		return NewIngestibleRelationship(
			IngestibleSource{
				Source:     containingPrincipal.ObjectIdentifier,
				SourceType: containingPrincipal.Kind(),
			},
			IngestibleTarget{
				Target:     item.ObjectIdentifier,
				TargetType: itemType,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.Contains,
			},
		)
	}

	// TODO: Decide if we even want empty rels in the first place
	return NewIngestibleRelationship(IngestibleSource{}, IngestibleTarget{}, IngestibleRel{})
}

func ParsePrimaryGroup(item IngestBase, itemType graph.Kind, primaryGroupSid string) IngestibleRelationship {
	if primaryGroupSid != "" {
		return NewIngestibleRelationship(
			IngestibleSource{
				Source:     item.ObjectIdentifier,
				SourceType: itemType,
			},
			IngestibleTarget{
				Target:     primaryGroupSid,
				TargetType: ad.Group,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false, "isprimarygroup": true},
				RelType:  ad.MemberOf,
			},
		)
	}

	// TODO: Decide if we even want empty rels in the first place
	return NewIngestibleRelationship(IngestibleSource{}, IngestibleTarget{}, IngestibleRel{})
}

func ParseGroupMembershipData(group Group) ParsedGroupMembershipData {
	result := ParsedGroupMembershipData{}
	for _, member := range group.Members {
		if strings.HasPrefix(member.ObjectIdentifier, "DN=") {
			result.DistinguishedNameMembers = append(result.DistinguishedNameMembers, NewIngestibleRelationship(
				IngestibleSource{
					Source:     member.ObjectIdentifier,
					SourceType: member.Kind(),
				},
				IngestibleTarget{
					Target:     group.ObjectIdentifier,
					TargetType: ad.Group,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): false, "isprimarygroup": false},
					RelType:  ad.MemberOf,
				},
			))
		} else {
			result.RegularMembers = append(result.RegularMembers, NewIngestibleRelationship(
				IngestibleSource{
					Source:     member.ObjectIdentifier,
					SourceType: member.Kind(),
				},
				IngestibleTarget{
					Target:     group.ObjectIdentifier,
					TargetType: ad.Group,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): false, "isprimarygroup": false},
					RelType:  ad.MemberOf,
				},
			))
		}
	}

	return result
}

type WriteOwnerLimitedPrincipal struct {
	SourceData  IngestibleSource
	IsInherited bool
}

// getFromPropertyMap attempts to look up a given key in the given properties map. If the value does not exist this
// function returns false and the zero-value of type T. If the value does exist but the type of the value does not
// correctly cast to type T this function returns false and the zero-value of type T. Lastly, if the value does exist
// and converts to type T, this function returns true and the typed value.
func getFromPropertyMap[T any](props map[string]any, keyName string) (T, bool) {
	var empty T

	if value, exists := props[keyName]; !exists {
		return empty, false
	} else if typedValue, typeOK := value.(T); !typeOK {
		return empty, false
	} else {
		return typedValue, true
	}
}

// ParseACEData parses Windows Access Control Entries (ACE) and also handles the first half of post-processing work
// for the following edges: Owns, OwnsLimitRights, WriteOwner WriteOwnerLimitedRights.
//
// Part of the goal of this function was to make it backwards compatible with older collectors (or third-party
// collectors). As such, this function will attempt to translate data as it comes in.
func ParseACEData(targetNode IngestibleNode, aces []ACE, targetID string, targetType graph.Kind) []IngestibleRelationship {
	var (
		ownerPrincipalInfo                   IngestibleSource
		ownerLimitedPrivs                    = make([]string, 0)
		writeOwnerLimitedPrivs               = make([]string, 0)
		potentialWriteOwnerLimitedPrincipals = make([]WriteOwnerLimitedPrincipal, 0)
		converted                            = make([]IngestibleRelationship, 0)
	)

	for _, ace := range aces {
		if ace.PrincipalSID == targetID {
			continue
		}

		if rightKind, err := analysis.ParseKind(ace.RightName); err != nil {
			slog.Error(fmt.Sprintf("Error during ParseACEData: %v", err))
			continue
		} else if !ad.IsACLKind(rightKind) {
			slog.Error(fmt.Sprintf("Non-ace edge type given to process aces: %s", ace.RightName))
			continue
		} else if rightKind.Is(ad.Owns) || rightKind.Is(ad.OwnsRaw) {
			// Get Owner SID from ACE granting Owns permission
			ownerPrincipalInfo = ace.GetCachedValue().SourceData
		} else if rightKind.Is(ad.WriteOwner) || rightKind.Is(ad.WriteOwnerRaw) {
			// Don't convert every WriteOwner permission to an edge, as they are not always abusable
			// Cache ACEs where WriteOwner permission is granted
			potentialWriteOwnerLimitedPrincipals = append(potentialWriteOwnerLimitedPrincipals, ace.GetCachedValue())
		} else if strings.HasSuffix(ace.PrincipalSID, "S-1-3-4") {
			// Cache ACEs where the OWNER RIGHTS SID is granted explicit abusable permissions
			ownerLimitedPrivs = append(ownerLimitedPrivs, rightKind.String())

			if ace.IsInherited {
				// If the ACE is inherited, it is abusable by principals with WriteOwner permission
				writeOwnerLimitedPrivs = append(writeOwnerLimitedPrivs, rightKind.String())
			} else {
				// If the ACE is not inherited, it not abusable after abusing WriteOwner
				continue
			}
		} else {
			// Create edges for all other ACEs
			converted = append(converted, NewIngestibleRelationship(
				IngestibleSource{
					Source:     ace.PrincipalSID,
					SourceType: ace.Kind(),
				},
				IngestibleTarget{
					Target:     targetID,
					TargetType: targetType,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): true, common.IsInherited.String(): ace.IsInherited},
					RelType:  rightKind,
				},
			))
		}
	}

	// TODO: When inheritance hashes are added, add them to these aces

	// Process abusable permissions granted to the OWNER RIGHTS SID if any were found
	if len(ownerLimitedPrivs) > 0 {

		// Create an OwnsLimitedRights edge containing all abusable permissions granted to the OWNER RIGHTS SID
		if ownerPrincipalInfo.Source != "" {
			converted = append(converted, NewIngestibleRelationship(
				ownerPrincipalInfo,
				IngestibleTarget{
					Target:     targetID,
					TargetType: targetType,
				},

				// Owns is never inherited
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): true, common.IsInherited.String(): false, "privileges": ownerLimitedPrivs},
					RelType:  ad.OwnsLimitedRights,
				},
			))
		}

		if len(writeOwnerLimitedPrivs) > 0 {
			// For every principal granted WriteOwner permission, create a WriteOwnerLimitedRights edge
			for _, limitedPrincipal := range potentialWriteOwnerLimitedPrincipals {
				converted = append(converted, NewIngestibleRelationship(
					limitedPrincipal.SourceData,
					IngestibleTarget{
						Target:     targetID,
						TargetType: targetType,
					},

					// Create an edge property containing an array of all INHERITED abusable permissions granted to the OWNER RIGHTS SID
					IngestibleRel{
						RelProps: map[string]any{ad.IsACL.String(): true, common.IsInherited.String(): limitedPrincipal.IsInherited, "privileges": writeOwnerLimitedPrivs},
						RelType:  ad.WriteOwnerLimitedRights,
					},
				))
			}
		} else {
			// We're dealing with a case where the OWNER RIGHTS SID is granted uninherited abusable permissions
			// We can tell if any non-abusable ACE is present and inherited by checking the DoesAnyInheritedAceGrantOwnerRights property
			// If there are no inherited abusable permissions but there are inherited non-abusable permissions,
			// the non-abusable permissions will NOT be deleted on ownership change, so WriteOwner will NOT be abusable
			if doesAnyInheritedAceGrantOwnerRights, hasValidProperty := getFromPropertyMap[bool](targetNode.PropertyMap, ad.DoesAnyInheritedAceGrantOwnerRights.String()); hasValidProperty && doesAnyInheritedAceGrantOwnerRights {
				return converted
			}
			// If the non-abusable rights were NOT inherited, they are deleted on ownership change and WriteOwner may be abusable
			// Post will determine if WriteOwner is abusable based on dSHeuristics:BlockOwnerImplicitRights enforcement and object type

			// If there are abusable permissions granted to the OWNER RIGHTS SID, but they are not inherited,
			// they will be deleted on ownership change and WriteOwner may be abusable, so create a WriteOwnerRaw
			// edge for post-processing so we can check BlockOwnerImplicitRights enforcement and object type
			for _, limitedPrincipal := range potentialWriteOwnerLimitedPrincipals {
				converted = append(converted, NewIngestibleRelationship(
					limitedPrincipal.SourceData,
					IngestibleTarget{
						Target:     targetID,
						TargetType: targetType,
					},
					IngestibleRel{
						RelProps: map[string]any{ad.IsACL.String(): true, common.IsInherited.String(): limitedPrincipal.IsInherited},
						RelType:  ad.WriteOwnerRaw,
					},
				))
			}
		}

	} else {
		// If no abusable permissions in the ACL are granted to the OWNER RIGHTS SID check whether any permissions were
		// granted to the OWNER RIGHTS SID that are not abusable
		if doesAnyAceGrantOwnerRights, hasValidProperty := getFromPropertyMap[bool](targetNode.PropertyMap, ad.DoesAnyAceGrantOwnerRights.String()); hasValidProperty && doesAnyAceGrantOwnerRights {
			// If the non-abusable rights were inherited, they are not deleted on ownership change and WriteOwner is not abusable
			// Do NOT create the OwnsRaw or WriteOwnerRaw edges if the OWNER RIGHTS SID has inherited, non-abusable permissions
			if doesAnyInheritedAceGrantOwnerRights, hasValidProperty := getFromPropertyMap[bool](targetNode.PropertyMap, ad.DoesAnyInheritedAceGrantOwnerRights.String()); hasValidProperty && doesAnyInheritedAceGrantOwnerRights {
				return converted
			} else {
				// If the non-abusable rights were NOT inherited, they are deleted on ownership change and WriteOwner may be abusable
				// Post will determine if WriteOwner is abusable based on dSHeuristics:BlockOwnerImplicitRights enforcement and object type
				// Create a non-traversable WriteOwnerRaw edge for post-processing
				for _, limitedPrincipal := range potentialWriteOwnerLimitedPrincipals {
					converted = append(converted, NewIngestibleRelationship(
						limitedPrincipal.SourceData,
						IngestibleTarget{
							Target:     targetID,
							TargetType: targetType,
						},
						IngestibleRel{
							RelProps: map[string]any{ad.IsACL.String(): true, common.IsInherited.String(): limitedPrincipal.IsInherited},
							RelType:  ad.WriteOwnerRaw,
						},
					))
				}

				// Don't add the OwnsRaw edge if the OWNER RIGHTS SID has uninherited, non-abusable permissions
				return converted
			}
		}

		// When the SharpHound collection does not include the doesanyacegrantownerrights property
		// Or when the doesanyinheritedacegrantownerrights property is false

		// Create a non-traversable OwnsRaw edge for post-processing
		if ownerPrincipalInfo.Source != "" {
			converted = append(converted, NewIngestibleRelationship(
				ownerPrincipalInfo,
				IngestibleTarget{
					Target:     targetID,
					TargetType: targetType,
				},

				// Owns is never inherited
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): true, common.IsInherited.String(): false},
					RelType:  ad.OwnsRaw,
				},
			))
		}

		// Create a non-traversable WriteOwnerRaw edge for post-processing
		for _, limitedPrincipal := range potentialWriteOwnerLimitedPrincipals {
			converted = append(converted, NewIngestibleRelationship(
				limitedPrincipal.SourceData,
				IngestibleTarget{
					Target:     targetID,
					TargetType: targetType,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): true, common.IsInherited.String(): limitedPrincipal.IsInherited},
					RelType:  ad.WriteOwnerRaw,
				},
			))
		}
	}

	return converted
}

func convertSPNData(spns []SPNTarget, sourceID string) []IngestibleRelationship {
	converted := make([]IngestibleRelationship, 0, len(spns))

	for _, s := range spns {
		if kind, err := analysis.ParseKind(s.Service); err != nil {
			slog.Error(fmt.Sprintf("Error during processSPNTargets: %v", err))
		} else {
			converted = append(converted, NewIngestibleRelationship(
				IngestibleSource{
					Source:     sourceID,
					SourceType: ad.User,
				},
				IngestibleTarget{
					Target:     s.ComputerSID,
					TargetType: ad.Computer,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): true, "port": s.Port},
					RelType:  kind,
				},
			))
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
		data = append(data, NewIngestibleRelationship(
			IngestibleSource{
				Source:     user.ObjectIdentifier,
				SourceType: ad.User,
			},
			IngestibleTarget{
				Target:     target.ObjectIdentifier,
				TargetType: target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.AllowedToDelegate,
			},
		))
	}

	for _, target := range user.HasSIDHistory {
		data = append(data, NewIngestibleRelationship(
			IngestibleSource{
				Source:     user.ObjectIdentifier,
				SourceType: ad.User,
			},
			IngestibleTarget{
				Target:     target.ObjectIdentifier,
				TargetType: target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.HasSIDHistory,
			},
		))
	}

	// CoerceToTGT / unconstrained delegation
	var (
		userProps        = graph.AsProperties(user.Properties)
		uncondel, _      = userProps.GetOrDefault(ad.UnconstrainedDelegation.String(), user.UnconstrainedDelegation).Bool() // SH v2.5.7 and earlier have unconstraineddelegation under 'Properties' only
		domainsid, _     = userProps.GetOrDefault(ad.DomainSID.String(), user.DomainSID).String()                           // SH v2.5.7 and earlier have domainsid under 'Properties' only
		validCoerceToTGT = uncondel && domainsid != ""
	)

	if validCoerceToTGT {
		data = append(data, NewIngestibleRelationship(
			IngestibleSource{
				Source:     user.ObjectIdentifier,
				SourceType: ad.User,
			},
			IngestibleTarget{
				Target:     domainsid,
				TargetType: ad.Domain,
			},
			IngestibleRel{
				RelProps: map[string]any{"isacl": false},
				RelType:  ad.CoerceToTGT,
			},
		))
	}

	return data
}

func ParseChildObjects(data []TypedPrincipal, containerId string, containerType graph.Kind) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0, len(data))
	for _, childObject := range data {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     containerId,
				SourceType: containerType,
			},
			IngestibleTarget{
				Target:     childObject.ObjectIdentifier,
				TargetType: childObject.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.Contains,
			},
		))
	}

	return relationships
}
func ParseGpLinks(links []GPLink, itemIdentifier string, itemType graph.Kind) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0, len(links))
	for _, gpLink := range links {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     gpLink.Guid,
				SourceType: ad.GPO,
			},
			IngestibleTarget{
				Target:     itemIdentifier,
				TargetType: itemType,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false, "enforced": gpLink.IsEnforced},
				RelType:  ad.GPLink,
			},
		))
	}

	return relationships
}

func ParseDomainTrusts(domain Domain) ParsedDomainTrustData {
	parsedData := ParsedDomainTrustData{}
	for _, trust := range domain.Trusts {
		var finalTrustAttributes int
		switch converted := trust.TrustAttributes.(type) {
		case string:
			if i, err := strconv.Atoi(converted); err != nil {
				slog.Error(fmt.Sprintf("Error converting trust attributes %s to int", converted))
				finalTrustAttributes = 0
			} else {
				finalTrustAttributes = i
			}
		case int:
			finalTrustAttributes = converted
		default:
			slog.Error(fmt.Sprintf("Error converting trust attributes %s to int", converted))
			finalTrustAttributes = 0
		}

		parsedData.ExtraNodeProps = append(parsedData.ExtraNodeProps, IngestibleNode{
			PropertyMap: map[string]any{"name": trust.TargetDomainName},
			ObjectID:    trust.TargetDomainSid,
			Label:       ad.Domain,
		})

		var dir = trust.TrustDirection
		if dir == TrustDirectionInbound || dir == TrustDirectionBidirectional {
			parsedData.TrustRelationships = append(parsedData.TrustRelationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     domain.ObjectIdentifier,
					SourceType: ad.Domain,
				},
				IngestibleTarget{
					Target:     trust.TargetDomainSid,
					TargetType: ad.Domain,
				},
				IngestibleRel{
					RelProps: map[string]any{
						ad.IsACL.String():      false,
						"sidfiltering":         trust.SidFilteringEnabled,
						"tgtdelegationenabled": trust.TGTDelegationEnabled,
						"trustattributes":      finalTrustAttributes,
						"trusttype":            trust.TrustType,
						"transitive":           trust.IsTransitive},
					RelType: ad.TrustedBy,
				},
			))
		}

		if dir == TrustDirectionOutbound || dir == TrustDirectionBidirectional {
			parsedData.TrustRelationships = append(parsedData.TrustRelationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     trust.TargetDomainSid,
					SourceType: ad.Domain,
				},
				IngestibleTarget{
					Target:     domain.ObjectIdentifier,
					TargetType: ad.Domain,
				},
				IngestibleRel{
					RelProps: map[string]any{
						ad.IsACL.String():      false,
						"sidfiltering":         trust.SidFilteringEnabled,
						"tgtdelegationenabled": trust.TGTDelegationEnabled,
						"trustattributes":      finalTrustAttributes,
						"trusttype":            trust.TrustType,
						"transitive":           trust.IsTransitive},
					RelType: ad.TrustedBy,
				},
			))
		}
	}

	return parsedData
}

// ParseComputerMiscData parses AllowedToDelegate, AllowedToAct, HasSIDHistory, DumpSMSAPassword, DCFor, Sessions, and CoerceToTGT
func ParseComputerMiscData(computer Computer) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, target := range computer.AllowedToDelegate {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     computer.ObjectIdentifier,
				SourceType: ad.Computer,
			},
			IngestibleTarget{
				Target:     target.ObjectIdentifier,
				TargetType: target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.AllowedToDelegate,
			},
		))
	}

	for _, actor := range computer.AllowedToAct {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     actor.ObjectIdentifier,
				SourceType: actor.Kind(),
			},
			IngestibleTarget{
				Target:     computer.ObjectIdentifier,
				TargetType: ad.Computer,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.AllowedToAct,
			},
		))
	}

	for _, target := range computer.DumpSMSAPassword {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     computer.ObjectIdentifier,
				SourceType: ad.Computer,
			},
			IngestibleTarget{
				Target:     target.ObjectIdentifier,
				TargetType: target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.DumpSMSAPassword,
			},
		))
	}

	for _, target := range computer.HasSIDHistory {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     computer.ObjectIdentifier,
				SourceType: ad.Computer,
			},
			IngestibleTarget{
				Target:     target.ObjectIdentifier,
				TargetType: target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.HasSIDHistory,
			},
		))
	}

	if computer.Sessions.Collected {
		for _, session := range computer.Sessions.Results {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     session.ComputerSID,
					SourceType: ad.Computer,
				},
				IngestibleTarget{
					Target:     session.UserSID,
					TargetType: ad.User,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): false},
					RelType:  ad.HasSession,
				},
			))
		}
	}

	if computer.PrivilegedSessions.Collected {
		for _, session := range computer.PrivilegedSessions.Results {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     session.ComputerSID,
					SourceType: ad.Computer,
				},
				IngestibleTarget{
					Target:     session.UserSID,
					TargetType: ad.User,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): false},
					RelType:  ad.HasSession,
				},
			))
		}
	}

	if computer.RegistrySessions.Collected {
		for _, session := range computer.RegistrySessions.Results {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     session.ComputerSID,
					SourceType: ad.Computer,
				},
				IngestibleTarget{
					Target:     session.UserSID,
					TargetType: ad.User,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): false},
					RelType:  ad.HasSession,
				},
			))
		}
	}

	if computer.IsDC && computer.DomainSID != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     computer.ObjectIdentifier,
				SourceType: ad.Computer,
			},
			IngestibleTarget{
				Target:     computer.DomainSID,
				TargetType: ad.Domain,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.DCFor,
			},
		))
	} else { // We do not want CoerceToTGT edges from DCs
		var (
			computerProps    = graph.AsProperties(computer.Properties)
			uncondel, _      = computerProps.GetOrDefault(ad.UnconstrainedDelegation.String(), computer.UnconstrainedDelegation).Bool() // SH v2.5.7 and earlier have unconstraineddelegation under 'Properties' only
			domainsid, _     = computerProps.GetOrDefault(ad.DomainSID.String(), computer.DomainSID).String()                           // SH schema version 5 and earlier have domainsid under 'Properties' only
			validCoerceToTGT = uncondel && domainsid != ""
		)

		if validCoerceToTGT {
			relationships = append(relationships, NewIngestibleRelationship(
				IngestibleSource{
					Source:     computer.ObjectIdentifier,
					SourceType: ad.Computer,
				},
				IngestibleTarget{
					Target:     domainsid,
					TargetType: ad.Domain,
				},
				IngestibleRel{
					RelProps: map[string]any{"isacl": false},
					RelType:  ad.CoerceToTGT,
				},
			))
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

	parsedData.Relationships = append(parsedData.Relationships, NewIngestibleRelationship(
		IngestibleSource{
			Source:     localGroup.ObjectIdentifier,
			SourceType: ad.LocalGroup,
		},
		IngestibleTarget{
			Target:     computer.ObjectIdentifier,
			TargetType: ad.Computer,
		},
		IngestibleRel{
			RelProps: map[string]any{ad.IsACL.String(): false},
			RelType:  ad.LocalToComputer,
		},
	))

	for _, member := range localGroup.Results {
		parsedData.Relationships = append(parsedData.Relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     member.ObjectIdentifier,
				SourceType: member.Kind(),
			},
			IngestibleTarget{
				Target:     localGroup.ObjectIdentifier,
				TargetType: ad.LocalGroup,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.MemberOfLocalGroup,
			},
		))
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
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     grant.ObjectIdentifier,
				SourceType: grant.Kind(),
			},
			IngestibleTarget{
				Target:     computer.ObjectIdentifier,
				TargetType: ad.Computer,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  right,
			},
		))
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

	// RoleSeparationEnabled
	if enterpriseCA.CARegistryData.RoleSeparationEnabled.Collected {
		propMap[ad.RoleSeparationEnabled.String()] = enterpriseCA.CARegistryData.RoleSeparationEnabled.Value
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
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     actor.ObjectIdentifier,
				SourceType: ad.CertTemplate,
			},
			IngestibleTarget{
				Target:     enterpriseCA.ObjectIdentifier,
				TargetType: ad.EnterpriseCA,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.PublishedTo,
			},
		))
	}

	if enterpriseCA.HostingComputer != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     enterpriseCA.HostingComputer,
				SourceType: ad.Computer,
			},
			IngestibleTarget{
				Target:     enterpriseCA.ObjectIdentifier,
				TargetType: ad.EnterpriseCA,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.HostsCAService,
			},
		))
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
					relationships = append(relationships, NewIngestibleRelationship(
						IngestibleSource{
							Source:     restriction.Agent.ObjectIdentifier,
							SourceType: restriction.Agent.Kind(),
						},
						IngestibleTarget{
							Target:     template,
							TargetType: ad.CertTemplate,
						},
						IngestibleRel{
							RelProps: map[string]any{ad.IsACL.String(): false},
							RelType:  ad.DelegatedEnrollmentAgent,
						},
					))

				}
			}
		}
	}

	return relationships
}

func handleEnterpriseCASecurity(enterpriseCA EnterpriseCA, relationships []IngestibleRelationship) []IngestibleRelationship {

	// Get IngesibleNode for EnterpriceCA
	baseNodeProp := ConvertObjectToNode(enterpriseCA.IngestBase, ad.EnterpriseCA)

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
		relationships = append(relationships, ParseACEData(baseNodeProp, combinedData, enterpriseCA.ObjectIdentifier, ad.EnterpriseCA)...)

	} else {
		relationships = append(relationships, ParseACEData(baseNodeProp, enterpriseCA.Aces, enterpriseCA.ObjectIdentifier, ad.EnterpriseCA)...)
	}

	return relationships
}

func ParseRootCAMiscData(rootCA RootCA) []IngestibleRelationship {
	var (
		relationships = make([]IngestibleRelationship, 0)
		domainsid     = rootCA.DomainSID
	)

	if domainsid != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     rootCA.ObjectIdentifier,
				SourceType: ad.RootCA,
			},
			IngestibleTarget{
				Target:     domainsid,
				TargetType: ad.Domain,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.RootCAFor,
			},
		))
	}

	return relationships
}

func ParseNTAuthStoreData(ntAuthStore NTAuthStore) []IngestibleRelationship {
	var (
		relationships = make([]IngestibleRelationship, 0)
		domainsid     = ntAuthStore.DomainSID
	)

	if domainsid != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleSource{
				Source:     ntAuthStore.ObjectIdentifier,
				SourceType: ad.NTAuthStore,
			},
			IngestibleTarget{
				Target:     domainsid,
				TargetType: ad.Domain,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.NTAuthStoreFor,
			},
		))
	}

	return relationships
}

type CertificateMappingMethod int

const (
	RegistryValueDoesNotExist                                                 = -1
	CertificateMappingOneToOne                       CertificateMappingMethod = 1
	CertificateMappingManytoMany                     CertificateMappingMethod = 1 << 1
	CertificateMappingUserPrincipalName              CertificateMappingMethod = 1 << 2
	CertificateMappingKerberosS4UCertificate         CertificateMappingMethod = 1 << 3
	CertificateMappingKerberosS4UExplicitCertificate CertificateMappingMethod = 1 << 4
)

// Prettified definitions for DCRegistryData
const (
	RegValNotExisting = "Registry value does not exist"

	PrettyCertMappingOneToOne                       = "0x01: One-to-one (subject/issuer)"
	PrettyCertMappingManyToOne                      = "0x02: Many-to-one (issuer certificate)"
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

		if computer.DCRegistryData.CertificateMappingMethods.Value == RegistryValueDoesNotExist {
			prettyMappings = append(prettyMappings, RegValNotExisting)
		} else {
			if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingOneToOne) != 0 {
				prettyMappings = append(prettyMappings, PrettyCertMappingOneToOne)
			}
			if computer.DCRegistryData.CertificateMappingMethods.Value&int(CertificateMappingManytoMany) != 0 {
				prettyMappings = append(prettyMappings, PrettyCertMappingManyToOne)
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
