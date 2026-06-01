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
	"io"
	"time"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/fs.go -package=mocks . Storage,FileService

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

// Storage serves as a storage abstraction that can be used to store and manage files
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

// FileService serves as an abstraction to handle files with different storage backends. This serves as
// a list of general functions that each file service must implement.
type FileService interface {
	// GetFile returns a io.ReadCloser and FileInfo for the named filed that is requested.
	GetFile(ctx context.Context, name string) (io.ReadCloser, FileInfo, error)

	// ReadFile returns the byte information of the file that is requested.
	ReadFile(ctx context.Context, name string) ([]byte, error)

	// WriteFile takes the name and byte information as well as WriteOptions to write to the
	// storage backend.
	WriteFile(ctx context.Context, name string, data []byte, opts WriteOptions) error

	// WriteFileFromReader takes the name, io.Reader, and WriteOptions to write to the
	// storage backend.
	WriteFileFromReader(ctx context.Context, name string, reader io.Reader, opts WriteOptions) error

	// DeleteFile deletes a file at a specific name from the storage backend. If the file
	// is not found, no error is returned.
	DeleteFile(ctx context.Context, name string) error

	// WriteTempFile handles the creation of a temp file when given an io.Reader. A prefix
	// can also be used to define how the temp file is created. WriteOptions can also be
	// specified.
	WriteTempFile(ctx context.Context, prefix string, reader io.Reader, opts WriteOptions) (string, error)

	// MoveFile takes a srcName and dstName to move a file from one location to another on
	// the storage backend. WriteOptions can also be specified in the case of collisions.
	MoveFile(ctx context.Context, srcName, dstName string, opts WriteOptions) error

	// ListFiles lists the files at a given location in the storage backend. This can be done
	// recursively, or with a limit on the specified directory.
	ListFiles(ctx context.Context, name string, opts ListOptions) ([]FileInfo, error)
}

type StorageFileService struct {
	Storage Storage
}

func NewFileService(storage Storage) *StorageFileService {
	return &StorageFileService{Storage: storage}
}

func randomID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func (s *StorageFileService) GetFile(ctx context.Context, name string) (io.ReadCloser, FileInfo, error) {
	return s.Storage.Get(ctx, name)
}

func (s *StorageFileService) ReadFile(ctx context.Context, name string) ([]byte, error) {
	rc, _, err := s.Storage.Get(ctx, name)
	if err != nil {
		return nil, err
	}
	data, readErr := io.ReadAll(rc)
	if closeErr := rc.Close(); closeErr != nil {
		if readErr != nil {
			return nil, errors.Join(readErr, closeErr)
		}
		return nil, closeErr
	}
	return data, readErr
}

func (s *StorageFileService) WriteFile(ctx context.Context, name string, data []byte, opts WriteOptions) error {
	return s.Storage.Put(ctx, name, bytes.NewReader(data), opts)
}

func (s *StorageFileService) WriteFileFromReader(ctx context.Context, name string, reader io.Reader, opts WriteOptions) error {
	return s.Storage.Put(ctx, name, reader, opts)
}

func (s *StorageFileService) DeleteFile(ctx context.Context, name string) error {
	return s.Storage.Delete(ctx, name)
}

func (s *StorageFileService) WriteTempFile(ctx context.Context, prefix string, reader io.Reader, opts WriteOptions) (string, error) {
	id, err := randomID()
	if err != nil {
		return "", err
	}

	tempPath := prefix + "tmp-" + id
	if err := s.Storage.Put(ctx, tempPath, reader, opts); err != nil {
		return "", err
	}

	return tempPath, nil
}

func (s *StorageFileService) MoveFile(ctx context.Context, srcName, dstName string, options WriteOptions) error {
	return s.Storage.Move(ctx, srcName, dstName, options)
}

func (s *StorageFileService) ListFiles(ctx context.Context, name string, options ListOptions) ([]FileInfo, error) {
	return s.Storage.List(ctx, name, options)
}

func MoveFileBetweenServices(
	ctx context.Context,
	sourceService FileService,
	destinationService FileService,
	sourceName string,
	destinationName string,
	opts WriteOptions,
) error {
	sourceFile, _, err := sourceService.GetFile(ctx, sourceName)
	if err != nil {
		return err
	}

	if err := destinationService.WriteFileFromReader(ctx, destinationName, sourceFile, opts); err != nil {
		if closeErr := sourceFile.Close(); closeErr != nil {
			return errors.Join(err, closeErr)
		}
		return err
	}

	if err := sourceFile.Close(); err != nil {
		return err
	}

	return sourceService.DeleteFile(ctx, sourceName)
}
