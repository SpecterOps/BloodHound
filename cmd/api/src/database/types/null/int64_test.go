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
	"errors"
	"math"
	"strconv"
	"testing"
)

var (
	int64JSON       = []byte(`12345`)
	int64StringJSON = []byte(`"12345"`)
	nullInt64JSON   = []byte(`{"Int64":12345,"Valid":true}`)
)

func TestInt64From(t *testing.T) {
	i := Int64From(12345)
	assertInt64(t, i, "Int64From()")

	zero := Int64From(0)
	if !zero.Valid {
		t.Error("Int64From(0)", "is invalid, but should be valid")
	}
}

func TestInt64FromPtr(t *testing.T) {
	n := int64(12345)
	iptr := &n
	i := Int64FromPtr(iptr)
	assertInt64(t, i, "Int64FromPtr()")

	null := Int64FromPtr(nil)
	assertNullInt64(t, null, "Int64FromPtr(nil)")
}

func TestUnmarshalInt64(t *testing.T) {
	var i Int64
	err := json.Unmarshal(int64JSON, &i)
	maybePanic(err)
	assertInt64(t, i, "int json")

	var si Int64
	err = json.Unmarshal(int64StringJSON, &si)
	maybePanic(err)
	assertInt64(t, si, "int string json")

	var ni Int64
	err = json.Unmarshal(nullInt64JSON, &ni)
	if err == nil {
		panic("err should not be nill")
	}

	var bi Int64
	err = json.Unmarshal(floatBlankJSON, &bi)
	if err == nil {
		panic("err should not be nill")
	}

	var null Int64
	err = json.Unmarshal(nullJSON, &null)
	maybePanic(err)
	assertNullInt64(t, null, "null json")

	var badType Int64
	err = json.Unmarshal(boolJSON, &badType)
	if err == nil {
		panic("err should not be nil")
	}
	assertNullInt64(t, badType, "wrong type json")

	var invalid Int64
	err = invalid.UnmarshalJSON(invalidJSON)
	var syntaxError *json.SyntaxError
	if !errors.As(err, &syntaxError) {
		t.Errorf("expected wrapped json.SyntaxError, not %T", err)
	}
	assertNullInt64(t, invalid, "invalid json")
}

func TestUnmarshalNonIntegerNumberToInt64(t *testing.T) {
	var i Int64
	err := json.Unmarshal(floatJSON, &i)
	if err == nil {
		panic("err should be present; non-integer number coerced to int")
	}
}

func TestUnmarshalInt64OverflowToInt64(t *testing.T) {
	int64Overflow := uint64(math.MaxInt64)

	// Max int64 should decode successfully
	var i Int64
	err := json.Unmarshal([]byte(strconv.FormatUint(int64Overflow, 10)), &i)
	maybePanic(err)

	// Attempt to overflow
	int64Overflow++
	err = json.Unmarshal([]byte(strconv.FormatUint(int64Overflow, 10)), &i)
	if err == nil {
		panic("err should be present; decoded value overflows int64")
	}
}

func TestTextUnmarshalInt64(t *testing.T) {
	var i Int64
	err := i.UnmarshalText([]byte("12345"))
	maybePanic(err)
	assertInt64(t, i, "UnmarshalText() int")

	var blank Int64
	err = blank.UnmarshalText([]byte(""))
	maybePanic(err)
	assertNullInt64(t, blank, "UnmarshalText() empty int")

	var null Int64
	err = null.UnmarshalText([]byte("null"))
	maybePanic(err)
	assertNullInt64(t, null, `UnmarshalText() "null"`)

	var invalid Int64
	err = invalid.UnmarshalText([]byte("hello world"))
	if err == nil {
		panic("expected error")
	}
}

func TestMarshalInt64(t *testing.T) {
	i := Int64From(12345)
	data, err := json.Marshal(i)
	maybePanic(err)
	assertJSONEquals(t, data, "12345", "non-empty json marshal")

	// invalid values should be encoded as null
	null := NewInt64(0, false)
	data, err = json.Marshal(null)
	maybePanic(err)
	assertJSONEquals(t, data, "null", "null json marshal")
}

func TestMarshalInt64Text(t *testing.T) {
	i := Int64From(12345)
	data, err := i.MarshalText()
	maybePanic(err)
	assertJSONEquals(t, data, "12345", "non-empty text marshal")

	// invalid values should be encoded as null
	null := NewInt64(0, false)
	data, err = null.MarshalText()
	maybePanic(err)
	assertJSONEquals(t, data, "", "null text marshal")
}

func TestInt64Pointer(t *testing.T) {
	i := Int64From(12345)
	ptr := i.Ptr()
	if *ptr != 12345 {
		t.Errorf("bad %s int: %#v ≠ %d\n", "pointer", ptr, 12345)
	}

	null := NewInt64(0, false)
	ptr = null.Ptr()
	if ptr != nil {
		t.Errorf("bad %s int: %#v ≠ %s\n", "nil pointer", ptr, "nil")
	}
}

func TestInt64IsZero(t *testing.T) {
	i := Int64From(12345)
	if i.IsZero() {
		t.Errorf("IsZero() should be false")
	}

	null := NewInt64(0, false)
	if !null.IsZero() {
		t.Errorf("IsZero() should be true")
	}

	zero := NewInt64(0, true)
	if zero.IsZero() {
		t.Errorf("IsZero() should be false")
	}
}

func TestInt64SetValid(t *testing.T) {
	change := NewInt64(0, false)
	assertNullInt64(t, change, "SetValid()")
	change.SetValid(12345)
	assertInt64(t, change, "SetValid()")
}

func TestInt64Scan(t *testing.T) {
	var i Int64
	err := i.Scan(12345)
	maybePanic(err)
	assertInt64(t, i, "scanned int")

	var null Int64
	err = null.Scan(nil)
	maybePanic(err)
	assertNullInt64(t, null, "scanned null")
}

func TestInt64ValueOrZero(t *testing.T) {
	valid := NewInt64(12345, true)
	if valid.ValueOrZero() != 12345 {
		t.Error("unexpected ValueOrZero", valid.ValueOrZero())
	}

	invalid := NewInt64(12345, false)
	if invalid.ValueOrZero() != 0 {
		t.Error("unexpected ValueOrZero", invalid.ValueOrZero())
	}
}

func TestInt64Equal(t *testing.T) {
	int1 := NewInt64(10, false)
	int2 := NewInt64(10, false)
	assertInt64EqualIsTrue(t, int1, int2)

	int1 = NewInt64(10, false)
	int2 = NewInt64(20, false)
	assertInt64EqualIsTrue(t, int1, int2)

	int1 = NewInt64(10, true)
	int2 = NewInt64(10, true)
	assertInt64EqualIsTrue(t, int1, int2)

	int1 = NewInt64(10, true)
	int2 = NewInt64(10, false)
	assertInt64EqualIsFalse(t, int1, int2)

	int1 = NewInt64(10, false)
	int2 = NewInt64(10, true)
	assertInt64EqualIsFalse(t, int1, int2)

	int1 = NewInt64(10, true)
	int2 = NewInt64(20, true)
	assertInt64EqualIsFalse(t, int1, int2)
}

func assertInt64(t *testing.T, i Int64, from string) {
	if i.Int64 != 12345 {
		t.Errorf("bad %s int: %d ≠ %d\n", from, i.Int64, 12345)
	}
	if !i.Valid {
		t.Error(from, "is invalid, but should be valid")
	}
}

func assertNullInt64(t *testing.T, i Int64, from string) {
	if i.Valid {
		t.Error(from, "is valid, but should be invalid")
	}
}

func assertInt64EqualIsTrue(t *testing.T, a, b Int64) {
	t.Helper()
	if !a.Equal(b) {
		t.Errorf("Equal() of Int64{%v, Valid:%t} and Int64{%v, Valid:%t} should return true", a.Int64, a.Valid, b.Int64, b.Valid)
	}
}

func assertInt64EqualIsFalse(t *testing.T, a, b Int64) {
	t.Helper()
	if a.Equal(b) {
		t.Errorf("Equal() of Int64{%v, Valid:%t} and Int64{%v, Valid:%t} should return false", a.Int64, a.Valid, b.Int64, b.Valid)
	}
}
