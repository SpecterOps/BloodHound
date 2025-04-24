// Copyright 2025 Specter Ops, Inc.
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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . IngestData
package ingest

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/specterops/bloodhound/bomenc"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/ein"
	"github.com/specterops/bloodhound/graphschema/ad"
	"github.com/specterops/bloodhound/graphschema/azure"
	"github.com/specterops/bloodhound/graphschema/common"
	"github.com/specterops/bloodhound/src/config"
	"github.com/specterops/bloodhound/src/database"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
)

// The IngestData interface is designed to manage the lifecycle of ingestion tasks and jobs in a system that processes graph-based data
type IngestData interface {
	// Task handlers
	CreateIngestTask(ctx context.Context, task model.IngestTask) (model.IngestTask, error)
	DeleteAllIngestTasks(ctx context.Context) error
	CreateCompositionInfo(ctx context.Context, nodes model.EdgeCompositionNodes, edges model.EdgeCompositionEdges) (model.EdgeCompositionNodes, model.EdgeCompositionEdges, error)

	// Job handlers
	CreateIngestJob(ctx context.Context, job model.IngestJob) (model.IngestJob, error)
	UpdateIngestJob(ctx context.Context, job model.IngestJob) error
	GetIngestJob(ctx context.Context, id int64) (model.IngestJob, error)
	GetAllIngestJobs(ctx context.Context, skip int, limit int, order string, filter model.SQLFilter) ([]model.IngestJob, int, error)
	GetIngestJobsWithStatus(ctx context.Context, status model.JobStatus) ([]model.IngestJob, error)
	DeleteAllIngestJobs(ctx context.Context) error
	CancelAllIngestJobs(ctx context.Context) error
}

type IngestService struct {
	db      database.Database
	graphDB graph.Database
	cfg     config.Configuration
}

func NewIngestService(db database.Database, graphDb graph.Database, cfg config.Configuration) IngestService {
	return IngestService{db: db, graphDB: graphDb, cfg: cfg}
}

func (s IngestService) ProcessIngestTasks(ctx context.Context) error {
	if ingestTasks, err := s.db.GetAllIngestTasks(ctx); err != nil {
		return fmt.Errorf("get all ingest tasks: %v", err)
	} else if err := s.db.SetDatapipeStatus(ctx, model.DatapipeStatusIngesting, false); err != nil {
		return fmt.Errorf("set datapipe status ingesting: %v", err)
	} else {
		var errs = make([]error, 0)

		for _, ingestTask := range ingestTasks {
			ingestTaskLogger := slog.Default().With(
				slog.Group("ingest_task",
					slog.Int64("id", ingestTask.ID),
					slog.String("file_name", ingestTask.FileName),
				),
			)

			// Check the context to see if we should continue processing ingest tasks. This has to be explicit since error
			// handling assumes that all failures should be logged and not returned.
			if ctx.Err() != nil {
				errs = append(errs, fmt.Errorf("context error encountered: %v", err))
				return errors.Join(errs...)
			}

			if paths, failed, err := preProcessIngestFile(ctx, s.cfg.TempDirectory(), ingestTask); errors.Is(err, fs.ErrNotExist) {
				ingestTaskLogger.WarnContext(
					ctx,
					"File does not exist for ingest task",
					slog.String("err", err.Error()),
				)
			} else if err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to preprocess ingest file",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("preprocess ingest file: %v", err))
			} else if total, failed, err := processIngestFile(ctx, s.graphDB, paths, failed); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to process ingest file",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("process ingest file: %v", err))
			} else if job, err := s.db.GetIngestJob(ctx, ingestTask.TaskID); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to get ingest job",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("get ingest job: %v", err))
			} else if err := updateIngestJob(ctx, s.db, job, total, failed); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to update file completion for ingest job",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("update ingest job: %v", err))
			}

			if err := s.db.DeleteIngestTask(ctx, ingestTask); err != nil {
				ingestTaskLogger.ErrorContext(
					ctx,
					"Failed to remove ingest task",
					slog.String("err", err.Error()),
				)
				errs = append(errs, fmt.Errorf("delete ingest task: %v", err))
			}
		}
		return errors.Join(errs...)
	}
}

func updateIngestJob(ctx context.Context, db database.Database, job model.IngestJob, total int, failed int) error {
	job.TotalFiles = total
	job.FailedFiles += failed

	if err := db.UpdateIngestJob(ctx, job); err != nil {
		return fmt.Errorf("could not update file completion for ingest job id %d: %w", job.ID, err)
	} else {
		return nil
	}
}

func preProcessIngestFile(ctx context.Context, tmpDir string, ingestTask model.IngestTask) ([]string, int, error) {
	if ingestTask.FileType == model.FileTypeJson {
		//If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []string{ingestTask.FileName}, 0, nil
	} else if archive, err := zip.OpenReader(ingestTask.FileName); err != nil {
		return []string{}, 0, err
	} else {
		var (
			errs      = util.NewErrorCollector()
			failed    = 0
			filePaths = make([]string, len(archive.File))
		)

		for i, f := range archive.File {
			//skip directories
			if f.FileInfo().IsDir() {
				continue
			}
			// Break out if temp file creation fails
			// Collect errors for other failures within the archive
			if tempFile, err := os.CreateTemp(tmpDir, "bh"); err != nil {
				return []string{}, 0, err
			} else if srcFile, err := f.Open(); err != nil {
				errs.Add(fmt.Errorf("error opening file %s in archive %s: %v", f.Name, ingestTask.FileName, err))
				failed++
			} else if normFile, err := bomenc.NormalizeToUTF8(srcFile); err != nil {
				errs.Add(fmt.Errorf("error normalizing file %s to UTF8 in archive %s: %v", f.Name, ingestTask.FileName, err))
				failed++
			} else if _, err := io.Copy(tempFile, normFile); err != nil {
				errs.Add(fmt.Errorf("error extracting file %s in archive %s: %v", f.Name, ingestTask.FileName, err))
				failed++
			} else if err := tempFile.Close(); err != nil {
				errs.Add(fmt.Errorf("error closing temp file %s: %v", f.Name, err))
				failed++
			} else {
				filePaths[i] = tempFile.Name()
			}
		}

		//Close the archive and delete it
		if err := archive.Close(); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error closing archive %s: %v", ingestTask.FileName, err))
		} else if err := os.Remove(ingestTask.FileName); err != nil {
			slog.ErrorContext(ctx, fmt.Sprintf("Error deleting archive %s: %v", ingestTask.FileName, err))
		}

		return filePaths, failed, errs.Combined()
	}
}

func processIngestFile(ctx context.Context, graphDB graph.Database, paths []string, failed int) (int, int, error) {
	return len(paths), failed, graphDB.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, filePath := range paths {
			file, err := os.Open(filePath)
			if err != nil {
				failed++
				return err
			} else if err := ReadFileForIngest(batch, file); err != nil {
				failed++
				slog.ErrorContext(ctx, fmt.Sprintf("Error reading ingest file %s: %v", filePath, err))
			}

			if err := file.Close(); err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error closing ingest file %s: %v", filePath, err))
			} else if err := os.Remove(filePath); errors.Is(err, fs.ErrNotExist) {
				slog.WarnContext(ctx, fmt.Sprintf("Removing ingest file %s: %v", filePath, err))
			} else if err != nil {
				slog.ErrorContext(ctx, fmt.Sprintf("Error removing ingest file %s: %v", filePath, err))
			}
		}

		return nil
	})
}

const (
	IngestCountThreshold = 500
	ReconcileProperty    = "reconcile"
)

func IngestWrapper(batch graph.Batch, reader io.ReadSeeker, meta ingest.Metadata) error {
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

func ReadFileForIngest(batch graph.Batch, reader io.ReadSeeker) error {
	if meta, err := ValidateMetaTag(reader, false); err != nil {
		return fmt.Errorf("error validating meta tag: %w", err)
	} else {
		return IngestWrapper(batch, reader, meta)
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
