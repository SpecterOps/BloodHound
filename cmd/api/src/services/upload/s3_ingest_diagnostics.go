// Copyright 2026 Specter Ops, Inc.
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

package upload

import (
	"context"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

const s3IngestDiagnostic = "s3_ingest"

type ingestUploadDiagnostic struct {
	ctx       context.Context
	requestID string
	jobID     int64
	fileType  model.FileType
	startedAt time.Time
}

type ingestStorageWriteDiagnostic struct {
	ctx       context.Context
	requestID string
	prefix    string
	startedAt time.Time
}

type ingestTaskCreationDiagnostic struct {
	ctx            context.Context
	requestID      string
	jobID          int64
	storedFileName string
	startedAt      time.Time
}

// The helpers in this file provide temporary, removable diagnostics for S3-backed ingest integration.
func startIngestUploadDiagnostic(ctx context.Context, jobID int64, fileType model.FileType) ingestUploadDiagnostic {
	requestID := bhctx.Get(ctx).RequestID

	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"S3 ingest diagnostic: ingest upload started",
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("request_id", requestID),
		slog.Int64("job_id", jobID),
		slog.String("file_type", fileType.String()),
	)

	return ingestUploadDiagnostic{
		ctx:       ctx,
		requestID: requestID,
		jobID:     jobID,
		fileType:  fileType,
		startedAt: time.Now(),
	}
}

func (s ingestUploadDiagnostic) finish(tempFileName string, err error) {
	var attributes = []slog.Attr{
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("request_id", s.requestID),
		slog.Int64("job_id", s.jobID),
		slog.String("file_type", s.fileType.String()),
		slog.String("stored_file_name", tempFileName),
		slog.Duration("duration", time.Since(s.startedAt)),
	}

	if err != nil {
		attributes = append(attributes, slog.Any("error", err))
		slog.LogAttrs(s.ctx, slog.LevelError, "S3 ingest diagnostic: ingest upload failed", attributes...)
		return
	}

	slog.LogAttrs(s.ctx, slog.LevelInfo, "S3 ingest diagnostic: ingest upload stored and validated", attributes...)
}

func startIngestStorageWriteDiagnostic(ctx context.Context, prefix string) ingestStorageWriteDiagnostic {
	var (
		requestID  = bhctx.Get(ctx).RequestID
		attributes = []slog.Attr{
			slog.String("diagnostic", s3IngestDiagnostic),
			slog.String("request_id", requestID),
			slog.String("prefix", prefix),
		}
		deadline    time.Time
		hasDeadline bool
	)

	deadline, hasDeadline = ctx.Deadline()
	attributes = append(attributes, slog.Bool("context_has_deadline", hasDeadline))
	if hasDeadline {
		attributes = append(attributes, slog.Time("context_deadline", deadline))
	}

	slog.LogAttrs(ctx, slog.LevelInfo, "S3 ingest diagnostic: ingest storage write started", attributes...)

	return ingestStorageWriteDiagnostic{
		ctx:       ctx,
		requestID: requestID,
		prefix:    prefix,
		startedAt: time.Now(),
	}
}

func (s ingestStorageWriteDiagnostic) finish(tempFileName string, err error) {
	var attributes = []slog.Attr{
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("request_id", s.requestID),
		slog.String("prefix", s.prefix),
		slog.String("stored_file_name", tempFileName),
		slog.Duration("duration", time.Since(s.startedAt)),
	}

	if err != nil {
		attributes = append(attributes, slog.Any("error", err))
		slog.LogAttrs(s.ctx, slog.LevelError, "S3 ingest diagnostic: ingest storage write returned", attributes...)
		return
	}

	slog.LogAttrs(s.ctx, slog.LevelInfo, "S3 ingest diagnostic: ingest storage write returned", attributes...)
}

func startIngestTaskCreationDiagnostic(ctx context.Context, parameters IngestTaskParams) ingestTaskCreationDiagnostic {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"S3 ingest diagnostic: creating ingest task",
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("request_id", parameters.RequestID),
		slog.Int64("job_id", parameters.JobID),
		slog.String("stored_file_name", parameters.Filename),
	)

	return ingestTaskCreationDiagnostic{
		ctx:            ctx,
		requestID:      parameters.RequestID,
		jobID:          parameters.JobID,
		storedFileName: parameters.Filename,
		startedAt:      time.Now(),
	}
}

func (s ingestTaskCreationDiagnostic) finish(err error) {
	var attributes = []slog.Attr{
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("request_id", s.requestID),
		slog.Int64("job_id", s.jobID),
		slog.String("stored_file_name", s.storedFileName),
		slog.Duration("duration", time.Since(s.startedAt)),
	}

	if err != nil {
		attributes = append(attributes, slog.Any("error", err))
		slog.LogAttrs(s.ctx, slog.LevelError, "S3 ingest diagnostic: ingest task creation failed", attributes...)
		return
	}

	slog.LogAttrs(s.ctx, slog.LevelInfo, "S3 ingest diagnostic: ingest task created", attributes...)
}
