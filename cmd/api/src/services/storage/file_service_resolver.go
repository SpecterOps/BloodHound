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

type fileServiceResolver struct {
	services map[storage.FileServiceName]storage.FileService
}

func NewFileServiceResolver(services map[storage.FileServiceName]storage.FileService) (FileServiceResolver, error) {
	var (
		copiedServices = make(map[storage.FileServiceName]storage.FileService, len(services))
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

// NewDefaultFileServices creates the file services that should be considered default with
// BloodHound. Additional FileServices can still be created prior to creating a Resolver.
func NewDefaultFileServices(cfg config.Configuration) (map[storage.FileServiceName]storage.FileService, error) {
	var (
		fileServices = make(map[storage.FileServiceName]storage.FileService)
		openedStores []*storage.LocalStore
	)

	closeOpened := func() {
		for _, store := range openedStores {
			_ = store.Close()
		}
	}

	workStore, err := storage.NewLocalStore(cfg.WorkDir)
	if err != nil {
		return nil, err
	}
	openedStores = append(openedStores, workStore)

	ingestStore, err := storage.NewLocalStore(cfg.TempDirectory())
	if err != nil {
		closeOpened()
		return nil, err
	}
	openedStores = append(openedStores, ingestStore)

	retainStore, err := storage.NewLocalStore(cfg.RetainedFilesDirectory())
	if err != nil {
		closeOpened()
		return nil, err
	}
	openedStores = append(openedStores, retainStore)

	collectorsStore, err := storage.NewLocalStore(cfg.CollectorsBasePath)
	if err != nil {
		closeOpened()
		return nil, err
	}
	openedStores = append(openedStores, collectorsStore)

	workService := storage.NewFileService(workStore)
	fileServices[storage.FileServiceWork] = workService

	ingestService := storage.NewFileService(ingestStore)
	fileServices[storage.FileServiceIngest] = ingestService

	retainService := storage.NewFileService(retainStore)
	fileServices[storage.FileServiceRetained] = retainService

	collectorsService := storage.NewFileService(collectorsStore)
	fileServices[storage.FileServiceCollectors] = collectorsService

	return fileServices, nil
}
