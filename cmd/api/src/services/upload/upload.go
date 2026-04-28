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
	"github.com/specterops/chow/pkg/validator"
)

var ErrInvalidJSON = errors.New("file is not valid json")

func SaveIngestFile(location string, request *http.Request, ingestSchema validator.IngestSchema, jobID int64) (IngestTaskParams, validator.ValidationReport, error) {
	var (
		fileData     = request.Body
		fileType     model.FileType
		tempFileName string
		report       validator.ValidationReport
		err          error
	)

	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		fileType = model.FileTypeJson
		if tempFileName, report, err = WriteAndValidateJSON(fileData, location, jobID, ingestSchema); err != nil {
			return IngestTaskParams{}, report, err
		}
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
		fileType = model.FileTypeZip
		if tempFileName, err = WriteAndValidateZip(fileData, location, jobID); err != nil {
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

func WriteAndValidateJSON(fileData io.Reader, location string, jobID int64, ingestSchema validator.IngestSchema) (string, validator.ValidationReport, error) {
	tempFile, err := os.CreateTemp(location, fmt.Sprintf("file_upload_job%d_", jobID))
	if err != nil {
		return "", validator.ValidationReport{}, fmt.Errorf("error creating ingest file: %w", err)
	}

	var (
		tempFileName     = tempFile.Name()
		report           validator.ValidationReport
		validationErr    error
		normalizedReader io.Reader
	)

	normalizedReader, normErr := bomenc.NormalizeToUTF8(fileData)
	if normErr != nil {
		validationErr = fmt.Errorf("%w: %w", ErrInvalidJSON, normErr)
	} else {
		var (
			teeReader       = io.TeeReader(normalizedReader, tempFile)
			ingestValidator = validator.NewValidator(teeReader, ingestSchema)
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
