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

package services

//go:generate go tool mockery

import (
	"context"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/opengraphschema"
)

type Database interface {
	ListAggregations(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems, skip, limit int) ([]Aggregation, int, error)
}

type OpenGraphSchema interface {
	GetExtension(ctx context.Context, id int32) (opengraphschema.Extension, error)
}

type Service struct {
	database        Database
	openGraphSchema OpenGraphSchema
}

func NewService(databaseInterface Database, openGraphSchema OpenGraphSchema) *Service {
	if databaseInterface == nil || openGraphSchema == nil {
		panic("dataquality: service requires database and opengraph schema")
	}
	return &Service{database: databaseInterface, openGraphSchema: openGraphSchema}
}
