// Copyright 2024 Specter Ops, Inc.
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

package pgsql

const (
	//
	FunctionASPHarness             Identifier = "asp_harness"
	FunctionIntArrayUnique         Identifier = "uniq"
	FunctionIntArraySort           Identifier = "sort"
	FunctionJSONBToTextArray       Identifier = "jsonb_to_text_array"
	FunctionJSONBArrayElementsText Identifier = "jsonb_array_elements_text"
	FunctionJSONBBuildObject       Identifier = "jsonb_build_object"
	FunctionArrayLength            Identifier = "array_length"
	FunctionArrayAggregate         Identifier = "array_agg"
	FunctionMin                    Identifier = "min"
	FunctionMax                    Identifier = "max"
	FunctionLocalTimestamp         Identifier = "localtimestamp"
	FunctionLocalTime              Identifier = "localtime"
	FunctionCurrentTime            Identifier = "current_time"
	FunctionCurrentDate            Identifier = "current_date"
	FunctionNow                    Identifier = "now"
	FunctionToLower                Identifier = "lower"
	FunctionToUpper                Identifier = "upper"
	FunctionCoalesce               Identifier = "coalesce"
	FunctionUnnest                 Identifier = "unnest"
	FunctionJSONBSet               Identifier = "jsonb_set"
	FunctionCount                  Identifier = "count"
	FunctionStringToArray          Identifier = "string_to_array"
	FunctionEdgesToPath            Identifier = "edges_to_path"
)
