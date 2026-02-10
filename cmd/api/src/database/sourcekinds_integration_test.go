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

package database_test

import (
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/dawgs/graph"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterSourceKind(t *testing.T) {
	type args struct {
		sourceKind graph.Kind
	}
	type want struct {
		err         error
		sourceKinds []database.SourceKind
	}
	tests := []struct {
		name  string
		args  args
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: Empty Kind",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				sourceKind: graph.StringKind(""),
			},
			want: want{
				err: nil,
				// the v8.0.0 migration initializes the source_kinds table with Base, AZBase, so we're
				// simply testing the default returned source_kinds
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: Register new source kind harnessEdge.Kind",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				sourceKind: graph.StringKind("harnessEdge.Kind"),
			},
			want: want{
				err: nil,
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
					{
						ID:     3,
						Name:   graph.StringKind("harnessEdge.Kind"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: Re-activate inactive source kind",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)

				// Register kind so we can deactivate it
				err := testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("Kind"))
				require.NoError(t, err)

				// Deactivate kind prior to re-activating it
				kind := new(graph.Kinds).Add(graph.StringKind("Kind"))
				require.NoError(t, testSuite.BHDatabase.DeactivateSourceKindsByName(testSuite.Context, kind))
				return testSuite
			},
			args: args{
				sourceKind: graph.StringKind("Kind"),
			},
			want: want{
				err: nil,
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
					{
						ID:     3,
						Name:   graph.StringKind("Kind"),
						Active: true,
					},
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			err := testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(testCase.args.sourceKind)
			if testCase.want.err != nil {
				assert.EqualError(t, err, testCase.want.err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Retrieve Source Kinds back to validate registration
			sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(testSuite.Context)
			assert.NoError(t, err)

			assert.Equal(t, testCase.want.sourceKinds, sourceKinds)
		})
	}
}

func TestGetSourceKinds(t *testing.T) {
	type want struct {
		err         error
		sourceKinds []database.SourceKind
	}
	tests := []struct {
		name  string
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: Retrieves Source Kinds - Ascending order by name",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			want: want{
				err: nil,
				// the v8.0.0 migration initializes the source_kinds table with Base, AZBase, so we're
				// simply testing the default returned source_kinds
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(testSuite.Context)
			if testCase.want.err != nil {
				assert.EqualError(t, err, testCase.want.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want.sourceKinds, sourceKinds)
			}
		})
	}
}

func TestGetSourceKindByName(t *testing.T) {
	type args struct {
		name string
	}
	type want struct {
		err        error
		sourceKind database.SourceKind
	}
	tests := []struct {
		name  string
		args  args
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: Retrieves Source Kinds by Name",
			args: args{
				name: "AZBase",
			},
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			want: want{
				err: nil,
				// the v8.0.0 migration initializes the source_kinds table with Base, AZBase, so we're
				// simply testing the default returned source_kinds
				sourceKind: database.SourceKind{
					ID:     2,
					Name:   graph.StringKind("AZBase"),
					Active: true,
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, testCase.args.name)
			if testCase.want.err != nil {
				assert.EqualError(t, err, testCase.want.err.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, testCase.want.sourceKind, sourceKind)
			}
		})
	}
}

func TestBloodhoundDB_GetSourceKindById(t *testing.T) {
	var (
		testSuite = setupIntegrationTestSuite(t)
	)
	defer teardownIntegrationTestSuite(t, &testSuite)

	tests := []struct {
		name    string
		setup   func(t *testing.T) database.SourceKind
		wantErr error
	}{
		{
			name: "fail - unknown source kind",
			setup: func(t *testing.T) database.SourceKind {
				return database.SourceKind{
					ID: 123,
				}
			},
			wantErr: database.ErrNotFound,
		},
		{
			name: "fail - inactive source kind",
			setup: func(t *testing.T) database.SourceKind {
				err := testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("SourceKind"))
				require.NoError(t, err)
				sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, "SourceKind")
				require.NoError(t, err)
				err = testSuite.BHDatabase.DeactivateSourceKindsByName(testSuite.Context, graph.StringsToKinds([]string{"SourceKind"}))
				require.NoError(t, err)
				return sourceKind
			},
			wantErr: database.ErrNotFound,
		},
		{
			name: "success",
			setup: func(t *testing.T) database.SourceKind {
				err := testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("SourceKind"))
				require.NoError(t, err)
				sourceKind, err := testSuite.BHDatabase.GetSourceKindByName(testSuite.Context, "SourceKind")
				require.NoError(t, err)
				return sourceKind
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createdSourceKind := tt.setup(t)

			if got, err := testSuite.BHDatabase.GetSourceKindById(testSuite.Context, createdSourceKind.ID); tt.wantErr != nil {
				require.EqualErrorf(t, err, tt.wantErr.Error(), "error not equal")
			} else {
				require.NoError(t, err)
				assert.Equal(t, createdSourceKind, got)
			}
		})
	}

}

func TestDeactivateSourceKindsByName(t *testing.T) {
	type args struct {
		sourceKind graph.Kinds
	}
	type want struct {
		err         error
		sourceKinds []database.SourceKind
	}
	tests := []struct {
		name  string
		args  args
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: No kinds - nothing is done",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				sourceKind: graph.Kinds{},
			},
			want: want{
				err: nil,
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: Base excluded from deactivation",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				sourceKind: new(graph.Kinds).Add(graph.StringKind("Base")),
			},
			want: want{
				err: nil,
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: AZBase excluded from deactivation",
			setup: func() IntegrationTestSuite {
				return setupIntegrationTestSuite(t)
			},
			args: args{
				sourceKind: new(graph.Kinds).Add(graph.StringKind("AZBase")),
			},
			want: want{
				err: nil,
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: Deactivate single source kind",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)

				// Register Kind so we can deactivate it
				err := testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("Kind"))
				require.NoError(t, err)

				// Register Kind so we can deactivate it
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("AnotherKind"))
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				sourceKind: new(graph.Kinds).Add(graph.StringKind("Kind")),
			},
			want: want{
				err: nil,
				sourceKinds: []database.SourceKind{
					{
						ID:     4,
						Name:   graph.StringKind("AnotherKind"),
						Active: true,
					},
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: Deactivate multiple source kinds",
			setup: func() IntegrationTestSuite {
				testSuite := setupIntegrationTestSuite(t)

				// Register Kind so we can deactivate it
				err := testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("Kind"))
				require.NoError(t, err)

				// Register Kind so we can deactivate it
				err = testSuite.BHDatabase.RegisterSourceKind(testSuite.Context)(graph.StringKind("AnotherKind"))
				require.NoError(t, err)

				return testSuite
			},
			args: args{
				sourceKind: new(graph.Kinds).Add(graph.StringKind("Kind")).Add(graph.StringKind("AnotherKind")),
			},
			want: want{
				err: nil,
				sourceKinds: []database.SourceKind{
					{
						ID:     2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID:     1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
				},
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			testSuite := testCase.setup()
			defer teardownIntegrationTestSuite(t, &testSuite)

			err := testSuite.BHDatabase.DeactivateSourceKindsByName(testSuite.Context, testCase.args.sourceKind)
			if testCase.want.err != nil {
				assert.EqualError(t, err, testCase.want.err.Error())
			} else {
				assert.NoError(t, err)
			}

			// Retrieve Source Kinds back to validate deactivation
			sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(testSuite.Context)
			assert.NoError(t, err)

			assert.Equal(t, testCase.want.sourceKinds, sourceKinds)
		})
	}
}
