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
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

const (
	IngestCountThreshold = 500
)

var (
	ErrMetaTagNotFound     = errors.New("no valid meta tag found")
	ErrDataTagNotFound     = errors.New("no data tag found")
	ErrNoTagFound          = errors.New("no valid meta tag or data tag found")
	ErrInvalidDataTag      = errors.New("invalid data tag found")
	ErrJSONDecoderInternal = errors.New("json decoder internal error")
)

func ReadFileForIngest(batch graph.Batch, reader io.ReadSeeker) error {
	if meta, err := ValidateMetaTag(reader); err != nil {
		return fmt.Errorf("error validating meta tag: %w", err)
	} else {
		return IngestWrapper(batch, reader, meta)
	}
}

func IngestBasicData(batch graph.Batch, converted ConvertedData) {
	IngestNodes(batch, ad.Entity, converted.NodeProps)
	IngestRelationships(batch, ad.Entity, converted.RelProps)
}

func IngestGroupData(batch graph.Batch, converted ConvertedGroupData) {
	IngestNodes(batch, ad.Entity, converted.NodeProps)
	IngestRelationships(batch, ad.Entity, converted.RelProps)
	IngestDNRelationships(batch, converted.DistinguishedNameProps)
}

func IngestAzureData(batch graph.Batch, converted ConvertedAzureData) {
	IngestNodes(batch, azure.Entity, converted.NodeProps)
	IngestNodes(batch, ad.Entity, converted.OnPremNodes)
	IngestRelationships(batch, azure.Entity, converted.RelProps)
}

func IngestWrapper(batch graph.Batch, reader io.ReadSeeker, meta Metadata) error {
	switch meta.Type {
	case DataTypeComputer:
		if meta.Version >= 5 {
			return decodeBasicData(batch, reader, convertComputerData)
		}
	case DataTypeUser:
		return decodeBasicData(batch, reader, convertUserData)
	case DataTypeGroup:
		return decodeGroupData(batch, reader)
	case DataTypeDomain:
		return decodeBasicData(batch, reader, convertDomainData)
	case DataTypeGPO:
		return decodeBasicData(batch, reader, convertGPOData)
	case DataTypeOU:
		return decodeBasicData(batch, reader, convertOUData)
	case DataTypeSession:
		return decodeSessionData(batch, reader)
	case DataTypeContainer:
		return decodeBasicData(batch, reader, convertContainerData)
	case DataTypeAIACA:
		return decodeBasicData(batch, reader, convertAIACAData)
	case DataTypeRootCA:
		return decodeBasicData(batch, reader, convertRootCAData)
	case DataTypeEnterpriseCA:
		return decodeBasicData(batch, reader, convertEnterpriseCAData)
	case DataTypeNTAuthStore:
		return decodeBasicData(batch, reader, convertNTAuthStoreData)
	case DataTypeCertTemplate:
		return decodeBasicData(batch, reader, convertCertTemplateData)
	case DataTypeAzure:
		return decodeAzureData(batch, reader)
	}

	return nil
}

func IngestNode(batch graph.Batch, nowUTC time.Time, identityKind graph.Kind, nextNode ein.IngestibleNode) error {
	//Ensure object id is upper case
	nextNode.ObjectID = strings.ToUpper(nextNode.ObjectID)

	nextNode.PropertyMap[common.LastSeen.String()] = nowUTC
	nextNode.PropertyMap[common.ObjectID.String()] = nextNode.ObjectID

	//Ensure that name, operatingsystem, and distinguishedname properties are upper case
	if rawName, hasName := nextNode.PropertyMap[common.Name.String()]; hasName && rawName != nil {
		if name, typeMatches := rawName.(string); typeMatches {
			nextNode.PropertyMap[common.Name.String()] = strings.ToUpper(name)
		} else {
			log.Errorf("Bad type found for node name property during ingest. Expected string, got %T", rawName)
		}
	}

	if rawOS, hasOS := nextNode.PropertyMap[common.OperatingSystem.String()]; hasOS && rawOS != nil {
		if os, typeMatches := rawOS.(string); typeMatches {
			nextNode.PropertyMap[common.OperatingSystem.String()] = strings.ToUpper(os)
		} else {
			log.Errorf("Bad type found for node operating system property during ingest. Expected string, got %T", rawOS)
		}
	}

	if rawDN, hasDN := nextNode.PropertyMap[ad.DistinguishedName.String()]; hasDN && rawDN != nil {
		if dn, typeMatches := rawDN.(string); typeMatches {
			nextNode.PropertyMap[ad.DistinguishedName.String()] = strings.ToUpper(dn)
		} else {
			log.Errorf("Bad type found for node distinguished name property during ingest. Expected string, got %T", rawDN)
		}
	}

	return batch.UpdateNodeBy(graph.NodeUpdate{
		Node:         graph.PrepareNode(graph.AsProperties(nextNode.PropertyMap), nextNode.Label),
		IdentityKind: identityKind,
		IdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestNodes(batch graph.Batch, identityKind graph.Kind, nodes []ein.IngestibleNode) {
	nowUTC := time.Now().UTC()

	for _, next := range nodes {
		if err := IngestNode(batch, nowUTC, identityKind, next); err != nil {
			log.Errorf("Error ingesting node: %v", err)
		}
	}
}

func IngestRelationship(batch graph.Batch, nowUTC time.Time, nodeIDKind graph.Kind, nextRel ein.IngestibleRelationship) error {
	nextRel.RelProps[common.LastSeen.String()] = nowUTC
	nextRel.Source = strings.ToUpper(nextRel.Source)
	nextRel.Target = strings.ToUpper(nextRel.Target)

	return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(nextRel.RelProps), nextRel.RelType),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Source,
			common.LastSeen: nowUTC,
		}), nextRel.SourceType),
		StartIdentityKind: nodeIDKind,
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Target,
			common.LastSeen: nowUTC,
		}), nextRel.TargetType),
		EndIdentityKind: nodeIDKind,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestRelationships(batch graph.Batch, nodeIDKind graph.Kind, relationships []ein.IngestibleRelationship) {
	nowUTC := time.Now().UTC()

	for _, next := range relationships {
		if err := IngestRelationship(batch, nowUTC, nodeIDKind, next); err != nil {
			log.Errorf("Error ingesting relationship from basic data : %v ", err)
		}
	}
}

func ingestDNRelationship(batch graph.Batch, nowUTC time.Time, nextRel ein.IngestibleRelationship) error {
	nextRel.RelProps[common.LastSeen.String()] = nowUTC
	nextRel.Source = strings.ToUpper(nextRel.Source)
	nextRel.Target = strings.ToUpper(nextRel.Target)

	return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(nextRel.RelProps), nextRel.RelType),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			ad.DistinguishedName: nextRel.Source,
			common.LastSeen:      nowUTC,
		}), nextRel.SourceType),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			ad.DistinguishedName.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Target,
			common.LastSeen: nowUTC,
		}), nextRel.TargetType),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestDNRelationships(batch graph.Batch, relationships []ein.IngestibleRelationship) {
	nowUTC := time.Now().UTC()

	for _, next := range relationships {
		if err := ingestDNRelationship(batch, nowUTC, next); err != nil {
			log.Errorf("Error ingesting relationship: %v", err)
		}
	}
}

func ingestSession(batch graph.Batch, nowUTC time.Time, nextSession ein.IngestibleSession) error {
	nextSession.Target = strings.ToUpper(nextSession.Target)
	nextSession.Source = strings.ToUpper(nextSession.Source)

	return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(graph.PropertyMap{
			common.LastSeen: nowUTC,
			ad.LogonType:    nextSession.LogonType,
		}), ad.HasSession),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Source,
			common.LastSeen: nowUTC,
		}), ad.Computer),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Target,
			common.LastSeen: nowUTC,
		}), ad.User),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestSessions(batch graph.Batch, sessions []ein.IngestibleSession) {
	nowUTC := time.Now().UTC()

	for _, next := range sessions {
		if err := ingestSession(batch, nowUTC, next); err != nil {
			log.Errorf("Error ingesting sessions: %v", err)
		}
	}
}
