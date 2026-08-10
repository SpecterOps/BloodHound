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
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsConfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/packages/go/storage"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file=../../../../../LICENSE.header -destination=./mocks/fileserviceresolver.go -package=mocks . FileServiceResolver

// FileServiceResolver is an interface that is used to resolve the actual filestorage.FileService needed for
// a specific use case. This is ultimately map backed.
type FileServiceResolver interface {
	// Resolve returns a filestorage.FileService interface if a filestorage.FileService is found with the given name.
	// Otherwise, an error is returned.
	Resolve(name storage.FileServiceName) (storage.FileService, error)
}

type FileServiceMap map[storage.FileServiceName]storage.FileService

// FileServiceDefinition describes a named file service and its local storage path.
type FileServiceDefinition struct {
	Name      storage.FileServiceName
	LocalPath string
}

type fileServiceResolver struct {
	services FileServiceMap
}

type fileServiceProvider string

const (
	fileServiceProviderLocal fileServiceProvider = "local"
	fileServiceProviderS3    fileServiceProvider = "s3"
)

func NewFileServiceResolver(services FileServiceMap) (FileServiceResolver, error) {
	copiedServices := make(FileServiceMap, len(services))

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

func createS3Store(bucket, prefix string, client *s3.Client) storage.FileService {
	return storage.NewFileService(storage.NewS3Store(bucket, prefix, client))
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

func parseFileServiceProvider(provider string) (fileServiceProvider, error) {
	provider = strings.TrimSpace(provider)

	switch fileServiceProvider(provider) {
	case fileServiceProviderLocal, fileServiceProviderS3:
		return fileServiceProvider(provider), nil
	default:
		return "", fmt.Errorf("unsupported file service provider %q", provider)
	}
}

func normalizeS3Prefix(prefix string) (string, error) {
	prefix = strings.Trim(strings.TrimSpace(prefix), "/")
	if prefix == "" {
		return "", errors.New("s3 prefix is required")
	}

	if strings.Contains(prefix, "\\") {
		return "", errors.New("s3 prefix contains a backslash")
	}

	for _, segment := range strings.Split(prefix, "/") {
		if segment == "." || segment == ".." {
			return "", errors.New("s3 prefix contains a relative segment")
		}
	}

	return prefix, nil
}

func s3PrefixesOverlap(left, right string) bool {
	return left == right || strings.HasPrefix(left, right+"/") || strings.HasPrefix(right, left+"/")
}

type resolvedFileServiceDefinition struct {
	definition FileServiceDefinition
	provider   fileServiceProvider
	prefix     string
}

func resolveFileServiceDefinitions(cfg config.Configuration, definitions []FileServiceDefinition) ([]resolvedFileServiceDefinition, bool, error) {
	var (
		definitionsByName   = make(map[storage.FileServiceName]struct{}, len(definitions))
		resolvedDefinitions = make([]resolvedFileServiceDefinition, 0, len(definitions))
		s3Prefixes          = make(map[string]storage.FileServiceName)
		s3Required          bool
	)

	for _, definition := range definitions {
		var (
			provider             = fileServiceProviderLocal
			prefix               string
			serviceConfiguration config.FileServiceConfiguration
			configured           bool
			err                  error
		)

		if definition.Name == "" {
			return nil, false, errors.New("file service definition name is required")
		}
		if _, found := definitionsByName[definition.Name]; found {
			return nil, false, fmt.Errorf("duplicate file service definition %q", definition.Name)
		}
		definitionsByName[definition.Name] = struct{}{}

		if serviceConfiguration, configured = cfg.Storage.FileServices[string(definition.Name)]; configured {
			provider, err = parseFileServiceProvider(serviceConfiguration.Provider)
			if err != nil {
				return nil, false, fmt.Errorf("file service %q: %w", definition.Name, err)
			}
		}

		if provider == fileServiceProviderS3 {
			s3Required = true
			if prefix, err = normalizeS3Prefix(serviceConfiguration.Prefix); err != nil {
				return nil, false, fmt.Errorf("file service %q: %w", definition.Name, err)
			}

			for existingPrefix, existingServiceName := range s3Prefixes {
				if s3PrefixesOverlap(prefix, existingPrefix) {
					return nil, false, fmt.Errorf("s3 prefixes for file services %q and %q overlap", existingServiceName, definition.Name)
				}
			}
			s3Prefixes[prefix] = definition.Name
		}

		resolvedDefinitions = append(resolvedDefinitions, resolvedFileServiceDefinition{
			definition: definition,
			provider:   provider,
			prefix:     prefix,
		})
	}

	for configuredServiceName := range cfg.Storage.FileServices {
		if _, found := definitionsByName[storage.FileServiceName(configuredServiceName)]; !found {
			return nil, false, fmt.Errorf("configuration references unknown file service %q", configuredServiceName)
		}
	}

	return resolvedDefinitions, s3Required, nil
}

func createS3Client(ctx context.Context, bucketConfiguration config.BucketConfiguration) (*s3.Client, error) {
	var (
		awsConfiguration aws.Config
		loadOptions      []func(*awsConfig.LoadOptions) error
		err              error
	)

	if strings.TrimSpace(bucketConfiguration.Name) == "" {
		return nil, errors.New("storage.instance_bucket.name is required when an S3 file service is configured")
	}

	if region := strings.TrimSpace(bucketConfiguration.Region); region == "" {
		return nil, errors.New("storage.instance_bucket.region is required when an S3 file service is configured")
	} else {
		loadOptions = append(loadOptions, awsConfig.WithRegion(region))
	}

	awsConfiguration, err = awsConfig.LoadDefaultConfig(ctx, loadOptions...)
	if err != nil {
		return nil, fmt.Errorf("load AWS configuration for file services: %w", err)
	}

	return s3.NewFromConfig(awsConfiguration), nil
}

// NewDefaultFileServices creates the file services that should be considered default with
// BloodHound. Additional file service definitions may be supplied by callers before creating a resolver.
func NewDefaultFileServices(ctx context.Context, cfg config.Configuration, additionalDefinitions ...FileServiceDefinition) (FileServiceMap, error) {
	var (
		definitions = append([]FileServiceDefinition{
			{Name: storage.FileServiceWork, LocalPath: cfg.WorkDir},
			{Name: storage.FileServiceIngest, LocalPath: cfg.TempDirectory()},
			{Name: storage.FileServiceRetained, LocalPath: cfg.RetainedFilesDirectory()},
			{Name: storage.FileServiceCollectors, LocalPath: cfg.CollectorsBasePath},
		}, additionalDefinitions...)
		fileServices        = make(FileServiceMap, len(definitions))
		openedStores        []*storage.LocalStore
		resolvedDefinitions []resolvedFileServiceDefinition
		fileService         storage.FileService
		localStore          *storage.LocalStore
		s3Client            *s3.Client
		s3Required          bool
		err                 error
	)

	resolvedDefinitions, s3Required, err = resolveFileServiceDefinitions(cfg, definitions)
	if err != nil {
		return nil, err
	}

	if s3Required {
		if s3Client, err = createS3Client(ctx, cfg.Storage.InstanceBucket); err != nil {
			return nil, err
		}
	}

	for _, resolvedDefinition := range resolvedDefinitions {
		switch resolvedDefinition.provider {
		case fileServiceProviderS3:
			fileServices[resolvedDefinition.definition.Name] = createS3Store(
				strings.TrimSpace(cfg.Storage.InstanceBucket.Name),
				resolvedDefinition.prefix,
				s3Client,
			)
		case fileServiceProviderLocal:
			localStore, fileService, err = createLocalStore(resolvedDefinition.definition.LocalPath)
			if err != nil {
				return nil, errors.Join(err, closeLocalStores(openedStores))
			}

			openedStores = append(openedStores, localStore)
			fileServices[resolvedDefinition.definition.Name] = fileService
		}
	}

	return fileServices, nil
}
