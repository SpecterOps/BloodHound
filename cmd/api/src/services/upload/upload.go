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

func SaveIngestFile(ctx context.Context, fileService storage.FileService, request *http.Request, ingestSchema payload.Schema, jobID int64) (IngestTaskParams, payload.ValidationReport, error) {
	var (
		uploadDiagnostic  ingestUploadDiagnostic
		storageDiagnostic ingestStorageWriteDiagnostic
		fileType          model.FileType
		fileData          = request.Body
		pendingFile       storage.PendingFile
		fileName          string
		report            payload.ValidationReport
		writeOptions      storage.WriteOptions
		stopCancellation  func() bool
		contextErr        error
		abortErr          error
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

	if fileData == nil {
		fileData = http.NoBody
	}

	uploadDiagnostic = startIngestUploadDiagnostic(ctx, jobID, fileType)
	storageDiagnostic = startIngestStorageWriteDiagnostic(ctx, ingestFileTempPrefix(jobID))
	defer func() {
		storageDiagnostic.finish(fileName, err)
		uploadDiagnostic.finish(fileName, err)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to save ingest file", attr.Error(err))
			metrics.RecordIngestTask(metrics.IngestCollectorManual, fileFormatFromFileType(fileType), metrics.IngestTaskStatusFailed)
		} else {
			slog.InfoContext(ctx, "File written and validated", slog.String("temp_file_name", fileName))
		}
	}()

	if request.ContentLength > 0 {
		writeOptions.SizeHint = request.ContentLength
	}

	pendingFile, err = fileService.BeginTempFile(ctx, ingestFileTempPrefix(jobID), writeOptions)
	if err != nil {
		return IngestTaskParams{}, payload.ValidationReport{}, err
	}

	defer func() {
		abortErr = pendingFile.Abort()
		if abortErr != nil {
			slog.ErrorContext(ctx, "Failed to abort unpublished ingest file", slog.Any("error", abortErr))
		}
	}()

	stopCancellation = context.AfterFunc(ctx, func() {
		_ = fileData.Close()
	})
	defer stopCancellation()

	switch fileType {
	case model.FileTypeJson:
		report, err = WriteAndValidateJSON(fileData, pendingFile, ingestSchema)
	case model.FileTypeZip:
		err = WriteAndValidateZip(fileData, pendingFile)
	}

	contextErr = ctx.Err()
	if contextErr != nil {
		report = payload.ValidationReport{}
		err = contextErr
	} else if errors.Is(err, storage.ErrWriteFailed) {
		report = payload.ValidationReport{}
	}

	if err != nil {
		return IngestTaskParams{}, report, err
	}

	fileName, err = pendingFile.Commit()
	if err != nil {
		report = payload.ValidationReport{}
		if ctx.Err() != nil {
			err = ctx.Err()
		}
		return IngestTaskParams{}, report, err
	}

	return IngestTaskParams{
		Filename: fileName,
		FileType: fileType,
	}, report, nil
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
		if errors.Is(err, storage.ErrWriteFailed) {
			return report, err
		}
		return report, fmt.Errorf("%w: %w", ErrInvalidJSON, err)
	}

	return report, nil
}

func ingestFileTempPrefix(jobID int64) string {
	return fmt.Sprintf("file_upload_job%d_", jobID)
}
