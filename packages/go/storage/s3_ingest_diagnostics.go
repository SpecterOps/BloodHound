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

package storage

import (
	"context"
	"errors"
	"io/fs"
	"log/slog"
	"time"
)

const s3IngestDiagnostic = "s3_ingest"

type s3OperationDiagnostic struct {
	ctx       context.Context
	operation string
	bucket    string
	key       string
	startedAt time.Time
}

// The helpers in this file provide temporary, removable diagnostics for S3-backed ingest integration.
func startS3OperationDiagnostic(ctx context.Context, operation, bucket, key string) s3OperationDiagnostic {
	var (
		attributes = []slog.Attr{
			slog.String("diagnostic", s3IngestDiagnostic),
			slog.String("operation", operation),
			slog.String("bucket", bucket),
			slog.String("key", key),
		}
		deadline    time.Time
		hasDeadline bool
	)

	deadline, hasDeadline = ctx.Deadline()
	attributes = append(attributes, slog.Bool("context_has_deadline", hasDeadline))
	if hasDeadline {
		attributes = append(attributes, slog.Time("context_deadline", deadline))
	}

	slog.LogAttrs(ctx, slog.LevelDebug, "S3 ingest diagnostic: S3 operation started", attributes...)

	return s3OperationDiagnostic{
		ctx:       ctx,
		operation: operation,
		bucket:    bucket,
		key:       key,
		startedAt: time.Now(),
	}
}

func (s s3OperationDiagnostic) finish(err error, attributes ...slog.Attr) {
	attributes = append([]slog.Attr{
		slog.String("diagnostic", s3IngestDiagnostic),
		slog.String("operation", s.operation),
		slog.String("bucket", s.bucket),
		slog.String("key", s.key),
		slog.Duration("duration", time.Since(s.startedAt)),
	}, attributes...)

	if errors.Is(err, fs.ErrNotExist) {
		attributes = append(attributes, slog.Bool("object_missing", true))
		slog.LogAttrs(s.ctx, slog.LevelDebug, "S3 ingest diagnostic: S3 object not found", attributes...)
	} else if err != nil {
		attributes = append(attributes, slog.Any("error", err))
		slog.LogAttrs(s.ctx, slog.LevelError, "S3 ingest diagnostic: S3 operation failed", attributes...)
	} else {
		slog.LogAttrs(s.ctx, slog.LevelDebug, "S3 ingest diagnostic: S3 operation completed", attributes...)
	}
}
