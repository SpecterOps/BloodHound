// Copyright 2026 Specter Ops, Inc.
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

package database

import (
	"context"
	"errors"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/assert"
)

type testInput struct {
	name  string
	value string
}

type testExisting struct {
	id    int32
	name  string
	value string
}

func TestReconcile(t *testing.T) {
	t.Parallel()

	type testData struct {
		name           string
		existing       []testExisting
		inputs         []testInput
		config         reconcileConfig[testInput, testExisting, string]
		expectedResult model.ReconcileResult[testExisting]
		expectedError  error
	}

	newConfig := func() reconcileConfig[testInput, testExisting, string] {
		return reconcileConfig[testInput, testExisting, string]{
			getInputKey:    func(input testInput) string { return input.name },
			getExistingKey: func(existing testExisting) string { return existing.name },
			create: func(_ context.Context, input testInput) (testExisting, error) {
				return testExisting{id: 100, name: input.name, value: "created:" + input.value}, nil
			},
			update: func(_ context.Context, existing testExisting, input testInput) (testExisting, error) {
				existing.value = "updated:" + input.value
				return existing, nil
			},
			delete: func(_ context.Context, _ testExisting) error {
				return nil
			},
		}
	}

	var (
		createError         = errors.New("create failed")
		updateError         = errors.New("update failed")
		deleteError         = errors.New("delete failed")
		createFailureConfig = newConfig()
		updateFailureConfig = newConfig()
		deleteFailureConfig = newConfig()
	)

	createFailureConfig.create = func(_ context.Context, _ testInput) (testExisting, error) {
		return testExisting{}, createError
	}
	updateFailureConfig.update = func(_ context.Context, _ testExisting, _ testInput) (testExisting, error) {
		return testExisting{}, updateError
	}
	deleteFailureConfig.delete = func(_ context.Context, _ testExisting) error {
		return deleteError
	}

	testCases := []testData{
		{
			name:   "success_-_all_inputs_are_new",
			inputs: []testInput{{name: "a", value: "1"}, {name: "b", value: "2"}},
			config: newConfig(),
			expectedResult: model.ReconcileResult[testExisting]{
				Created: []testExisting{
					{id: 100, name: "a", value: "created:1"},
					{id: 100, name: "b", value: "created:2"},
				},
			},
		},
		{
			name:     "success_-_all_inputs_match_and_update_existing",
			existing: []testExisting{{id: 1, name: "a", value: "old"}, {id: 2, name: "b", value: "old"}},
			inputs:   []testInput{{name: "a", value: "new-a"}, {name: "b", value: "new-b"}},
			config:   newConfig(),
			expectedResult: model.ReconcileResult[testExisting]{
				Updated: []testExisting{
					{id: 1, name: "a", value: "updated:new-a"},
					{id: 2, name: "b", value: "updated:new-b"},
				},
			},
		},
		{
			name:     "success_-_mixed_stale_row_deleted_existing_row_updated_new_row_created",
			existing: []testExisting{{id: 10, name: "keep", value: "old"}, {id: 20, name: "stale", value: "old"}},
			inputs:   []testInput{{name: "keep", value: "new"}, {name: "brand-new", value: "1"}},
			config:   newConfig(),
			expectedResult: model.ReconcileResult[testExisting]{
				Created: []testExisting{{id: 100, name: "brand-new", value: "created:1"}},
				Updated: []testExisting{{id: 10, name: "keep", value: "updated:new"}},
				Deleted: []testExisting{{id: 20, name: "stale", value: "old"}},
			},
		},
		{
			name:     "success_-_empty_inputs_all_existing_rows_deleted",
			existing: []testExisting{{id: 1, name: "a"}, {id: 2, name: "b"}},
			inputs:   []testInput{},
			config:   newConfig(),
			expectedResult: model.ReconcileResult[testExisting]{
				Deleted: []testExisting{{id: 1, name: "a"}, {id: 2, name: "b"}},
			},
		},
		{
			name:           "success_-_both_inputs_and_existing_empty_no_operations_performed",
			config:         newConfig(),
			expectedResult: model.ReconcileResult[testExisting]{},
		},
		{
			name:          "error_-_create_fails",
			inputs:        []testInput{{name: "a"}},
			config:        createFailureConfig,
			expectedError: createError,
		},
		{
			name:          "error_-_update_fails",
			existing:      []testExisting{{id: 1, name: "a"}},
			inputs:        []testInput{{name: "a"}},
			config:        updateFailureConfig,
			expectedError: updateError,
		},
		{
			name:          "error_-_delete_fails",
			existing:      []testExisting{{id: 1, name: "stale"}},
			config:        deleteFailureConfig,
			expectedError: deleteError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			result, err := reconcile(context.Background(), testCase.inputs, testCase.existing, testCase.config)
			if testCase.expectedError != nil {
				assert.ErrorIs(t, err, testCase.expectedError)
				return
			}

			assert.NoError(t, err)
			assert.ElementsMatch(t, testCase.expectedResult.Created, result.Created)
			assert.ElementsMatch(t, testCase.expectedResult.Updated, result.Updated)
			assert.ElementsMatch(t, testCase.expectedResult.Deleted, result.Deleted)
		})
	}
}
