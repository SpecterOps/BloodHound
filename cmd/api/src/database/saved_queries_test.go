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

//go:build integration
// +build integration

package database_test

import (
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/src/model"
	"github.com/specterops/bloodhound/src/test/integration"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSavedQueries_ListSavedQueries(t *testing.T) {
	var (
		dbInst = integration.OpenDatabase(t)

		savedQueriesFilter = model.QueryParameterFilter{
			Name:         "id",
			Operator:     model.GreaterThan,
			Value:        "4",
			IsStringData: false,
		}
		savedQueriesFilterMap = model.QueryParameterFilterMap{savedQueriesFilter.Name: model.QueryParameterFilters{savedQueriesFilter}}
	)

	if err := integration.Prepare(dbInst); err != nil {
		t.Fatalf("Failed preparing DB: %v", err)
	}

	userUUID, err := uuid.NewV4()
	require.Nil(t, err)

	for i := 0; i < 7; i++ {
		if _, err := dbInst.CreateSavedQuery(userUUID, fmt.Sprintf("saved_query_%d", i), ""); err != nil {
			t.Fatalf("Error creating audit log: %v", err)
		}
	}

	if _, count, err := dbInst.ListSavedQueries(userUUID, "", model.SQLFilter{}, 0, 10); err != nil {
		t.Fatalf("Failed to list all saved queries: %v", err)
	} else if count != 7 {
		t.Fatalf("Expected 7 saved queries to be returned")
	} else if filter, err := savedQueriesFilterMap.BuildSQLFilter(); err != nil {
		t.Fatalf("Failed to generate SQL Filter: %v", err)
		// Limit is set to 1 to verify that count is total filtered count, not response size
	} else if _, count, err = dbInst.ListSavedQueries(userUUID, "", filter, 0, 1); err != nil {
		t.Fatalf("Failed to list filtered saved queries: %v", err)
	} else if count != 3 {
		t.Fatalf("Expected 3 saved queries to be returned")
	}
}
