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
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"log/slog"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

const (
	minimumPendingPartSize int64 = 8 * 1024 * 1024
	maximumPendingPartSize int64 = 5 * 1024 * 1024 * 1024
	maximumPendingParts          = 10000
	pendingCleanupTimeout        = 30 * time.Second
)

func uploadPartSize(sizeHint int64) (int64, error) {
	var (
		partSize            = minimumPendingPartSize
		minimumForPartLimit int64
		partSizeUnits       int64
	)

	if sizeHint > maximumPendingPartSize*maximumPendingParts {
		return 0, fmt.Errorf("required S3 multipart size exceeds maximum part size %d", maximumPendingPartSize)
	}

	if sizeHint <= 0 {
		return partSize, nil
	}

	minimumForPartLimit = sizeHint / maximumPendingParts
	if sizeHint%maximumPendingParts != 0 {
		minimumForPartLimit++
	}

	if minimumForPartLimit > partSize {
		partSizeUnits = minimumForPartLimit / minimumPendingPartSize
		if minimumForPartLimit%minimumPendingPartSize != 0 {
			partSizeUnits++
		}

		partSize = partSizeUnits * minimumPendingPartSize
	}

	return partSize, nil
}

type s3PendingFile struct {
	store       *Store
	ctx         context.Context
	name        string
	key         string
	options     WriteOptions
	partSize    int64
	buffer      bytes.Buffer
	uploadID    string
	parts       []s3types.CompletedPart
	committed   bool
	aborted     bool
	terminalErr error
}

func (s *Store) BeginWrite(ctx context.Context, name string, options WriteOptions) (PendingFile, error) {
	var (
		key      string
		partSize int64
		err      error
	)

	err = ctx.Err()
	if err != nil {
		return nil, err
	}

	key, err = s.key(name)
	if err != nil {
		return nil, err
	}

	partSize, err = uploadPartSize(options.SizeHint)
	if err != nil {
		return nil, err
	}

	return &s3PendingFile{
		store:    s,
		ctx:      ctx,
		name:     name,
		key:      key,
		options:  options,
		partSize: partSize,
	}, nil
}

func (s *s3PendingFile) Write(data []byte) (int, error) {
	var (
		bytesWritten int
		remaining    int
		chunkSize    int
		err          error
	)

	if s.aborted || s.committed {
		return 0, io.ErrClosedPipe
	}

	if s.terminalErr != nil {
		return 0, s.terminalErr
	}

	if err = s.ctx.Err(); err != nil {
		return 0, s.fail(err)
	}

	for len(data) > 0 {
		if int64(s.buffer.Len()) == s.partSize {
			if err = s.ensureMultipart(); err != nil {
				return bytesWritten, s.fail(err)
			}
			if err = s.uploadBufferedPart(); err != nil {
				return bytesWritten, s.fail(err)
			}
		}

		remaining = int(s.partSize) - s.buffer.Len()
		chunkSize = min(remaining, len(data))
		if _, err = s.buffer.Write(data[:chunkSize]); err != nil {
			return bytesWritten, s.fail(err)
		}
		bytesWritten += chunkSize
		data = data[chunkSize:]
	}
	return bytesWritten, nil
}

func (s *s3PendingFile) fail(err error) error {
	if s.terminalErr == nil {
		s.terminalErr = fmt.Errorf("%w: %w", ErrWriteFailed, err)
	}
	return s.terminalErr
}

func (s *s3PendingFile) ensureMultipart() error {
	var (
		diagnostic s3OperationDiagnostic
		input      *s3.CreateMultipartUploadInput
		output     *s3.CreateMultipartUploadOutput
		err        error
	)

	if s.uploadID != "" {
		return nil
	}

	if err = s.ctx.Err(); err != nil {
		return err
	}

	diagnostic = startS3OperationDiagnostic(s.ctx, "create_multipart_upload", s.store.bucket, s.key)
	input = &s3.CreateMultipartUploadInput{
		Bucket:            aws.String(s.store.bucket),
		Key:               aws.String(s.key),
		ChecksumAlgorithm: s3types.ChecksumAlgorithmCrc32,
		ChecksumType:      s3types.ChecksumTypeComposite,
	}
	if s.options.ContentType != "" {
		input.ContentType = aws.String(s.options.ContentType)
	}
	if len(s.options.Metadata) > 0 {
		input.Metadata = s.options.Metadata
	}

	output, err = s.store.client.CreateMultipartUpload(s.ctx, input)
	if err != nil {
		diagnostic.finish(err)
		return err
	}

	s.uploadID = aws.ToString(output.UploadId)
	if s.uploadID == "" {
		err = errors.New("S3 returned an empty multipart upload ID")
	}
	diagnostic.finish(err)
	return err
}

func (s *s3PendingFile) uploadBufferedPart() error {
	var (
		partNumber int32
		partData   []byte
		checksum   string
		diagnostic s3OperationDiagnostic
		output     *s3.UploadPartOutput
		part       s3types.CompletedPart
		err        error
	)

	if len(s.parts) >= maximumPendingParts {
		return fmt.Errorf("S3 multipart upload exceeds the %d part limit", maximumPendingParts)
	}

	if err = s.ctx.Err(); err != nil {
		return err
	}

	partNumber = int32(len(s.parts) + 1)
	partData = s.buffer.Bytes()
	checksum = base64.StdEncoding.EncodeToString(crc32Checksum(partData))
	diagnostic = startS3OperationDiagnostic(s.ctx, "upload_part", s.store.bucket, s.key)
	output, err = s.store.client.UploadPart(s.ctx, &s3.UploadPartInput{
		Bucket:            aws.String(s.store.bucket),
		Key:               aws.String(s.key),
		UploadId:          aws.String(s.uploadID),
		PartNumber:        aws.Int32(partNumber),
		Body:              bytes.NewReader(partData),
		ChecksumAlgorithm: s3types.ChecksumAlgorithmCrc32,
		ChecksumCRC32:     aws.String(checksum),
	})
	diagnostic.finish(err, slog.Int("part_number", int(partNumber)), slog.Int64("part_size", int64(len(partData))))

	if err != nil {
		return err
	}
	part = s3types.CompletedPart{
		ETag:          output.ETag,
		PartNumber:    aws.Int32(partNumber),
		ChecksumCRC32: output.ChecksumCRC32,
	}
	if part.ChecksumCRC32 == nil {
		part.ChecksumCRC32 = aws.String(checksum)
	}
	s.parts = append(s.parts, part)
	s.buffer.Reset()

	return nil
}

func crc32Checksum(data []byte) []byte {
	var (
		checksum      uint32
		checksumBytes [4]byte
	)

	checksum = crc32.ChecksumIEEE(data)
	binary.BigEndian.PutUint32(checksumBytes[:], checksum)
	return checksumBytes[:]
}

func (s *s3PendingFile) Commit() (string, error) {
	var (
		input      *s3.CompleteMultipartUploadInput
		diagnostic s3OperationDiagnostic
		err        error
	)

	if s.aborted {
		return "", io.ErrClosedPipe
	}

	if s.committed {
		return s.name, nil
	}

	if s.terminalErr != nil {
		return "", s.terminalErr
	}

	if err = s.ctx.Err(); err != nil {
		s.terminalErr = err
		return "", err
	}

	if s.uploadID == "" {
		err = s.putSmallObject()
		if err != nil {
			s.terminalErr = err
			return "", err
		}

		s.committed = true
		return s.name, nil
	}

	if s.buffer.Len() > 0 {
		if err = s.uploadBufferedPart(); err != nil {
			return "", s.fail(err)
		}
	}
	if err = s.ctx.Err(); err != nil {
		s.terminalErr = err
		return "", err
	}

	input = &s3.CompleteMultipartUploadInput{
		Bucket:   aws.String(s.store.bucket),
		Key:      aws.String(s.key),
		UploadId: aws.String(s.uploadID),
		MultipartUpload: &s3types.CompletedMultipartUpload{
			Parts: s.parts,
		},
		ChecksumType: s3types.ChecksumTypeComposite,
	}
	if s.options.FailIfExists {
		input.IfNoneMatch = aws.String("*")
	}
	diagnostic = startS3OperationDiagnostic(s.ctx, "complete_multipart_upload", s.store.bucket, s.key)
	_, err = s.store.client.CompleteMultipartUpload(s.ctx, input)
	err = mapExistsError(err)
	diagnostic.finish(err, slog.Int("part_count", len(s.parts)))

	if err != nil {
		s.terminalErr = err
		return "", err
	}

	s.committed = true
	return s.name, nil
}

func (s *s3PendingFile) putSmallObject() error {
	var (
		data       []byte
		checksum   string
		input      *s3.PutObjectInput
		diagnostic s3OperationDiagnostic
		err        error
	)

	if err = s.ctx.Err(); err != nil {
		return err
	}

	data = s.buffer.Bytes()
	checksum = base64.StdEncoding.EncodeToString(crc32Checksum(data))
	input = &s3.PutObjectInput{
		Bucket:            aws.String(s.store.bucket),
		Key:               aws.String(s.key),
		Body:              bytes.NewReader(data),
		ChecksumAlgorithm: s3types.ChecksumAlgorithmCrc32,
		ChecksumCRC32:     aws.String(checksum),
	}
	if s.options.FailIfExists {
		input.IfNoneMatch = aws.String("*")
	}
	if s.options.ContentType != "" {
		input.ContentType = aws.String(s.options.ContentType)
	}
	if len(s.options.Metadata) > 0 {
		input.Metadata = s.options.Metadata
	}

	diagnostic = startS3OperationDiagnostic(s.ctx, "put_object", s.store.bucket, s.key)
	_, err = s.store.client.PutObject(s.ctx, input)
	err = mapExistsError(err)
	diagnostic.finish(err, slog.Int64("content_length", int64(len(data))))

	return err
}

func (s *s3PendingFile) Abort() error {
	var (
		cleanupCtx context.Context
		cancel     context.CancelFunc
		diagnostic s3OperationDiagnostic
		err        error
	)

	if s.aborted || s.committed {
		return nil
	}

	s.aborted = true

	if s.uploadID == "" {
		return nil
	}

	cleanupCtx, cancel = context.WithTimeout(context.WithoutCancel(s.ctx), pendingCleanupTimeout)
	defer cancel()
	diagnostic = startS3OperationDiagnostic(cleanupCtx, "abort_multipart_upload", s.store.bucket, s.key)
	_, err = s.store.client.AbortMultipartUpload(cleanupCtx, &s3.AbortMultipartUploadInput{
		Bucket:   aws.String(s.store.bucket),
		Key:      aws.String(s.key),
		UploadId: aws.String(s.uploadID),
	})
	diagnostic.finish(err)

	return err
}
