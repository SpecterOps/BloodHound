// Copyright 2026 Specter Ops, Inc.
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

package hybrid

import (
	"testing"

	adSchema "github.com/specterops/bloodhound/packages/go/graphschema/ad"
)

func TestAADObjectIDPropertyMatchesCollectorContract(t *testing.T) {
	const collectorProperty = "aadobjectid"

	if actual := adSchema.AADObjectID.String(); actual != collectorProperty {
		t.Fatalf("expected AADObjectID property %q, got %q", collectorProperty, actual)
	}
}

func TestNormalizeObjectID(t *testing.T) {
	for name, testCase := range map[string]struct {
		input    string
		expected string
	}{
		"lowercase":          {input: "69e33ede-7272-4893-ba72-18e6a92a0184", expected: "69E33EDE-7272-4893-BA72-18E6A92A0184"},
		"surrounding spaces": {input: " 69e33ede-7272-4893-ba72-18e6a92a0184 ", expected: "69E33EDE-7272-4893-BA72-18E6A92A0184"},
		"empty":              {input: "", expected: ""},
	} {
		t.Run(name, func(t *testing.T) {
			if actual := normalizeObjectID(testCase.input); actual != testCase.expected {
				t.Fatalf("expected normalized object ID %q, got %q", testCase.expected, actual)
			}
		})
	}
}
