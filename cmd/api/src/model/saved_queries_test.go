// Copyright 2023 Specter Ops, Inc.
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

package model_test

import (
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm/utils"
	"testing"
)

func TestSavedQueries_IsSortable(t *testing.T) {
	savedQueries := model.SavedQueries{}
	for _, column := range []string{"user_id", "name", "query", "id", "created_at", "updated_at", "deleted_at"} {
		require.True(t, savedQueries.IsSortable(column))
	}

	require.False(t, savedQueries.IsSortable("foobar"))
}

func TestSavedQueries_ValidFilters(t *testing.T) {
	savedQueries := model.SavedQueries{}
	validFilters := savedQueries.ValidFilters()
	require.Equal(t, 3, len(validFilters))

	for _, column := range []string{"user_id", "name", "query"} {
		operators, ok := validFilters[column]
		require.True(t, ok)
		require.Equal(t, 2, len(operators))
	}
}

func TestSavedQueries_GetValidFilterPredicatesAsStrings(t *testing.T) {
	savedQueries := model.SavedQueries{}
	for _, column := range []string{"user_id", "name", "query"} {
		predicates, err := savedQueries.GetValidFilterPredicatesAsStrings(column)
		require.Nil(t, err)
		require.Equal(t, 2, len(predicates))
		require.True(t, utils.Contains(predicates, string(model.Equals)))
		require.True(t, utils.Contains(predicates, string(model.NotEquals)))
	}
}

func TestSavedQueries_IsString(t *testing.T) {
	savedQueries := model.SavedQueries{}
	for _, column := range []string{"name", "query"} {
		require.True(t, savedQueries.IsString(column))
	}
}
