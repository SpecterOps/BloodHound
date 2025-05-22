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

package ingest

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"

	"github.com/specterops/bloodhound/src/database/types/null"
	"github.com/specterops/bloodhound/src/model"
)

type IngestTaskParams struct {
	Filename  string
	FileType  model.FileType
	RequestID string
	JobID     int64
	IsGeneric bool
}

func CreateIngestTask(ctx context.Context, db IngestData, params IngestTaskParams) (model.IngestTask, error) {
	newIngestTask := model.IngestTask{
		FileName:    params.Filename,
		RequestGUID: params.RequestID,
		TaskID:      null.Int64From(params.JobID),
		FileType:    params.FileType,
		IsGeneric:   params.IsGeneric,
	}

	return db.CreateIngestTask(ctx, newIngestTask)
}

func CreateCompositionInfo(ctx context.Context, db IngestData, nodes model.EdgeCompositionNodes, edges model.EdgeCompositionEdges) (model.EdgeCompositionNodes, model.EdgeCompositionEdges, error) {
	return db.CreateCompositionInfo(ctx, nodes, edges)
}

// Datapipe calls this function to process tasks
func (s *IngestService) ProcessIngestTasks() {

	if ingestTasks, err := s.db.GetAllIngestTasks(s.ctx); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Failed fetching available ingest tasks: %v", err))
	} else {
		for _, ingestTask := range ingestTasks {
			// Check the context to see if we should continue processing ingest tasks. This has to be explicit since error
			// handling assumes that all failures should be logged and not returned.
			if s.ctx.Err() != nil {
				return
			}

			if s.cfg.DisableIngest {
				slog.WarnContext(s.ctx, "Skipped processing of ingestTasks due to config flag.")
				return
			}

			total, failed, err := s.processIngestFile(s.ctx, ingestTask.FileName, ingestTask.FileType)
			if errors.Is(err, fs.ErrNotExist) {
				slog.WarnContext(s.ctx, fmt.Sprintf("Did not process ingest task %d with file %s: %v", ingestTask.ID, ingestTask.FileName, err))
			} else if err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed processing ingest task %d with file %s: %v", ingestTask.ID, ingestTask.FileName, err))

				//kpom-todo: Needs extraction into service.UpdateIngestJob
			} else if job, err := s.db.GetIngestJob(s.ctx, ingestTask.TaskID.ValueOrZero()); err != nil {
				slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to fetch job for ingest task %d: %v", ingestTask.ID, err))
			} else {
				job.TotalFiles = total
				job.FailedFiles += failed
				if err = s.db.UpdateIngestJob(s.ctx, job); err != nil {
					slog.ErrorContext(s.ctx, fmt.Sprintf("Failed to update number of failed files for ingest job ID %d: %v", job.ID, err))
				}
			}

			s.clearFileTask(ingestTask)
		}
	}
}

// clearFileTask removes a generic ingest task for ingested data.
func (s *IngestService) clearFileTask(ingestTask model.IngestTask) {
	if err := s.db.DeleteIngestTask(s.ctx, ingestTask); err != nil {
		slog.ErrorContext(s.ctx, fmt.Sprintf("Error removing ingest task from db: %v", err))
	}
}
