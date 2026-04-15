package database

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		name     string
		existing []testExisting
		inputs   []testInput
		// override callbacks to inject errors; nil means use the default tracking impl
		createErr error
		updateErr error
		deleteErr error
		assert    func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting)
	}

	tt := []testData{
		{
			name:     "Success: all inputs are new - all created, none deleted or updated",
			existing: nil,
			inputs:   []testInput{{name: "a", value: "1"}, {name: "b", value: "2"}},
			assert: func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, results, 2)
				assert.Empty(t, deleted)
				assert.Len(t, created, 2)
				assert.Empty(t, updated)
			},
		},
		{
			name:     "Success: all inputs match existing - all updated, none created or deleted",
			existing: []testExisting{{id: 1, name: "a", value: "old"}, {id: 2, name: "b", value: "old"}},
			inputs:   []testInput{{name: "a", value: "new"}, {name: "b", value: "new"}},
			assert: func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, results, 2)
				assert.Empty(t, deleted)
				assert.Empty(t, created)
				assert.Len(t, updated, 2)
				assert.Equal(t, "new", results[0].value)
				assert.Equal(t, "new", results[1].value)
			},
		},
		{
			name:     "Success: mixed - stale row deleted, existing row updated, new row created",
			existing: []testExisting{{id: 10, name: "keep", value: "old"}, {id: 20, name: "stale", value: "old"}},
			inputs:   []testInput{{name: "keep", value: "new"}, {name: "brand-new", value: "1"}},
			assert: func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting) {
				t.Helper()
				require.NoError(t, err)
				assert.Len(t, results, 2)
				assert.Equal(t, []testExisting{{id: 20, name: "stale", value: "old"}}, deleted)
				assert.Len(t, created, 1)
				assert.Equal(t, "brand-new", created[0].name)
				assert.Len(t, updated, 1)
				assert.Equal(t, "new", updated[0].value)
			},
		},
		{
			name:     "Success: empty inputs - all existing rows deleted",
			existing: []testExisting{{id: 1, name: "a"}, {id: 2, name: "b"}},
			inputs:   []testInput{},
			assert: func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting) {
				t.Helper()
				require.NoError(t, err)
				assert.Empty(t, results)
				assert.Equal(t, []testExisting{{id: 1, name: "a"}, {id: 2, name: "b"}}, deleted)
				assert.Empty(t, created)
				assert.Empty(t, updated)
			},
		},
		{
			name:      "Error: create fails - error propagates",
			createErr: errors.New("create failed"),
			inputs:    []testInput{{name: "a"}},
			assert: func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting) {
				t.Helper()
				require.ErrorContains(t, err, "create failed")
				assert.Nil(t, results)
			},
		},
		{
			name:      "Error: update fails - error propagates",
			existing:  []testExisting{{id: 1, name: "a"}},
			updateErr: errors.New("update failed"),
			inputs:    []testInput{{name: "a", value: "new"}},
			assert: func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting) {
				t.Helper()
				require.ErrorContains(t, err, "update failed")
				assert.Nil(t, results)
			},
		},
		{
			name:      "Error: delete fails - error propagates",
			existing:  []testExisting{{id: 1, name: "stale"}},
			deleteErr: errors.New("delete failed"),
			inputs:    []testInput{},
			assert: func(t *testing.T, results []testExisting, err error, deleted []testExisting, created []testInput, updated []testExisting) {
				t.Helper()
				require.ErrorContains(t, err, "delete failed")
				assert.Nil(t, results)
			},
		},
	}

	for _, testCase := range tt {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			var (
				deleted []testExisting
				created []testInput
				updated []testExisting
			)

			config := reconcileConfig[testInput, testExisting, string]{
				getInputKey:    func(input testInput) string { return input.name },
				getExistingKey: func(existing testExisting) string { return existing.name },
				create: func(ctx context.Context, input testInput) (testExisting, error) {
					if testCase.createErr != nil {
						return testExisting{}, testCase.createErr
					}
					created = append(created, input)
					return testExisting{id: int32(len(created)), name: input.name, value: input.value}, nil
				},
				update: func(ctx context.Context, existing testExisting, input testInput) (testExisting, error) {
					if testCase.updateErr != nil {
						return testExisting{}, testCase.updateErr
					}
					existing.value = input.value
					updated = append(updated, existing)
					return existing, nil
				},
				delete: func(ctx context.Context, existing testExisting) error {
					if testCase.deleteErr != nil {
						return testCase.deleteErr
					}
					deleted = append(deleted, existing)
					return nil
				},
			}

			results, err := reconcile(context.Background(), testCase.inputs, testCase.existing, config)
			testCase.assert(t, results, err, deleted, created, updated)
		})
	}
}
