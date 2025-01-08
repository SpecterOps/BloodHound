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
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/specterops/bloodhound/log"
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
	lock                  *sync.Mutex
	fileOps               FileOperations
	tempDirectoryRootPath string
}

func NewOrphanFileSweeper(fileOps FileOperations, tempDirectoryRootPath string) *OrphanFileSweeper {
	return &OrphanFileSweeper{
		lock:                  &sync.Mutex{},
		fileOps:               fileOps,
		tempDirectoryRootPath: strings.TrimSuffix(tempDirectoryRootPath, string(filepath.Separator)),
	}
}

// Clear takes a context and a list of expected file names. The function will list all directory entries within the
// configured tempDirectoryRootPath that the sweeper was instantiated with and compare the resulting list against the
// passed expected file names. The function will then call RemoveAll on each directory entry not found in the expected
// file name slice.
func (s *OrphanFileSweeper) Clear(ctx context.Context, expectedFileNames []string) {
	// Only allow one background thread to run for clearing orphaned data/
	if !s.lock.TryLock() {
		return
	}

	// Release the lock once finished
	defer s.lock.Unlock()

	log.Infof(fmt.Sprintf("Running OrphanFileSweeper for path %s", s.tempDirectoryRootPath))
	log.Debugf(fmt.Sprintf("OrphanFileSweeper expected names %v", expectedFileNames))

	if dirEntries, err := s.fileOps.ReadDir(s.tempDirectoryRootPath); err != nil {
		log.Errorf(fmt.Sprintf("Failed reading work directory %s: %v", s.tempDirectoryRootPath, err))
	} else {
		numDeleted := 0

		// Remove expected files from the deletion list
		for _, expectedFileName := range expectedFileNames {
			expectedDir, expectedFN := filepath.Split(expectedFileName)
			if expectedDir != "" {
				expectedDir = strings.TrimSuffix(expectedDir, string(filepath.Separator))
				if expectedDir != s.tempDirectoryRootPath {
					log.Warnf(fmt.Sprintf("directory '%s' for expectedFileName '%s' does not match tempDirectoryRootPath '%s': skipping", expectedDir, expectedFileName, s.tempDirectoryRootPath))
					continue
				}
			}
			for idx, dirEntry := range dirEntries {
				if expectedFN == dirEntry.Name() {
					log.Debugf(fmt.Sprintf("skipping expected file %s", expectedFN))
					dirEntries = append(dirEntries[:idx], dirEntries[idx+1:]...)
				}
			}
		}

		for _, orphanedDirEntry := range dirEntries {
			// Check for context cancellation before each file deletion
			if ctx.Err() != nil {
				break
			}

			log.Infof(fmt.Sprintf("Removing orphaned file %s", orphanedDirEntry.Name()))
			fullPath := filepath.Join(s.tempDirectoryRootPath, orphanedDirEntry.Name())

			if err := s.fileOps.RemoveAll(fullPath); err != nil {
				log.Errorf(fmt.Sprintf("Failed removing orphaned file %s: %v", fullPath, err))
			}

			numDeleted += 1
		}

		if numDeleted > 0 {
			log.Infof(fmt.Sprintf("Finished removing %d orphaned ingest files", numDeleted))
		}
	}
}
