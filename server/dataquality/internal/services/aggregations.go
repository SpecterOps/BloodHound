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

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/specterops/bloodhound/packages/go/params"
	"github.com/specterops/bloodhound/server/opengraphschema"
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

var (
	ErrExtensionNotFound = errors.New("extension not found")
	ErrInvalidQuery      = errors.New("invalid aggregation query")
)

func (s *Service) ListAggregations(ctx context.Context, queryFilters params.Filters, sortItems params.SortItems, skip, limit int) ([]Aggregation, int, error) {
	for _, filter := range queryFilters["schema_extension_id"] {
		extensionID, err := strconv.ParseInt(filter.Value, 10, 32)
		if err != nil || filter.Operator != params.Equals || (filter.SetOperator != "" && filter.SetOperator != params.FilterAnd) {
			return nil, 0, ErrInvalidQuery
		}
		if _, err := s.openGraphSchema.GetExtension(ctx, int32(extensionID)); errors.Is(err, opengraphschema.ErrExtensionNotFound) {
			return nil, 0, ErrExtensionNotFound
		} else if err != nil {
			return nil, 0, err
		}
	}
	return s.database.ListAggregations(ctx, queryFilters, sortItems, skip, limit)
}
