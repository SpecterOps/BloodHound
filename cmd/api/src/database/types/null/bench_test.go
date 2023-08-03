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
	"testing"
)

func BenchmarkIntUnmarshalJSON(b *testing.B) {
	input := []byte("123456")
	var nullable Int64
	for n := 0; n < b.N; n++ {
		nullable.UnmarshalJSON(input)
	}
}

func BenchmarkIntStringUnmarshalJSON(b *testing.B) {
	input := []byte(`"123456"`)
	var nullable String
	for n := 0; n < b.N; n++ {
		nullable.UnmarshalJSON(input)
	}
}

func BenchmarkNullIntUnmarshalJSON(b *testing.B) {
	input := []byte("null")
	var nullable Int64
	for n := 0; n < b.N; n++ {
		nullable.UnmarshalJSON(input)
	}
}

func BenchmarkStringUnmarshalJSON(b *testing.B) {
	input := []byte(`"hello"`)
	var nullable String
	for n := 0; n < b.N; n++ {
		nullable.UnmarshalJSON(input)
	}
}

func BenchmarkNullStringUnmarshalJSON(b *testing.B) {
	input := []byte("null")
	var nullable String
	for n := 0; n < b.N; n++ {
		nullable.UnmarshalJSON(input)
	}
}
