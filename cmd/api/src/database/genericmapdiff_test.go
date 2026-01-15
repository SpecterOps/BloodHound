// Copyright 2025 Specter Ops, Inc.
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
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

type testMapDiffStruct struct {
	foo string
	bar int
	z   string
}

func swapZTestFunc(src, dst *testMapDiffStruct) {
	src.z = dst.z
}

func TestDiffMapsToSyncActions(t *testing.T) {
	var (
		testStruct1       = testMapDiffStruct{foo: "foo_1", bar: 1, z: "z_1"}
		testStruct2       = testMapDiffStruct{foo: "foo_2", bar: 2, z: "z_2"}
		updatedTestStruct = testMapDiffStruct{foo: "foo_1", bar: 4, z: "z_4"}
		testStruct3       = testMapDiffStruct{foo: "foo_3", bar: 3, z: "z_3"}
	)

	type args[K comparable, V any] struct {
		dst      map[K]V
		src      map[K]V
		onUpsert func(*V, *V)
	}
	type testCase[K comparable, V any] struct {
		name string
		args args[K, V]
		want MapDiffActions[V]
	}
	tests := []testCase[string, testMapDiffStruct]{
		{
			name: "empty src",
			args: args[string, testMapDiffStruct]{
				dst: map[string]testMapDiffStruct{
					testStruct1.foo: testStruct1,
					testStruct2.foo: testStruct2,
				},
				src:      map[string]testMapDiffStruct{},
				onUpsert: swapZTestFunc,
			},
			want: MapDiffActions[testMapDiffStruct]{
				ItemsToDelete: []testMapDiffStruct{testStruct1, testStruct2},
				ItemsToUpdate: []testMapDiffStruct{},
				ItemsToInsert: []testMapDiffStruct{},
			},
		},
		{
			name: "empty dst",
			args: args[string, testMapDiffStruct]{
				dst: map[string]testMapDiffStruct{},
				src: map[string]testMapDiffStruct{
					updatedTestStruct.foo: updatedTestStruct,
					testStruct3.foo:       testStruct3,
				},
				onUpsert: swapZTestFunc,
			},
			want: MapDiffActions[testMapDiffStruct]{
				ItemsToDelete: []testMapDiffStruct{},
				ItemsToUpdate: []testMapDiffStruct{},
				ItemsToInsert: []testMapDiffStruct{updatedTestStruct, testStruct3},
			},
		},
		{
			name: "success - convertGraphSchemaNodeKinds",
			args: args[string, testMapDiffStruct]{
				dst: map[string]testMapDiffStruct{
					testStruct1.foo: testStruct1,
					testStruct2.foo: testStruct2,
				},
				src: map[string]testMapDiffStruct{
					updatedTestStruct.foo: updatedTestStruct,
					testStruct3.foo:       testStruct3,
				},
				onUpsert: swapZTestFunc,
			},
			want: MapDiffActions[testMapDiffStruct]{
				ItemsToDelete: []testMapDiffStruct{testStruct2},
				ItemsToUpdate: []testMapDiffStruct{{foo: "foo_1", bar: 4, z: "z_1"}},
				ItemsToInsert: []testMapDiffStruct{testStruct3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateMapDiffActions(tt.args.src, tt.args.dst, tt.args.onUpsert)
			compareTestMapDiffStructs(t, got.ItemsToInsert, tt.want.ItemsToInsert)
			compareTestMapDiffStructs(t, got.ItemsToUpdate, tt.want.ItemsToUpdate)
			compareTestMapDiffStructs(t, got.ItemsToDelete, tt.want.ItemsToDelete)
		})
	}
}

func compareTestMapDiffStructs(t *testing.T, got, want []testMapDiffStruct) {
	t.Helper()
	require.Equalf(t, len(want), len(got), "length mismatch of GraphSchemaEdgeKinds")
	require.ElementsMatchf(t, want, got, "GraphSchemaEdgeKinds mismatch")
}

func TestHandleMapDiffAction(t *testing.T) {
	var (
		testStruct1 = testMapDiffStruct{foo: "foo_1", bar: 1, z: "z_1"}
		testStruct2 = testMapDiffStruct{foo: "foo_2", bar: 2, z: "z_2"}
		testStruct3 = testMapDiffStruct{foo: "foo_3", bar: 3, z: "z_3"}
	)

	type args[V any] struct {
		ctx        context.Context
		actions    MapDiffActions[testMapDiffStruct]
		deleteFunc func(context.Context, V) error
		updateFunc func(context.Context, V) (V, error)
		insertFunc func(context.Context, V) (V, error)
	}
	type testCase[V any] struct {
		name    string
		args    args[V]
		wantErr error
		want    []testMapDiffStruct
	}
	tests := []testCase[testMapDiffStruct]{
		{
			name: "fail - error during delete func",
			args: args[testMapDiffStruct]{
				ctx: context.Background(),
				actions: MapDiffActions[testMapDiffStruct]{
					ItemsToDelete: []testMapDiffStruct{testStruct1},
					ItemsToUpdate: []testMapDiffStruct{testStruct2},
					ItemsToInsert: []testMapDiffStruct{testStruct3},
				},
				deleteFunc: testDeleteMapDiffStructFuncError,
				updateFunc: testMapDiffStructFunc,
				insertFunc: testMapDiffStructFunc,
			},
			wantErr: fmt.Errorf("test map diff func error")},
		{
			name: "fail - error during update func",
			args: args[testMapDiffStruct]{
				ctx: context.Background(),
				actions: MapDiffActions[testMapDiffStruct]{
					ItemsToDelete: []testMapDiffStruct{testStruct1},
					ItemsToUpdate: []testMapDiffStruct{testStruct2},
					ItemsToInsert: []testMapDiffStruct{testStruct3},
				},
				deleteFunc: testDeleteMapDiffStructFunc,
				updateFunc: testMapDiffStructFuncError,
				insertFunc: testMapDiffStructFunc,
			},
			wantErr: fmt.Errorf("test map diff func error")},
		{
			name: "fail - error during insert func",
			args: args[testMapDiffStruct]{
				ctx: context.Background(),
				actions: MapDiffActions[testMapDiffStruct]{
					ItemsToDelete: []testMapDiffStruct{testStruct1},
					ItemsToUpdate: []testMapDiffStruct{testStruct2},
					ItemsToInsert: []testMapDiffStruct{testStruct3},
				},
				deleteFunc: testDeleteMapDiffStructFunc,
				updateFunc: testMapDiffStructFunc,
				insertFunc: testMapDiffStructFuncError,
			},
			wantErr: fmt.Errorf("test map diff func error"),
		},
		{
			name: "success",
			args: args[testMapDiffStruct]{
				ctx: context.Background(),
				actions: MapDiffActions[testMapDiffStruct]{
					ItemsToDelete: []testMapDiffStruct{testStruct1},
					ItemsToUpdate: []testMapDiffStruct{testStruct2},
					ItemsToInsert: []testMapDiffStruct{testStruct3},
				},
				deleteFunc: testDeleteMapDiffStructFunc,
				updateFunc: testMapDiffStructFunc,
				insertFunc: testMapDiffStructFunc,
			},
			wantErr: nil,
			want:    []testMapDiffStruct{testStruct2, testStruct3},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if updatedItems, err := HandleMapDiffAction(tt.args.ctx, tt.args.actions, tt.args.deleteFunc,
				tt.args.updateFunc, tt.args.insertFunc); tt.wantErr != nil {
				require.EqualErrorf(t, err, tt.wantErr.Error(), "HandleMapDiffAction(%v, %v)", tt.args.ctx, tt.args.actions)
			} else {
				require.NoError(t, err)
				require.Equalf(t, updatedItems, tt.want, "HandleMapDiffAction(%v, %v)", tt.args.ctx, tt.args.actions)
			}
		})
	}
}

func testDeleteMapDiffStructFunc(_ context.Context, t testMapDiffStruct) error {
	return nil
}
func testDeleteMapDiffStructFuncError(_ context.Context, t testMapDiffStruct) error {
	return fmt.Errorf("test map diff func error")
}

func testMapDiffStructFunc(_ context.Context, t testMapDiffStruct) (testMapDiffStruct, error) {
	return t, nil
}

func testMapDiffStructFuncError(_ context.Context, t testMapDiffStruct) (testMapDiffStruct, error) {
	return t, fmt.Errorf("test map diff func error")
}
