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
	"path"
	"strings"
	"sync"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/services/storage"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/headers"
)

type IngestControl struct {
	retainedFileLock    *sync.RWMutex
	parameterService    appcfg.ParameterService
	retainedFileService storage.FileService
}

func NewIngestControlTool(parameterService appcfg.ParameterService, retainedFileService storage.FileService) IngestControl {
	return IngestControl{
		retainedFileLock:    &sync.RWMutex{},
		parameterService:    parameterService,
		retainedFileService: retainedFileService,
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

func retainedArchiveName(filePath string) string {
	filePath = path.Clean(strings.TrimPrefix(filePath, "/"))
	if filePath == "." || filePath == ".." || strings.HasPrefix(filePath, "../") {
		return path.Base(filePath)
	}
	return filePath
}

func (s *IngestControl) writeRetainedFileToTar(ctx context.Context, tarWriter *tar.Writer, filePath string) error {
	reader, fileInfo, err := s.retainedFileService.GetFile(ctx, filePath)
	if err != nil {
		return fmt.Errorf("opening retained file: %w", err)
	}
	defer reader.Close()

	tarHeader := &tar.Header{
		Name:    retainedArchiveName(fileInfo.Path),
		Size:    fileInfo.Size,
		Mode:    0o600,
		ModTime: fileInfo.LastModified,
	}

	if err := tarWriter.WriteHeader(tarHeader); err != nil {
		return fmt.Errorf("writing tar header: %w", err)
	}

	if _, err := io.Copy(tarWriter, reader); err != nil {
		return fmt.Errorf("copying retained file: %w", err)
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

	files, err := s.retainedFileService.ListFiles(request.Context(), "", storage.ListOptions{Recursive: true})
	if err != nil {
		slog.WarnContext(request.Context(), "Failed listing retained files", attr.Error(err))
		response.WriteHeader(http.StatusInternalServerError)
		return
	}

	response.Header().Set(headers.ContentType.String(), "application/gzip")
	response.Header().Set(headers.ContentEncoding.String(), "gzip")
	response.Header().Set(headers.ContentDisposition.String(), fmt.Sprintf(`attachment; filename="retained-ingest-%s.tar.gz"`, time.Now().UTC().Format("20060102T150405Z")))
	response.WriteHeader(http.StatusOK)

	gzipWriter := gzip.NewWriter(response)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	for _, file := range files {
		if file.IsDir {
			continue
		}

		if err := s.writeRetainedFileToTar(request.Context(), tarWriter, file.Path); err != nil {
			slog.WarnContext(request.Context(), "Failed writing retained file to response", slog.String("path", file.Path), attr.Error(err))
			break
		}
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
		s.retainedFileLock.Unlock()
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusAccepted)

	// Clear all retained files in a goroutine that releases the lock once finished. This is best effort as
	// it is still possible for a file to make it into the retained directory even after the parameter is set.
	cleanupCtx := context.WithoutCancel(request.Context())

	go func(ctx context.Context) {
		defer s.retainedFileLock.Unlock()

		files, err := s.retainedFileService.ListFiles(ctx, "", storage.ListOptions{Recursive: true})
		if err != nil {
			slog.WarnContext(ctx, "Failed listing retained files", attr.Error(err))
		}

		for _, file := range files {
			if file.IsDir {
				continue
			}

			if err := s.retainedFileService.DeleteFile(ctx, file.Path); err != nil {
				slog.WarnContext(ctx, "Failed removing retained file", slog.String("path", file.Path), attr.Error(err))
			}
		}
	}(cleanupCtx)
}
