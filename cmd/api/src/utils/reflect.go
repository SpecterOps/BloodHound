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

package utils

import (
	"reflect"
)

// NilErrorValue is the reflected zero value of errors.Error
var NilErrorValue = reflect.Zero(reflect.TypeOf((*error)(nil)).Elem())

// GenericClone is the implementation of a shallow clone function for dynamic types.
//
//	var cloneMap func(map[string]interface{}) map[string]interface{}
//	func init() {
//		Bind(&cloneMap, GenericClone)
//	}
//
//	func main() {
//		m1 := map[string]interface{} {
//			"foo": "bar",
//		}
//		m2 := cloneMap(m1)
//		fmt.Println(m2)
//		// Output: map[foo:bar]
//	}
func GenericClone(args []reflect.Value) []reflect.Value {
	param := args[0]
	var duplicate reflect.Value

	switch param.Kind() {
	case reflect.Map:
		duplicate = reflect.MakeMap(param.Type())
		for entry := param.MapRange(); entry.Next(); {
			duplicate.SetMapIndex(entry.Key(), entry.Value())
		}
	case reflect.Array, reflect.Slice:
		duplicate = reflect.MakeSlice(param.Type(), param.Len(), param.Cap())
		reflect.Copy(duplicate, param)
	case reflect.Struct:
		duplicate = reflect.New(param.Type()).Elem()
		for i := 0; i < param.NumField(); i++ {
			duplicate.Field(i).Set(param.Field(i))
		}
	}
	return []reflect.Value{duplicate}
}

// GenericAssign is the implementation of a variadic assign function for dynamic types.
// The function assigns the enumerable properties of each source object from left to right to the returned object.
// Subsequent source objects overwrite the property assignments of previous source object.
//
//	var assignMap func(...map[string]interface{}) map[string]interface{}
//	func init() {
//		Bind(&assignMap, GenericAssign)
//	}
//
//	func main() {
//		m1 := map[string]interface{} {
//			"foo": "bar",
//		}
//		m2 := map[string]interface{} {
//			"foo": "baz",
//		}
//		m3 := map[string]interface{} {
//			"bar": "buzz",
//		}
//		if m4, err := assignMap(m1, m2, m3); err == nil {
//			fmt.Println(m4)
//			// Output: map[bar:buzz foo:baz]
//		}
//	}
func GenericAssign(args []reflect.Value) []reflect.Value {
	variadicParams := args[0]
	param1 := variadicParams.Index(0)
	var dest reflect.Value

	switch param1.Kind() {
	case reflect.Map:
		dest = GenericClone([]reflect.Value{param1})[0]
		for i := 1; i < variadicParams.Len(); i++ {
			for entry := variadicParams.Index(i).MapRange(); entry.Next(); {
				dest.SetMapIndex(entry.Key(), entry.Value())
			}
		}
	case reflect.Array, reflect.Slice:
		length := 0
		capacity := 0
		for i := 0; i < variadicParams.Len(); i++ {
			slice := variadicParams.Index(i)
			if slice.Len() > length {
				length = slice.Len()
			}

			if slice.Cap() > capacity {
				capacity = slice.Cap()
			}
		}
		dest = reflect.MakeSlice(param1.Type(), length, capacity)

		for i := 0; i < variadicParams.Len(); i++ {
			slice := variadicParams.Index(i)
			for j := 0; j < slice.Len(); j++ {
				item := slice.Index(j)
				if !item.IsZero() {
					dest.Index(j).Set(slice.Index(j))
				}
			}
		}
	case reflect.Struct:
		dest = reflect.New(param1.Type()).Elem()
		for i := 0; i < variadicParams.Len(); i++ {
			instance := variadicParams.Index(i)
			for j := 0; j < instance.NumField(); j++ {
				item := instance.Field(j)
				if !item.IsZero() {
					dest.Field(j).Set(item)
				}
			}
		}
	}
	return []reflect.Value{dest}
}

// GenericMap is the implementation of a map function for dynamic types.
func GenericMap(args []reflect.Value) []reflect.Value {
	collection := args[0]
	iteratee := args[1]

	numParams := iteratee.Type().NumIn()
	sliceType := reflect.SliceOf(iteratee.Type().Out(0))
	result := reflect.MakeSlice(sliceType, 0, 0)

	switch collection.Kind() {
	case reflect.Map:
		for entry := collection.MapRange(); entry.Next(); {
			var item reflect.Value
			if numParams == 0 {
				item = iteratee.Call([]reflect.Value{})[0]
			} else if numParams == 1 {
				item = iteratee.Call([]reflect.Value{entry.Value()})[0]
			} else if numParams == 2 {
				item = iteratee.Call([]reflect.Value{entry.Value(), entry.Key()})[0]
			} else {
				item = iteratee.Call([]reflect.Value{entry.Value(), entry.Key(), collection})[0]
			}
			result = reflect.Append(result, item)
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < collection.Len(); i++ {
			var item reflect.Value
			if numParams == 0 {
				item = iteratee.Call([]reflect.Value{})[0]
			} else if numParams == 1 {
				item = iteratee.Call([]reflect.Value{collection.Index(i)})[0]
			} else if numParams == 2 {
				item = iteratee.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i)})[0]
			} else {
				item = iteratee.Call([]reflect.Value{collection.Index(i), reflect.ValueOf(i), collection})[0]
			}
			result = reflect.Append(result, item)
		}
	}

	return []reflect.Value{result}
}

// Bind sets a typed function pointer to a generic implementation of a function based on reflect.Value.
//
// Caveat Emptor: These functions do not benefit from compiler type-checking and can panic during runtime in the same
// way dynamically typed languages behave. Albeit small, there is a performance impact when creating functions with
// Bind. This is due to the additional steps it takes to determine the type of an interface.
func Bind(functionPtr any, function func([]reflect.Value) []reflect.Value) {
	ptr := reflect.ValueOf(functionPtr).Elem()
	ptr.Set(reflect.MakeFunc(ptr.Type(), function))
}
