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
