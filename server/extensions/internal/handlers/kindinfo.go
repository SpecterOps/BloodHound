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
package handlers

import (
	"encoding/json"
	"fmt"
)

// KindInfoView is the JSON shape of a kind info embedded in a NodeKindView.
type KindInfoView struct {
	Title    string       `json:"title"`
	Position int32        `json:"position"`
	Markdown MarkdownView `json:"markdown"`
}

// MarkdownView is the JSON shape of the flattened markdown content.
type MarkdownView struct {
	Content string `json:"content"`
}

// kindInfoContentView unwraps the stored content JSON object shaped
// {"markdown": {"content": "..."}} into its inner MarkdownView.
type kindInfoContentView struct {
	Markdown MarkdownView `json:"markdown"`
}

// buildMarkdownView unwraps the stored content object into its inner MarkdownView,
// returning an empty MarkdownView and an error when the content cannot be parsed.
func buildMarkdownView(content json.RawMessage) (MarkdownView, error) {
	var contentView kindInfoContentView

	if err := json.Unmarshal(content, &contentView); err != nil {
		return MarkdownView{}, fmt.Errorf("unmarshalling markdown content: %w", err)
	}

	return contentView.Markdown, nil
}
