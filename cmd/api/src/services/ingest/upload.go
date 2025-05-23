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

package ingest

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
	tempFile, err := os.CreateTemp(location, "bh")
	if err != nil {
		return IngestTaskParams{Filename: "", FileType: model.FileTypeJson}, fmt.Errorf("error creating ingest file: %w", err)
	}

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

	if metadata, err := writeAndValidateFile(fileData, tempFile, validationFn); err != nil {
		return IngestTaskParams{}, err
	} else {
		isGeneric := false
		if metadata.Type == ingest.DataTypeGeneric {
			isGeneric = true
		}
		return IngestTaskParams{
			Filename:  tempFile.Name(),
			FileType:  fileType,
			IsGeneric: isGeneric,
		}, nil
	}
}

func writeAndValidateFile(fileData io.ReadCloser, tempFile *os.File, validationFunc FileValidator) (ingest.Metadata, error) {
	if meta, err := validationFunc(fileData, tempFile); err != nil {
		if err := tempFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("Error closing temp file %s with failed validation: %v", tempFile.Name(), err))
		} else if err := os.Remove(tempFile.Name()); err != nil {
			slog.Error(fmt.Sprintf("Error deleting temp file %s: %v", tempFile.Name(), err))
		}
		return meta, err
	} else {
		if err := tempFile.Close(); err != nil {
			slog.Error(fmt.Sprintf("Error closing temp file with successful validation %s: %v", tempFile.Name(), err))
		}
		return meta, nil
	}
}
