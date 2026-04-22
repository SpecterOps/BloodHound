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
	"io/fs"
	"os"
	"path"
)

// OsStore is a Storage implementation backed by a sandboxed directory on the
// local filesystem. All operations are confined to the root supplied to
// NewOsStore; callers use forward-slash-separated logical paths (e.g.
// "archives/clients/abc/file") and the kernel enforces that resolution stays
// within the root, including through symlinks.
type OsStore struct {
	root *os.Root
}

func NewOsStore(root string) (*OsStore, error) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	return &OsStore{
		root: r,
	}, nil
}

func (s *OsStore) Close() error {
	return s.root.Close()
}

func (s *OsStore) Open(name string) (fs.File, error) {
	return s.root.Open(name) // *os.File Satisfies fs.File
}

func (s *OsStore) Get(ctx context.Context, name string) (io.ReadCloser, FileInfo, error) {
	file, err := s.root.Open(name)
	if err != nil {
		return nil, FileInfo{}, err
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, FileInfo{}, err
	}

	return file, FileInfo{
		Path:         name,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
	}, nil
}

func (s *OsStore) Put(ctx context.Context, name string, reader io.Reader, options WriteOptions) error {
	var (
		dir     = path.Dir(name)
		tmpName string
		tmp     *os.File
		closed  bool
		err     error
	)
	if dir != "." {
		if err := s.root.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	id, err := randomID()
	if err != nil {
		return err
	}
	tmpName = path.Join(dir, ".tmp-"+id)

	tmp, err = s.root.OpenFile(tmpName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640)
	if err != nil {
		return err
	}

	// Remove the temp file on any failure path. On the success path we rename it
	// before this runs, so Remove returns ErrNotExist and we ignore it.
	defer func() {
		if err != nil {
			if !closed {
				_ = tmp.Close()
			}
			_ = s.root.Remove(tmpName)
		}
	}()

	if _, err = io.Copy(tmp, reader); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil { // flush data blocks
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	closed = true

	err = s.root.Rename(tmpName, name)
	return err
}

func (s *OsStore) Stat(ctx context.Context, name string) (FileInfo, error) {
	stat, err := s.root.Stat(name)
	if err != nil {
		return FileInfo{}, err
	}

	return FileInfo{
		Path:         name,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
	}, nil
}

func (s *OsStore) Exists(ctx context.Context, name string) (bool, error) {
	_, err := s.root.Stat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (s *OsStore) Delete(ctx context.Context, name string) error {
	err := s.root.Remove(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *OsStore) List(ctx context.Context, name string, options ListOptions) ([]FileInfo, error) {
	fsys := s.root.FS()
	if options.Recursive {
		out := []FileInfo{}
		err := fs.WalkDir(fsys, name, func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) && p == name {
					return fs.SkipAll
				}
				return err
			}
			if p == name && d.IsDir() {
				return nil // skip the prefix directory itself
			}
			info, err := d.Info()
			if err != nil {
				return err
			}
			out = append(out, FileInfo{
				Path:         p,
				Size:         info.Size(),
				LastModified: info.ModTime(),
				IsDir:        d.IsDir(),
			})
			if options.Limit > 0 && len(out) >= options.Limit {
				return fs.SkipAll
			}
			return nil
		})
		if err != nil {
			return nil, err
		}
		return out, nil
	}
	entries, err := fs.ReadDir(fsys, name)
	if errors.Is(err, fs.ErrNotExist) {
		return []FileInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		out = append(out, FileInfo{
			Path:         path.Join(name, entry.Name()),
			Size:         info.Size(),
			LastModified: info.ModTime(),
			IsDir:        entry.IsDir(),
		})
		if options.Limit > 0 && len(out) >= options.Limit {
			break
		}
	}
	return out, nil
}

// Copy duplicates srcName to destName atomically: on failure the destination is unchanged;
// on success, callers observe either the old content or hew, never a partial write.
func (s *OsStore) Copy(ctx context.Context, srcName, dstName string) error {
	src, err := s.root.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	return s.Put(ctx, dstName, src, WriteOptions{})
}

func (s *OsStore) Move(ctx context.Context, srcName, dstName string) error {
	dir := path.Dir(dstName)
	if dir != "." {
		if err := s.root.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}
	return s.root.Rename(srcName, dstName)
}
