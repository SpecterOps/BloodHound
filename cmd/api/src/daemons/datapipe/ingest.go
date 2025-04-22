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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
	ingest_service "github.com/specterops/bloodhound/src/services/ingest"
)

const (
	IngestCountThreshold = 500
	ReconcileProperty    = "reconcile"
)

type ReadOptions struct {
	FileType     model.FileType // JSON or ZIP
	IngestSchema ingest_service.IngestSchema
	ADCSEnabled  bool
}

func ReadFileForIngest(batch graph.Batch, reader io.ReadSeeker, options ReadOptions) error {

	var (
		shouldValidateGeneric = false
		readToEnd             = false
	)

	// if filetype == ZIP, we need to validate against jsonschema because the archive bypassed validation controls at fileupload time
	if options.FileType == model.FileTypeZip {
		shouldValidateGeneric = true
		readToEnd = true
	}

	// todo: plumb shouldValidateGeneric for zips
	fmt.Println(">>> todo: ", shouldValidateGeneric)

	if meta, err := ingest_service.ValidateMetaTag(reader, options.IngestSchema, readToEnd); err != nil {
		return err
	} else {
		return IngestWrapper(batch, reader, meta, options.ADCSEnabled)
	}
}

func IngestBasicData(batch graph.Batch, identityKind graph.Kind, converted ConvertedData) error {
	errs := util.NewErrorCollector()

	if err := IngestNodes(batch, identityKind, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, identityKind, converted.RelProps); err != nil {
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
	if handler, ok := ingestHandlers[meta.Type]; !ok {
		return fmt.Errorf("no handler for ingest data type: %v", meta.Type)
	} else {
		return handler(batch, reader, meta)
	}
}

type ingestHandler func(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error

func defaultBasicHandler[T any](conversionFunc ConversionFunc[T]) ingestHandler {
	return func(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error {
		decoder, err := getDefaultDecoder(reader)
		if err != nil {
			return err
		}
		return decodeBasicData(batch, decoder, conversionFunc, ad.Entity)
	}
}

var ingestHandlers = map[ingest.DataType]ingestHandler{
	ingest.DataTypeGeneric: func(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error {
		decoder, err := CreateIngestDecoder(reader, "nodes", 2)
		if errors.Is(err, ingest.ErrDataTagNotFound) {
			slog.Debug("no nodes found in generic ingest payload; continuing to edges")
		} else if err != nil {
			return err
		} else {
			if err = decodeBasicData(batch, decoder, convertGenericNode, graph.EmptyKind); err != nil {
				return err
			}
		}

		decoder, err = CreateIngestDecoder(reader, "edges", 2)
		if errors.Is(err, ingest.ErrDataTagNotFound) {
			slog.Debug("no edges found in generic ingest payload")
		} else if err != nil {
			return err
		} else {
			return decodeBasicData(batch, decoder, convertGenericEdge, graph.EmptyKind)
		}

		return nil
	},
	ingest.DataTypeComputer: func(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else if meta.Version >= 5 {
			return decodeBasicData(batch, decoder, convertComputerData, ad.Entity)
		} else {
			return nil
		}
	},
	ingest.DataTypeGroup: func(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else {
			return decodeGroupData(batch, decoder)
		}
	},
	ingest.DataTypeSession: func(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else {
			return decodeSessionData(batch, decoder)
		}
	},
	ingest.DataTypeAzure: func(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else {
			return decodeAzureData(batch, decoder)
		}
	},
	ingest.DataTypeUser:           defaultBasicHandler(convertUserData),
	ingest.DataTypeDomain:         defaultBasicHandler(convertDomainData),
	ingest.DataTypeGPO:            defaultBasicHandler(convertGPOData),
	ingest.DataTypeOU:             defaultBasicHandler(convertOUData),
	ingest.DataTypeContainer:      defaultBasicHandler(convertContainerData),
	ingest.DataTypeAIACA:          defaultBasicHandler(convertAIACAData),
	ingest.DataTypeRootCA:         defaultBasicHandler(convertRootCAData),
	ingest.DataTypeEnterpriseCA:   defaultBasicHandler(convertEnterpriseCAData),
	ingest.DataTypeNTAuthStore:    defaultBasicHandler(convertNTAuthStoreData),
	ingest.DataTypeCertTemplate:   defaultBasicHandler(convertCertTemplateData),
	ingest.DataTypeIssuancePolicy: defaultBasicHandler(convertIssuancePolicy),
}

func getDefaultDecoder(reader io.ReadSeeker) (*json.Decoder, error) {
	return CreateIngestDecoder(reader, "data", 1)
}

func NormalizeEinNodeProperties(properties map[string]any, objectID string, nowUTC time.Time) map[string]any {
	if properties == nil {
		properties = make(map[string]any)
	}
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
	var (
		normalizedProperties = NormalizeEinNodeProperties(nextNode.PropertyMap, nextNode.ObjectID, nowUTC)
		nodeUpdate           = graph.NodeUpdate{
			Node: graph.PrepareNode(graph.AsProperties(normalizedProperties), nextNode.Labels...),
			IdentityProperties: []string{
				common.ObjectID.String(),
			},
		}
	)

	if identityKind != graph.EmptyKind {
		nodeUpdate.IdentityKind = identityKind
	}

	return batch.UpdateNodeBy(nodeUpdate)
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
	nextRel.Source.Value = strings.ToUpper(nextRel.Source.Value)
	nextRel.Target.Value = strings.ToUpper(nextRel.Target.Value)

	relationshipUpdate := graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(nextRel.RelProps), nextRel.RelType),
		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Source,
			common.LastSeen: nowUTC,
		}), nextRel.Source.Kind),
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},
		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Target,
			common.LastSeen: nowUTC,
		}), nextRel.Target.Kind),
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	}

	if nodeIDKind != graph.EmptyKind {
		relationshipUpdate.StartIdentityKind = nodeIDKind
		relationshipUpdate.EndIdentityKind = nodeIDKind
	}

	return batch.UpdateRelationshipBy(relationshipUpdate)
}

func IngestRelationships(batch graph.Batch, nodeIDKind graph.Kind, relationships []ein.IngestibleRelationship) error {
	var (
		nowUTC = time.Now().UTC()
		errs   = util.NewErrorCollector()
	)

	for _, next := range relationships {
		if next.Source.MatchBy == ein.MatchByID && next.Target.MatchBy == ein.MatchByID { // if no property to match against, do the usual objectid thang
			if err := IngestRelationship(batch, nowUTC, nodeIDKind, next); err != nil {
				slog.Error(fmt.Sprintf("Error ingesting relationship from %s to %s : %v", next.Source.Value, next.Target.Value, err))
				errs.Add(err)
			}
		} else {
			// todo: need to update signature
			if err := submitUpdate(batch, next); err != nil {
				errs.Add(err)
			}
		}

	}
	return errs.Combined()
}

// this func attempts to resolve objectids for a rel's source and target nodes.
// name (and optional kind filter) --> objectid --> submit to batch processor for standard processing flow
// TODO: how does this function cleanup if the feature is limited to just name matching, and not any arbitrary property?
// todo: what if only one of source/target requires resolution
// todo: change rel -> rels and refactor this to do resolution for many rels in one dawg query
func ResolveRelationshipByName(batch graph.Batch, rel ein.IngestibleRelationship) (graph.RelationshipUpdate, error) {
	var (
		nowUTC              = time.Now().UTC()
		matches             = map[string]string{}
		ambiguousResolution = false // if multiple nodes matched source/target
		filter              = func() graph.Criteria {
			var sourceCriteria, targetCriteria []graph.Criteria

			// always append name filter
			sourceCriteria = append(sourceCriteria, query.Equals(query.NodeProperty(common.Name.String()), rel.Source.Value))
			targetCriteria = append(targetCriteria, query.Equals(query.NodeProperty(common.Name.String()), rel.Target.Value))

			// optionally append kind filter
			if rel.Source.Kind != nil {
				sourceCriteria = append(sourceCriteria, query.Kind(query.Node(), rel.Source.Kind))
			}

			if rel.Target.Kind != nil {
				targetCriteria = append(targetCriteria, query.Kind(query.Node(), rel.Target.Kind))
			}

			// form the full filter
			return query.Or(
				query.And(sourceCriteria...),
				query.And(targetCriteria...),
			)
		}
		result graph.RelationshipUpdate
	)

	if rel.Source.Value == "" || rel.Target.Value == "" {
		return result, nil
	}

	if err := batch.Nodes().Filterf(filter).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
		for node := range cursor.Chan() {
			props := node.Properties

			if val, _ := props.Get(common.Name.String()).String(); strings.EqualFold(val, rel.Source.Value) {
				if oid, err := props.Get(string(common.ObjectID)).String(); err != nil {
					slog.Warn(fmt.Sprintf("matched source node missing objectid for %s", rel.Source.Value))
				} else {
					if _, hasExisting := matches["source"]; hasExisting {
						slog.Warn("ambiguous name match on source node. multiple results match.",
							slog.String("value", rel.Source.Value))
						ambiguousResolution = true
						return nil
					} else {
						matches["source"] = oid
					}
				}
			}

			if val, _ := props.Get(common.Name.String()).String(); strings.EqualFold(val, rel.Target.Value) {
				if oid, err := props.Get(string(common.ObjectID)).String(); err != nil {
					slog.Warn(fmt.Sprintf("matched target node missing objectid for %s", rel.Target.Value))
				} else {
					if _, hasExisting := matches["target"]; hasExisting {
						slog.Warn("ambiguous property match on target node. multiple results match.",
							slog.String("value", rel.Target.Value))
						ambiguousResolution = true
						return nil
					} else {
						matches["target"] = oid
					}
				}
			}
		}
		return nil
	}); err != nil {
		return result, err
	}

	if ambiguousResolution {
		return result, nil
	}

	srcID, srcOk := matches["source"]
	targetID, targetOk := matches["target"]

	if !srcOk || !targetOk {
		slog.Warn("failed to resolve both nodes by name",
			slog.String("source", rel.Source.Value),
			slog.String("target", rel.Target.Value),
			slog.Bool("resolved_source", srcOk),
			slog.Bool("resolved_target", targetOk))
		return result, nil
	}

	start := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
		common.ObjectID: srcID,
		common.LastSeen: nowUTC,
	}), rel.Source.Kind)

	end := graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
		common.ObjectID: targetID,
		common.LastSeen: nowUTC,
	}), rel.Target.Kind)

	result = graph.RelationshipUpdate{
		Start: start,
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},
		End: end,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
		Relationship: graph.PrepareRelationship(graph.AsProperties(rel.RelProps), rel.RelType),
		// note: no need to set start/end identitykind because this code path is generic-ingest only which has no base kind.
	}

	return result, nil

}

func submitUpdate(batch graph.Batch, rel ein.IngestibleRelationship) error {
	if update, err := ResolveRelationshipByName(batch, rel); err != nil {
		return err
	} else {
		return batch.UpdateRelationshipBy(update)
	}
}

func ingestDNRelationship(batch graph.Batch, nowUTC time.Time, nextRel ein.IngestibleRelationship) error {
	nextRel.RelProps[common.LastSeen.String()] = nowUTC
	nextRel.Source.Value = strings.ToUpper(nextRel.Source.Value)
	nextRel.Target.Value = strings.ToUpper(nextRel.Target.Value)

	return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(nextRel.RelProps), nextRel.RelType),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			ad.DistinguishedName: nextRel.Source,
			common.LastSeen:      nowUTC,
		}), nextRel.Source.Kind),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			ad.DistinguishedName.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Target,
			common.LastSeen: nowUTC,
		}), nextRel.Target.Kind),
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
