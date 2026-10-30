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

	"github.com/specterops/bloodhound/server/dataquality/internal/services"
	"github.com/specterops/bloodhound/server/dataquality/internal/services/mocks"
	"github.com/specterops/bloodhound/server/opengraphschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestListAggregations(t *testing.T) {
	t.Parallel()
	var (
		ctx             = context.Background()
		extensionID     = int32(42)
		query           = services.AggregationQuery{EnvironmentKindID: 10, ExtensionID: &extensionID, Limit: 1000}
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
			database.EXPECT().ListAggregations(ctx, query).Return(records, 7, nil).Once()
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
			actual, count, err := service.ListAggregations(ctx, query)
			require.ErrorIs(t, err, testCase.err)
			assert.Equal(t, testCase.expected, actual)
			assert.Equal(t, testCase.count, count)
		})
	}
}

func TestListAggregationsWithoutExtension(t *testing.T) {
	t.Parallel()
	var (
		ctx        = context.Background()
		database   = mocks.NewMockDatabase(t)
		extensions = mocks.NewMockExtensions(t)
		query      = services.AggregationQuery{EnvironmentKindID: 10, Limit: 1000}
		expected   = []services.Aggregation{{ID: 1}}
	)
	database.EXPECT().ListAggregations(ctx, query).Return(expected, 1, nil).Once()
	records, count, err := services.NewService(database, extensions).ListAggregations(ctx, query)
	require.NoError(t, err)
	assert.Equal(t, expected, records)
	assert.Equal(t, 1, count)
}
