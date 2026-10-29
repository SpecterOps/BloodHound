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
	"errors"
	"time"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/opengraphschema"
)

var (
	ErrExtensionNotFound = errors.New("extension not found")
	ErrInvalidQuery      = errors.New("invalid aggregation query")
)

type Aggregation struct {
	ID                      int32
	RunID                   string
	SchemaExtensionID       int32
	SchemaEnvironmentKindID int32
	MetricType              string
	MetricName              string
	MetricValue             float64
	KindID                  *int32
	CreatedAt               time.Time
	UpdatedAt               time.Time
}

type AggregationQuery struct {
	EnvironmentKindID int32
	ExtensionID       *int32
	Start             time.Time
	End               time.Time
	Sort              params.SortItems
	Skip              int
	Limit             int
}

type Database interface {
	ListAggregations(ctx context.Context, query AggregationQuery) ([]Aggregation, int, error)
}

// Extensions observes extensions through their owning slice.
type Extensions interface {
	GetExtension(ctx context.Context, id int32) (opengraphschema.Extension, error)
}

type Service struct {
	database   Database
	extensions Extensions
}

func NewService(databaseInterface Database, extensions Extensions) *Service {
	if databaseInterface == nil || extensions == nil {
		panic("dataquality: service requires database and extensions")
	}
	return &Service{database: databaseInterface, extensions: extensions}
}

func (s *Service) ListAggregations(ctx context.Context, query AggregationQuery) ([]Aggregation, int, error) {
	if query.ExtensionID != nil {

		if _, err := s.extensions.GetExtension(ctx, *query.ExtensionID); errors.Is(err, opengraphschema.ErrExtensionNotFound) {
			return nil, 0, ErrExtensionNotFound
		} else if err != nil {
			return nil, 0, err
		}
	}
	return s.database.ListAggregations(ctx, query)
}
