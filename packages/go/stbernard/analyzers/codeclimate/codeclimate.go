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

package codeclimate

// Entry represents a simple CodeClimate-like entry
type Entry struct {
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Location    Location `json:"location"`
}

// Location represents a code path and lines for an entry
type Location struct {
	Path  string `json:"path"`
	Lines Lines  `json:"lines"`
}

// Lines represents only the beginning line for the location
type Lines struct {
	Begin uint64 `json:"begin"`
}
