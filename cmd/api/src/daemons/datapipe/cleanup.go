// Copyright 2024 Specter Ops, Inc.
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

package datapipe

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/cleanup.go -package=mocks . FileOperations

import (
	"context"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/storage"
)

const (
	orphanMinimumAge  = 24 * time.Hour
	scratchMinimumAge = 24 * time.Hour
)

// FileOperations is an interface for describing filesystem actions. This implementation exists due to deficiencies in
// the fs package where the fs.FS interface does not support destructive operations like RemoveAll.
type FileOperations interface {
	ReadDir(path string) ([]os.DirEntry, error)
	RemoveAll(path string) error
}

// osFileOperations is a passthrough implementation of the FileOperations interface that uses the os package
// functions.
type osFileOperations struct{}

func NewOSFileOperations() FileOperations {
	return osFileOperations{}
}

func (s osFileOperations) ReadDir(path string) ([]os.DirEntry, error) {
	return os.ReadDir(path)
}

func (s osFileOperations) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// OrphanFileSweeper is a file cleanup utility that allows only one goroutine to attempt file cleanup at a time.
type OrphanFileSweeper struct {
	lock                     *sync.Mutex
	fileOps                  FileOperations
	tempDirectoryRootPath    string
	scratchDirectoryRootPath string
	excludedStoragePrefixes  []string
}

func normalizeStoragePrefixes(prefixes []string) []string {
	var normalizedPrefixes []string

	for _, prefix := range prefixes {
		prefix = strings.TrimSpace(filepath.ToSlash(prefix))
		prefix = strings.Trim(prefix, "/")
		if prefix == "" {
			continue
		}

		prefix = path.Clean(prefix)
		if prefix == "." || prefix == ".." || strings.HasPrefix(prefix, "../") {
			continue
		}

		normalizedPrefixes = append(normalizedPrefixes, prefix)
	}

	return normalizedPrefixes
}

func normalizeStoragePath(logicalPath string) string {
	return path.Clean(strings.TrimLeft(filepath.ToSlash(logicalPath), "/"))
}

func NewOrphanFileSweeper(fileOps FileOperations, tempDirectoryRootPath string, scratchDirectoryRootPath string, excludedStoragePrefixes ...string) *OrphanFileSweeper {
	return &OrphanFileSweeper{
		lock:                     &sync.Mutex{},
		fileOps:                  fileOps,
		tempDirectoryRootPath:    strings.TrimSuffix(tempDirectoryRootPath, string(filepath.Separator)),
		scratchDirectoryRootPath: strings.TrimSuffix(scratchDirectoryRootPath, string(filepath.Separator)),
		excludedStoragePrefixes:  normalizeStoragePrefixes(excludedStoragePrefixes),
	}
}

func ClearLocalIngestScratch(ctx context.Context, scratchDirectory string, minimumAge time.Duration) {
	entries, err := os.ReadDir(scratchDirectory)
	if err != nil {
		slog.ErrorContext(
			ctx,
			"Error reading scratch directory",
			slog.String("scratch_directory", scratchDirectory),
			attr.Error(err),
		)
		return
	}

	cutoff := time.Now().Add(-minimumAge)

	for _, entry := range entries {
		if ctx.Err() != nil {
			return
		}

		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(cutoff) {
			continue
		}

		_ = os.Remove(filepath.Join(scratchDirectory, entry.Name()))
	}
}

func (s *OrphanFileSweeper) isExcludedStoragePath(logicalPath string) bool {
	logicalPath = normalizeStoragePath(logicalPath)

	for _, excludedPrefix := range s.excludedStoragePrefixes {
		if logicalPath == excludedPrefix || strings.HasPrefix(logicalPath, excludedPrefix+"/") {
			return true
		}
	}

	return false
}

// addExpectedLocalPath records an expected file path for the legacy local temp-directory cleanup pass.
// Expected names may be legacy absolute filesystem paths or file-service logical paths. Logical paths are
// normalized, resolved under tempDirectoryRootPath, and ignored if they are empty or would escape the temp root.
func (s *OrphanFileSweeper) addExpectedLocalPath(expectedLocalFiles map[string]struct{}, expectedFileName string) {
	expectedFileName = strings.TrimSpace(expectedFileName)
	if expectedFileName == "" {
		return
	}

	if filepath.IsAbs(expectedFileName) {
		expectedLocalFiles[filepath.Clean(expectedFileName)] = struct{}{}
		return
	}

	logicalPath := path.Clean(filepath.ToSlash(expectedFileName))
	if logicalPath == "." || logicalPath == ".." || strings.HasPrefix(logicalPath, "../") {
		return
	}

	tempDirectoryRootPath := filepath.Clean(s.tempDirectoryRootPath)
	localPath := filepath.Clean(filepath.Join(tempDirectoryRootPath, filepath.FromSlash(logicalPath)))

	relativePath, err := filepath.Rel(tempDirectoryRootPath, localPath)
	if err != nil || relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) || filepath.IsAbs(relativePath) {
		return
	}

	expectedLocalFiles[localPath] = struct{}{}
}

// addExpectedStoragePath handles saved file paths that were saved in absolute prior to this release
func (s *OrphanFileSweeper) addExpectedStoragePath(expectedFiles map[string]struct{}, expectedFileName string) {
	expectedFileName = strings.TrimSpace(expectedFileName)
	if expectedFileName == "" {
		return
	}

	if filepath.IsAbs(expectedFileName) {
		relativePath, err := filepath.Rel(s.tempDirectoryRootPath, filepath.Clean(expectedFileName))
		if err != nil {
			return
		}

		if relativePath == "." || relativePath == ".." || strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) || filepath.IsAbs(relativePath) {
			return
		}

		expectedFiles[path.Clean(filepath.ToSlash(relativePath))] = struct{}{}
		return
	}

	logicalPath := normalizeStoragePath(expectedFileName)
	if logicalPath == "." || logicalPath == ".." || strings.HasPrefix(logicalPath, "../") {
		return
	}

	expectedFiles[logicalPath] = struct{}{}
}

func (s *OrphanFileSweeper) clearStoredIngestFiles(ctx context.Context, ingestFileService storage.FileService, expectedFileNames []string) {
	expectedFiles := map[string]struct{}{}

	for _, expectedFileName := range expectedFileNames {
		s.addExpectedStoragePath(expectedFiles, expectedFileName)
	}

	files, err := ingestFileService.ListFiles(ctx, "", storage.ListOptions{Recursive: true})
	if err != nil {
		slog.ErrorContext(ctx, "Failed listing ingest files", attr.Error(err))
		return
	}

	for _, file := range files {
		if ctx.Err() != nil {
			return
		}
		if file.IsDir {
			continue
		}

		logicalPath := normalizeStoragePath(file.Path)
		if s.isExcludedStoragePath(logicalPath) {
			continue
		}

		if _, ok := expectedFiles[logicalPath]; ok {
			continue
		}

		if !file.LastModified.IsZero() && time.Since(file.LastModified) < orphanMinimumAge {
			continue
		}

		if err := ingestFileService.DeleteFile(ctx, logicalPath); err != nil {
			slog.WarnContext(
				ctx,
				"Failed deleting orphaned ingest file",
				slog.String("path", logicalPath),
				attr.Error(err),
			)
		}
	}
}

func isExpectedLocalEntry(expectedLocalFiles map[string]struct{}, localPath string, isDirectory bool) bool {
	localPath = filepath.Clean(localPath)
	if _, ok := expectedLocalFiles[localPath]; ok {
		return true
	}

	if !isDirectory {
		return false
	}

	for expectedLocalFile := range expectedLocalFiles {
		relativePath, err := filepath.Rel(localPath, expectedLocalFile)
		if err != nil {
			continue
		}

		if relativePath != "." && relativePath != ".." && !strings.HasPrefix(relativePath, ".."+string(filepath.Separator)) && !filepath.IsAbs(relativePath) {
			return true
		}
	}

	return false
}

func (s *OrphanFileSweeper) clearLegacyLocalIngestFiles(ctx context.Context, expectedFileNames []string) {
	expectedLocalFiles := map[string]struct{}{}

	for _, expectedFileName := range expectedFileNames {
		s.addExpectedLocalPath(expectedLocalFiles, expectedFileName)
	}

	entries, err := s.fileOps.ReadDir(s.tempDirectoryRootPath)
	if err != nil {
		slog.WarnContext(ctx, "Failed reading legacy ingest temp directory", attr.Error(err))
		return
	}

	cutoff := time.Now().Add(-orphanMinimumAge)

	for _, entry := range entries {
		if ctx.Err() != nil {
			return
		}

		fullPath := filepath.Join(s.tempDirectoryRootPath, entry.Name())
		if isExpectedLocalEntry(expectedLocalFiles, fullPath, entry.IsDir()) {
			continue
		}

		if entry.IsDir() {
			_ = s.fileOps.RemoveAll(fullPath)
			continue
		}

		info, err := entry.Info()
		if err != nil || info.ModTime().After(cutoff) {
			continue
		}

		_ = s.fileOps.RemoveAll(fullPath)
	}
}

// Clear takes a context and a list of expected file names. The function will list all directory entries within the
// configured tempDirectoryRootPath and scratchDirectoryRootPath that the sweeper was instantiated with and compare the
// resulting list against the passed expected file names. The function will then call RemoveAll on each directory entry not
// found in the expected file name slice. In addition, ingested files are also deleted using the supplied file service.
func (s *OrphanFileSweeper) Clear(ctx context.Context, ingestFileService storage.FileService, expectedFileNames []string) {
	if !s.lock.TryLock() {
		return
	}
	defer s.lock.Unlock()

	s.clearStoredIngestFiles(ctx, ingestFileService, expectedFileNames)
	s.clearLegacyLocalIngestFiles(ctx, expectedFileNames)
	ClearLocalIngestScratch(ctx, s.scratchDirectoryRootPath, scratchMinimumAge)
}
