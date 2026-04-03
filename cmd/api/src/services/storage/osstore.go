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
	"io"
	"os"
	"path/filepath"
)

// OsStore is a Storage implementation that operates directly on the OS filesystem
// using absolute paths. It is intended for use in CE deployments where all paths
// are already fully-qualified (e.g. temp directory paths from config).
type OsStore struct{}

func NewOsStore() *OsStore {
	return &OsStore{}
}

func (s *OsStore) Put(ctx context.Context, path string, reader io.Reader, options WriteOptions) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	cleanup := func() {
		tmp.Close()
		os.Remove(tmpName)
	}

	if _, err := io.Copy(tmp, reader); err != nil {
		cleanup()
		return err
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, path)
}

func (s *OsStore) Get(ctx context.Context, path string) (io.ReadCloser, FileInfo, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, FileInfo{}, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, FileInfo{}, err
	}

	return file, FileInfo{
		Path:         path,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
	}, nil
}

func (s *OsStore) Stat(ctx context.Context, path string) (FileInfo, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:         path,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
	}, nil
}

func (s *OsStore) Exists(ctx context.Context, path string) (bool, error) {
	_, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *OsStore) Delete(ctx context.Context, path string) error {
	err := os.Remove(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *OsStore) List(ctx context.Context, path string, options ListOptions) ([]FileInfo, error) {
	entries, err := os.ReadDir(path)
	if errors.Is(err, os.ErrNotExist) {
		return []FileInfo{}, nil
	}
	if err != nil {
		return nil, err
	}

	var out []FileInfo
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		out = append(out, FileInfo{
			Path:         filepath.Join(path, entry.Name()),
			Size:         info.Size(),
			LastModified: info.ModTime(),
			IsDir:        entry.IsDir(),
		})
	}
	return out, nil
}

func (s *OsStore) Copy(ctx context.Context, srcPath, dstPath string) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	return s.Put(ctx, dstPath, src, WriteOptions{})
}

func (s *OsStore) Move(ctx context.Context, srcPath, dstPath string) error {
	return os.Rename(srcPath, dstPath)
}
