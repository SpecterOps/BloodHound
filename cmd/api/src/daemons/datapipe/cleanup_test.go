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

package datapipe_test

import (
	"context"
	"github.com/specterops/bloodhound/src/daemons/datapipe"
	"github.com/specterops/bloodhound/src/daemons/datapipe/mocks"
	"go.uber.org/mock/gomock"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
)

type dirEntry struct {
	name    string
	isDir   bool
	mode    fs.FileMode
	info    fs.FileInfo
	infoErr error
}

func (s dirEntry) Name() string {
	return s.name
}

func (s dirEntry) IsDir() bool {
	return s.isDir
}

func (s dirEntry) Type() fs.FileMode {
	return s.mode
}

func (s dirEntry) Info() (fs.FileInfo, error) {
	return s.info, s.infoErr
}

func TestOrphanFileSweeper_Clear(t *testing.T) {
	const workDir = "/fake/work/dir"

	t.Run("Allow Only One Goroutine", func(t *testing.T) {
		var (
			mockCtrl       = gomock.NewController(t)
			mockFileOps    = mocks.NewMockFileOperations(mockCtrl)
			sweeper        = datapipe.NewOrphanFileSweeper(mockFileOps, workDir)
			wgCoordination = &sync.WaitGroup{}
			wgReadDir      = &sync.WaitGroup{}
		)

		defer mockCtrl.Finish()

		// Prep the wait groups for coordination
		wgCoordination.Add(1)
		wgReadDir.Add(1)

		mockFileOps.EXPECT().ReadDir(workDir).DoAndReturn(func(path string) ([]os.DirEntry, error) {
			// Release the coordination wait group
			wgCoordination.Done()

			// Block on the readDir wait group
			wgReadDir.Wait()

			return nil, nil
		})

		// Launch the clear function in a goroutine. The wait group will cause this call to block
		go sweeper.Clear(context.Background(), []string{})

		// Wait for the go routine to reach the ReadDir function
		wgCoordination.Wait()

		// Run the clear function in the current thread context as this should exit without blocking
		sweeper.Clear(context.Background(), []string{})

		// Release the wait group to complete the test
		wgReadDir.Done()
	})

	t.Run("Clear Orphan Files", func(t *testing.T) {
		var (
			mockCtrl    = gomock.NewController(t)
			mockFileOps = mocks.NewMockFileOperations(mockCtrl)
			sweeper     = datapipe.NewOrphanFileSweeper(mockFileOps, workDir)
		)

		defer mockCtrl.Finish()

		mockFileOps.EXPECT().ReadDir(workDir).Return([]os.DirEntry{
			dirEntry{
				name: "1",
			},
		}, nil)

		mockFileOps.EXPECT().RemoveAll(filepath.Join(workDir, "1")).Return(nil)

		sweeper.Clear(context.Background(), []string{})
	})

	t.Run("Exclude Expected Files", func(t *testing.T) {
		var (
			mockCtrl    = gomock.NewController(t)
			mockFileOps = mocks.NewMockFileOperations(mockCtrl)
			sweeper     = datapipe.NewOrphanFileSweeper(mockFileOps, workDir)
		)

		defer mockCtrl.Finish()

		mockFileOps.EXPECT().ReadDir(workDir).Return([]os.DirEntry{
			dirEntry{
				name: "1",
			},
		}, nil)

		// This one is a negative assertion. Because we're passing in "1" to be excluded the listed dirEntries will be
		// empty and therefore the sweeper MUST NOT call RemoveAll() on the FileOperations mock. If RemoveAll() is
		// called then this test MUST fail.
		sweeper.Clear(context.Background(), []string{"1"})
	})

	t.Run("Exit on Context Cancellation", func(t *testing.T) {
		var (
			mockCtrl    = gomock.NewController(t)
			mockFileOps = mocks.NewMockFileOperations(mockCtrl)
			sweeper     = datapipe.NewOrphanFileSweeper(mockFileOps, workDir)
		)

		defer mockCtrl.Finish()

		mockFileOps.EXPECT().ReadDir(workDir).Return([]os.DirEntry{
			dirEntry{
				name: "1",
			},
		}, nil)

		// Create a cancellable context and cancel it right away
		ctx, done := context.WithCancel(context.Background())
		done()

		// When passed in with the cancelled context the sweeper should not call os.RemoveAll("1")
		sweeper.Clear(ctx, []string{})
	})
}
