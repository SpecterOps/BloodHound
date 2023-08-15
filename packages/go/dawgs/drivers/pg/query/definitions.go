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

package query

import "regexp"

var (
	pgPropertyIndexRegex = regexp.MustCompile(`(?i)^create\s+(unique)?(?:\s+)?index\s+([^ ]+)\s+on\s+\S+\s+using\s+([^ ]+)\s+\(+properties\s+->>\s+'([^:]+)::.+$`)
	pgColumnIndexRegex   = regexp.MustCompile(`(?i)^create\s+(unique)?(?:\s+)?index\s+([^ ]+)\s+on\s+\S+\s+using\s+([^ ]+)\s+\(([^)]+)\)$`)
)

const (
	pgIndexRegexGroupUnique       = 1
	pgIndexRegexGroupName         = 2
	pgIndexRegexGroupIndexType    = 3
	pgIndexRegexGroupFields       = 4
	pgIndexRegexNumExpectedGroups = 5

	pgIndexTypeBTree   = "btree"
	pgIndexTypeGIN     = "gin"
	pgIndexUniqueStr   = "unique"
	pgPropertiesColumn = "properties"
)
