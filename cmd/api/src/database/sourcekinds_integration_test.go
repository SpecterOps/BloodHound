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
)

func TestRegisterSourceKind(t *testing.T) {
	type args struct {
		sourceKind graph.Kind
	}
	type want struct {
		err        error
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
						ID: 1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
					{
						ID: 2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: Registered harnessEdge.Kind",
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
						ID: 1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
					{
						ID: 2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
					{
						ID: 3,
						Name:   graph.StringKind("harnessEdge.Kind"),
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
				assert.Equal(t, err, testCase.want)
			}

			// Retrieve Source Kinds back to validate registration
			sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(testSuite.Context)
			assert.NoError(t, err)

			assert.Equal(t, sourceKinds, testCase.want.sourceKinds)
		})
	}
}

func TestGetSourceKinds(t *testing.T) {
	type args struct {
		sourceKind graph.Kind
	}
	type want struct {
		err        error
		sourceKinds []database.SourceKind
	}
	tests := []struct {
		name  string
		args  args
		setup func() IntegrationTestSuite
		want  want
	}{
		{
			name: "Success: Retrieves Source Kinds",
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
						ID: 1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
					{
						ID: 2,
						Name:   graph.StringKind("AZBase"),
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
				assert.EqualError(t, testCase.want.err, err.Error())
			} else {
				assert.NoError(t, testCase.want.err)
				assert.Equal(t, sourceKinds, testCase.want.sourceKinds)
			}
		})
	}
}

func TestDeactivateSourceKindsByName(t *testing.T) {
	type args struct {
		sourceKind graph.Kinds
	}
	type want struct {
		err        error
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
						ID: 1,
						Name:   graph.StringKind("Base"),
						Active: true,
					},
					{
						ID: 2,
						Name:   graph.StringKind("AZBase"),
						Active: true,
					},
				},
			},
		},
		{
			name: "Success: Deactivated Base Source Kind - should no longer show in results",
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
						ID: 2,
						Name:   graph.StringKind("AZBase"),
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
				assert.EqualError(t, testCase.want.err, err.Error())
			}

			// Retrieve Source Kinds back to validate deactivation
			sourceKinds, err := testSuite.BHDatabase.GetSourceKinds(testSuite.Context)
			assert.NoError(t, err)

			assert.Equal(t, sourceKinds, testCase.want.sourceKinds)
		})
	}
}
