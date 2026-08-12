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
	"bytes"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// markdownsanitize.go validates that markdown supplied in extension uploads does
// not contain unsafe HTML. Markdown is rendered to HTML (goldmark, raw HTML
// preserved), sanitized (bluemonday with link-rewriting disabled), then BOTH the
// rendered and sanitized HTML are canonicalized via x/net/html and compared. Any
// difference (or any render/parse error) rejects the content. Canonicalizing both
// sides collapses cosmetic serialization drift (entity form, void-element
// self-closing, boolean attributes) so only genuine removals cause rejection.
//
// Two goldmark GFM attributes are emitted that UGCPolicy strips (which would cause
// safe content to be rejected): table cell alignment and the fenced-code language
// class. Tables are configured with TableCellAlignNone so no align/style attribute
// is emitted, and the policy allows the constrained language class on <code> so both
// compared sides retain it. Neither attribute is user-facing since this HTML is only
// used to validate the markdown; the original markdown is what gets stored.

type markdownValidator struct {
	renderer goldmark.Markdown
	policy   *bluemonday.Policy
}

func newMarkdownValidator() *markdownValidator {
	var policy = bluemonday.UGCPolicy()
	policy.RequireNoFollowOnLinks(false)
	policy.RequireNoFollowOnFullyQualifiedLinks(false)
	// Allow goldmark's fenced-code language hint. The value is class-only and
	// non-executable, and the anchored alnum regex forbids smuggling extra attrs.
	policy.AllowAttrs("class").
		Matching(regexp.MustCompile(`^language-[a-zA-Z0-9]+$`)).
		OnElements("code")
	return &markdownValidator{
		renderer: goldmark.New(
			// Mirrors extension.GFM (Linkify, Table, Strikethrough, TaskList) but
			// substitutes a TableCellAlignNone table so no align/style attribute is
			// emitted. Keep this list in sync if goldmark's GFM bundle changes.
			goldmark.WithExtensions(
				extension.Linkify,
				extension.Strikethrough,
				extension.TaskList,
				extension.NewTable(extension.WithTableCellAlignMethod(extension.TableCellAlignNone)),
			),
			goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()),
		),
		policy: policy,
	}
}

func (s *markdownValidator) render(markdownContent string) (string, error) {
	var buf bytes.Buffer
	if err := s.renderer.Convert([]byte(markdownContent), &buf); err != nil {
		return "", fmt.Errorf("rendering markdown to html: %w", err)
	}
	return buf.String(), nil
}

var errUnsafeMarkdownContent = errors.New("markdown content contains unsafe or disallowed HTML")

func (s *markdownValidator) validate(markdownContent string) error {
	rendered, err := s.render(markdownContent)
	if err != nil {
		return fmt.Errorf("rendering markdown: %w", err)
	}
	sanitized := s.policy.Sanitize(rendered)
	normalizedRendered, err := normalizeHTML(rendered)
	if err != nil {
		return fmt.Errorf("normalizing rendered html: %w", err)
	}
	normalizedSanitized, err := normalizeHTML(sanitized)
	if err != nil {
		return fmt.Errorf("normalizing sanitized html: %w", err)
	}
	if normalizedRendered != normalizedSanitized {
		return errUnsafeMarkdownContent
	}
	return nil
}

func normalizeHTML(fragment string) (string, error) {
	var (
		contextNode = &xhtml.Node{Type: xhtml.ElementNode, Data: "body", DataAtom: atom.Body}
		buf         bytes.Buffer
	)
	if nodes, err := xhtml.ParseFragment(strings.NewReader(fragment), contextNode); err != nil {
		return "", fmt.Errorf("parse fragment: %w", err)
	} else {
		for _, node := range nodes {
			if err := xhtml.Render(&buf, node); err != nil {
				return "", fmt.Errorf("render node: %w", err)
			}
		}
	}
	return buf.String(), nil
}
