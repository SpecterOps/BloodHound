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

package graphify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/util"
)

const (
	IngestCountThreshold = 500
	ReconcileProperty    = "reconcile"
)

// registrationFn persists a kind encountered in the ingest payload, if it hasn't already been registered in the source_kinds table.
// (e.g., "Base", "AZBase", "GithubBase")
type registrationFn func(kind graph.Kind) error

type ReadOptions struct {
	FileType           model.FileType // JSON or ZIP
	IngestSchema       upload.IngestSchema
	registerSourceKind registrationFn
}

type TimestampedBatch struct {
	Batch      graph.Batch
	IngestTime time.Time
}

func NewTimestampedBatch(batch graph.Batch, ingestTime time.Time) *TimestampedBatch {
	return &TimestampedBatch{
		Batch:      batch,
		IngestTime: ingestTime,
	}
}

// ReadFileForIngest orchestrates the ingestion of a file into the graph database,
// performing any necessary metadata validation and schema enforcement before
// delegating to the core ingest logic.
//
// If the file type is ZIP, additional validation is performed using JSON Schema,
// and the full stream is consumed to enable downstream readers to function correctly.
// Zip files are validated here and not at file upload time because it would be expensive to
// decompress the entire zip into memory.
// Files that fail this validation step will not be processed further.
//
// Returns an error if metadata validation or ingestion fails.
func ReadFileForIngest(batch *TimestampedBatch, reader io.ReadSeeker, options ReadOptions) error {

	var (
		shouldValidateGraph = false
	)

	// TODO: Should this be moved into the upload service. The comment here is helpful, but more
	// discovery required.
	// if filetype == ZIP, we need to validate against jsonschema because
	// the archive bypassed validation controls at file upload time, as opposed to JSON files,
	// which were validated at file upload time
	if options.FileType == model.FileTypeZip {
		shouldValidateGraph = true
	}

	if meta, err := upload.ParseAndValidatePayload(reader, options.IngestSchema, shouldValidateGraph, shouldValidateGraph); err != nil {
		return err
	} else {
		// Because we gave the reader to ParseAndValidatePayload above, if they read the whole
		// thing, we need to make sure we're starting at the front. Be kind, Rewind.
		if _, err := reader.Seek(0, io.SeekStart); err != nil {
			return fmt.Errorf("rewind failed: %w", err)
		}
		return IngestWrapper(batch, reader, meta, options)
	}
}

func IngestBasicData(batch *TimestampedBatch, converted ConvertedData) error {
	errs := util.NewErrorCollector()

	if err := IngestNodes(batch, ad.Entity, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, ad.Entity, converted.RelProps); err != nil {
		errs.Add(err)
	}

	return errs.Combined()
}

// IngestGenericData writes generic graph data into the database using the provided batch.
// It attempts to ingest all nodes and relationships from the ConvertedData object.
//
// Because generic entities do not have a predefined base kind (unlike AZ or AD), this function passes
// graph.EmptyKind to the node and relationship ingestion functions. This indicates that no
// base kind should be applied uniformly to all ingested entities, and instead the kind(s)
// defined directly on each node or edge (if any) are used as-is.
func IngestGenericData(batch *TimestampedBatch, sourceKind graph.Kind, converted ConvertedData) error {
	errs := util.NewErrorCollector()

	if err := IngestNodes(batch, sourceKind, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, sourceKind, converted.RelProps); err != nil {
		errs.Add(err)
	}

	return errs.Combined()
}

func IngestGroupData(batch *TimestampedBatch, converted ConvertedGroupData) error {
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

func IngestAzureData(batch *TimestampedBatch, converted ConvertedAzureData) error {
	defer measure.ContextLogAndMeasure(context.TODO(), slog.LevelDebug, "ingest azure data")()
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

// IngestWrapper dispatches the ingest process based on the metadata's type.
func IngestWrapper(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata, readOpts ReadOptions) error {
	// Source-kind-aware handler
	if handler, ok := sourceKindHandlers[meta.Type]; ok {
		if readOpts.registerSourceKind == nil {
			return fmt.Errorf("missing source kind registration function for data type: %v", meta.Type)
		}
		return handler(batch, reader, meta, readOpts.registerSourceKind)
	}

	// Basic handler
	if handler, ok := basicHandlers[meta.Type]; ok {
		return handler(batch, reader, meta)
	}

	return fmt.Errorf("no handler for ingest data type: %v", meta.Type)
}

// basicIngestHandler defines the function signature for all ingest paths except for the OpenGraph
type basicIngestHandler func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata) error

// sourceKindIngestHandler defines the function signature for ingest handlers that require
// additional logic — specifically, registration of a sourceKind before decoding data.
// This is only used for ingest payloads within OpenGraph, which may specify new source kinds that we want to track (e.g. Base, AZBase, GithubBase).
type sourceKindIngestHandler func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata, register registrationFn) error

func defaultBasicHandler[T any](conversionFunc ConversionFuncWithTime[T]) basicIngestHandler {
	return func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata) error {
		decoder, err := getDefaultDecoder(reader)
		if err != nil {
			return err
		}
		return decodeBasicData(batch, decoder, conversionFunc)
	}
}

var basicHandlers = map[ingest.DataType]basicIngestHandler{
	ingest.DataTypeComputer: func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else if meta.Version >= 5 {
			return decodeBasicData(batch, decoder, convertComputerData)
		} else {
			return nil
		}
	},
	ingest.DataTypeGroup: func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else {
			return decodeGroupData(batch, decoder)
		}
	},
	ingest.DataTypeSession: func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else {
			return decodeSessionData(batch, decoder)
		}
	},
	ingest.DataTypeAzure: func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata) error {
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

var sourceKindHandlers = map[ingest.DataType]sourceKindIngestHandler{
	ingest.DataTypeOpenGraph: func(batch *TimestampedBatch, reader io.ReadSeeker, meta ingest.Metadata, registerSourceKind registrationFn) error {
		sourceKind := graph.EmptyKind

		// decode metadata, if present
		if decoder, err := CreateIngestDecoder(reader, "metadata", 1); err != nil {
			if !errors.Is(err, ingest.ErrDataTagNotFound) {
				return err
			}
			slog.Debug("no metadata found in opengraph payload; continuing to nodes")
		} else {
			var meta ein.GenericMetadata
			if err := decoder.Decode(&meta); err != nil {
				return fmt.Errorf("failed to parse opengraph metadata tag: %w", err)
			}

			sourceKind = graph.StringKind(meta.SourceKind)
			if err := registerSourceKind(sourceKind); err != nil {
				return fmt.Errorf("failed to register sourceKind: %w", err)
			}
		}

		// decode nodes, if present
		if decoder, err := CreateIngestDecoder(reader, "nodes", 2); err != nil {
			if !errors.Is(err, ingest.ErrDataTagNotFound) {
				return err
			}
			slog.Debug("no nodes found in opengraph payload; continuing to edges")
		} else if err := decodeGenericData(batch, decoder, sourceKind, convertGenericNode); err != nil {
			return err
		}

		// decode edges, if present
		if decoder, err := CreateIngestDecoder(reader, "edges", 2); err != nil {
			if !errors.Is(err, ingest.ErrDataTagNotFound) {
				return err
			}
			slog.Debug("no edges found in opengraph payload")
		} else {
			return decodeGenericData(batch, decoder, sourceKind, convertGenericEdge)
		}

		return nil
	},
}

func getDefaultDecoder(reader io.ReadSeeker) (*json.Decoder, error) {
	return CreateIngestDecoder(reader, "data", 1)
}

func NormalizeEinNodeProperties(properties map[string]any, objectID string, ingestTime time.Time) map[string]any {
	if properties == nil {
		properties = make(map[string]any)
	}
	delete(properties, ReconcileProperty)
	properties[common.LastSeen.String()] = ingestTime
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

func IngestNode(batch *TimestampedBatch, baseKind graph.Kind, nextNode ein.IngestibleNode) error {
	var (
		nodeKinds            = mergeBaseKind(baseKind, nextNode.Labels...)
		normalizedProperties = NormalizeEinNodeProperties(nextNode.PropertyMap, nextNode.ObjectID, batch.IngestTime)
		nodeUpdate           = graph.NodeUpdate{
			Node:         graph.PrepareNode(graph.AsProperties(normalizedProperties), nodeKinds...),
			IdentityKind: baseKind, // todo: when this is empty kind i think it gets saved to the kinds property
			IdentityProperties: []string{
				common.ObjectID.String(),
			},
		}
	)

	return batch.Batch.UpdateNodeBy(nodeUpdate)
}

func IngestNodes(batch *TimestampedBatch, baseKind graph.Kind, nodes []ein.IngestibleNode) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range nodes {
		if err := IngestNode(batch, baseKind, next); err != nil {
			slog.Error(fmt.Sprintf("Error ingesting node ID %s: %v", next.ObjectID, err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}

// IngestRelationships resolves and writes a batch of ingestible relationships to the graph.
//
// This function first calls resolveRelationships to resolve node identifiers based on name and kind.
//
// Each resolved relationship update is applied to the graph via batch.UpdateRelationshipBy.
// Errors encountered during resolution or update are collected and returned as a single combined error.
func IngestRelationships(batch *TimestampedBatch, baseKind graph.Kind, relationships []ein.IngestibleRelationship) error {
	var (
		errs = util.NewErrorCollector()
	)

	updates, err := resolveRelationships(batch, relationships, baseKind)
	if err != nil {
		errs.Add(err)
	}

	for _, update := range updates {
		if err := batch.Batch.UpdateRelationshipBy(update); err != nil {
			errs.Add(err)
		}
	}

	return errs.Combined()
}

func ingestDNRelationship(batch *TimestampedBatch, nextRel ein.IngestibleRelationship) error {
	nextRel.RelProps[common.LastSeen.String()] = batch.IngestTime
	nextRel.Source.Value = strings.ToUpper(nextRel.Source.Value)
	nextRel.Target.Value = strings.ToUpper(nextRel.Target.Value)

	return batch.Batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(nextRel.RelProps), nextRel.RelType),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			ad.DistinguishedName: nextRel.Source,
			common.LastSeen:      batch.IngestTime,
		}), nextRel.Source.Kind),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			ad.DistinguishedName.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextRel.Target,
			common.LastSeen: batch.IngestTime,
		}), nextRel.Target.Kind),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestDNRelationships(batch *TimestampedBatch, relationships []ein.IngestibleRelationship) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range relationships {
		if err := ingestDNRelationship(batch, next); err != nil {
			slog.Error(fmt.Sprintf("Error ingesting relationship: %v", err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}

func ingestSession(batch *TimestampedBatch, nextSession ein.IngestibleSession) error {
	nextSession.Target = strings.ToUpper(nextSession.Target)
	nextSession.Source = strings.ToUpper(nextSession.Source)

	return batch.Batch.UpdateRelationshipBy(graph.RelationshipUpdate{
		Relationship: graph.PrepareRelationship(graph.AsProperties(graph.PropertyMap{
			common.LastSeen: batch.IngestTime,
			ad.LogonType:    nextSession.LogonType,
		}), ad.HasSession),

		Start: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Source,
			common.LastSeen: batch.IngestTime,
		}), ad.Computer),
		StartIdentityKind: ad.Entity,
		StartIdentityProperties: []string{
			common.ObjectID.String(),
		},

		End: graph.PrepareNode(graph.AsProperties(graph.PropertyMap{
			common.ObjectID: nextSession.Target,
			common.LastSeen: batch.IngestTime,
		}), ad.User),
		EndIdentityKind: ad.Entity,
		EndIdentityProperties: []string{
			common.ObjectID.String(),
		},
	})
}

func IngestSessions(batch *TimestampedBatch, sessions []ein.IngestibleSession) error {
	var (
		errs = util.NewErrorCollector()
	)

	for _, next := range sessions {
		if err := ingestSession(batch, next); err != nil {
			slog.Error(fmt.Sprintf("Error ingesting sessions: %v", err))
			errs.Add(err)
		}
	}
	return errs.Combined()
}
