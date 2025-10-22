// Copyright 2025 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
//
//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/ingest.go -package=mocks -source=ingest.go
package graphify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/daemons/changelog"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/services/upload"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/bloodhound/packages/go/errorlist"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
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
	RegisterSourceKind registrationFn
}

// IngestContext is a container for dependencies needed by ingest
type IngestContext struct {
	Ctx context.Context
	// Batch is the buffering/flushing mechanism that writes entities to the graph database
	Batch BatchUpdater
	// IngestTime is a single timestamp assigned to the lastseen property of every entity ignested per ingest run
	IngestTime time.Time
	// Manager is the caching layer that deduplicates ingest payloads across ingest runs
	Manager ChangeManager
	// RetainIngestedFiles determines if the service should clean up working files after ingest
	RetainIngestedFiles bool
}

func NewIngestContext(ctx context.Context, opts ...IngestOption) *IngestContext {
	ic := &IngestContext{
		Ctx: ctx,
	}

	for _, opt := range opts {
		opt(ic)
	}

	// avoid a zero IngestTime as it breaks lastseen semantics.
	if ic.IngestTime.IsZero() {
		ic.IngestTime = time.Now()
	}

	return ic
}

// option helpers
type IngestOption func(*IngestContext)

func WithIngestTime(ingestTime time.Time) IngestOption {
	return func(s *IngestContext) {
		s.IngestTime = ingestTime
	}
}

func WithIngestRetentionConfig(shouldRetainIngestedFiles bool) IngestOption {
	return func(s *IngestContext) {
		s.RetainIngestedFiles = shouldRetainIngestedFiles
	}
}

func WithChangeManager(manager ChangeManager) IngestOption {
	return func(s *IngestContext) {
		s.Manager = manager
	}
}

func WithBatchUpdater(batchUpdater BatchUpdater) IngestOption {
	return func(s *IngestContext) {
		s.Batch = batchUpdater
	}
}

func (s *IngestContext) BindBatchUpdater(batch BatchUpdater) {
	s.Batch = batch
}

func (s *IngestContext) HasChangelog() bool {
	return s.Manager != nil
}

// ChangeManager represents the ingestion-facing API for the changelog daemon.
//
// It provides three responsibilities:
//   - Deduplication: ResolveChange determines whether a proposed change is new or modified
//     and therefore requires persistence, or whether it has already been seen.
//   - Submission: Submit enqueues a change for asynchronous processing by the changelog loop.
//   - Metrics: FlushStats logs and resets internal cache hit/miss statistics,
//     allowing callers to observe deduplication efficiency over time.
//
// To generate mocks for this interface for unit testing seams in the application
// please use:
//
// mockgen -source=ingest.go -destination=mocks/ingest.go -package=mocks
type ChangeManager interface {
	ResolveChange(change changelog.Change) (bool, error)
	Submit(ctx context.Context, change changelog.Change) bool
	FlushStats()
	ClearCache(ctx context.Context)
}

// BatchUpdater represents the ingestion-facing API for a dawgs BatchOperation
type BatchUpdater interface {
	UpdateNodeBy(update graph.NodeUpdate) error
	UpdateRelationshipBy(update graph.RelationshipUpdate) error
	Nodes() graph.NodeQuery
	Relationships() graph.RelationshipQuery
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
func ReadFileForIngest(batch *IngestContext, reader io.ReadSeeker, options ReadOptions) error {

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

// IngestWrapper dispatches the ingest process based on the metadata's type.
func IngestWrapper(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata, readOpts ReadOptions) error {
	// Source-kind-aware handler
	if handler, ok := sourceKindHandlers[meta.Type]; ok {
		if readOpts.RegisterSourceKind == nil {
			return fmt.Errorf("missing source kind registration function for data type: %v", meta.Type)
		}
		return handler(batch, reader, meta, readOpts.RegisterSourceKind)
	}

	// Basic handler
	if handler, ok := basicHandlers[meta.Type]; ok {
		return handler(batch, reader, meta)
	}

	return fmt.Errorf("no handler for ingest data type: %v", meta.Type)
}

func IngestBasicData(batch *IngestContext, converted ConvertedData) error {
	errs := errorlist.NewBuilder()

	if err := IngestNodes(batch, ad.Entity, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, ad.Entity, converted.RelProps); err != nil {
		errs.Add(err)
	}

	return errs.Build()
}

// IngestGenericData writes generic graph data into the database using the provided batch.
// It attempts to ingest all nodes and relationships from the ConvertedData object.
//
// Because generic entities do not have a predefined base kind (unlike AZ or AD), this function passes
// graph.EmptyKind to the node and relationship ingestion functions. This indicates that no
// base kind should be applied uniformly to all ingested entities, and instead the kind(s)
// defined directly on each node or edge (if any) are used as-is.
func IngestGenericData(batch *IngestContext, sourceKind graph.Kind, converted ConvertedData) error {
	errs := errorlist.NewBuilder()

	if err := IngestNodes(batch, sourceKind, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, sourceKind, converted.RelProps); err != nil {
		errs.Add(err)
	}

	return errs.Build()
}

func IngestGroupData(batch *IngestContext, converted ConvertedGroupData) error {
	errs := errorlist.NewBuilder()

	if err := IngestNodes(batch, ad.Entity, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, ad.Entity, converted.RelProps); err != nil {
		errs.Add(err)
	}

	if err := IngestDNRelationships(batch, converted.DistinguishedNameProps); err != nil {
		errs.Add(err)
	}

	return errs.Build()
}

func IngestAzureData(batch *IngestContext, converted ConvertedAzureData) error {
	defer measure.ContextLogAndMeasure(context.TODO(), slog.LevelDebug, "ingest azure data")()
	errs := errorlist.NewBuilder()

	if err := IngestNodes(batch, azure.Entity, converted.NodeProps); err != nil {
		errs.Add(err)
	}

	if err := IngestNodes(batch, ad.Entity, converted.OnPremNodes); err != nil {
		errs.Add(err)
	}

	if err := IngestRelationships(batch, azure.Entity, converted.RelProps); err != nil {
		errs.Add(err)
	}

	return errs.Build()
}

// basicIngestHandler defines the function signature for all ingest paths except for the OpenGraph
type basicIngestHandler func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata) error

// sourceKindIngestHandler defines the function signature for ingest handlers that require
// additional logic — specifically, registration of a sourceKind before decoding data.
// This is only used for ingest payloads within OpenGraph, which may specify new source kinds that we want to track (e.g. Base, AZBase, GithubBase).
type sourceKindIngestHandler func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata, register registrationFn) error

func defaultBasicHandler[T any](conversionFunc ConversionFuncWithTime[T]) basicIngestHandler {
	return func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata) error {
		decoder, err := getDefaultDecoder(reader)
		if err != nil {
			return err
		}
		return decodeBasicData(batch, decoder, conversionFunc)
	}
}

var basicHandlers = map[ingest.DataType]basicIngestHandler{
	ingest.DataTypeComputer: func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else if meta.Version >= 5 {
			return decodeBasicData(batch, decoder, convertComputerData)
		} else {
			return nil
		}
	},
	ingest.DataTypeGroup: func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else {
			return decodeGroupData(batch, decoder)
		}
	},
	ingest.DataTypeSession: func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata) error {
		if decoder, err := getDefaultDecoder(reader); err != nil {
			return err
		} else {
			return decodeSessionData(batch, decoder)
		}
	},
	ingest.DataTypeAzure: func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata) error {
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
	ingest.DataTypeOpenGraph: func(batch *IngestContext, reader io.ReadSeeker, meta ingest.Metadata, registerSourceKind registrationFn) error {
		sourceKind := graph.EmptyKind

		// decode metadata, if present
		if decoder, err := CreateIngestDecoder(reader, "metadata", 1); err != nil {
			if !errors.Is(err, ingest.ErrDataTagNotFound) {
				return err
			}
			slog.Debug("No metadata found in opengraph payload; continuing to nodes")
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
			slog.Debug("No nodes found in opengraph payload; continuing to edges")
		} else if err := DecodeGenericData(batch, decoder, sourceKind, ConvertGenericNode); err != nil {
			return err
		}

		// decode edges, if present
		if decoder, err := CreateIngestDecoder(reader, "edges", 2); err != nil {
			if !errors.Is(err, ingest.ErrDataTagNotFound) {
				return err
			}
			slog.Debug("No edges found in opengraph payload")
		} else {
			return DecodeGenericData(batch, decoder, sourceKind, ConvertGenericEdge)
		}

		return nil
	},
}

func getDefaultDecoder(reader io.ReadSeeker) (*json.Decoder, error) {
	return CreateIngestDecoder(reader, "data", 1)
}
