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
	"encoding/json"
	"fmt"
	"github.com/specterops/bloodhound/errors"
	"io"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/log"
)

const (
	delimOpenBracket        = json.Delim('{')
	delimCloseBracket       = json.Delim('}')
	delimOpenSquareBracket  = json.Delim('[')
	delimCloseSquareBracket = json.Delim(']')
	ingestCountThreshold    = 500
)

var (
	ErrMetaNotFound   = errors.New("no valid meta tag found")
	ErrDataNotFound   = errors.New("no data tag found")
	ErrNoTagFound     = errors.New("no valid meta tag or data tag found")
	ErrInvalidDataTag = errors.New("invalid data tag found")
)

func ReadFileForIngest(batch graph.Batch, reader io.ReadSeeker) error {
	if meta, err := validateMetaTag(reader); err != nil {
		return fmt.Errorf("error validating meta tag: %w", err)
	} else {
		return IngestWrapper(batch, reader, meta)
	}
}

func validateMetaTag(reader io.ReadSeeker) (Metadata, error) {
	_, err := reader.Seek(0, io.SeekStart)
	if err != nil {
		return Metadata{}, fmt.Errorf("error seeking to start of file: %w", err)
	}
	depth := 0
	decoder := json.NewDecoder(reader)
	dataTagFound := false
	metaTagFound := false
	var meta Metadata
	for {
		if dataTagFound && metaTagFound {
			return meta, nil
		}
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				if !metaTagFound && !dataTagFound {
					return Metadata{}, ErrNoTagFound
				} else if !metaTagFound {
					return Metadata{}, ErrDataNotFound
				} else {
					return Metadata{}, ErrMetaNotFound
				}
			}
			return Metadata{}, err
		} else {
			switch typed := token.(type) {
			case json.Delim:
				switch typed {
				case delimCloseBracket, delimCloseSquareBracket:
					depth--
				case delimOpenBracket, delimOpenSquareBracket:
					depth++
				}
			case string:
				if !metaTagFound && depth == 1 && typed == "meta" {
					if err := decoder.Decode(&meta); err != nil {
						return Metadata{}, err
					} else if meta.IsValid() {
						metaTagFound = true
					}
				}

				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}
			}
		}
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
		if meta.Version > 5 {
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

func seekToDataTag(decoder *json.Decoder) error {
	depth := 0
	dataTagFound := false
	for {
		if token, err := decoder.Token(); err != nil {
			if errors.Is(err, io.EOF) {
				return ErrDataNotFound
			}

			return err
		} else {
			//Break here to allow for one more token read, which should take us to the "[" token, exactly where we need to be
			if dataTagFound {
				//Do some extra checks
				if typed, ok := token.(json.Delim); !ok {
					return ErrInvalidDataTag
				} else if typed != delimOpenSquareBracket {
					return ErrInvalidDataTag
				}
				//Break out of our loop if we're in a good spot
				return nil
			}
			switch typed := token.(type) {
			case json.Delim:
				switch typed {
				case delimCloseBracket, delimCloseSquareBracket:
					depth--
				case delimOpenBracket, delimOpenSquareBracket:
					depth++
				}
			case string:
				if !dataTagFound && depth == 1 && typed == "data" {
					dataTagFound = true
				}
			}
		}
	}
}

func createIngestDecoder(reader io.ReadSeeker) (*json.Decoder, error) {
	if _, err := reader.Seek(0, io.SeekStart); err != nil {
		return nil, fmt.Errorf("error seeking to start of file: %w", err)
	} else {
		decoder := json.NewDecoder(reader)
		if err := seekToDataTag(decoder); err != nil {
			return nil, fmt.Errorf("error seeking to data tag: %w", err)
		} else {
			return decoder, nil
		}
	}
}

type ConversionFunc[T any] func(decoded T, converted *ConvertedData)

func decodeBasicData[T any](batch graph.Batch, reader io.ReadSeeker, conversionFunc ConversionFunc[T]) error {
	decoder, err := createIngestDecoder(reader)
	if err != nil {
		return err
	}

	var (
		count         = 0
		batchCount    = 0
		convertedData ConvertedData
	)

	for decoder.More() {
		var decodeTarget T
		if err := decoder.Decode(&decodeTarget); err != nil {
			log.Errorf("Error decoding %T object: %v", decodeTarget, err)
		} else {
			count++
			conversionFunc(decodeTarget, &convertedData)
		}

		if count == ingestCountThreshold {
			batchCount++
			log.Infof("Sending batch %d of %d objects", batchCount, ingestCountThreshold)
			IngestBasicData(batch, convertedData)
			convertedData.Clear()
			count = 0
		}
	}

	if count > 0 {
		IngestBasicData(batch, convertedData)
	}

	return nil
}

func decodeGroupData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := createIngestDecoder(reader)
	if err != nil {
		return err
	}

	convertedData := ConvertedGroupData{}
	var group ein.Group
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&group); err != nil {
			log.Errorf("Error decoding group object: %v", err)
		} else {
			count++
			convertGroupData(group, &convertedData)
			if count == ingestCountThreshold {
				IngestGroupData(batch, convertedData)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		IngestGroupData(batch, convertedData)
	}

	return nil
}

func decodeSessionData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := createIngestDecoder(reader)
	if err != nil {
		return err
	}

	convertedData := ConvertedSessionData{}
	var session ein.Session
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&session); err != nil {
			log.Errorf("Error decoding session object: %v", err)
		} else {
			count++
			convertSessionData(session, &convertedData)
			if count == ingestCountThreshold {
				IngestSessions(batch, convertedData.SessionProps)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		IngestSessions(batch, convertedData.SessionProps)
	}

	return nil
}

func decodeAzureData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := createIngestDecoder(reader)
	if err != nil {
		return err
	}

	convertedData := ConvertedAzureData{}
	var data AzureBase
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&data); err != nil {
			log.Errorf("Error decoding azure object: %v", err)
		} else {
			convert := getKindConverter(data.Kind)
			convert(data.Data, &convertedData)
			count++
			if count == ingestCountThreshold {
				IngestAzureData(batch, convertedData)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		IngestAzureData(batch, convertedData)
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
