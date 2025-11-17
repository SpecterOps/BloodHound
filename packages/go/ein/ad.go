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
	"time"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/bloodhound/packages/go/slicesext"
	"github.com/specterops/dawgs/graph"
)

type LocalGroup struct {
	RID     string
	Members []TypedPrincipal
}

func ConvertSessionObject(session Session) IngestibleSession {
	return IngestibleSession{
		Source:    session.ComputerSID,
		Target:    session.UserSID,
		LogonType: session.LogonType,
	}
}

func ConvertObjectToNode(item IngestBase, itemType graph.Kind, ingestTime time.Time) IngestibleNode {
	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: getBaseProperties(item, ingestTime),
		Labels:      []graph.Kind{itemType},
	}
}

func ConvertDomainToNode(item Domain, ingestTime time.Time) IngestibleNode {
	itemProps := getBaseProperties(item.IngestBase, ingestTime)
	convertInvalidDomainProperties(itemProps)

	if len(item.InheritanceHashes) > 0 {
		itemProps[ad.InheritanceHashes.String()] = item.InheritanceHashes
	}

	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: itemProps,
		Labels:      []graph.Kind{ad.Domain},
	}
}

func ConvertOUToNode(item OU, ingestTime time.Time) IngestibleNode {
	itemProps := getBaseProperties(item.IngestBase, ingestTime)

	if len(item.InheritanceHashes) > 0 {
		itemProps[ad.InheritanceHashes.String()] = item.InheritanceHashes
	}

	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: itemProps,
		Labels:      []graph.Kind{ad.OU},
	}
}

func ConvertContainerToNode(item Container, ingestTime time.Time) IngestibleNode {
	itemProps := getBaseProperties(item.IngestBase, ingestTime)

	if len(item.InheritanceHashes) > 0 {
		itemProps[ad.InheritanceHashes.String()] = item.InheritanceHashes
	}

	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: itemProps,
		Labels:      []graph.Kind{ad.Container},
	}
}

func ConvertComputerToNode(item Computer, ingestTime time.Time) IngestibleNode {
	itemProps := getBaseProperties(item.IngestBase, ingestTime)

	if item.IsWebClientRunning.Collected {
		itemProps[ad.WebClientRunning.String()] = item.IsWebClientRunning.Result
	}

	if item.SmbInfo.Collected {
		itemProps[ad.SMBSigning.String()] = item.SmbInfo.Result.SigningEnabled
	}

	if item.NTLMRegistryData.Collected {
		// If a registry value doesn't exist, assign its item prop to nil to clear it from the node
		itemProps[ad.RestrictOutboundNTLM.String()] = nil
		itemProps[ad.RestrictReceivingNTLMTraffic.String()] = nil
		itemProps[ad.RequireSecuritySignature.String()] = nil
		itemProps[ad.EnableSecuritySignature.String()] = nil
		itemProps[ad.NTLMMinClientSec.String()] = nil
		itemProps[ad.NTLMMinServerSec.String()] = nil
		itemProps[ad.LMCompatibilityLevel.String()] = nil
		itemProps[ad.UseMachineID.String()] = nil
		itemProps[ad.ClientAllowedNTLMServers.String()] = nil

		/*
			RestrictSendingNtlmTraffic is sent to us as an uint if sent at all
			The possible values are
				0: Allow All
				1: Audit All
				2: Deny All
		*/
		if item.NTLMRegistryData.Result.RestrictSendingNtlmTraffic != nil {
			itemProps[ad.RestrictOutboundNTLM.String()] = *item.NTLMRegistryData.Result.RestrictSendingNtlmTraffic == 2
		}
		if item.NTLMRegistryData.Result.RestrictReceivingNTLMTraffic != nil {
			itemProps[ad.RestrictReceivingNTLMTraffic.String()] = *item.NTLMRegistryData.Result.RestrictReceivingNTLMTraffic == 2
		}
		if item.NTLMRegistryData.Result.RequireSecuritySignature != nil {
			itemProps[ad.RequireSecuritySignature.String()] = *item.NTLMRegistryData.Result.RequireSecuritySignature != 0
		}
		if item.NTLMRegistryData.Result.EnableSecuritySignature != nil {
			itemProps[ad.EnableSecuritySignature.String()] = *item.NTLMRegistryData.Result.EnableSecuritySignature != 0
		}
		if item.NTLMRegistryData.Result.NtlmMinClientSec != nil {
			itemProps[ad.NTLMMinClientSec.String()] = *item.NTLMRegistryData.Result.NtlmMinClientSec
		}
		if item.NTLMRegistryData.Result.NtlmMinServerSec != nil {
			itemProps[ad.NTLMMinServerSec.String()] = *item.NTLMRegistryData.Result.NtlmMinServerSec
		}
		if item.NTLMRegistryData.Result.LmCompatibilityLevel != nil {
			itemProps[ad.LMCompatibilityLevel.String()] = *item.NTLMRegistryData.Result.LmCompatibilityLevel
		}
		if item.NTLMRegistryData.Result.UseMachineId != nil {
			itemProps[ad.UseMachineID.String()] = *item.NTLMRegistryData.Result.UseMachineId != 0
		}
		if item.NTLMRegistryData.Result.ClientAllowedNTLMServers != nil {
			itemProps[ad.ClientAllowedNTLMServers.String()] = *item.NTLMRegistryData.Result.ClientAllowedNTLMServers
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
		Labels:      []graph.Kind{ad.Computer},
	}
}

func ConvertEnterpriseCAToNode(item EnterpriseCA, ingestTime time.Time) IngestibleNode {
	itemProps := getBaseProperties(item.IngestBase, ingestTime)

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
		itemProps[ad.HTTPEnrollmentEndpoints.String()] = httpEndpoints
		itemProps[ad.HasVulnerableEndpoint.String()] = true
	} else if hasCollectedData {
		// If we have collected data but no endpoints, we can mark this enterprise CA as not having a vulnerable endpoint
		itemProps[ad.HasVulnerableEndpoint.String()] = false
	}

	return IngestibleNode{
		ObjectID:    item.ObjectIdentifier,
		PropertyMap: itemProps,
		Labels:      []graph.Kind{ad.EnterpriseCA},
	}
}

func getBaseProperties(item IngestBase, ingestTime time.Time) map[string]any {
	itemProps := item.Properties
	if itemProps == nil {
		itemProps = make(map[string]any)
	}

	itemProps[common.LastCollected.String()] = ingestTime

	convertOwnsEdgeToProperty(item, itemProps)

	return itemProps
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
		case float32:
			itemProps[keyName] = int(converted)
		case float64:
			itemProps[keyName] = int(converted)
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
			IngestibleEndpoint{
				Value: containingPrincipal.ObjectIdentifier,
				Kind:  containingPrincipal.Kind(),
			},
			IngestibleEndpoint{
				Value: item.ObjectIdentifier,
				Kind:  itemType,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.Contains,
			},
		)
	}

	// TODO: Decide if we even want empty rels in the first place
	return NewIngestibleRelationship(IngestibleEndpoint{}, IngestibleEndpoint{}, IngestibleRel{})
}

func ParsePrimaryGroup(item IngestBase, itemType graph.Kind, primaryGroupSid string) IngestibleRelationship {
	if primaryGroupSid != "" {
		return NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: item.ObjectIdentifier,
				Kind:  itemType,
			},
			IngestibleEndpoint{
				Value: primaryGroupSid,
				Kind:  ad.Group,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false, "isprimarygroup": true},
				RelType:  ad.MemberOf,
			},
		)
	}

	// TODO: Decide if we even want empty rels in the first place
	return NewIngestibleRelationship(IngestibleEndpoint{}, IngestibleEndpoint{}, IngestibleRel{})
}

// ParseGroupMiscData parses HasSIDHistory
func ParseGroupMiscData(group Group) []IngestibleRelationship {
	data := make([]IngestibleRelationship, 0)

	for _, target := range group.HasSIDHistory {
		data = append(data, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: group.ObjectIdentifier,
				Kind:  ad.Group,
			},
			IngestibleEndpoint{
				Value: target.ObjectIdentifier,
				Kind:  target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.HasSIDHistory,
			},
		))
	}

	return data
}

func ParseGroupMembershipData(group Group) ParsedGroupMembershipData {
	result := ParsedGroupMembershipData{}
	for _, member := range group.Members {
		if strings.HasPrefix(member.ObjectIdentifier, "DN=") {
			result.DistinguishedNameMembers = append(result.DistinguishedNameMembers, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: member.ObjectIdentifier,
					Kind:  member.Kind(),
				},
				IngestibleEndpoint{
					Value: group.ObjectIdentifier,
					Kind:  ad.Group,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): false, "isprimarygroup": false},
					RelType:  ad.MemberOf,
				},
			))
		} else {
			result.RegularMembers = append(result.RegularMembers, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: member.ObjectIdentifier,
					Kind:  member.Kind(),
				},
				IngestibleEndpoint{
					Value: group.ObjectIdentifier,
					Kind:  ad.Group,
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
	SourceData      IngestibleEndpoint
	IsInherited     bool
	InheritanceHash string
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
		ownerPrincipalInfo                   IngestibleEndpoint
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
				IngestibleEndpoint{
					Value: ace.PrincipalSID,
					Kind:  ace.Kind(),
				},
				IngestibleEndpoint{
					Value: targetID,
					Kind:  targetType,
				},
				IngestibleRel{
					RelProps: map[string]any{
						ad.IsACL.String():           true,
						common.IsInherited.String(): ace.IsInherited,
						ad.InheritanceHash.String(): ace.InheritanceHash,
					},
					RelType: rightKind,
				},
			))
		}
	}

	// TODO: When inheritance hashes are added, add them to these aces

	// Process abusable permissions granted to the OWNER RIGHTS SID if any were found
	if len(ownerLimitedPrivs) > 0 {

		// Create an OwnsLimitedRights edge containing all abusable permissions granted to the OWNER RIGHTS SID
		if ownerPrincipalInfo.Value != "" {
			converted = append(converted, NewIngestibleRelationship(
				ownerPrincipalInfo,
				IngestibleEndpoint{
					Value: targetID,
					Kind:  targetType,
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
					IngestibleEndpoint{
						Value: targetID,
						Kind:  targetType,
					},

					// Create an edge property containing an array of all INHERITED abusable permissions granted to the OWNER RIGHTS SID
					IngestibleRel{
						RelProps: map[string]any{
							ad.IsACL.String():           true,
							common.IsInherited.String(): limitedPrincipal.IsInherited,
							ad.InheritanceHash.String(): limitedPrincipal.InheritanceHash,
							"privileges":                writeOwnerLimitedPrivs,
						},
						RelType: ad.WriteOwnerLimitedRights,
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
					IngestibleEndpoint{
						Value: targetID,
						Kind:  targetType,
					},
					IngestibleRel{
						RelProps: map[string]any{
							ad.IsACL.String():           true,
							common.IsInherited.String(): limitedPrincipal.IsInherited,
							ad.InheritanceHash.String(): limitedPrincipal.InheritanceHash,
						},
						RelType: ad.WriteOwnerRaw,
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
						IngestibleEndpoint{
							Value: targetID,
							Kind:  targetType,
						},
						IngestibleRel{
							RelProps: map[string]any{
								ad.IsACL.String():           true,
								common.IsInherited.String(): limitedPrincipal.IsInherited,
								ad.InheritanceHash.String(): limitedPrincipal.InheritanceHash,
							},
							RelType: ad.WriteOwnerRaw,
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
		if ownerPrincipalInfo.Value != "" {
			converted = append(converted, NewIngestibleRelationship(
				ownerPrincipalInfo,
				IngestibleEndpoint{
					Value: targetID,
					Kind:  targetType,
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
				IngestibleEndpoint{
					Value: targetID,
					Kind:  targetType,
				},
				IngestibleRel{
					RelProps: map[string]any{
						ad.IsACL.String():           true,
						common.IsInherited.String(): limitedPrincipal.IsInherited,
						ad.InheritanceHash.String(): limitedPrincipal.InheritanceHash,
					},
					RelType: ad.WriteOwnerRaw,
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
				IngestibleEndpoint{
					Value: sourceID,
					Kind:  ad.User,
				},
				IngestibleEndpoint{
					Value: s.ComputerSID,
					Kind:  ad.Computer,
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
			IngestibleEndpoint{
				Value: user.ObjectIdentifier,
				Kind:  ad.User,
			},
			IngestibleEndpoint{
				Value: target.ObjectIdentifier,
				Kind:  target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.AllowedToDelegate,
			},
		))
	}

	for _, target := range user.HasSIDHistory {
		data = append(data, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: user.ObjectIdentifier,
				Kind:  ad.User,
			},
			IngestibleEndpoint{
				Value: target.ObjectIdentifier,
				Kind:  target.Kind(),
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
			IngestibleEndpoint{
				Value: user.ObjectIdentifier,
				Kind:  ad.User,
			},
			IngestibleEndpoint{
				Value: domainsid,
				Kind:  ad.Domain,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
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
			IngestibleEndpoint{
				Value: containerId,
				Kind:  containerType,
			},
			IngestibleEndpoint{
				Value: childObject.ObjectIdentifier,
				Kind:  childObject.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.Contains,
			},
		))
	}

	return relationships
}

func ParseGPOChanges(changes GPOChanges) ParsedLocalGroupData {
	parsedData := ParsedLocalGroupData{}

	groups := []LocalGroup{
		{"544", changes.LocalAdmins},
		{"551", changes.BackupOperators},		
		{"555", changes.RemoteDesktopUsers},
		{"562", changes.DcomUsers},
		{"580", changes.PSRemoteUsers},
	}

	for _, computer := range changes.AffectedComputers {
		for _, group := range groups {
			groupID := computer.ObjectIdentifier + "-" + group.RID
			if len(group.Members) > 0 {
				parsedData.Nodes = append(parsedData.Nodes, IngestibleNode{
					ObjectID:    groupID,
					PropertyMap: map[string]any{},
					Labels:      []graph.Kind{ad.LocalGroup},
				})

				parsedData.Relationships = append(parsedData.Relationships, NewIngestibleRelationship(
					IngestibleEndpoint{
						Value: groupID,
						Kind:  ad.LocalGroup,
					},
					IngestibleEndpoint{
						Value: computer.ObjectIdentifier,
						Kind:  ad.Computer,
					},
					IngestibleRel{
						RelProps: map[string]any{ad.IsACL.String(): false},
						RelType:  ad.LocalToComputer,
					},
				))
			}

			for _, member := range group.Members {
				parsedData.Relationships = append(parsedData.Relationships, NewIngestibleRelationship(
					IngestibleEndpoint{
						Value: member.ObjectIdentifier,
						Kind:  member.Kind(),
					},
					IngestibleEndpoint{
						Value: groupID,
						Kind:  ad.LocalGroup,
					},
					IngestibleRel{
						RelProps: map[string]any{ad.IsACL.String(): false},
						RelType:  ad.MemberOfLocalGroup,
					},
				))
			}
		}
	}

	return parsedData
}

func ParseGpLinks(links []GPLink, itemIdentifier string, itemType graph.Kind) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0, len(links))
	for _, gpLink := range links {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: gpLink.Guid,
				Kind:  ad.GPO,
			},
			IngestibleEndpoint{
				Value: itemIdentifier,
				Kind:  itemType,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false, "enforced": gpLink.IsEnforced},
				RelType:  ad.GPLink,
			},
		))
	}

	return relationships
}

// ParseDomainTrusts converts the marshalled value of the domain's trust attributes to a valid int or nil
// and sets the trust relationships for the domain
func ParseDomainTrusts(domain Domain) ParsedDomainTrustData {
	parsedData := ParsedDomainTrustData{}

	for _, trust := range domain.Trusts {
		var (
			convertedTrustAttributes int
			invalidTrustAttribute    bool
			finalTrustAttributes     any
		)

		// The type of the TrustAttributes is decided during decoding due to the `any` type usage
		// We need to switch on the type and convert it to an int, or nil when there was an error parsing the value
		switch converted := trust.TrustAttributes.(type) {
		case string:
			if i, err := strconv.Atoi(converted); err != nil {
				slog.Warn(fmt.Sprintf("Error converting trust attributes with a string value of %s to an int", converted))
				invalidTrustAttribute = true
			} else {
				convertedTrustAttributes = i
			}
		case int:
			convertedTrustAttributes = converted
		case float32:
			convertedTrustAttributes = int(converted)
		case float64:
			convertedTrustAttributes = int(converted)
		default:
			slog.Warn(fmt.Sprintf("Unexpected trust attributes type of %T, failed to convert to an int", converted))
			invalidTrustAttribute = true
		}

		if invalidTrustAttribute {
			finalTrustAttributes = nil
		} else {
			finalTrustAttributes = convertedTrustAttributes
		}

		parsedData.ExtraNodeProps = append(parsedData.ExtraNodeProps, IngestibleNode{
			PropertyMap: map[string]any{"name": trust.TargetDomainName, ad.DomainSID.String(): trust.TargetDomainSid},
			ObjectID:    trust.TargetDomainSid,
			Labels:      []graph.Kind{ad.Domain},
		})

		// Determine edge type
		edgeType := ad.CrossForestTrust
		if trust.TrustType == "ParentChild" || trust.TrustType == "TreeRoot" || trust.TrustType == "CrossLink" {
			edgeType = ad.SameForestTrust
		}

		var dir = trust.TrustDirection
		if dir == TrustDirectionInbound || dir == TrustDirectionBidirectional {
			realProps := map[string]any{
				ad.IsACL.String():                  false,
				ad.TrustType.String():              trust.TrustType,
				ad.Transitive.String():             trust.IsTransitive,
				ad.TrustAttributesInbound.String(): finalTrustAttributes,
			}
			// Only add TGT delegation if trustattributes collected
			if finalTrustAttributes != nil {
				realProps[ad.TGTDelegation.String()] = trust.TGTDelegationEnabled // collection of tgtdelegation was added with trustattributes
			}

			parsedData.TrustRelationships = append(parsedData.TrustRelationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: trust.TargetDomainSid,
					Kind:  ad.Domain,
				},
				IngestibleEndpoint{
					Value: domain.ObjectIdentifier,
					Kind:  ad.Domain,
				},
				IngestibleRel{
					RelProps: realProps,
					RelType:  edgeType,
				},
			))

			if edgeType == ad.CrossForestTrust && trust.TGTDelegationEnabled {
				parsedData.TrustRelationships = append(parsedData.TrustRelationships, NewIngestibleRelationship(
					IngestibleEndpoint{
						Value: trust.TargetDomainSid,
						Kind:  ad.Domain,
					},
					IngestibleEndpoint{
						Value: domain.ObjectIdentifier,
						Kind:  ad.Domain,
					},
					IngestibleRel{
						RelProps: map[string]any{
							ad.IsACL.String():     false,
							ad.TrustType.String(): trust.TrustType,
						},
						RelType: ad.AbuseTGTDelegation,
					},
				))
			}
		}

		if dir == TrustDirectionOutbound || dir == TrustDirectionBidirectional {
			realProps := map[string]any{
				ad.IsACL.String():                   false,
				ad.SpoofSIDHistoryBlocked.String():  trust.SidFilteringEnabled,
				ad.TrustType.String():               trust.TrustType,
				ad.Transitive.String():              trust.IsTransitive,
				ad.TrustAttributesOutbound.String(): finalTrustAttributes,
			}

			parsedData.TrustRelationships = append(parsedData.TrustRelationships, NewIngestibleRelationship(
				IngestibleEndpoint{
					Value: domain.ObjectIdentifier,
					Kind:  ad.Domain,
				},
				IngestibleEndpoint{
					Value: trust.TargetDomainSid,
					Kind:  ad.Domain,
				},
				IngestibleRel{
					RelProps: realProps,
					RelType:  edgeType,
				},
			))

			if edgeType == ad.CrossForestTrust && !trust.SidFilteringEnabled {
				parsedData.TrustRelationships = append(parsedData.TrustRelationships, NewIngestibleRelationship(
					IngestibleEndpoint{
						Value: trust.TargetDomainSid,
						Kind:  ad.Domain,
					},
					IngestibleEndpoint{
						Value: domain.ObjectIdentifier,
						Kind:  ad.Domain,
					},
					IngestibleRel{
						RelProps: map[string]any{
							ad.IsACL.String():     false,
							ad.TrustType.String(): trust.TrustType,
						},
						RelType: ad.SpoofSIDHistory,
					},
				))
			}
		}
	}

	return parsedData
}

// ParseComputerMiscData parses AllowedToDelegate, AllowedToAct, HasSIDHistory, DumpSMSAPassword, DCFor, Sessions, and CoerceToTGT
func ParseComputerMiscData(computer Computer) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)
	for _, target := range computer.AllowedToDelegate {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: computer.ObjectIdentifier,
				Kind:  ad.Computer,
			},
			IngestibleEndpoint{
				Value: target.ObjectIdentifier,
				Kind:  target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.AllowedToDelegate,
			},
		))
	}

	for _, actor := range computer.AllowedToAct {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: actor.ObjectIdentifier,
				Kind:  actor.Kind(),
			},
			IngestibleEndpoint{
				Value: computer.ObjectIdentifier,
				Kind:  ad.Computer,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.AllowedToAct,
			},
		))
	}

	for _, target := range computer.DumpSMSAPassword {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: computer.ObjectIdentifier,
				Kind:  ad.Computer,
			},
			IngestibleEndpoint{
				Value: target.ObjectIdentifier,
				Kind:  target.Kind(),
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.DumpSMSAPassword,
			},
		))
	}

	for _, target := range computer.HasSIDHistory {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: computer.ObjectIdentifier,
				Kind:  ad.Computer,
			},
			IngestibleEndpoint{
				Value: target.ObjectIdentifier,
				Kind:  target.Kind(),
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
				IngestibleEndpoint{
					Value: session.ComputerSID,
					Kind:  ad.Computer,
				},
				IngestibleEndpoint{
					Value: session.UserSID,
					Kind:  ad.User,
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
				IngestibleEndpoint{
					Value: session.ComputerSID,
					Kind:  ad.Computer,
				},
				IngestibleEndpoint{
					Value: session.UserSID,
					Kind:  ad.User,
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
				IngestibleEndpoint{
					Value: session.ComputerSID,
					Kind:  ad.Computer,
				},
				IngestibleEndpoint{
					Value: session.UserSID,
					Kind:  ad.User,
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
			IngestibleEndpoint{
				Value: computer.ObjectIdentifier,
				Kind:  ad.Computer,
			},
			IngestibleEndpoint{
				Value: computer.DomainSID,
				Kind:  ad.Domain,
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
				IngestibleEndpoint{
					Value: computer.ObjectIdentifier,
					Kind:  ad.Computer,
				},
				IngestibleEndpoint{
					Value: domainsid,
					Kind:  ad.Domain,
				},
				IngestibleRel{
					RelProps: map[string]any{ad.IsACL.String(): false},
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
			Labels: []graph.Kind{ad.LocalGroup},
		})
	}

	parsedData.Relationships = append(parsedData.Relationships, NewIngestibleRelationship(
		IngestibleEndpoint{
			Value: localGroup.ObjectIdentifier,
			Kind:  ad.LocalGroup,
		},
		IngestibleEndpoint{
			Value: computer.ObjectIdentifier,
			Kind:  ad.Computer,
		},
		IngestibleRel{
			RelProps: map[string]any{ad.IsACL.String(): false},
			RelType:  ad.LocalToComputer,
		},
	))

	for _, member := range localGroup.Results {
		parsedData.Relationships = append(parsedData.Relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: member.ObjectIdentifier,
				Kind:  member.Kind(),
			},
			IngestibleEndpoint{
				Value: localGroup.ObjectIdentifier,
				Kind:  ad.LocalGroup,
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
			Labels: []graph.Kind{ad.Entity},
		})
	}

	return parsedData
}

func ParseUserRightData(userRight UserRightsAssignmentAPIResult, computer Computer, right graph.Kind) []IngestibleRelationship {
	relationships := make([]IngestibleRelationship, 0)

	for _, grant := range userRight.Results {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: grant.ObjectIdentifier,
				Kind:  grant.Kind(),
			},
			IngestibleEndpoint{
				Value: computer.ObjectIdentifier,
				Kind:  ad.Computer,
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
		Labels:      []graph.Kind{ad.EnterpriseCA},
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
			IngestibleEndpoint{
				Value: actor.ObjectIdentifier,
				Kind:  ad.CertTemplate,
			},
			IngestibleEndpoint{
				Value: enterpriseCA.ObjectIdentifier,
				Kind:  ad.EnterpriseCA,
			},
			IngestibleRel{
				RelProps: map[string]any{ad.IsACL.String(): false},
				RelType:  ad.PublishedTo,
			},
		))
	}

	if enterpriseCA.HostingComputer != "" {
		relationships = append(relationships, NewIngestibleRelationship(
			IngestibleEndpoint{
				Value: enterpriseCA.HostingComputer,
				Kind:  ad.Computer,
			},
			IngestibleEndpoint{
				Value: enterpriseCA.ObjectIdentifier,
				Kind:  ad.EnterpriseCA,
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
						IngestibleEndpoint{
							Value: restriction.Agent.ObjectIdentifier,
							Kind:  restriction.Agent.Kind(),
						},
						IngestibleEndpoint{
							Value: template,
							Kind:  ad.CertTemplate,
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
	baseNodeProp := ConvertObjectToNode(enterpriseCA.IngestBase, ad.EnterpriseCA, time.Now().UTC())

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
			IngestibleEndpoint{
				Value: rootCA.ObjectIdentifier,
				Kind:  ad.RootCA,
			},
			IngestibleEndpoint{
				Value: domainsid,
				Kind:  ad.Domain,
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
			IngestibleEndpoint{
				Value: ntAuthStore.ObjectIdentifier,
				Kind:  ad.NTAuthStore,
			},
			IngestibleEndpoint{
				Value: domainsid,
				Kind:  ad.Domain,
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

	propMap[ad.VulnerableNetlogonSecurityDescriptorCollected.String()] = computer.DCRegistryData.VulnerableNetlogonSecurityDescriptor.Collected
	if computer.DCRegistryData.VulnerableNetlogonSecurityDescriptor.Collected {
		propMap[ad.VulnerableNetlogonSecurityDescriptor.String()] = computer.DCRegistryData.VulnerableNetlogonSecurityDescriptor.Value
	}

	return IngestibleNode{
		ObjectID:    computer.ObjectIdentifier,
		PropertyMap: propMap,
		Labels:      []graph.Kind{ad.Computer},
	}
}

// Prettified definitions for GPOStatus
const (
	PrettyGPOStatusNotExisting                   = "GPO Status does not exist"
	PrettyGPOStatusEnabled                       = "Enabled"
	PrettyGPOStatusUserConfigurationDisabled     = "User Configuration Disabled"
	PrettyGPOStatusComputerConfigurationDisabled = "Computer Configuration Disabled"
	PrettyGPOStatusDisabled                      = "Disabled"
)

func ParseGPOData(gpo GPO) IngestibleNode {
	propMap := make(map[string]any)

	if status, ok := gpo.Properties[ad.GPOStatus.String()]; ok {
		propMap[ad.GPOStatusRaw.String()] = strings.TrimSpace(fmt.Sprint(status))

		switch propMap[ad.GPOStatusRaw.String()] {
		case "0":
			propMap[ad.GPOStatus.String()] = PrettyGPOStatusEnabled
		case "1":
			propMap[ad.GPOStatus.String()] = PrettyGPOStatusUserConfigurationDisabled
		case "2":
			propMap[ad.GPOStatus.String()] = PrettyGPOStatusComputerConfigurationDisabled
		case "3":
			propMap[ad.GPOStatus.String()] = PrettyGPOStatusDisabled
		default:
			propMap[ad.GPOStatus.String()] = PrettyGPOStatusNotExisting
		}
	}

	return IngestibleNode{
		ObjectID:    gpo.ObjectIdentifier,
		PropertyMap: propMap,
		Labels:      []graph.Kind{ad.GPO},
	}
}
