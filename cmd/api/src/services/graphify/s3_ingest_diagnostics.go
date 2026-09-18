// Copyright 2026 Specter Ops, Inc.
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

package graphify

import (
	"context"
	"log/slog"
	"time"

	"github.com/specterops/bloodhound/packages/go/storage"
)

const s3IngestDiagnostic = "s3_ingest"

type ingestStorageStreamDiagnostic struct {
	ctx            context.Context
	storedFileName string
	startedAt      time.Time
}

// The helpers in this file provide temporary, removable diagnostics for S3-backed ingest integration.
func logStoredIngestFileOpenStarted(ctx context.Context, storedFileName string) {
	slog.LogAttrs(
		ctx,
		slog.LevelDebug,
		"S3 ingest diagnostic: opening stored ingest file",
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("stored_file_name", storedFileName),
	)
}

func logStoredIngestFileOpenFinished(ctx context.Context, storedFileName string, fileInfo storage.FileInfo, err error) {
	var attributes = []slog.Attr{
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("stored_file_name", storedFileName),
		slog.Int64("content_length", fileInfo.Size),
	}

	if err != nil {
		attributes = append(attributes, slog.Any("error", err))
		slog.LogAttrs(ctx, slog.LevelError, "S3 ingest diagnostic: failed to open stored ingest file", attributes...)
		return
	}

	slog.LogAttrs(ctx, slog.LevelDebug, "S3 ingest diagnostic: stored ingest file opened", attributes...)
}

func startIngestStorageStreamDiagnostic(ctx context.Context, storedFileName string) ingestStorageStreamDiagnostic {
	slog.LogAttrs(
		ctx,
		slog.LevelDebug,
		"S3 ingest diagnostic: streaming stored ingest file to scratch",
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("stored_file_name", storedFileName),
	)

	return ingestStorageStreamDiagnostic{
		ctx:            ctx,
		storedFileName: storedFileName,
		startedAt:      time.Now(),
	}
}

func (s ingestStorageStreamDiagnostic) finish(bytesCopied int64, err error) {
	var attributes = []slog.Attr{
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("stored_file_name", s.storedFileName),
		slog.Int64("bytes_copied", bytesCopied),
		slog.Duration("duration", time.Since(s.startedAt)),
	}

	if err != nil {
		attributes = append(attributes, slog.Any("error", err))
		slog.LogAttrs(s.ctx, slog.LevelError, "S3 ingest diagnostic: failed streaming stored ingest file to scratch", attributes...)
		return
	}

	slog.LogAttrs(s.ctx, slog.LevelDebug, "S3 ingest diagnostic: stored ingest file streamed to scratch", attributes...)
}
