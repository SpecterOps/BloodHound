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

//TODO: Remove this file after release 2.0.8
package model

type DataQualityStat struct {
	Users                  int     `json:"users"`
	Groups                 int     `json:"groups"`
	Computers              int     `json:"computers"`
	Ous                    int     `json:"ous"`
	Gpos                   int     `json:"gpos"`
	Domains                int     `json:"domains"`
	Acls                   int     `json:"acls"`
	Sessions               int     `json:"sessions"`
	Relationships          int     `json:"relationships"`
	SessionCompleteness    float32 `json:"session_completeness"`
	LocalGroupCompleteness float32 `json:"local_group_completeness"`

	Serial
}
