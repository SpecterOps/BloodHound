// Copyright 2025 Specter Ops, Inc.
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
package tools

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/packages/go/headers"
)

type IngestControl struct {
	cfg              config.Configuration
	retainedFileLock *sync.RWMutex
	parameterService appcfg.ParameterService
}

func NewIngestControlTool(cfg config.Configuration, parameterService appcfg.ParameterService) IngestControl {
	return IngestControl{
		cfg:              cfg,
		retainedFileLock: &sync.RWMutex{},
		parameterService: parameterService,
	}
}

func (s *IngestControl) setIngestFileRetention(ctx context.Context, parameter appcfg.RetainIngestedFilesParameter) error {
	if val, err := types.NewJSONBObject(parameter); err != nil {
		return fmt.Errorf("failed to convert value to JSONBObject: %w", err)
	} else {
		enabledParameter := appcfg.Parameter{
			Key:         appcfg.RetainIngestedFilesKey,
			Name:        "Hold Ingested Files",
			Description: "Boolean switch for holding previously ingested files for operator inspection.",
			Value:       val,
		}

		if err := s.parameterService.SetConfigurationParameter(ctx, enabledParameter); err != nil {
			return fmt.Errorf("failed to set parameter: %w", err)
		}
	}

	return nil
}

func (s *IngestControl) FetchRetainedIngestFiles(response http.ResponseWriter, request *http.Request) {
	if !s.retainedFileLock.TryRLock() {
		// Unable to acquire a write lock at this time - notify of a conflict. User is expected to retry
		// the request later.
		response.WriteHeader(http.StatusConflict)
		return
	}

	defer s.retainedFileLock.RUnlock()

	if !appcfg.ShouldRetainIngestedFiles(request.Context(), s.parameterService) {
		// If the setting for ingest file retention is disabled then assume there is nothing to read.
		response.WriteHeader(http.StatusNotFound)
		return
	}

	retainedFilesDirectory := s.cfg.RetainedFilesDirectory()

	if retainedFilesDirEntries, err := os.ReadDir(retainedFilesDirectory); err != nil {
		// Unable to stat the retained files directory. Log a warning and inform the user that the
		// operation failed.
		slog.WarnContext(request.Context(), fmt.Sprintf("Failed reading retained files directory %s: %v", retainedFilesDirectory, err))
		response.WriteHeader(http.StatusInternalServerError)
	} else {
		// Author the response
		response.Header().Set(headers.ContentType.String(), "application/gzip")
		response.Header().Set(headers.ContentEncoding.String(), "gzip")
		response.Header().Set(headers.ContentDisposition.String(), fmt.Sprintf(`attachment; filename="retained-ingest-%s.tar.gz"`, time.Now().UTC().Format("20060102T150405Z")))
		response.WriteHeader(http.StatusOK)

		var (
			gzipWriter = gzip.NewWriter(response)
			tarWriter  = tar.NewWriter(gzipWriter)
		)

		for _, retainedFilesDirEntry := range retainedFilesDirEntries {
			retainedFilePath := filepath.Join(retainedFilesDirectory, retainedFilesDirEntry.Name())

			if retainedFilesDirEntry.IsDir() {
				// Log a warning for directory entries and skip reading it.
				slog.WarnContext(request.Context(), fmt.Sprintf("Unexpected directory %s in retained files directory %s: %v", retainedFilePath, retainedFilesDirectory, err))
				continue
			}

			if fileInfo, err := retainedFilesDirEntry.Info(); err != nil {
				slog.WarnContext(request.Context(), fmt.Sprintf("Unable to stat file %s: %v", retainedFilePath, err))
				break
			} else {
				if tarHeader, err := tar.FileInfoHeader(fileInfo, ""); err != nil {
					slog.WarnContext(request.Context(), fmt.Sprintf("Unable to convert file info for file %s: %v", retainedFilePath, err))
					break
				} else if err := tarWriter.WriteHeader(tarHeader); err != nil {
					slog.WarnContext(request.Context(), fmt.Sprintf("Failed writing tar file header from %s to response: %v", retainedFilePath, err))
					break
				}

				if fin, err := os.Open(retainedFilePath); err != nil {
					slog.WarnContext(request.Context(), fmt.Sprintf("Unexpected directory %s in retained files directory %s: %v", retainedFilePath, retainedFilesDirectory, err))
					break
				} else {
					// Copy and inspect the error only after closing the file.
					_, copyErr := io.Copy(tarWriter, fin)
					fin.Close()

					if copyErr != nil {
						slog.WarnContext(request.Context(), fmt.Sprintf("Failed writing file content from %s to response: %v", retainedFilePath, copyErr))
						break
					}
				}
			}
		}

		// Attempt to flush and close the tar.gz writers - best effort
		tarWriter.Close()
		gzipWriter.Close()
	}
}

func (s *IngestControl) EnableIngestFileRetention(response http.ResponseWriter, request *http.Request) {
	if !s.retainedFileLock.TryLock() {
		// Unable to acquire a write lock at this time - notify of a conflict. User is expected to retry
		// the request later.
		response.WriteHeader(http.StatusConflict)
		return
	}

	defer s.retainedFileLock.Unlock()

	// Set the parameter to inform ingest to retain ingested files.
	if err := s.setIngestFileRetention(request.Context(), appcfg.RetainIngestedFilesParameter{
		Enabled: true,
	}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
	} else {
		response.WriteHeader(http.StatusOK)
	}
}

func (s *IngestControl) DisableIngestFileRetention(response http.ResponseWriter, request *http.Request) {
	if !s.retainedFileLock.TryLock() {
		// Unable to acquire a write lock at this time - notify of a conflict. User is expected to retry
		// the request later.
		response.WriteHeader(http.StatusConflict)
		return
	}

	// Set the parameter to no longer retain ingest files
	if err := s.setIngestFileRetention(request.Context(), appcfg.RetainIngestedFilesParameter{
		Enabled: false,
	}); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusAccepted)

	// Clear all retained files in a goroutine that releases the lock once finished. This is best effort as
	// it is still possible for a file to make it into the retained directory even after the parameter is set.
	go func() {
		defer s.retainedFileLock.Unlock()

		retainedFilesDirectory := s.cfg.RetainedFilesDirectory()

		if retainedFilesDirEntries, err := os.ReadDir(retainedFilesDirectory); err != nil {
			slog.WarnContext(request.Context(), fmt.Sprintf("Failed reading retained files directory %s: %v", retainedFilesDirectory, err))
		} else {
			for _, retainedFilesDirEntry := range retainedFilesDirEntries {
				retainedFilePath := filepath.Join(retainedFilesDirectory, retainedFilesDirEntry.Name())

				if err := os.RemoveAll(retainedFilePath); err != nil {
					slog.WarnContext(request.Context(), fmt.Sprintf("Failed removing retained file %s: %v", retainedFilePath, err))
				}
			}
		}
	}()
}
