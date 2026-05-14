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
	"mime"
	"os"
	"path"
	"strings"
	"sync"
)

var ErrIsDirectory = errors.New("is a directory")

// ctxReader wraps an io.Reader so that context cancellation is observed between
// reads. io.Copy calls Read in a loop and each call checks ctx.Err() before delegating.
type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (s *ctxReader) Read(p []byte) (int, error) {
	if err := s.ctx.Err(); err != nil {
		return 0, err
	}
	return s.r.Read(p)
}

// LocalStore is a Storage implementation backed by a sandboxed directory on the
// local filesystem. All operations are confined to the root supplied to
// NewLocalStore. Callers use forward-slash-separated logical paths (e.g.
// "archives/clients/abc/file") and the kernel enforces that resolution stays
// within the root, including through symlinks. Passing a path with a leading slash
// will cause an error.
type LocalStore struct {
	root      *os.Root
	closeOnce sync.Once
	closeErr  error
}

func NewLocalStore(root string) (*LocalStore, error) {
	r, err := os.OpenRoot(root)
	if err != nil {
		return nil, err
	}
	return &LocalStore{
		root: r,
	}, nil
}

func (s *LocalStore) Close() error {
	s.closeOnce.Do(func() { s.closeErr = s.root.Close() })
	return s.closeErr
}

func detectContentType(name string) string {
	if ext := path.Ext(name); ext != "" {
		if ct := mime.TypeByExtension(ext); ct != "" {
			return ct
		}
	}
	return "application/octet-stream"
}

// writeAtomic streams src into a temp file under dir(name), then publishes it at name.
// If failIfExists is true, publish uses link+unlink and returns an error satisfying
// errors.Is(err, fs.ErrExist) on collision. Otherwise, publish uses rename and silently
// replaces any existing file at name. The temp file is removed on every failure path.
func (s *LocalStore) writeAtomic(ctx context.Context, name string, src io.Reader, failIfExists bool) error {
	var (
		dir     = path.Dir(name)
		tmpName string
		tmp     *os.File
		id      string
		closed  bool
		err     error
	)

	if err = ctx.Err(); err != nil {
		return err
	}

	if dir != "." {
		if err = s.root.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}

	if id, err = randomID(); err != nil {
		return err
	}
	tmpName = path.Join(dir, ".tmp-"+id)

	if tmp, err = s.root.OpenFile(tmpName, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o640); err != nil {
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

	if _, err = io.Copy(tmp, &ctxReader{ctx: ctx, r: src}); err != nil {
		return err
	}
	if err = tmp.Sync(); err != nil { // flush data blocks
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}
	closed = true

	if failIfExists {
		// Link fails with fs.ErrExist if name already exists. Atomic and race-free
		// against concurrent writers on the same filesystem.
		if err = s.root.Link(tmpName, name); err != nil {
			return err
		}
		// Publish is durable, a failed Remove leaks a .tmp-... that a sweeper can reclaim
		_ = s.root.Remove(tmpName)
		return nil
	}
	err = s.root.Rename(tmpName, name)
	return err
}

// Open returns an fs.File, which must be closed.
func (s *LocalStore) Open(name string) (fs.File, error) {
	return s.root.Open(name) // *os.File Satisfies fs.File
}

func (s *LocalStore) Stat(ctx context.Context, name string) (FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return FileInfo{}, err
	}
	stat, err := s.root.Stat(name)
	if err != nil {
		return FileInfo{}, err
	}
	if stat.IsDir() {
		return FileInfo{}, fmt.Errorf("stat %q: %w", name, ErrIsDirectory)
	}

	return FileInfo{
		Path:         name,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
		ContentType:  detectContentType(name),
	}, nil
}

func (s *LocalStore) Get(ctx context.Context, name string) (io.ReadCloser, FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, FileInfo{}, err
	}
	file, err := s.root.Open(name)
	if err != nil {
		return nil, FileInfo{}, err
	}
	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, FileInfo{}, err
	}
	if stat.IsDir() {
		_ = file.Close()
		return nil, FileInfo{}, fmt.Errorf("get: %q, %w", name, ErrIsDirectory)
	}
	if err := ctx.Err(); err != nil {
		_ = file.Close()
		return nil, FileInfo{}, err
	}

	return file, FileInfo{
		Path:         name,
		Size:         stat.Size(),
		LastModified: stat.ModTime(),
		IsDir:        stat.IsDir(),
		ContentType:  detectContentType(name),
	}, nil
}

func (s *LocalStore) Put(ctx context.Context, name string, reader io.Reader, options WriteOptions) error {
	return s.writeAtomic(ctx, name, reader, options.FailIfExists)
}

func (s *LocalStore) Exists(ctx context.Context, name string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	stat, err := s.root.Stat(name)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if stat.IsDir() {
		return false, nil
	}
	return true, nil
}

func (s *LocalStore) Delete(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	err := s.root.Remove(name)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (s *LocalStore) List(ctx context.Context, name string, options ListOptions) ([]FileInfo, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	listName := name
	if strings.TrimSpace(listName) == "" || listName == "/" {
		listName = "."
	}
	fsys := s.root.FS()
	if options.Recursive {
		out := []FileInfo{}
		err := fs.WalkDir(fsys, listName, func(p string, d fs.DirEntry, err error) error {
			if cerr := ctx.Err(); cerr != nil {
				return cerr
			}
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) && p == listName {
					return fs.SkipAll
				}
				return err
			}
			if d.IsDir() {
				return nil
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
				ContentType:  detectContentType(p),
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
	entries, err := fs.ReadDir(fsys, listName)
	if errors.Is(err, fs.ErrNotExist) {
		return []FileInfo{}, nil
	}
	if err != nil {
		return nil, err
	}
	out := make([]FileInfo, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() {
			continue
		}
		info, err := entry.Info()
		if err != nil {
			return nil, err
		}
		entryPath := path.Join(listName, entry.Name())
		if listName == "." {
			entryPath = entry.Name()
		}
		out = append(out, FileInfo{
			Path:         entryPath,
			Size:         info.Size(),
			LastModified: info.ModTime(),
			IsDir:        entry.IsDir(),
			ContentType:  detectContentType(entryPath),
		})
		if options.Limit > 0 && len(out) >= options.Limit {
			break
		}
	}
	return out, nil
}

// Copy duplicates srcName to destName atomically: on failure the destination is unchanged;
// on success, callers observe either the old content or new, never a partial write.
func (s *LocalStore) Copy(ctx context.Context, srcName, dstName string, options WriteOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	src, err := s.root.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()
	info, err := src.Stat()
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("copy %q: %w", srcName, ErrIsDirectory)
	}
	return s.writeAtomic(ctx, dstName, src, options.FailIfExists)
}

// Move is able to move a file from srcName to dstName using a two-step Link+Remove. A crash between link
// and unlink leaves both names present.
func (s *LocalStore) Move(ctx context.Context, srcName, dstName string, options WriteOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	dir := path.Dir(dstName)
	if dir != "." {
		if err := s.root.MkdirAll(dir, 0o750); err != nil {
			return err
		}
	}
	if options.FailIfExists {
		if err := s.root.Link(srcName, dstName); err != nil {
			return err
		}
		return s.root.Remove(srcName)
	}
	return s.root.Rename(srcName, dstName)
}
