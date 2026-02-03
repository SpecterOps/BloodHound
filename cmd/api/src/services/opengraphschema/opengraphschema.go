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

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// OpenGraphSchemaRepository -
//
//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/opengraphschema.go -package=mocks . OpenGraphSchemaRepository
type OpenGraphSchemaRepository interface {
	UpsertOpenGraphExtension(ctx context.Context, graphExtensionInput model.GraphExtensionInput) (bool, error)
	GetGraphSchemaExtensions(ctx context.Context, extensionFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error)
}

// GraphDBKindRepository -
//
//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphdbkindrepository.go -package=mocks . GraphDBKindRepository
type GraphDBKindRepository interface {
	// RefreshKinds refreshes the database and in memory kinds maps
	RefreshKinds(ctx context.Context) error
}

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
