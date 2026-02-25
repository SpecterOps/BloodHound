// Copyright 2026 Specter Ops, Inc.
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
package metrics

import (
	"testing"
)

func TestMetricKey(t *testing.T) {
	tests := []struct {
		testName  string
		name      string
		namespace string
		labels    map[string]string
		expected  string
	}{
		{
			testName:  "empty inputs",
			namespace: "",
			labels:    nil,
			expected:  "||",
		},
		{
			testName:  "single label",
			namespace: "ns1",
			name:      "counter1",
			labels:    map[string]string{"label1": "val1"},
			expected:  "ns1|counter1|label1val1",
		},
		{
			testName:  "multiple labels",
			namespace: "ns2",
			name:      "counter2",
			labels:    map[string]string{"b": "v2", "a": "v1"},
			expected:  "ns2|counter2|av1|bv2",
		},
		{
			testName:  "no labels",
			namespace: "ns4",
			name:      "counter4",
			labels:    nil,
			expected:  "ns4|counter4|",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			key := metricKey(tc.name, tc.namespace, tc.labels)
			if key != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, key)
			}
		})
	}
}

func TestMetricVecKey(t *testing.T) {
	tests := []struct {
		testName       string
		namespace      string
		name           string
		labels         map[string]string
		variableLabels []string
		expected       string
	}{
		{
			testName:       "empty inputs",
			namespace:      "",
			labels:         nil,
			variableLabels: nil,
			expected:       "|||vec",
		},
		{
			testName:       "single variable label",
			namespace:      "ns1",
			name:           "counter1",
			labels:         map[string]string{"label1": "val1"},
			variableLabels: []string{"label2"},
			expected:       "ns1|counter1|label1val1|label2vec",
		},
		{
			testName:       "multiple variable labels",
			namespace:      "ns2",
			name:           "counter2",
			labels:         map[string]string{"b": "v2", "a": "v1"},
			variableLabels: []string{"c", "d"},
			expected:       "ns2|counter2|av1|bv2|c|dvec",
		},
		{
			testName:       "unsorted variable labels",
			namespace:      "ns3",
			name:           "counter3",
			labels:         map[string]string{"c": "v3", "b": "v2", "a": "v1"},
			variableLabels: []string{"e", "g", "f"},
			expected:       "ns3|counter3|av1|bv2|cv3|e|f|gvec",
		},
		{
			testName:       "no variable labels",
			namespace:      "ns4",
			name:           "counter4",
			labels:         map[string]string{"x": "v"},
			variableLabels: nil,
			expected:       "ns4|counter4|xv|vec",
		},
	}

	for _, tc := range tests {
		t.Run(tc.testName, func(t *testing.T) {
			key := metricVecKey(tc.name, tc.namespace, tc.labels, tc.variableLabels)
			if key != tc.expected {
				t.Errorf("expected %q, got %q", tc.expected, key)
			}
		})
	}
}
