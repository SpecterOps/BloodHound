package opengraphschema

import (
	"bytes"
	"errors"
	"fmt"
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

type markdownValidator struct {
	renderer goldmark.Markdown
	policy   *bluemonday.Policy
}

func newMarkdownValidator() *markdownValidator {
	var policy = bluemonday.UGCPolicy()
	policy.RequireNoFollowOnLinks(false)
	policy.RequireNoFollowOnFullyQualifiedLinks(false)
	return &markdownValidator{
		renderer: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
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
