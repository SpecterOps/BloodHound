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

package pgsql

import (
	"bytes"
	"encoding/json"
	"github.com/jackc/pgtype"
	"github.com/specterops/bloodhound/dawgs/graph"
)

func ValueToJSONB(value any) (pgtype.JSONB, error) {
	var jsonbArgument pgtype.JSONB

	return jsonbArgument, jsonbArgument.Set(value)
}

func Int32SliceToInt4Array(value []int32) (pgtype.Int4Array, error) {
	var pgInt4Array pgtype.Int4Array

	return pgInt4Array, pgInt4Array.Set(value)
}

func IDSliceToInt8Array(value []graph.ID) (pgtype.Int8Array, error) {
	var pgInt8Array pgtype.Int8Array

	return pgInt8Array, pgInt8Array.Set(value)
}

func StringSliceToTextArray(values []string) (pgtype.TextArray, error) {
	var pgTextArray pgtype.TextArray
	return pgTextArray, pgTextArray.Set(values)
}

func MapStringAnyToJSONB(value map[string]any) (pgtype.JSONB, error) {
	var jsonb pgtype.JSONB

	return jsonb, jsonb.Set(value)
}

func PropertiesToJSONB(properties *graph.Properties) (pgtype.JSONB, error) {
	return MapStringAnyToJSONB(properties.MapOrEmpty())
}

func JSONBToProperties(jsonb pgtype.JSONB) (*graph.Properties, error) {
	propertiesMap := make(map[string]any)

	if err := jsonb.AssignTo(&propertiesMap); err != nil {
		return nil, err
	}

	return graph.AsProperties(propertiesMap), nil
}

func MatcherAsJSONB(fieldName string, value any) (pgtype.JSONB, error) {
	var (
		matcher      = bytes.Buffer{}
		jsonbMatcher = pgtype.JSONB{}
	)

	// Prepare the JSONB matcher
	if marshalledValue, err := json.Marshal(value); err != nil {
		return jsonbMatcher, err
	} else {
		matcher.WriteString(`{"`)
		matcher.WriteString(fieldName)
		matcher.WriteString(`":`)
		matcher.Write(marshalledValue)
		matcher.WriteString(`}`)
	}

	return ValueToJSONB(matcher.Bytes())
}

func MustMatcherAsJSONB(fieldName string, value any) pgtype.JSONB {
	if jsonbMatcher, err := MatcherAsJSONB(fieldName, value); err != nil {
		panic(err)
	} else {
		return jsonbMatcher
	}
}
