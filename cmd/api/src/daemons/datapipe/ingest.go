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

func (s *Daemon) ReadWrapper(batch graph.Batch, reader io.ReadSeeker) error {
	if meta, err := validateMetaTag(reader); err != nil {
		return err
	} else {

	}
	var wrapper DataWrapper

	if err := json.NewDecoder(reader).Decode(&wrapper); err != nil {
		return err
	}

	return s.IngestWrapper(batch, wrapper)
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

func (s *Daemon) IngestBasicData(batch graph.Batch, converted ConvertedData) {
	IngestNodes(batch, ad.Entity, converted.NodeProps)
	IngestRelationships(batch, ad.Entity, converted.RelProps)
}

func (s *Daemon) IngestGroupData(batch graph.Batch, converted ConvertedGroupData) {
	IngestNodes(batch, ad.Entity, converted.NodeProps)
	IngestRelationships(batch, ad.Entity, converted.RelProps)
	IngestDNRelationships(batch, converted.DistinguishedNameProps)
}

func (s *Daemon) IngestAzureData(batch graph.Batch, converted ConvertedAzureData) {
	IngestNodes(batch, azure.Entity, converted.NodeProps)
	IngestNodes(batch, ad.Entity, converted.OnPremNodes)
	IngestRelationships(batch, azure.Entity, converted.RelProps)
}

func (s *Daemon) IngestWrapperNew(batch graph.Batch, reader io.ReadSeeker, meta Metadata) error {
	switch meta.Type {
	case DataTypeComputer:
		if meta.Version > 5 {
			return s.decodeComputerData(batch, reader)
		}
	case DataTypeUser:
		return s.decodeUserData(batch, reader)
	case DataTypeGroup:

	}
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

func (s *Daemon) decodeComputerData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := createIngestDecoder(reader)
	if err != nil {
		return err
	}

	convertedData := ConvertedData{}
	var computer ein.Computer
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&computer); err != nil {
			log.Errorf("Error decoding computer object: %v", err)
		} else {
			count++
			convertComputerData(computer, &convertedData)
			if count == ingestCountThreshold {
				s.IngestBasicData(batch, convertedData)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		s.IngestBasicData(batch, convertedData)
	}

	return nil
}

func (s *Daemon) decodeUserData(batch graph.Batch, reader io.ReadSeeker) error {
	decoder, err := createIngestDecoder(reader)
	if err != nil {
		return err
	}

	convertedData := ConvertedData{}
	var user ein.User
	count := 0
	for decoder.More() {
		if err := decoder.Decode(&user); err != nil {
			log.Errorf("Error decoding user object: %v", err)
		} else {
			count++
			convertUserData(user, &convertedData)
			if count == ingestCountThreshold {
				s.IngestBasicData(batch, convertedData)
				convertedData.Clear()
				count = 0
			}
		}
	}

	if count > 0 {
		s.IngestBasicData(batch, convertedData)
	}

	return nil
}

func (s *Daemon) IngestWrapper(batch graph.Batch, wrapper DataWrapper) error {
	switch wrapper.Metadata.Type {
	case DataTypeComputer:
		// We should not be getting anything with Version < 5 at this point, and we don't want to ingest it if we do as post-processing will blow it away anyways
		if wrapper.Metadata.Version >= 5 {
			var computerData []ein.Computer

			if err := json.Unmarshal(wrapper.Payload, &computerData); err != nil {
				return err
			} else {
				converted := convertComputerData(computerData)
				s.IngestBasicData(batch, converted)
			}
		}

	case DataTypeUser:
		var userData []ein.User
		if err := json.Unmarshal(wrapper.Payload, &userData); err != nil {
			return err
		} else {
			converted := convertUserData(userData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeGroup:
		var groupData []ein.Group
		if err := json.Unmarshal(wrapper.Payload, &groupData); err != nil {
			return err
		} else {
			converted := convertGroupData(groupData)
			s.IngestGroupData(batch, converted)
		}

	case DataTypeDomain:
		var domainData []ein.Domain
		if err := json.Unmarshal(wrapper.Payload, &domainData); err != nil {
			return err
		} else {
			converted := convertDomainData(domainData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeGPO:
		var gpoData []ein.GPO
		if err := json.Unmarshal(wrapper.Payload, &gpoData); err != nil {
			return err
		} else {
			converted := convertGPOData(gpoData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeOU:
		var ouData []ein.OU
		if err := json.Unmarshal(wrapper.Payload, &ouData); err != nil {
			return err
		} else {
			converted := convertOUData(ouData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeSession:
		var sessionData []ein.Session
		if err := json.Unmarshal(wrapper.Payload, &sessionData); err != nil {
			return err
		} else {
			IngestSessions(batch, convertSessionData(sessionData).SessionProps)
		}

	case DataTypeContainer:
		var containerData []ein.Container
		if err := json.Unmarshal(wrapper.Payload, &containerData); err != nil {
			return err
		} else {
			converted := convertContainerData(containerData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeAIACA:
		var aiacaData []ein.AIACA
		if err := json.Unmarshal(wrapper.Payload, &aiacaData); err != nil {
			return err
		} else {
			converted := convertAIACAData(aiacaData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeRootCA:
		var rootcaData []ein.RootCA
		if err := json.Unmarshal(wrapper.Payload, &rootcaData); err != nil {
			return err
		} else {
			converted := convertRootCAData(rootcaData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeEnterpriseCA:
		var enterprisecaData []ein.EnterpriseCA
		if err := json.Unmarshal(wrapper.Payload, &enterprisecaData); err != nil {
			return err
		} else {
			converted := convertEnterpriseCAData(enterprisecaData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeNTAuthStore:
		var ntauthstoreData []ein.NTAuthStore
		if err := json.Unmarshal(wrapper.Payload, &ntauthstoreData); err != nil {
			return err
		} else {
			converted := convertNTAuthStoreData(ntauthstoreData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeCertTemplate:
		var certtemplateData []ein.CertTemplate
		if err := json.Unmarshal(wrapper.Payload, &certtemplateData); err != nil {
			return err
		} else {
			converted := convertCertTemplateData(certtemplateData)
			s.IngestBasicData(batch, converted)
		}

	case DataTypeAzure:
		var azureData []json.RawMessage
		if err := json.Unmarshal(wrapper.Payload, &azureData); err != nil {
			return err
		} else {
			converted := convertAzureData(azureData)
			s.IngestAzureData(batch, converted)
		}
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
