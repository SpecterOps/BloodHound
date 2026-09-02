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
package dogtags

// Typed flag keys - each type can only be used with its matching getter
type BoolDogTag string
type StringDogTag string
type IntDogTag string

type BoolDogTagSpec struct {
	Description string `json:"description,omitempty"`
	Default     bool   `json:"default"`
}

type StringDogTagSpec struct {
	Description string `json:"description,omitempty"`
	Default     string `json:"default"`
}

type IntDogTagSpec struct {
	Description string `json:"description,omitempty"`
	Default     int64  `json:"default"`
}

const (
	PZ_MULTI_TIER_ANALYSIS BoolDogTag = "privilege_zones.multi_tier_analysis"
	PZ_TIER_LIMIT          IntDogTag  = "privilege_zones.tier_limit"
	PZ_LABEL_LIMIT         IntDogTag  = "privilege_zones.label_limit"

	ETAC_ENABLED BoolDogTag = "auth.environment_targeted_access_control"
)

var AllBoolDogTags = map[BoolDogTag]BoolDogTagSpec{
	PZ_MULTI_TIER_ANALYSIS: {Description: "PZ Multi Tier Analysis", Default: false},
	ETAC_ENABLED:           {Description: "ETAC Enabled", Default: false},
}

var AllIntDogTags = map[IntDogTag]IntDogTagSpec{
	PZ_TIER_LIMIT:  {Description: "PZ Tier Limit", Default: 1},
	PZ_LABEL_LIMIT: {Description: "PZ Label Limit", Default: 0},
}

var AllStringDogTags = map[StringDogTag]StringDogTagSpec{
	// none yet
}
