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

package version

import "testing"

func TestParseVersion(t *testing.T) {
	if parsed, err := Parse("v1.2.3"); err != nil {
		t.Fatalf("Parsing version failed: %v", err)
	} else if parsed.String() != "v1.2.3" {
		t.Fatalf("Expected version to match %s but got %s", "v1.2.3", parsed.String())
	}

	if parsed, err := Parse("v1.2.3-rc1"); err != nil {
		t.Fatalf("Parsing version failed: %v", err)
	} else if parsed.String() != "v1.2.3-rc1" {
		t.Fatalf("Expected version to match %s but got %s", "v1.2.3", parsed.String())
	}
}

func TestVersion_GreaterThan(t *testing.T) {
	var (
		greater = Version{
			Major: 2,
			Minor: 2,
			Patch: 2,
		}

		lesser = Version{
			Major: 1,
			Minor: 1,
			Patch: 1,
		}
	)

	if !greater.GreaterThan(lesser) {
		t.Fatalf("Expected version %s to be greater than %s", greater, lesser)
	}

	if !lesser.LessThan(greater) {
		t.Fatalf("Expected version %s to be less than %s", lesser, greater)
	}

	lesser.Major = greater.Major
	if !greater.GreaterThan(lesser) {
		t.Fatalf("Expected version %s to be greater than %s", greater, lesser)
	}

	if !lesser.LessThan(greater) {
		t.Fatalf("Expected version %s to be less than %s", lesser, greater)
	}

	lesser.Minor = greater.Minor
	if !greater.GreaterThan(lesser) {
		t.Fatalf("Expected version %s to be greater than %s", greater, lesser)
	}

	if !lesser.LessThan(greater) {
		t.Fatalf("Expected version %s to be less than %s", lesser, greater)
	}

	lesser.Patch = greater.Patch
	if greater.GreaterThan(lesser) {
		t.Fatalf("Expected version %s to no longer be greater than %s", greater, lesser)
	}

	if lesser.LessThan(greater) {
		t.Fatalf("Expected version %s to no longer be less than %s", lesser, greater)
	}

	if !greater.Equals(lesser) {
		t.Fatalf("Expected version %s to match %s", greater, lesser)
	}
}
