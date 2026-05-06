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
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

var ErrInvalidJSON = errors.New("file is not valid json")

func SaveIngestFile(ctx context.Context, fileService storage.FileService, request *http.Request, validator IngestValidator, jobID int64) (IngestTaskParams, error) {
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

	if tempFileName, err := WriteAndValidateFile(ctx, fileService, fileData, fmt.Sprintf("tmp/file_upload_job%d_", jobID), validationFn); err != nil {
		return IngestTaskParams{}, err
	} else {
		return IngestTaskParams{
			Filename: tempFileName,
			FileType: fileType,
		}, nil
	}
}

func WriteAndValidateFile(ctx context.Context, fileService storage.FileService, fileData io.Reader, prefix string, validationFunc FileValidator) (string, error) {
	// Create a pipe: pr (read end) and pw (write end).
	// Data written to pw can be read from pr.
	pr, pw := io.Pipe()

	// validationErrCh carries the result of the validation goroutine.
	// Using a buffered channel (size 1) ensures the goroutine never blocks on send,
	// and gives the main goroutine a synchronization point to wait for the result.
	validationErrCh := make(chan error, 1)

	// Start validation in a separate goroutine.
	// validationFunc reads from pr, writes to io.Discard.
	// This validates the stream without storing the data twice.
	go func() {
		_, err := validationFunc(pr, io.Discard)
		pr.Close()
		validationErrCh <- err
	}()

	// TeeReader: as we read fileData, a copy is written to pw and flows to the goroutine via pr.
	teeReader := io.TeeReader(fileData, pw)

	// Write to storage while validation happens concurrently.
	tempFileName, writeErr := fileService.WriteTempFile(ctx, prefix, teeReader, storage.WriteOptions{})

	// Closing pw signals EOF to the validation goroutine so it can finish.
	pw.Close()

	// Wait for the validation goroutine to finish and collect its result.
	validationErr := <-validationErrCh

	// Check if validation failed first — the temp file should be cleaned up.
	if validationErr != nil {
		slog.ErrorContext(ctx, "Validation failed", slog.String("tempFileName", tempFileName), attr.Error(validationErr))
		fileService.DeleteFile(ctx, tempFileName)
		return "", validationErr
	}

	// Check if writing failed — the temp file should be cleaned up.
	if writeErr != nil {
		slog.ErrorContext(ctx, "Write failed", slog.String("tempFileName", tempFileName), attr.Error(writeErr))
		fileService.DeleteFile(ctx, tempFileName)
		return "", writeErr
	}

	slog.InfoContext(ctx, "File written and validated", slog.String("tempFileName", tempFileName))
	return tempFileName, nil
}
