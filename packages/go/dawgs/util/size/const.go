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

package size

const (
	iecUnitFactor = 1024

	Bytes    Size = 1
	Kibibyte      = Bytes * iecUnitFactor
	Mebibyte      = Kibibyte * iecUnitFactor
	Gibibyte      = Mebibyte * iecUnitFactor
	Tebibyte      = Gibibyte * iecUnitFactor
	Pebibyte      = Tebibyte * iecUnitFactor
)

type Size uintptr

func (s Size) Bytes() uintptr {
	return uintptr(s)
}

func (s Size) Kibibytes() float64 {
	return float64(s) / float64(Kibibyte)
}

func (s Size) Mebibytes() float64 {
	return float64(s) / float64(Mebibyte)
}

func (s Size) Gibibyte() float64 {
	return float64(s) / float64(Gibibyte)
}

func (s Size) Tebibyte() float64 {
	return float64(s) / float64(Tebibyte)
}

func (s Size) Pebibyte() float64 {
	return float64(s) / float64(Pebibyte)
}
