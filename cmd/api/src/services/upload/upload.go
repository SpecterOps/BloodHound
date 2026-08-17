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

package upload

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/bloodhound/packages/go/metrics"
	"github.com/specterops/bloodhound/packages/go/storage"
	"github.com/specterops/chow/pkg/payload"
)

var ErrInvalidJSON = errors.New("file is not valid json")

// FileValidator defines the interface for ingest file validation.
// It receives a source reader (src) and a destination writer (dst).
// Implementations are responsible for validating the input stream,
// while simultaneously copying it to the destination for persistence.
// This abstraction supports format-agnostic payloads (e.g., JSON, ZIP).
type FileValidator func(src io.Reader, dst io.Writer) (ingest.OriginalMetadata, error)

func SaveIngestFile(ctx context.Context, fileService storage.FileService, request *http.Request, ingestSchema payload.Schema, jobID int64) (IngestTaskParams, payload.ValidationReport, error) {
	var (
		fileData     = request.Body
		fileType     model.FileType
		report       payload.ValidationReport
		validationFn FileValidator
	)

	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		fileType = model.FileTypeJson
		validationFn = func(src io.Reader, dst io.Writer) (ingest.OriginalMetadata, error) {
			var err error
			if report, err = WriteAndValidateJSON(src, dst, ingestSchema); err != nil {
				return ingest.OriginalMetadata{}, err
			}

			return ingest.OriginalMetadata{}, nil
		}
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
		fileType = model.FileTypeZip
		validationFn = WriteAndValidateZip
	default:
		return IngestTaskParams{}, report, fmt.Errorf("invalid content type for ingest file")
	}

	if tempFileName, err := WriteAndValidateFile(ctx, fileService, fileData, ingestFileTempPrefix(jobID), validationFn); err != nil {
		metrics.RecordIngestTask(metrics.IngestCollectorManual, fileFormatFromFileType(fileType), metrics.IngestTaskStatusFailed)
		return IngestTaskParams{}, report, err
	} else {
		return IngestTaskParams{
			Filename: tempFileName,
			FileType: fileType,
		}, report, nil
	}
}

func WriteAndValidateZip(fileData io.Reader, destination io.Writer) (ingest.OriginalMetadata, error) {
	teeReader := io.TeeReader(fileData, destination)
	return ingest.OriginalMetadata{}, ValidateZipFile(teeReader)
}

func WriteAndValidateJSON(fileData io.Reader, destination io.Writer, ingestSchema payload.Schema) (payload.ValidationReport, error) {
	var report payload.ValidationReport

	normalizedReader, err := bomenc.NormalizeToUTF8(fileData)
	if err != nil {
		return report, fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	}

	teeReader := io.TeeReader(normalizedReader, destination)
	ingestValidator := payload.NewValidator(teeReader, ingestSchema)
	if _, report, err = ingestValidator.ParseAndValidate(); err != nil {
		return report, fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	}

	return report, nil
}

func ingestFileTempPrefix(jobID int64) string {
	return fmt.Sprintf("file_upload_job%d_", jobID)
}

func cleanupTempFile(ctx context.Context, fileService storage.FileService, tempFileName string) {
	if tempFileName == "" {
		return
	}

	cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()

	if err := fileService.DeleteFile(cleanupCtx, tempFileName); err != nil {
		slog.ErrorContext(
			cleanupCtx,
			"Failed to delete temp file",
			slog.String("temp_file_name", tempFileName),
			attr.Error(err),
		)
	}
}

func WriteAndValidateFile(ctx context.Context, fileService storage.FileService, fileData io.Reader, prefix string, validationFunc FileValidator) (string, error) {
	if validationFunc == nil {
		return "", fmt.Errorf("validation function is required")
	}

	// Create a pipe: pr (read end) and pw (write end).
	// Data written to pw can be read from pr.
	pr, pw := io.Pipe()

	// validationErrCh carries the result of the validation goroutine.
	// Using a buffered channel (size 1) ensures the goroutine never blocks on send,
	// and gives the main goroutine a synchronization point to wait for the result.
	validationErrCh := make(chan error, 1)

	// Start validation in a separate goroutine.
	// validationFunc reads from the request body and writes the validated output to pw.
	// WriteTempFile reads from pr and persists that validated output.
	go func() {
		_, err := validationFunc(fileData, pw)
		_ = pw.CloseWithError(err)
		validationErrCh <- err
	}()

	// Write to storage while validation happens concurrently.
	tempFileName, writeErr := fileService.WriteTempFile(ctx, prefix, pr, storage.WriteOptions{})
	if writeErr != nil {
		_ = pr.CloseWithError(writeErr)
	}

	var validationErr error
	select {
	case validationErr = <-validationErrCh:
		// Context cancellation wins over validation errors when both are ready.
		if err := ctx.Err(); err != nil {
			cleanupTempFile(ctx, fileService, tempFileName)
			return "", err
		}
	case <-ctx.Done():
		_ = pr.CloseWithError(ctx.Err())
		cleanupTempFile(ctx, fileService, tempFileName)
		return "", ctx.Err()
	}

	// Check if validation failed, which should win over write errors.
	if validationErr != nil {
		slog.ErrorContext(
			ctx,
			"Validation failed",
			slog.String("temp_file_name", tempFileName),
			attr.Error(validationErr),
		)
		cleanupTempFile(ctx, fileService, tempFileName)
		return "", validationErr
	}

	// Check if writing failed; the temp file should be cleaned up.
	if writeErr != nil {
		slog.ErrorContext(
			ctx,
			"Write failed",
			slog.String("temp_file_name", tempFileName),
			attr.Error(writeErr),
		)
		cleanupTempFile(ctx, fileService, tempFileName)
		return "", writeErr
	}

	slog.InfoContext(ctx, "File written and validated", slog.String("temp_file_name", tempFileName))
	return tempFileName, nil
}
