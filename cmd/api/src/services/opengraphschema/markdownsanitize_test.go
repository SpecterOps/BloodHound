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

package opengraphschema

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarkdownValidator_Policy(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{name: "success_-_link_survives_unchanged", input: `<a href="http://x">y</a>`, expected: `<a href="http://x">y</a>`},
		{name: "success_-_script_tag_stripped", input: `<script>bad()</script>ok`, expected: `ok`},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, testCase.expected, newMarkdownValidator().policy.Sanitize(testCase.input))
		})
	}
}

func TestMarkdownValidator_Render(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name       string
		input      string
		wantSubstr string
	}{
		{name: "success_-_bold_renders_strong", input: "**bold**", wantSubstr: "<strong>bold</strong>"},
		{name: "success_-_raw_html_preserved", input: "<script>alert(1)</script>", wantSubstr: "<script>alert(1)</script>"},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			output, err := newMarkdownValidator().render(testCase.input)
			require.NoError(t, err)
			assert.Contains(t, output, testCase.wantSubstr)
		})
	}
}

func TestNormalizeHTML(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name      string
		left      string
		right     string
		wantEqual bool
	}{
		{name: "success_-_cosmetic_entities_equal", left: `<p>a &quot;b&quot;</p>`, right: `<p>a &#34;b&#34;</p>`, wantEqual: true},
		{name: "success_-_structural_diff_not_equal", left: `<a href="x" onclick="y">z</a>`, right: `<a href="x">z</a>`, wantEqual: false},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			normalizedLeft, err := normalizeHTML(testCase.left)
			require.NoError(t, err)
			normalizedRight, err := normalizeHTML(testCase.right)
			require.NoError(t, err)
			if testCase.wantEqual {
				assert.Equal(t, normalizedLeft, normalizedRight)
			} else {
				assert.NotEqual(t, normalizedLeft, normalizedRight)
			}
		})
	}
}

func TestMarkdownValidator_Validate(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{name: "success_-_plain_markdown_safe", input: "**hello** [x](http://y)", wantErr: false},
		{name: "success_-_aligned_table_safe", input: "| a | b |\n|:--|--:|\n| 1 | 2 |\n", wantErr: false},
		{name: "success_-_fenced_code_with_language_safe", input: "```go\nfmt.Println(1)\n```\n", wantErr: false},
		{name: "error_-_script_tag_rejected", input: "<script>alert(1)</script>", wantErr: true},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			validator := newMarkdownValidator()
			err := validator.validate(testCase.input)
			if testCase.wantErr {
				assert.ErrorIs(t, err, errUnsafeMarkdownContent)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
