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

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/opengraphschema.go -package=mocks . OpenGraphSchemaRepository

import (
	"context"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

// OpenGraphSchemaRepository -
type OpenGraphSchemaRepository interface {
	UpsertGraphSchemaExtension(ctx context.Context, extensionID int32, environments []database.EnvironmentInput, findings []database.FindingInput) error
	GetGraphSchemaExtensions(ctx context.Context, extensionFilters model.Filters, sort model.Sort, skip, limit int) (model.GraphSchemaExtensions, int, error)
}

type OpenGraphSchemaService struct {
	openGraphSchemaRepository OpenGraphSchemaRepository
}

func NewOpenGraphSchemaService(openGraphSchemaRepository OpenGraphSchemaRepository) *OpenGraphSchemaService {
	return &OpenGraphSchemaService{
		openGraphSchemaRepository: openGraphSchemaRepository,
	}
}
