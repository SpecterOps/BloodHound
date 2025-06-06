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

	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/ingest"
	"github.com/specterops/bloodhound/src/utils"
)

var ErrInvalidJSON = errors.New("file is not valid json")

func SaveIngestFile(location string, request *http.Request, validator IngestValidator) (IngestTaskParams, error) {
	fileData := request.Body

	var (
		fileType     model.FileType
		validationFn FileValidator
	)

	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		fileType = model.FileTypeJson
		validationFn = validator.WriteAndValidateJSON
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
		fileType = model.FileTypeZip
		validationFn = WriteAndValidateZip
	default:
		return IngestTaskParams{}, fmt.Errorf("invalid content type for ingest file")
	}

	if tempFileName, err := WriteAndValidateFile(fileData, location, validationFn); err != nil {
		return IngestTaskParams{}, err
	} else {
		return IngestTaskParams{
			Filename: tempFileName,
			FileType: fileType,
		}, nil
	}

}

func WriteAndValidateFile(fileData io.Reader, location string, validationFunc FileValidator) (string, error) {
	// Write a temp file. If it passes validation, keep it around and return the filename. Otherwise destroy it.
	tempFile, err := os.CreateTemp(location, "bh")
	if err != nil {
		return "", fmt.Errorf("error creating ingest file: %w", err)
	}

	// Save this for later
	tempFileName := tempFile.Name()

	// Run validation on the file to see if we even want to keep it
	_, validationErr := validationFunc(fileData, tempFile)

	// We close the file next, not last. We can't defer this if we might want to delete it.
	// Note: fileData does not need to be closed because the HTTP server manages it's lifecyle
	if closeErr := tempFile.Close(); closeErr != nil {
		slog.Error(fmt.Sprintf("Error closing temp file %s: %v", tempFileName, closeErr))
	}

	// If the validation was not successful, after we close the file, we remove it and return the error
	if validationErr != nil {
		if removeErr := os.Remove(tempFileName); removeErr != nil {
			slog.Error(fmt.Sprintf("Error deleting temp file %s: %v", tempFileName, removeErr))
		}
		return "", validationErr
	}

	// Assuming no other errors, return the name of the closed temp file
	return tempFileName, nil
}
