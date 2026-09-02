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

package storage

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	smithyMiddleware "github.com/aws/smithy-go/middleware"
	"github.com/specterops/bloodhound/cmd/api/src/config"
)

const s3IngestDiagnostic = "s3_ingest"

type awsS3OperationDiagnostic struct {
	ctx        context.Context
	attributes []slog.Attr
	startedAt  time.Time
}

// The helpers in this file provide temporary, removable diagnostics for S3-backed ingest integration.
func s3IngestDiagnosticAttributes(attributes ...slog.Attr) []slog.Attr {
	return append([]slog.Attr{slog.String("diagnostic", s3IngestDiagnostic)}, attributes...)
}

func addS3IngestDiagnosticMiddleware(stack *smithyMiddleware.Stack) error {
	return stack.Initialize.Add(awsS3OperationDiagnosticMiddleware{}, smithyMiddleware.Before)
}

type awsS3OperationDiagnosticMiddleware struct{}

func (s awsS3OperationDiagnosticMiddleware) ID() string {
	return "S3IngestDiagnostic"
}

func (s awsS3OperationDiagnosticMiddleware) HandleInitialize(
	ctx context.Context,
	input smithyMiddleware.InitializeInput,
	next smithyMiddleware.InitializeHandler,
) (smithyMiddleware.InitializeOutput, smithyMiddleware.Metadata, error) {
	var (
		diagnostic = startAWSS3OperationDiagnostic(ctx, smithyMiddleware.GetOperationName(ctx), input.Parameters)
		output     smithyMiddleware.InitializeOutput
		metadata   smithyMiddleware.Metadata
		err        error
	)

	output, metadata, err = next.HandleInitialize(ctx, input)
	diagnostic.finish(err)

	return output, metadata, err
}

func awsS3BucketAndKeyAttributes(bucket, key *string) []slog.Attr {
	return []slog.Attr{
		slog.String("bucket", aws.ToString(bucket)),
		slog.String("key", aws.ToString(key)),
	}
}

func awsS3OperationRequestAttributes(parameters any) []slog.Attr {
	switch input := parameters.(type) {
	case *s3.PutObjectInput:
		return awsS3BucketAndKeyAttributes(input.Bucket, input.Key)
	case *s3.CreateMultipartUploadInput:
		return awsS3BucketAndKeyAttributes(input.Bucket, input.Key)
	case *s3.UploadPartInput:
		return append(awsS3BucketAndKeyAttributes(input.Bucket, input.Key), slog.Int("part_number", int(aws.ToInt32(input.PartNumber))))
	case *s3.CompleteMultipartUploadInput:
		return awsS3BucketAndKeyAttributes(input.Bucket, input.Key)
	case *s3.AbortMultipartUploadInput:
		return awsS3BucketAndKeyAttributes(input.Bucket, input.Key)
	case *s3.GetObjectInput:
		return awsS3BucketAndKeyAttributes(input.Bucket, input.Key)
	case *s3.HeadObjectInput:
		return awsS3BucketAndKeyAttributes(input.Bucket, input.Key)
	case *s3.DeleteObjectInput:
		return awsS3BucketAndKeyAttributes(input.Bucket, input.Key)
	case *s3.ListObjectsV2Input:
		return []slog.Attr{
			slog.String("bucket", aws.ToString(input.Bucket)),
			slog.String("key_prefix", aws.ToString(input.Prefix)),
		}
	case *s3.CopyObjectInput:
		return append(awsS3BucketAndKeyAttributes(input.Bucket, input.Key), slog.String("copy_source", aws.ToString(input.CopySource)))
	case *s3.UploadPartCopyInput:
		return append(
			awsS3BucketAndKeyAttributes(input.Bucket, input.Key),
			slog.String("copy_source", aws.ToString(input.CopySource)),
			slog.Int("part_number", int(aws.ToInt32(input.PartNumber))),
		)
	default:
		return nil
	}
}

func startAWSS3OperationDiagnostic(ctx context.Context, operation string, parameters any) awsS3OperationDiagnostic {
	var (
		attributes = []slog.Attr{
			slog.String("aws_operation", operation),
		}
		startAttributes []slog.Attr
		deadline        time.Time
		hasDeadline     bool
	)

	attributes = append(attributes, awsS3OperationRequestAttributes(parameters)...)
	startAttributes = append([]slog.Attr{}, attributes...)
	deadline, hasDeadline = ctx.Deadline()
	startAttributes = append(startAttributes, slog.Bool("context_has_deadline", hasDeadline))
	if hasDeadline {
		startAttributes = append(startAttributes, slog.Time("context_deadline", deadline))
	}

	slog.LogAttrs(
		ctx,
		slog.LevelDebug,
		"S3 ingest diagnostic: AWS S3 API operation started",
		s3IngestDiagnosticAttributes(startAttributes...)...,
	)

	return awsS3OperationDiagnostic{
		ctx:        ctx,
		attributes: attributes,
		startedAt:  time.Now(),
	}
}

func (s awsS3OperationDiagnostic) finish(err error) {
	var attributes = append([]slog.Attr{}, s.attributes...)
	attributes = append(attributes, slog.Duration("duration", time.Since(s.startedAt)))

	if err != nil {
		attributes = append(attributes, slog.Any("error", err))
		slog.LogAttrs(
			s.ctx,
			slog.LevelError,
			"S3 ingest diagnostic: AWS S3 API operation failed",
			s3IngestDiagnosticAttributes(attributes...)...,
		)
		return
	}

	slog.LogAttrs(
		s.ctx,
		slog.LevelDebug,
		"S3 ingest diagnostic: AWS S3 API operation completed",
		s3IngestDiagnosticAttributes(attributes...)...,
	)
}

func logAWSConfigurationLoading(ctx context.Context, bucketConfiguration config.BucketConfiguration) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"S3 ingest diagnostic: loading AWS configuration",
		s3IngestDiagnosticAttributes(
			slog.String("bucket", strings.TrimSpace(bucketConfiguration.Name)),
			slog.String("region", strings.TrimSpace(bucketConfiguration.Region)),
		)...,
	)
}

func logS3ClientInitialized(ctx context.Context, bucketConfiguration config.BucketConfiguration) {
	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"S3 ingest diagnostic: S3 client initialized",
		s3IngestDiagnosticAttributes(
			slog.String("bucket", strings.TrimSpace(bucketConfiguration.Name)),
			slog.String("region", strings.TrimSpace(bucketConfiguration.Region)),
		)...,
	)
}

func logFileServiceInitialized(ctx context.Context, bucketConfiguration config.BucketConfiguration, definition resolvedFileServiceDefinition) {
	attributes := []slog.Attr{
		slog.String("file_service", string(definition.definition.Name)),
		slog.String("provider", string(definition.provider)),
	}

	switch definition.provider {
	case fileServiceProviderS3:
		attributes = append(
			attributes,
			slog.String("bucket", strings.TrimSpace(bucketConfiguration.Name)),
			slog.String("region", strings.TrimSpace(bucketConfiguration.Region)),
			slog.String("prefix", definition.prefix),
		)
	case fileServiceProviderLocal:
		attributes = append(attributes, slog.String("local_path", definition.definition.LocalPath))
	}

	slog.LogAttrs(
		ctx,
		slog.LevelInfo,
		"S3 ingest diagnostic: file service initialized",
		s3IngestDiagnosticAttributes(attributes...)...,
	)
}
