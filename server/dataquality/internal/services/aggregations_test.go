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

package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/packages/go/params"

	"github.com/specterops/bloodhound/server/dataquality/internal/services"
	"github.com/specterops/bloodhound/server/dataquality/internal/services/mocks"
	"github.com/specterops/bloodhound/server/opengraphschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestServices_ListAggregations(t *testing.T) {
	t.Parallel()
	var (
		ctx             = context.Background()
		extensionID     = int32(42)
		query           = params.Filters{"schema_extension_id": {{Field: "schema_extension_id", Operator: params.Equals, Value: "42"}}}
		records         = []services.Aggregation{{ID: 1, SchemaExtensionID: extensionID}}
		unexpectedError = errors.New("extension lookup failed")
	)
	testCases := []struct {
		name      string
		setupMock func(*mocks.MockDatabase, *mocks.MockExtensions)
		expected  []services.Aggregation
		count     int
		err       error
	}{
		{name: "existing extension", setupMock: func(database *mocks.MockDatabase, extensions *mocks.MockExtensions) {
			extensions.EXPECT().GetExtension(ctx, extensionID).Return(opengraphschema.Extension{ID: extensionID}, nil).Once()
			database.EXPECT().ListAggregations(ctx, query, params.SortItems(nil), 0, 1000).Return(records, 7, nil).Once()
		}, expected: records, count: 7},
		{name: "missing extension", setupMock: func(database *mocks.MockDatabase, extensions *mocks.MockExtensions) {
			extensions.EXPECT().GetExtension(ctx, extensionID).Return(opengraphschema.Extension{}, opengraphschema.ErrExtensionNotFound).Once()
		}, err: services.ErrExtensionNotFound},
		{name: "extension lookup fails", setupMock: func(database *mocks.MockDatabase, extensions *mocks.MockExtensions) {
			extensions.EXPECT().GetExtension(ctx, extensionID).Return(opengraphschema.Extension{}, unexpectedError).Once()
		}, err: unexpectedError},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				database   = mocks.NewMockDatabase(t)
				extensions = mocks.NewMockExtensions(t)
				service    = services.NewService(database, extensions)
			)
			testCase.setupMock(database, extensions)
			actual, count, err := service.ListAggregations(ctx, query, params.SortItems(nil), 0, 1000)
			require.ErrorIs(t, err, testCase.err)
			assert.Equal(t, testCase.expected, actual)
			assert.Equal(t, testCase.count, count)
		})
	}
}

func TestServices_ListAggregations_WithoutExtension(t *testing.T) {
	t.Parallel()
	var (
		ctx        = context.Background()
		database   = mocks.NewMockDatabase(t)
		extensions = mocks.NewMockExtensions(t)
		query      = params.Filters{"schema_environment_kind_id": {{Field: "schema_environment_kind_id", Operator: params.Equals, Value: "10"}}}
		expected   = []services.Aggregation{{ID: 1}}
	)
	database.EXPECT().ListAggregations(ctx, query, params.SortItems(nil), 0, 1000).Return(expected, 1, nil).Once()
	records, count, err := services.NewService(database, extensions).ListAggregations(ctx, query, params.SortItems(nil), 0, 1000)
	require.NoError(t, err)
	assert.Equal(t, expected, records)
	assert.Equal(t, 1, count)
}

func TestListAggregationsRejectsInvalidExtensionFilters(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name   string
		filter params.Filter
	}{
		{name: "invalid ID", filter: params.Filter{Field: "schema_extension_id", Operator: params.Equals, Value: "bad"}},
		{name: "ID overflow", filter: params.Filter{Field: "schema_extension_id", Operator: params.Equals, Value: "2147483648"}},
		{name: "unsupported operator", filter: params.Filter{Field: "schema_extension_id", Operator: params.NotEquals, Value: "42"}},
		{name: "OR filter", filter: params.Filter{Field: "schema_extension_id", Operator: params.Equals, Value: "42", SetOperator: params.FilterOr}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			var (
				database   = mocks.NewMockDatabase(t)
				extensions = mocks.NewMockExtensions(t)
				service    = services.NewService(database, extensions)
			)
			records, count, err := service.ListAggregations(context.Background(), params.Filters{"schema_extension_id": {testCase.filter}}, nil, 0, 1000)
			require.ErrorIs(t, err, services.ErrInvalidQuery)
			assert.Nil(t, records)
			assert.Zero(t, count)
		})
	}
}
