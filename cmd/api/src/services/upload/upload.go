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
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/bloodhound/packages/go/metrics"
	"github.com/specterops/chow/pkg/payload"
)

var ErrInvalidJSON = errors.New("file is not valid json")

func SaveIngestFile(location string, request *http.Request, ingestSchema payload.Schema, jobID int64) (IngestTaskParams, payload.ValidationReport, error) {
	var (
		fileData     = request.Body
		fileType     model.FileType
		tempFileName string
		report       payload.ValidationReport
		err          error
	)

	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		fileType = model.FileTypeJson
		if tempFileName, report, err = WriteAndValidateJSON(fileData, location, jobID, ingestSchema); err != nil {
			metrics.RecordIngestTask(metrics.IngestCollectorManual, fileFormatFromFileType(fileType), metrics.IngestTaskStatusFailed)
			return IngestTaskParams{}, report, err
		}
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
		fileType = model.FileTypeZip
		if tempFileName, err = WriteAndValidateZip(fileData, location, jobID); err != nil {
			metrics.RecordIngestTask(metrics.IngestCollectorManual, fileFormatFromFileType(fileType), metrics.IngestTaskStatusFailed)
			return IngestTaskParams{}, report, err
		}
	default:
		return IngestTaskParams{}, report, fmt.Errorf("invalid content type for ingest file")
	}

	return IngestTaskParams{
		Filename: tempFileName,
		FileType: fileType,
	}, report, nil
}

func WriteAndValidateZip(fileData io.Reader, location string, jobID int64) (string, error) {
	tempFile, err := os.CreateTemp(location, fmt.Sprintf("file_upload_job%d_", jobID))
	if err != nil {
		return "", fmt.Errorf("error creating ingest file: %w", err)
	}

	var (
		tempFileName  = tempFile.Name()
		teeReader     = io.TeeReader(fileData, tempFile)
		validationErr = ValidateZipFile(teeReader)
	)

	if err := tempFile.Close(); err != nil {
		slog.Error(
			"Error closing temp file",
			slog.String("file", tempFileName),
			attr.Error(err),
		)
	}

	if validationErr != nil {
		if removeErr := os.Remove(tempFileName); removeErr != nil {
			slog.Error(
				"Error deleting temp file",
				slog.String("file", tempFileName),
				attr.Error(removeErr),
			)
		}
		return "", validationErr
	}

	return tempFileName, nil
}

func WriteAndValidateJSON(fileData io.Reader, location string, jobID int64, ingestSchema payload.Schema) (string, payload.ValidationReport, error) {
	tempFile, err := os.CreateTemp(location, fmt.Sprintf("file_upload_job%d_", jobID))
	if err != nil {
		return "", payload.ValidationReport{}, fmt.Errorf("error creating ingest file: %w", err)
	}

	var (
		tempFileName     = tempFile.Name()
		report           payload.ValidationReport
		validationErr    error
		normalizedReader io.Reader
	)

	normalizedReader, normErr := bomenc.NormalizeToUTF8(fileData)
	if normErr != nil {
		validationErr = fmt.Errorf("%w: %w", ErrInvalidJSON, normErr)
	} else {
		var (
			teeReader       = io.TeeReader(normalizedReader, tempFile)
			ingestValidator = payload.NewValidator(teeReader, ingestSchema)
		)
		if _, report, validationErr = ingestValidator.ParseAndValidate(); validationErr != nil {
			validationErr = fmt.Errorf("%w: %w", ErrInvalidJSON, validationErr)
		}
	}

	// We close the file next, not last. We can't defer this if we might want to delete it.
	// Note: fileData does not need to be closed because the HTTP server manages it's lifecyle
	if closeErr := tempFile.Close(); closeErr != nil {
		slog.Error(
			"Error closing temp file",
			slog.String("file", tempFileName),
			attr.Error(closeErr),
		)
	}

	// If the validation was not successful, after we close the file, we remove it and return the error
	if validationErr != nil {
		if removeErr := os.Remove(tempFileName); removeErr != nil {
			slog.Error(
				"Error deleting temp file",
				slog.String("file", tempFileName),
				attr.Error(removeErr),
			)
		}
		return "", report, validationErr
	}

	return tempFileName, report, nil
}

// FileValidator defines the interface for ingest file validation.
// It receives a source reader (src) and a destination writer (dst).
// Implementations are responsible for validating the input stream,
// while simultaneously copying it to the destination for persistence.
// This abstraction supports format-agnostic payloads (e.g., JSON, ZIP)
type FileValidator func(src io.Reader, dst io.Writer) (ingest.OriginalMetadata, error)

func ingestFileTempPrefix(jobID int64) string {
	return fmt.Sprintf("file_upload_job%d_", jobID)
}

func WriteAndValidateFile(fileData io.Reader, location string, jobID int64, validationFunc FileValidator) (string, error) {
	return WriteAndValidateFileWithPrefix(fileData, location, ingestFileTempPrefix(jobID), validationFunc)
}

func WriteAndValidateFileWithPrefix(fileData io.Reader, location string, tempFilePrefix string, validationFunc FileValidator) (string, error) {
	if validationFunc == nil {
		return "", fmt.Errorf("validation function is required")
	}

	var (
		tempFile      *os.File
		tempFileName  string
		err           error
		validationErr error
	)

	// Write a temp file. If it passes validation, keep it around and return the filename. Otherwise destroy it.
	if tempFile, err = os.CreateTemp(location, tempFilePrefix); err != nil {
		return "", fmt.Errorf("error creating ingest file: %w", err)
	}

	// Save this for later
	tempFileName = tempFile.Name()

	// Run validation on the file to see if we even want to keep it
	_, validationErr = validationFunc(fileData, tempFile)

	// We close the file next, not last. We can't defer this if we might want to delete it.
	// Note: fileData does not need to be closed because the HTTP server manages it's lifecyle
	if closeErr := tempFile.Close(); closeErr != nil {
		slog.Error(
			"Error closing temp file",
			slog.String("file", tempFileName),
			attr.Error(closeErr),
		)
	}

	// If the validation was not successful, after we close the file, we remove it and return the error
	if validationErr != nil {
		if removeErr := os.Remove(tempFileName); removeErr != nil {
			slog.Error(
				"Error deleting temp file",
				slog.String("file", tempFileName),
				attr.Error(removeErr),
			)
		}
		return "", validationErr
	}

	// Assuming no other errors, return the name of the closed temp file
	return tempFileName, nil
}
