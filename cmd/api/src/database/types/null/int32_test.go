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

package null

import (
	"encoding/json"
	"math"
	"strconv"
	"testing"

	"github.com/specterops/bloodhound/errors"
)

var (
	int32JSON       = []byte(`12345`)
	int32StringJSON = []byte(`"12345"`)
	nullInt32JSON   = []byte(`{"Int32":12345,"Valid":true}`)
)

func TestIntFrom(t *testing.T) {
	i := Int32From(12345)
	assertInt(t, i, "Int32From()")

	zero := Int32From(0)
	if !zero.Valid {
		t.Error("Int32From(0)", "is invalid, but should be valid")
	}
}

func TestIntFromPtr(t *testing.T) {
	n := int32(12345)
	iptr := &n
	i := Int32FromPtr(iptr)
	assertInt(t, i, "Int32FromPtr()")

	null := Int32FromPtr(nil)
	assertNullInt(t, null, "Int32FromPtr(nil)")
}

func TestUnmarshalInt(t *testing.T) {
	var i Int32
	err := json.Unmarshal(int32JSON, &i)
	maybePanic(err)
	assertInt(t, i, "int json")

	var si Int32
	err = json.Unmarshal(int32StringJSON, &si)
	maybePanic(err)
	assertInt(t, si, "int string json")

	var ni Int32
	err = json.Unmarshal(nullInt32JSON, &ni)
	if err == nil {
		panic("err should not be nill")
	}

	var bi Int32
	err = json.Unmarshal(floatBlankJSON, &bi)
	if err == nil {
		panic("err should not be nill")
	}

	var null Int32
	err = json.Unmarshal(nullJSON, &null)
	maybePanic(err)
	assertNullInt(t, null, "null json")

	var badType Int32
	err = json.Unmarshal(boolJSON, &badType)
	if err == nil {
		panic("err should not be nil")
	}
	assertNullInt(t, badType, "wrong type json")

	var invalid Int32
	err = invalid.UnmarshalJSON(invalidJSON)
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Errorf("expected wrapped json.SyntaxError, not %T", err)
	}
	assertNullInt(t, invalid, "invalid json")
}

func TestUnmarshalNonIntegerNumber(t *testing.T) {
	var i Int32
	err := json.Unmarshal(floatJSON, &i)
	if err == nil {
		panic("err should be present; non-integer number coerced to int")
	}
}

func TestUnmarshalInt64Overflow(t *testing.T) {
	int64Overflow := uint64(math.MaxInt32)

	// Max int64 should decode successfully
	var i Int32
	err := json.Unmarshal([]byte(strconv.FormatUint(int64Overflow, 10)), &i)
	maybePanic(err)

	// Attempt to overflow
	int64Overflow++
	err = json.Unmarshal([]byte(strconv.FormatUint(int64Overflow, 10)), &i)
	if err == nil {
		panic("err should be present; decoded value overflows int64")
	}
}

func TestTextUnmarshalInt(t *testing.T) {
	var i Int32
	err := i.UnmarshalText([]byte("12345"))
	maybePanic(err)
	assertInt(t, i, "UnmarshalText() int")

	var blank Int32
	err = blank.UnmarshalText([]byte(""))
	maybePanic(err)
	assertNullInt(t, blank, "UnmarshalText() empty int")

	var null Int32
	err = null.UnmarshalText([]byte("null"))
	maybePanic(err)
	assertNullInt(t, null, `UnmarshalText() "null"`)

	var invalid Int32
	err = invalid.UnmarshalText([]byte("hello world"))
	if err == nil {
		panic("expected error")
	}
}

func TestMarshalInt(t *testing.T) {
	i := Int32From(12345)
	data, err := json.Marshal(i)
	maybePanic(err)
	assertJSONEquals(t, data, "12345", "non-empty json marshal")

	// invalid values should be encoded as null
	null := NewInt32(0, false)
	data, err = json.Marshal(null)
	maybePanic(err)
	assertJSONEquals(t, data, "null", "null json marshal")
}

func TestMarshalIntText(t *testing.T) {
	i := Int32From(12345)
	data, err := i.MarshalText()
	maybePanic(err)
	assertJSONEquals(t, data, "12345", "non-empty text marshal")

	// invalid values should be encoded as null
	null := NewInt32(0, false)
	data, err = null.MarshalText()
	maybePanic(err)
	assertJSONEquals(t, data, "", "null text marshal")
}

func TestIntPointer(t *testing.T) {
	i := Int32From(12345)
	ptr := i.Ptr()
	if *ptr != 12345 {
		t.Errorf("bad %s int: %#v ≠ %d\n", "pointer", ptr, 12345)
	}

	null := NewInt32(0, false)
	ptr = null.Ptr()
	if ptr != nil {
		t.Errorf("bad %s int: %#v ≠ %s\n", "nil pointer", ptr, "nil")
	}
}

func TestIntIsZero(t *testing.T) {
	i := Int32From(12345)
	if i.IsZero() {
		t.Errorf("IsZero() should be false")
	}

	null := NewInt32(0, false)
	if !null.IsZero() {
		t.Errorf("IsZero() should be true")
	}

	zero := NewInt32(0, true)
	if zero.IsZero() {
		t.Errorf("IsZero() should be false")
	}
}

func TestIntSetValid(t *testing.T) {
	change := NewInt32(0, false)
	assertNullInt(t, change, "SetValid()")
	change.SetValid(12345)
	assertInt(t, change, "SetValid()")
}

func TestIntScan(t *testing.T) {
	var i Int32
	err := i.Scan(12345)
	maybePanic(err)
	assertInt(t, i, "scanned int")

	var null Int32
	err = null.Scan(nil)
	maybePanic(err)
	assertNullInt(t, null, "scanned null")
}

func TestIntValueOrZero(t *testing.T) {
	valid := NewInt32(12345, true)
	if valid.ValueOrZero() != 12345 {
		t.Error("unexpected ValueOrZero", valid.ValueOrZero())
	}

	invalid := NewInt32(12345, false)
	if invalid.ValueOrZero() != 0 {
		t.Error("unexpected ValueOrZero", invalid.ValueOrZero())
	}
}

func TestIntEqual(t *testing.T) {
	int1 := NewInt32(10, false)
	int2 := NewInt32(10, false)
	assertIntEqualIsTrue(t, int1, int2)

	int1 = NewInt32(10, false)
	int2 = NewInt32(20, false)
	assertIntEqualIsTrue(t, int1, int2)

	int1 = NewInt32(10, true)
	int2 = NewInt32(10, true)
	assertIntEqualIsTrue(t, int1, int2)

	int1 = NewInt32(10, true)
	int2 = NewInt32(10, false)
	assertIntEqualIsFalse(t, int1, int2)

	int1 = NewInt32(10, false)
	int2 = NewInt32(10, true)
	assertIntEqualIsFalse(t, int1, int2)

	int1 = NewInt32(10, true)
	int2 = NewInt32(20, true)
	assertIntEqualIsFalse(t, int1, int2)
}

func assertInt(t *testing.T, i Int32, from string) {
	if i.Int32 != 12345 {
		t.Errorf("bad %s int: %d ≠ %d\n", from, i.Int32, 12345)
	}
	if !i.Valid {
		t.Error(from, "is invalid, but should be valid")
	}
}

func assertNullInt(t *testing.T, i Int32, from string) {
	if i.Valid {
		t.Error(from, "is valid, but should be invalid")
	}
}

func assertIntEqualIsTrue(t *testing.T, a, b Int32) {
	t.Helper()
	if !a.Equal(b) {
		t.Errorf("Equal() of Int32{%v, Valid:%t} and Int32{%v, Valid:%t} should return true", a.Int32, a.Valid, b.Int32, b.Valid)
	}
}

func assertIntEqualIsFalse(t *testing.T, a, b Int32) {
	t.Helper()
	if a.Equal(b) {
		t.Errorf("Equal() of Int32{%v, Valid:%t} and Int32{%v, Valid:%t} should return false", a.Int32, a.Valid, b.Int32, b.Valid)
	}
}
