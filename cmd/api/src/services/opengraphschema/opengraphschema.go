// Copyright 2025 Specter Ops, Inc.
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
package opengraphschema

// Mocks

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/opengraphschemarepository.go -package=mocks . OpenGraphSchemaRepository
//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphdbkindrepository.go -package=mocks . GraphDBKindRepository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
)

// OpenGraphSchemaRepository -
type OpenGraphSchemaRepository interface {
	UpsertGraphSchemaExtension(ctx context.Context, graphSchema model.GraphSchema) (bool, error)
}

// GraphDBKindRepository -
type GraphDBKindRepository interface {
	// RefreshKinds refreshes the in memory kinds maps
	RefreshKinds(ctx context.Context) error
}

// OpenGraphSchemaService -
type OpenGraphSchemaService struct {
	openGraphSchemaRepository OpenGraphSchemaRepository
	graphDBKindRepository     GraphDBKindRepository
}

func NewOpenGraphSchemaService(openGraphSchemaExtensionRepository OpenGraphSchemaRepository, graphDBKindRepository GraphDBKindRepository) *OpenGraphSchemaService {
	return &OpenGraphSchemaService{
		openGraphSchemaRepository: openGraphSchemaExtensionRepository,
		graphDBKindRepository:     graphDBKindRepository,
	}
}

// UpsertGraphSchemaExtension - upserts the provided graph schema.
func (o *OpenGraphSchemaService) UpsertGraphSchemaExtension(ctx context.Context, graphSchema model.GraphSchema) (bool, error) {
	var (
		err          error
		schemaExists bool
	)

	if err = validateGraphSchemaModel(graphSchema); err != nil {
		return schemaExists, fmt.Errorf("graph schema validation error: %w", err)
	} else if schemaExists, err = o.openGraphSchemaRepository.UpsertGraphSchemaExtension(ctx, graphSchema); err != nil {
		return schemaExists, err
	} else if err = o.graphDBKindRepository.RefreshKinds(ctx); err != nil {
		slog.WarnContext(ctx, "OpenGraphSchema: refreshing graph kind maps failed", attr.Error(err))
	}
	return schemaExists, nil
}

func validateGraphSchemaModel(graphSchema model.GraphSchema) error {
	if graphSchema.GraphSchemaExtension.Name == "" {
		return errors.New("graph schema extension name is required")
	} else if len(graphSchema.GraphSchemaNodeKinds) == 0 {
		return errors.New("graph schema node kinds is required")
	} // TODO: Put all edge and node kinds into a map and verify there's no duplicates
	return nil
}
