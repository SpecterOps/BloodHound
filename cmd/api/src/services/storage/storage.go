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

package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/config"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/fs.go -package=mocks . FileService

type FileServiceName string

const (
	FileServiceIngest     FileServiceName = "ingest"
	FileServiceRetained   FileServiceName = "retained"
	FileServiceCollectors FileServiceName = "collectors"
	FileServiceJobLogs    FileServiceName = "job_logs"
	FileServiceWork       FileServiceName = "work"
)

var ErrFileServiceNotFound = errors.New("file service not found")

type FileInfo struct {
	Path         string
	Size         int64
	ContentType  string
	ETag         string
	LastModified time.Time
	IsDir        bool
}

type WriteOptions struct {
	ContentType string
	Metadata    map[string]string
	// FailIfExists causes write to return an error wrapping fs.ErrExist when the
	// destination already exists, instead of silently replacing it.
	FailIfExists bool
}

type ListOptions struct {
	Recursive bool
	Limit     int
}

// Serves as a storage abstraction that can be used to store and manage files
// in a variety of storage backends.
type Storage interface {
	// Put writes a file at the given path.
	Put(ctx context.Context, name string, reader io.Reader, options WriteOptions) error

	// Get opens a file for reading.
	Get(ctx context.Context, name string) (io.ReadCloser, FileInfo, error)

	// Stat returns metadata for a given path.
	Stat(ctx context.Context, name string) (FileInfo, error)

	// Delete removes a file.
	Delete(ctx context.Context, name string) error

	// Exists checks whether a file exists.
	Exists(ctx context.Context, name string) (bool, error)

	// List returns a list of files in a given path.
	List(ctx context.Context, name string, options ListOptions) ([]FileInfo, error)

	// Copy duplicates an object.
	Copy(ctx context.Context, srcName, dstName string, options WriteOptions) error

	// Move moves an object. Is done by a copy and a delete.
	Move(ctx context.Context, srcName, dstName string, options WriteOptions) error
}

type FileService interface {
	GetFile(ctx context.Context, name string) (io.ReadCloser, FileInfo, error)
	ReadFile(ctx context.Context, name string) ([]byte, error)
	WriteFile(ctx context.Context, name string, data []byte, opts WriteOptions) error
	DeleteFile(ctx context.Context, name string) error
	WriteTempFile(ctx context.Context, prefix string, reader io.Reader, opts WriteOptions) (string, error)
	MoveFile(ctx context.Context, srcName, dstName string, opts WriteOptions) error
	ListFiles(ctx context.Context, name string, opts ListOptions) ([]FileInfo, error)
}

type LocalFileService struct {
	Storage Storage
}

func NewFileService(storage Storage) *LocalFileService {
	return &LocalFileService{Storage: storage}
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *LocalFileService) GetFile(ctx context.Context, name string) (io.ReadCloser, FileInfo, error) {
	return s.Storage.Get(ctx, name)
}

func (s *LocalFileService) ReadFile(ctx context.Context, name string) ([]byte, error) {
	rc, _, err := s.Storage.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}

func (s *LocalFileService) WriteFile(ctx context.Context, name string, data []byte, opts WriteOptions) error {
	return s.Storage.Put(ctx, name, bytes.NewReader(data), opts)
}

func (s *LocalFileService) DeleteFile(ctx context.Context, name string) error {
	return s.Storage.Delete(ctx, name)
}

func (s *LocalFileService) WriteTempFile(ctx context.Context, prefix string, reader io.Reader, opts WriteOptions) (string, error) {
	id, err := randomID()
	if err != nil {
		return "", err
	}

	tempPath := path.Join(prefix, "tmp-"+id)
	if err := s.Storage.Put(ctx, tempPath, reader, opts); err != nil {
		return "", err
	}

	return tempPath, nil
}

func (s *LocalFileService) MoveFile(ctx context.Context, srcName, dstName string, options WriteOptions) error {
	return s.Storage.Move(ctx, srcName, dstName, options)
}

func (s *LocalFileService) ListFiles(ctx context.Context, name string, options ListOptions) ([]FileInfo, error) {
	return s.Storage.List(ctx, name, options)
}

// FileServiceResolver is an interface that is used to resolve the actual FileService needed for
// a specific use case. This is ultimately map backed.
type FileServiceResolver interface {
	Resolve(name FileServiceName) (FileService, error)
}

type fileServiceResolver struct {
	services map[FileServiceName]FileService
}

func NewFileServiceResolver(services map[FileServiceName]FileService) (FileServiceResolver, error) {
	var (
		serviceName    FileServiceName
		fileService    FileService
		copiedServices = make(map[FileServiceName]FileService, len(services))
	)

	for serviceName, fileService = range services {
		if serviceName == "" {
			return nil, errors.New("file service name is required")
		}
		if fileService == nil {
			return nil, fmt.Errorf("file service %q is nil", serviceName)
		}

		copiedServices[serviceName] = fileService
	}

	return &fileServiceResolver{
		services: copiedServices,
	}, nil
}

func (s *fileServiceResolver) Resolve(name FileServiceName) (FileService, error) {
	var (
		fileService FileService
		found       bool
	)

	if name == "" {
		return nil, fmt.Errorf("%w: empty name", ErrFileServiceNotFound)
	}

	fileService, found = s.services[name]
	if !found {
		return nil, fmt.Errorf("%w: %s", ErrFileServiceNotFound, name)
	}

	return fileService, nil
}

func NewDefaultFileServices(cfg config.Configuration) (map[FileServiceName]FileService, error) {
	var (
		fileServices = make(map[FileServiceName]FileService)
	)
	workStore, err := NewLocalStore(cfg.WorkDir)
	if err != nil {
		return nil, err
	}

	ingestStore, err := NewLocalStore(cfg.TempDirectory())
	if err != nil {
		return nil, err
	}

	retainStore, err := NewLocalStore(cfg.RetainedFilesDirectory())
	if err != nil {
		return nil, err
	}

	collectorsStore, err := NewLocalStore(cfg.CollectorsBasePath)
	if err != nil {
		return nil, err
	}

	workService := NewFileService(workStore)
	fileServices[FileServiceWork] = workService

	ingestService := NewFileService(ingestStore)
	fileServices[FileServiceIngest] = ingestService

	retainService := NewFileService(retainStore)
	fileServices[FileServiceRetained] = retainService

	collectersService := NewFileService(collectorsStore)
	fileServices[FileServiceCollectors] = collectersService

	return fileServices, nil
}
