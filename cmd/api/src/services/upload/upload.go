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
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/bloodhound/packages/go/chow/payload"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/bloodhound/packages/go/metrics"
	"github.com/specterops/bloodhound/packages/go/storage"
)

var ErrInvalidJSON = errors.New("file is not valid json")

type ingestValidationResult struct {
	report payload.ValidationReport
	err    error
}

func SaveIngestFile(ctx context.Context, fileService storage.FileService, request *http.Request, ingestSchema payload.Schema, jobID int64) (IngestTaskParams, payload.ValidationReport, error) {
	var (
		uploadDiagnostic  ingestUploadDiagnostic
		storageDiagnostic ingestStorageWriteDiagnostic
		fileType          model.FileType
		fileData          = request.Body
		fileName          string
		report            payload.ValidationReport
		err               error
	)

	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		fileType = model.FileTypeJson
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), AllowedZipFileUploadTypes()...):
		fileType = model.FileTypeZip
	default:
		return IngestTaskParams{}, payload.ValidationReport{}, fmt.Errorf("invalid content type for ingest file")
	}

	if err = ctx.Err(); err != nil {
		return IngestTaskParams{}, payload.ValidationReport{}, err
	}

	if fileData == nil {
		fileData = http.NoBody
	}

	uploadDiagnostic = startIngestUploadDiagnostic(ctx, jobID, fileType)
	defer func() {
		uploadDiagnostic.finish(fileName, err)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to save ingest file", attr.Error(err))
			metrics.RecordIngestTask(metrics.IngestCollectorManual, fileFormatFromFileType(fileType), metrics.IngestTaskStatusFailed)
		} else {
			slog.InfoContext(ctx, "File written and validated", slog.String("temp_file_name", fileName))
		}
	}()

	pipeReader, pipeWriter := io.Pipe()
	defer pipeReader.Close()

	stopCancellation := context.AfterFunc(ctx, func() {
		_ = pipeReader.CloseWithError(ctx.Err())
		_ = pipeWriter.CloseWithError(ctx.Err())
		_ = fileData.Close()
	})
	defer stopCancellation()

	validationResults := make(chan ingestValidationResult, 1)
	go func() {
		var result ingestValidationResult

		switch fileType {
		case model.FileTypeJson:
			result.report, result.err = WriteAndValidateJSON(fileData, pipeWriter, ingestSchema)
		case model.FileTypeZip:
			result.err = WriteAndValidateZip(fileData, pipeWriter)
		}

		_ = pipeWriter.CloseWithError(result.err)
		validationResults <- result
	}()

	storageDiagnostic = startIngestStorageWriteDiagnostic(ctx, ingestFileTempPrefix(jobID))
	fileName, writeErr := fileService.WriteTempFile(ctx, ingestFileTempPrefix(jobID), pipeReader, storage.WriteOptions{})
	storageDiagnostic.finish(fileName, writeErr)
	_ = pipeReader.CloseWithError(writeErr)
	if writeErr != nil {
		_ = fileData.Close()
	}

	validationResult := <-validationResults
	report = validationResult.report
	switch {
	case ctx.Err() != nil:
		err = ctx.Err()
		report = payload.ValidationReport{}
	case writeErr != nil && !errors.Is(writeErr, validationResult.err):
		// A backend failure can interrupt validation through the pipe.
		err = writeErr
		report = payload.ValidationReport{}
	default:
		err = validationResult.err
	}

	if err != nil {
		cleanupTempFile(ctx, fileService, fileName)
		return IngestTaskParams{}, report, err
	}

	return IngestTaskParams{
		Filename: fileName,
		FileType: fileType,
	}, report, nil
}

func cleanupTempFile(ctx context.Context, fileService storage.FileService, fileName string) {
	if fileName == "" {
		return
	}

	cleanupContext, cancel := context.WithTimeout(context.WithoutCancel(ctx), 30*time.Second)
	defer cancel()

	if err := fileService.DeleteFile(cleanupContext, fileName); err != nil {
		slog.ErrorContext(
			cleanupContext,
			"Failed to delete ingest file",
			slog.String("temp_file_name", fileName),
			attr.Error(err),
		)
	}
}

func WriteAndValidateZip(fileData io.Reader, destination io.Writer) error {
	teeReader := io.TeeReader(fileData, destination)
	return ValidateZipFile(teeReader)
}

func WriteAndValidateJSON(fileData io.Reader, destination io.Writer, ingestSchema payload.Schema) (payload.ValidationReport, error) {
	var (
		report           payload.ValidationReport
		normalizedReader io.Reader
		teeReader        io.Reader
		ingestValidator  payload.Validator
		err              error
	)

	normalizedReader, err = bomenc.NormalizeToUTF8(fileData)
	if err != nil {
		return report, fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	}

	teeReader = io.TeeReader(normalizedReader, destination)
	ingestValidator = payload.NewValidator(teeReader, ingestSchema)
	_, report, err = ingestValidator.ParseAndValidate()
	if err != nil {
		return report, fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	}

	return report, nil
}

func ingestFileTempPrefix(jobID int64) string {
	return fmt.Sprintf("file_upload_job%d_", jobID)
}
