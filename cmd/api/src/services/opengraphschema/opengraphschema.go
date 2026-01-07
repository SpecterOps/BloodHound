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
	"github.com/specterops/dawgs/graph"
)

// OpenGraphSchemaRepository -
type OpenGraphSchemaRepository interface {
	// Kinds
	GetKindByName(ctx context.Context, name string) (model.Kind, error)

	// Environment
	CreateSchemaEnvironment(ctx context.Context, schemaExtensionId int32, environmentKindId int32, sourceKindId int32) (model.SchemaEnvironment, error)
	GetSchemaEnvironmentById(ctx context.Context, schemaEnvironmentId int32) (model.SchemaEnvironment, error)
	DeleteSchemaEnvironment(ctx context.Context, schemaEnvironmentId int32) error

	// Source Kinds
	RegisterSourceKind(ctx context.Context) func(sourceKind graph.Kind) error
	GetSourceKindByName(ctx context.Context, name string) (database.SourceKind, error)

	// Principal Kinds
	CreateSchemaEnvironmentPrincipalKind(ctx context.Context, environmentId int32, principalKind int32) (model.SchemaEnvironmentPrincipalKind, error)
	GetSchemaEnvironmentPrincipalKindsByEnvironmentId(ctx context.Context, environmentId int32) (model.SchemaEnvironmentPrincipalKinds, error)
	DeleteSchemaEnvironmentPrincipalKind(ctx context.Context, environmentId int32, principalKind int32) error
}

type OpenGraphSchemaService struct {
	openGraphSchemaRepository OpenGraphSchemaRepository
	transactor                Transactor
}

type Transactor interface {
	WithTransaction(ctx context.Context, fn func(repo OpenGraphSchemaRepository) error) error
}

func NewOpenGraphSchemaService(openGraphSchemaRepository OpenGraphSchemaRepository) *OpenGraphSchemaService {
	return &OpenGraphSchemaService{
		openGraphSchemaRepository: openGraphSchemaRepository,
	}
}
