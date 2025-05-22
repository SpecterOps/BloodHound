// Copyright 2025 Specter Ops, Inc.
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

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/mock.go -package=mocks . IngestData
package ingest

import (
	"archive/zip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"

	"github.com/specterops/bloodhound/bomenc"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/util"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/model/appcfg"
)

const (
	IngestCountThreshold = 500
	ReconcileProperty    = "reconcile"
)

// processIngestFile reads the files at the path supplied, and returns the total number of files in the
// archive, the number of files that failed to ingest as JSON, and an error
func (s *IngestService) processIngestFile(ctx context.Context, path string, fileType model.FileType) (int, int, error) {

	adcsEnabled := appcfg.GetTieringEnabled(ctx, s.db)
	if paths, failed, err := s.preProcessIngestFile(path, fileType); err != nil {
		return 0, failed, err
	} else {
		//kpom-todo: Why do we reset to zero here? This feels like a bug
		failed = 0

		return len(paths), failed, s.graphdb.BatchOperation(ctx, func(batch graph.Batch) error {
			for _, filePath := range paths {
				file, err := os.Open(filePath)
				if err != nil {
					failed++
					return err
				} else if err := readFileForIngest(batch, file, s.schema, adcsEnabled); err != nil {
					failed++
					slog.ErrorContext(ctx, fmt.Sprintf("Error reading ingest file %s: %v", filePath, err))
				}

				if err := file.Close(); err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error closing ingest file %s: %v", filePath, err))
				} else if err := os.Remove(filePath); errors.Is(err, fs.ErrNotExist) {
					slog.WarnContext(ctx, fmt.Sprintf("Removing ingest file %s: %v", filePath, err))
				} else if err != nil {
					slog.ErrorContext(ctx, fmt.Sprintf("Error removing ingest file %s: %v", filePath, err))
				}
			}

			return nil
		})
	}
}

func readFileForIngest(batch graph.Batch, reader io.ReadSeeker, ingestSchema IngestSchema, adcsEnabled bool) error {
	if meta, err := ValidateMetaTag(reader, ingestSchema, false); err != nil {
		return fmt.Errorf("error validating meta tag: %w", err)
	} else {
		return IngestWrapper(batch, reader, meta, adcsEnabled)
	}
}

// preProcessIngestFile will take a path and extract zips if necessary, returning the paths for files to process
// along with any errors and the number of failed files (in the case of a zip archive)
func (s *IngestService) preProcessIngestFile(path string, fileType model.FileType) ([]string, int, error) {
	if fileType == model.FileTypeJson {
		//If this isn't a zip file, just return a slice with the path in it and let stuff process as normal
		return []string{path}, 0, nil
	} else if archive, err := zip.OpenReader(path); err != nil {
		return []string{}, 0, err
	} else {
		var (
			errs      = util.NewErrorCollector()
			failed    = 0
			filePaths = make([]string, len(archive.File))
		)

		for i, f := range archive.File {
			//skip directories
			if f.FileInfo().IsDir() {
				continue
			}
			// Break out if temp file creation fails
			// Collect errors for other failures within the archive
			if tempFile, err := os.CreateTemp(s.cfg.TempDirectory(), "bh"); err != nil {
				return []string{}, 0, err
			} else if srcFile, err := f.Open(); err != nil {
				errs.Add(fmt.Errorf("error opening file %s in archive %s: %v", f.Name, path, err))
				failed++
			} else if normFile, err := bomenc.NormalizeToUTF8(srcFile); err != nil {
				errs.Add(fmt.Errorf("error normalizing file %s to UTF8 in archive %s: %v", f.Name, path, err))
				failed++
			} else if _, err := io.Copy(tempFile, normFile); err != nil {
				errs.Add(fmt.Errorf("error extracting file %s in archive %s: %v", f.Name, path, err))
				failed++
			} else if err := tempFile.Close(); err != nil {
				errs.Add(fmt.Errorf("error closing temp file %s: %v", f.Name, err))
				failed++
			} else {
				filePaths[i] = tempFile.Name()
			}
		}

		//Close the archive and delete it
		if err := archive.Close(); err != nil {
			slog.ErrorContext(s.ctx, fmt.Sprintf("Error closing archive %s: %v", path, err))
		} else if err := os.Remove(path); err != nil {
			slog.ErrorContext(s.ctx, fmt.Sprintf("Error deleting archive %s: %v", path, err))
		}

		return filePaths, failed, errs.Combined()
	}
}
