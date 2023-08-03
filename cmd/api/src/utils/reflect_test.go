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

package utils_test

import (
	"reflect"
	"testing"

	"github.com/specterops/bloodhound/src/utils"
)

func TestGenericClone(t *testing.T) {

	type Foo struct {
		Bar string
	}

	var cloneMap func(map[string]string) map[string]string
	utils.Bind(&cloneMap, utils.GenericClone)
	var cloneSlice func([]string) []string
	utils.Bind(&cloneSlice, utils.GenericClone)
	var cloneStruct func(Foo) Foo
	utils.Bind(&cloneStruct, utils.GenericClone)

	m1 := map[string]string{"foo": "bar"}
	m2 := cloneMap(m1)

	if !reflect.DeepEqual(m1, m2) {
		t.Errorf("expected map values to be equal")
	}
	if &m1 == &m2 {
		t.Errorf("expected map addresses to be different")
	}

	s1 := []string{"foo", "bar"}
	s2 := cloneSlice(s1)
	if !reflect.DeepEqual(s1, s2) {
		t.Errorf("expected slice values to be equal")
	}
	if &s1 == &s2 {
		t.Errorf("expected slice addresses to be different")
	}

	f1 := Foo{"baz"}
	f2 := cloneStruct(f1)

	if f1 != f2 {
		t.Errorf("expected slice values to be equal")
	}
	if &f1 == &f2 {
		t.Errorf("expected slice addresses to be different")
	}
}

func TestGenericAssign(t *testing.T) {

	var assignMaps func(...map[string]string) map[string]string
	utils.Bind(&assignMaps, utils.GenericAssign)

	m1 := map[string]string{"foo": "bar"}
	m2 := map[string]string{"baz": "buzz"}
	m3 := map[string]string{"bar": "bazinga"}
	expectedMap := map[string]string{
		"foo": "bar",
		"baz": "buzz",
		"bar": "bazinga",
	}
	actualMap := assignMaps(m1, m2, m3)

	if !reflect.DeepEqual(actualMap, expectedMap) {
		t.Errorf("failed to correctly assign maps: got %v, want %v", actualMap, expectedMap)
	}

	var assignSlices func(...[]any) []any
	utils.Bind(&assignSlices, utils.GenericAssign)

	s1 := []any{"foo", 1}
	s2 := []any{nil, 2, struct{ Foo string }{}}
	s3 := []any{nil, nil, nil, 3}
	expectedSlice := []any{"foo", 2, struct{ Foo string }{}, 3}
	actualSlice := assignSlices(s1, s2, s3)

	if !reflect.DeepEqual(actualSlice, expectedSlice) {
		t.Errorf("failed to correctly assign slices: got %v, want %v", actualSlice, expectedSlice)
	}

	type Foo struct {
		Bar, Baz, Buzz string
	}

	var assignStructs func(...Foo) Foo
	utils.Bind(&assignStructs, utils.GenericAssign)

	f1 := Foo{Bar: "Baz"}
	f2 := Foo{Baz: "bazzle"}
	f3 := Foo{Buzz: "buzzle"}
	expectedFoo := Foo{
		Bar:  "Baz",
		Baz:  "bazzle",
		Buzz: "buzzle",
	}
	actualFoo := assignStructs(f1, f2, f3)

	if actualFoo != expectedFoo {
		t.Errorf("failed to correctly assign structs: got %v, want %v", actualFoo, expectedFoo)
	}
}

func TestGenericMap(t *testing.T) {

	expected := []any{nil, nil, nil, nil, nil}

	var mapArray0 func(collection []int, iteratee func() any) []any
	var mapArray1 func(collection []int, iteratee func(value int) any) []any
	var mapArray2 func(collection []int, iteratee func(value int, idx int) any) []any
	var mapArray3 func(collection []int, iteratee func(value int, idx int, collection []int) any) []any
	utils.Bind(&mapArray0, utils.GenericMap)
	utils.Bind(&mapArray1, utils.GenericMap)
	utils.Bind(&mapArray2, utils.GenericMap)
	utils.Bind(&mapArray3, utils.GenericMap)

	ints := []int{1, 2, 3, 4, 5}

	actual0 := mapArray0(ints, func() any { return nil })
	actual1 := mapArray1(ints, func(value int) any { return nil })
	actual2 := mapArray2(ints, func(value int, idx int) any { return nil })
	actual3 := mapArray3(ints, func(value int, idx int, collection []int) any { return nil })

	if !reflect.DeepEqual(actual0, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual0, expected)
	}

	if !reflect.DeepEqual(actual1, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual0, expected)
	}

	if !reflect.DeepEqual(actual2, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual0, expected)
	}

	if !reflect.DeepEqual(actual3, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual0, expected)
	}

	var mapMap0 func(collection map[string]int, iteratee func() any) []any
	var mapMap1 func(collection map[string]int, iteratee func(value int) any) []any
	var mapMap2 func(collection map[string]int, iteratee func(value int, key string) any) []any
	var mapMap3 func(collection map[string]int, iteratee func(value int, key string, collection map[string]int) any) []any
	utils.Bind(&mapMap0, utils.GenericMap)
	utils.Bind(&mapMap1, utils.GenericMap)
	utils.Bind(&mapMap2, utils.GenericMap)
	utils.Bind(&mapMap3, utils.GenericMap)

	collection := map[string]int{
		"a": 1,
		"b": 2,
		"c": 3,
		"d": 4,
		"e": 5,
	}

	actualMap0 := mapMap0(collection, func() any { return nil })
	actualMap1 := mapMap1(collection, func(value int) any { return nil })
	actualMap2 := mapMap2(collection, func(value int, key string) any { return nil })
	actualMap3 := mapMap3(collection, func(value int, kye string, collection map[string]int) any { return nil })

	if !reflect.DeepEqual(actualMap0, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual2, expected)
	}

	if !reflect.DeepEqual(actualMap1, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual2, expected)
	}

	if !reflect.DeepEqual(actualMap2, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual2, expected)
	}

	if !reflect.DeepEqual(actualMap3, expected) {
		t.Errorf("failed to correctly map ints: got %v, want %v", actual2, expected)
	}
}
