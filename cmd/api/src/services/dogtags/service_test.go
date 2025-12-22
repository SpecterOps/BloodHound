// Copyright 2025 Specter Ops, Inc.
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
// service_test.go
package dogtags

import (
	"testing"
)

func TestService(t *testing.T) {
	svc := NewDefaultService()

	// Test typed getters return defaults
	tierLimit := svc.GetFlagAsInt(PZ_TIER_LIMIT)
	if tierLimit != 1 {
		t.Errorf("expected 1, got %d", tierLimit)
	}

	labelLimit := svc.GetFlagAsInt(PZ_LABEL_LIMIT)
	if labelLimit != 10 {
		t.Errorf("expected 10, got %d", labelLimit)
	}

	multiTier := svc.GetFlagAsBool(PZ_MULTI_TIER_ANALYSIS)
	if multiTier != false {
		t.Errorf("expected false, got %v", multiTier)
	}

	// Test GetAllDogTags
	all := svc.GetAllDogTags()
	t.Logf("All dogtags: %+v", all)
}
