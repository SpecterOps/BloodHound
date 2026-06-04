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
	"errors"
	"fmt"

	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/packages/go/storage"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/file_service_resolver.go -package=mocks . FileServiceResolver

// FileServiceResolver is an interface that is used to resolve the actual filestorage.FileService needed for
// a specific use case. This is ultimately map backed.
type FileServiceResolver interface {
	// Resolve returns a filestorage.FileService interface if a filestorage.FileService is found with the given name.
	// Otherwise, an error is returned.
	Resolve(name storage.FileServiceName) (storage.FileService, error)
}

type FileServiceMap map[storage.FileServiceName]storage.FileService

type fileServiceResolver struct {
	services FileServiceMap
}

func NewFileServiceResolver(services FileServiceMap) (FileServiceResolver, error) {
	var (
		copiedServices = make(FileServiceMap, len(services))
	)

	for serviceName, fileService := range services {
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

func (s *fileServiceResolver) Resolve(name storage.FileServiceName) (storage.FileService, error) {
	var (
		fileService storage.FileService
		found       bool
	)

	if name == "" {
		return nil, fmt.Errorf("%w: empty name", storage.ErrFileServiceNotFound)
	}

	fileService, found = s.services[name]
	if !found {
		return nil, fmt.Errorf("%w: %s", storage.ErrFileServiceNotFound, name)
	}

	return fileService, nil
}

// createLocalStore takes a location to create the storage.LocalStore, and wraps that in a
// storage.FileService. Both are returned. If there is an error in this process, nil is
// returned for both structs, and the error is returned.
func createLocalStore(location string) (*storage.LocalStore, storage.FileService, error) {
	var (
		localStore *storage.LocalStore
		err        error
	)

	if localStore, err = storage.NewLocalStore(location); err != nil {
		return nil, nil, err
	}

	return localStore, storage.NewFileService(localStore), nil
}

// closeLocalStores contains the functionality to close any storage.LocalStore that has been opened
// if there was an error. Errors from the close are joined together and returned as well.
func closeLocalStores(localStores []*storage.LocalStore) error {
	var closeErr error

	for _, localStore := range localStores {
		closeErr = errors.Join(closeErr, localStore.Close())
	}

	return closeErr
}

// NewDefaultFileServices creates the file services that should be considered default with
// BloodHound. Additional FileServices can still be created prior to creating a Resolver.
func NewDefaultFileServices(cfg config.Configuration) (FileServiceMap, error) {
	var (
		fileServices = make(FileServiceMap, 4)
		openedStores []*storage.LocalStore
		definitions  = []struct {
			name     storage.FileServiceName
			location string
		}{
			{name: storage.FileServiceWork, location: cfg.WorkDir},
			{name: storage.FileServiceIngest, location: cfg.TempDirectory()},
			{name: storage.FileServiceRetained, location: cfg.RetainedFilesDirectory()},
			{name: storage.FileServiceCollectors, location: cfg.CollectorsBasePath},
		}
	)

	for _, definition := range definitions {
		localStore, fileService, err := createLocalStore(definition.location)
		if err != nil {
			return nil, errors.Join(err, closeLocalStores(openedStores))
		}

		openedStores = append(openedStores, localStore)
		fileServices[definition.name] = fileService
	}

	return fileServices, nil
}
