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
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model/ingest"
	ingest_service "github.com/specterops/bloodhound/src/services/ingest"
)

const (
	IngestCountThreshold = 500
	ReconcileProperty    = "reconcile"
)

func ReadFileForIngest(batch graph.Batch, reader io.ReadSeeker, ingestSchema ingest_service.IngestSchema, adcsEnabled bool) error {
	if meta, err := ingest_service.ValidateMetaTag(reader, ingestSchema, false); err != nil {
		return fmt.Errorf("error validating meta tag: %w", err)
	} else {
		return IngestWrapper(batch, reader, meta, adcsEnabled)
	}
}

func IngestBasicData(batch graph.Batch, converted ConvertedData) error {
	errs := util.NewErrorCollector()

	if err := IngestNodes(batch, ad.Entity, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, ad.Entity, converted.RelProps); err != nil {
		errs.Add(err)
	}

	return errs.Combined()
}

func IngestGroupData(batch graph.Batch, converted ConvertedGroupData) error {
	errs := util.NewErrorCollector()

	if err := IngestNodes(batch, ad.Entity, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, ad.Entity, converted.RelProps); err != nil {
		errs.Add(err)
	}

	if err := IngestDNRelationships(batch, converted.DistinguishedNameProps); err != nil {
		errs.Add(err)
	}

	return errs.Combined()
}

func IngestAzureData(batch graph.Batch, converted ConvertedAzureData) error {
	errs := util.NewErrorCollector()

	if err := IngestNodes(batch, azure.Entity, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestNodes(batch, ad.Entity, converted.OnPremNodes); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, azure.Entity, converted.RelProps); err != nil {
		errs.Add(err)
	}

	return errs.Combined()
}

func IngestWrapper(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata, adcsEnabled bool) error {
	switch meta.Type {
	case ingest.DataTypeComputer:
		if meta.Version >= 5 {
			return decodeBasicData(batch, reader, convertComputerData)
		}
	case ingest.DataTypeUser:
		return decodeBasicData(batch, reader, convertUserData)
	case ingest.DataTypeGroup:
		return decodeGroupData(batch, reader)
	case ingest.DataTypeDomain:
		return decodeBasicData(batch, reader, convertDomainData)
	case ingest.DataTypeGPO:
		return decodeBasicData(batch, reader, convertGPOData)
	case ingest.DataTypeOU:
		return decodeBasicData(batch, reader, convertOUData)
	case ingest.DataTypeSession:
		return decodeSessionData(batch, reader)
	case ingest.DataTypeContainer:
		return decodeBasicData(batch, reader, convertContainerData)
	case ingest.DataTypeAIACA:
		return decodeBasicData(batch, reader, convertAIACAData)
	case ingest.DataTypeRootCA:
		return decodeBasicData(batch, reader, convertRootCAData)
	case ingest.DataTypeEnterpriseCA:
		return decodeBasicData(batch, reader, convertEnterpriseCAData)
	case ingest.DataTypeNTAuthStore:
		return decodeBasicData(batch, reader, convertNTAuthStoreData)
	case ingest.DataTypeCertTemplate:
		return decodeBasicData(batch, reader, convertCertTemplateData)
	case ingest.DataTypeAzure:
		return decodeAzureData(batch, reader)
	case ingest.DataTypeIssuancePolicy:
		return decodeBasicData(batch, reader, convertIssuancePolicy)
	}

	return nil
}

func NormalizeEinNodeProperties(properties map[string]any, objectID string, nowUTC time.Time) map[string]any {
	delete(properties, ReconcileProperty)
	properties[common.LastSeen.String()] = nowUTC
	properties[common.ObjectID.String()] = strings.ToUpper(objectID)

	// Ensure that name, operatingsystem, and distinguishedname properties are upper case
	if rawName, hasName := properties[common.Name.String()]; hasName && rawName != nil {
		if name, typeMatches := rawName.(string); typeMatches {
			properties[common.Name.String()] = strings.ToUpper(name)
		} else {
			slog.Error(fmt.Sprintf("Bad type found for node name property during ingest. Expected string, got %T", rawName))
		}
	}

	if rawOS, hasOS := properties[common.OperatingSystem.String()]; hasOS && rawOS != nil {
		if os, typeMatches := rawOS.(string); typeMatches {
			properties[common.OperatingSystem.String()] = strings.ToUpper(os)
		} else {
			slog.Error(fmt.Sprintf("Bad type found for node operating system property during ingest. Expected string, got %T", rawOS))
		}
	}

	if rawDN, hasDN := properties[ad.DistinguishedName.String()]; hasDN && rawDN != nil {
		if dn, typeMatches := rawDN.(string); typeMatches {
			properties[ad.DistinguishedName.String()] = strings.ToUpper(dn)
		} else {
			slog.Error(fmt.Sprintf("Bad type found for node distinguished name property during ingest. Expected string, got %T", rawDN))
		}
	}

	return properties
}

func IngestNode(batch graph.Batch, nowUTC time.Time, identityKind graph.Kind, nextNode ein.IngestibleNode) error {
	normalizedProperties := NormalizeEinNodeProperties(nextNode.PropertyMap, nextNode.ObjectID, nowUTC)

	return batch.UpdateNodeBy(graph.NodeUpdate{
		Node:         graph.PrepareNode(graph.AsProperties(normalizedProperties), nextNode.Label),
		IdentityKind: identityKind,
		IdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestNodes(batch graph.Batch, identityKind graph.Kind, nodes []ein.IngestibleNode) error {
	var (
		nowUTC = time.Now().UTC()
		errs   = util.NewErrorCollector()
	)

	for _, next := range nodes {
		if err := IngestNode(batch, nowUTC, identityKind, next); err != nil {
			slog.Error(fmt.Sprintf("Error ingesting node ID %s: %v", next.ObjectID, err))
			errs.Add(err)
		}
	}
	return errs.Combined()
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

func IngestRelationships(batch graph.Batch, nodeIDKind graph.Kind, relationships []ein.IngestibleRelationship) error {
	var (
		nowUTC = time.Now().UTC()
		errs   = util.NewErrorCollector()
	)

	for _, next := range relationships {
		if err := IngestRelationship(batch, nowUTC, nodeIDKind, next); err != nil {
			slog.Error(fmt.Sprintf("Error ingesting relationship from %s to %s : %v", next.Source, next.Target, err))
			errs.Add(err)
		}
	}
	return errs.Combined()
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

func IngestDNRelationships(batch graph.Batch, relationships []ein.IngestibleRelationship) error {
	var (
		nowUTC = time.Now().UTC()
		errs   = util.NewErrorCollector()
	)

	for _, next := range relationships {
		if err := ingestDNRelationship(batch, nowUTC, next); err != nil {
			slog.Error(fmt.Sprintf("Error ingesting relationship: %v", err))
			errs.Add(err)
		}
	}
	return errs.Combined()
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

func IngestSessions(batch graph.Batch, sessions []ein.IngestibleSession) error {
	var (
		nowUTC = time.Now().UTC()
		errs   = util.NewErrorCollector()
	)

	for _, next := range sessions {
		if err := ingestSession(batch, nowUTC, next); err != nil {
			slog.Error(fmt.Sprintf("Error ingesting sessions: %v", err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}
