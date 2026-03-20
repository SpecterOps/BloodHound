// Copyright 2026 Specter Ops, Inc.
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
package storage

import (
	"context"
	"errors"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"
)

type LocalStore struct {
	root string
}

func NewLocalStore(root string) *LocalStore {
	return &LocalStore{
		root: root,
	}
}

func (s *LocalStore) fullPath(path string) (string, error) {
	path = filepath.ToSlash(strings.TrimSpace(path))
	path = strings.TrimPrefix(path, "/")

	cleaned := filepath.Clean(path)

	if cleaned == "." {
		return "", errors.New("invalid empty path")
	}

	// Prevent root escape via ../
	if cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", errors.New("path escapes storage root")
	}

	full := filepath.Join(s.root, cleaned)
	return full, nil
}

func normalizeLogicalPath(path string) string {
	path = filepath.ToSlash(strings.TrimSpace(path))
	path = strings.TrimPrefix(path, "/")
	return filepath.Clean(path)
}

func detectContentType(path string) string {
	ext := filepath.Ext(path)
	if ext == "" {
		return "application/octet-stream"
	}

	ct := mime.TypeByExtension(ext)
	if ct == "" {
		return "application/octet-stream"
	}

	return ct
}

func (s *LocalStore) Put(ctx context.Context, path string, reader io.Reader, options WriteOptions) error {
	full, err := s.fullPath(path)
	if err != nil {
		return err
	}

	dir := filepath.Dir(full)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	cleanup := func() {
		tmp.Close()
		os.Remove(tmpName)
	}

	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			cleanup()
			return ctx.Err()
		default:
		}

		n, readErr := reader.Read(buf)
		if n > 0 {
			if _, err := tmp.Write(buf[:n]); err != nil {
				cleanup()
				return err
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			cleanup()
			return readErr
		}
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	if err := os.Rename(tmpName, full); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return nil
}

func (s *LocalStore) Get(ctx context.Context, path string) (io.ReadCloser, FileInfo, error) {
	full, err := s.fullPath(path)
	if err != nil {
		return nil, FileInfo{}, err
	}

	file, err := os.Open(full)
	if err != nil {
		return nil, FileInfo{}, err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, FileInfo{}, err
	}

	info := FileInfo{
		Path:         normalizeLogicalPath(path),
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
		ContentType:  detectContentType(path),
	}

	select {
	case <-ctx.Done():
		_ = file.Close()
		return nil, FileInfo{}, ctx.Err()
	default:
	}

	return file, info, nil
}

func (s *LocalStore) Stat(ctx context.Context, path string) (FileInfo, error) {
	full, err := s.fullPath(path)
	if err != nil {
		return FileInfo{}, err
	}

	select {
	case <-ctx.Done():
		return FileInfo{}, ctx.Err()
	default:
	}

	stat, err := os.Stat(full)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:         normalizeLogicalPath(path),
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
		ContentType:  detectContentType(path),
	}, nil
}

func (s *LocalStore) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.Stat(ctx, path)
	if err != nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (s *LocalStore) Delete(ctx context.Context, path string) error {
	full, err := s.fullPath(path)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	err = os.Remove(full)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *LocalStore) Move(ctx context.Context, srcPath, dstPath string) error {
	srcFull, err := s.fullPath(srcPath)
	if err != nil {
		return err
	}

	dstFull, err := s.fullPath(dstPath)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(dstFull), 0o755); err != nil {
		return err
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	return os.Rename(srcFull, dstFull)
}

func (s *LocalStore) listRecursive(ctx context.Context, logicalPrefix, fullPrefix string) ([]FileInfo, error) {
	var out []FileInfo

	err := filepath.WalkDir(fullPrefix, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(s.root, path)
		if err != nil {
			return err
		}

		out = append(out, FileInfo{
			Path:         normalizeLogicalPath(relPath),
			Size:         info.Size(),
			ContentType:  detectContentType(path),
			LastModified: info.ModTime(),
		})

		return nil
	})

	if errors.Is(err, os.ErrNotExist) {
		return []FileInfo{}, nil
	}
	if err != nil {
		return nil, err
	}

	return out, nil
}

func (s *LocalStore) listShallow(ctx context.Context, logicalPrefix, fullPrefix string) ([]FileInfo, error) {
	entries, err := os.ReadDir(fullPrefix)
	if errors.Is(err, os.ErrNotExist) {
		return []FileInfo{}, nil
	}
	if err != nil {
		return nil, err
	}

	out := make([]FileInfo, 0, len(entries))

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		full := filepath.Join(fullPrefix, entry.Name())
		rel, err := filepath.Rel(s.root, full)
		if err != nil {
			return nil, err
		}

		out = append(out, FileInfo{
			Path:         normalizeLogicalPath(rel),
			Size:         info.Size(),
			ContentType:  detectContentType(full),
			LastModified: info.ModTime(),
		})
	}

	return out, nil
}

func (s *LocalStore) List(ctx context.Context, path string, options ListOptions) ([]FileInfo, error) {
	full, err := s.fullPath(path)
	if err != nil {
		return nil, err
	}

	if !options.Recursive {
		// List only the contents of the prefix directory
		return s.listShallow(ctx, path, full)
	}

	return s.listShallow(ctx, path, full)
}

func (s *LocalStore) Copy(ctx context.Context, srcPath, dstPath string) error {
	srcFull, err := s.fullPath(srcPath)
	if err != nil {
		return err
	}

	dstFull, err := s.fullPath(dstPath)
	if err != nil {
		return err
	}

	srcFile, err := os.Open(srcFull)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}
	if srcInfo.IsDir() {
		return errors.New("cannot copy a directory")
	}

	if err := os.MkdirAll(filepath.Dir(dstFull), 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(filepath.Dir(dstFull), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()

	cleanup := func() {
		_ = tmp.Close()
		_ = os.Remove(tmpName)
	}

	buf := make([]byte, 32*1024)
	for {
		select {
		case <-ctx.Done():
			cleanup()
			return ctx.Err()
		default:
		}

		n, readErr := srcFile.Read(buf)
		if n > 0 {
			if _, err := tmp.Write(buf[:n]); err != nil {
				cleanup()
				return err
			}
		}

		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			cleanup()
			return readErr
		}
	}

	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpName)
		return err
	}

	return os.Rename(tmpName, dstFull)
}
