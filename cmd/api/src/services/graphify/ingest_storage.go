// Copyright 2026 Specter Ops, Inc.
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

package graphify

import (
	"archive/zip"
	"context"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/bloodhound/packages/go/errorlist"
	"github.com/specterops/chow/pkg/payload"
)

type IngestFileData struct {
	Name         string
	ParentFile   string
	Path         string
	Errors       []string
	UserDataErrs []string
}

// ExtractIngestFiles takes a stored ingest file path and extracts ZIP archives when necessary, returning the paths for
// files to process. ZIP archives are deleted after extraction, and extracted files are written to the configured temp
// directory with tempFilePrefix.
func ExtractIngestFiles(ctx context.Context, configuration config.Configuration, storedFileName string, providedFileName string, fileType model.FileType, tempFilePrefix string) ([]IngestFileData, error) {
	if fileType == model.FileTypeJson {
		// If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []IngestFileData{
			{
				Name:   providedFileName,
				Path:   storedFileName,
				Errors: []string{},
			},
		}, nil
	} else if archive, err := zip.OpenReader(storedFileName); err != nil {
		return []IngestFileData{}, err
	} else {
		var (
			errs     = errorlist.NewBuilder()
			fileData = make([]IngestFileData, 0)
		)

		defer func() {
			if err := archive.Close(); err != nil {
				slog.ErrorContext(
					ctx,
					"Error closing archive",
					slog.String("path", storedFileName),
					attr.Error(err),
				)
			}
			if err := os.Remove(storedFileName); err != nil {
				slog.ErrorContext(
					ctx,
					"Error deleting archive",
					slog.String("path", storedFileName),
					attr.Error(err),
				)
			}
		}()

		for _, file := range archive.File {
			// skip directories
			if file.FileInfo().IsDir() {
				continue
			}

			fileName, err := extractToTempFile(configuration, file, tempFilePrefix)
			if err != nil {
				fileData = append(fileData, IngestFileData{
					Name:       file.Name,
					ParentFile: providedFileName,
					Errors:     []string{err.Error()},
				})

				errs.Add(err)
			} else {
				fileData = append(fileData, IngestFileData{
					Name:       file.Name,
					ParentFile: providedFileName,
					Path:       fileName,
				})
			}
		}

		return fileData, errs.Build()
	}
}

func extractToTempFile(configuration config.Configuration, file *zip.File, tempFilePrefix string) (string, error) {
	var (
		tempFile            *os.File
		tempFileName        string
		sourceFile          io.ReadCloser
		normalizedFile      io.Reader
		extractionSucceeded bool
		err                 error
	)

	if tempFile, err = os.CreateTemp(configuration.TempDirectory(), tempFilePrefix); err != nil {
		return "", err
	}

	tempFileName = tempFile.Name()
	defer func() {
		if !extractionSucceeded {
			tempFile.Close()
			os.Remove(tempFileName)
		}
	}()

	if sourceFile, err = file.Open(); err != nil {
		return "", err
	}
	defer sourceFile.Close()

	// this creates a normalized file to feed to the copy
	if normalizedFile, err = bomenc.NormalizeToUTF8(sourceFile); err != nil {
		return "", err
		// and this is what actually copies it to disk
	} else if _, err = io.Copy(tempFile, normalizedFile); err != nil {
		return "", err
		// try closing the temp file
	} else if err = tempFile.Close(); err != nil {
		return "", err
	} else {
		// let the deferred method above know we shouldn't delete it and return the filename
		extractionSucceeded = true
		return tempFileName, nil
	}
}

func processSingleFile(ctx context.Context, fileData IngestFileData, ingestContext *IngestContext, readOpts ReadOptions) (payload.ValidationReport, error) {
	defer measure.ContextLogAndMeasureWithThreshold(ctx, slog.LevelDebug, "processing single file for ingest", slog.String("filepath", fileData.Path))()

	file, err := os.Open(fileData.Path)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Error opening ingest file",
			slog.String("path", fileData.Path),
			attr.Error(err),
		)
		return payload.ValidationReport{}, err
	}

	defer func() {
		file.Close()

		// Always remove the file after attempting to ingest it. Even if it failed
		if err := os.Remove(fileData.Path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			slog.ErrorContext(
				ctx,
				"Error removing ingest file",
				slog.String("path", fileData.Path),
				attr.Error(err),
			)
		}
	}()

	report, err := ReadFileForIngest(ingestContext, file, readOpts)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Error reading ingest file",
			slog.String("storage_path", fileData.Path),
			attr.Error(err),
		)
		return report, err
	}

	return report, nil
}
