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

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
)

type localPendingFile struct {
	store       *LocalStore
	ctx         context.Context
	name        string
	tempName    string
	tempFile    *os.File
	failIfExist bool
	committed   bool
	aborted     bool
	terminalErr error
}

func (s *LocalStore) BeginWrite(ctx context.Context, name string, options WriteOptions) (PendingFile, error) {
	var (
		directory  = path.Dir(name)
		identifier string
		tempName   string
		tempFile   *os.File
		err        error
	)

	if err = ctx.Err(); err != nil {
		return nil, err
	}

	if directory != "." {
		if err = s.root.MkdirAll(directory, 0o750); err != nil {
			return nil, err
		}
	}

	if identifier, err = randomID(); err != nil {
		return nil, err
	}

	tempName = path.Join(directory, ".tmp-"+identifier)
	if tempFile, err = s.root.OpenFile(tempName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640); err != nil {
		return nil, err
	}

	return &localPendingFile{
		store:       s,
		ctx:         ctx,
		name:        name,
		tempName:    tempName,
		tempFile:    tempFile,
		failIfExist: options.FailIfExists,
	}, nil
}

func (s *localPendingFile) Write(data []byte) (int, error) {
	var (
		bytesWritten int
		err          error
	)

	if s.aborted || s.committed {
		return 0, io.ErrClosedPipe
	}

	if s.terminalErr != nil {
		return 0, s.terminalErr
	}

	if err = s.ctx.Err(); err != nil {
		s.terminalErr = fmt.Errorf("%w: %w", ErrWriteFailed, err)
		return 0, s.terminalErr
	}

	bytesWritten, err = s.tempFile.Write(data)
	if err != nil {
		s.terminalErr = fmt.Errorf("%w: %w", ErrWriteFailed, err)
		return bytesWritten, s.terminalErr
	}

	if bytesWritten != len(data) {
		s.terminalErr = fmt.Errorf("%w: %w", ErrWriteFailed, io.ErrShortWrite)
		return bytesWritten, s.terminalErr
	}

	return bytesWritten, nil
}

func (s *localPendingFile) Commit() (string, error) {
	var (
		directory      string
		publicationErr error
		err            error
	)

	if s.aborted {
		return "", io.ErrClosedPipe
	}

	if s.committed {
		return s.name, nil
	}

	if s.terminalErr != nil {
		return "", s.terminalErr
	}

	if err = s.ctx.Err(); err != nil {
		s.terminalErr = err
		return "", err
	}

	if err = s.tempFile.Sync(); err == nil {
		err = s.tempFile.Close()
	}
	if err != nil {
		s.terminalErr = err
		return "", err
	}

	s.tempFile = nil

	directory = path.Dir(s.name)
	if s.failIfExist {
		publicationErr = s.store.root.Link(s.tempName, s.name)
	} else {
		publicationErr = s.store.root.Rename(s.tempName, s.name)
	}

	if publicationErr != nil {
		s.terminalErr = publicationErr
		return "", publicationErr
	}

	if err = syncDir(s.store.root, directory); err != nil {
		s.terminalErr = err
		return "", err
	}

	if s.failIfExist {
		_ = s.store.root.Remove(s.tempName)
	}

	s.committed = true

	return s.name, nil
}

func (s *localPendingFile) Abort() error {
	var (
		cleanupErrors []error
		err           error
	)

	if s.aborted || s.committed {
		return nil
	}

	s.aborted = true

	if s.tempFile != nil {
		if err = s.tempFile.Close(); err != nil {
			cleanupErrors = append(cleanupErrors, err)
		}
		s.tempFile = nil
	}

	if err = s.store.root.Remove(s.tempName); err != nil && !errors.Is(err, fs.ErrNotExist) {
		cleanupErrors = append(cleanupErrors, err)
	}
	return errors.Join(cleanupErrors...)
}
