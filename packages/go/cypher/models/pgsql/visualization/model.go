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

package visualization

type Graph struct {
	Title         string         `json:"-"`
	Style         Style          `json:"style"`
	Nodes         []Node         `json:"nodes"`
	Relationships []Relationship `json:"relationships"`
}

type Node struct {
	Caption    string         `json:"caption"`
	ID         string         `json:"id"`
	Labels     []string       `json:"labels"`
	Properties map[string]any `json:"properties"`
	Position   Position       `json:"position"`
	Style      Style          `json:"style"`
}

type Relationship struct {
	ID         string         `json:"id"`
	FromID     string         `json:"fromId"`
	ToID       string         `json:"toId"`
	Type       string         `json:"type"`
	Properties map[string]any `json:"properties"`
	Style      Style          `json:"style"`
}

type Style struct{}

type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}
