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

package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
)

// NodeContext is the template representation of a node.
// It intentionally exposes only the node data supported by the template
// contract rather than the complete Node domain object.
type NodeContext struct {
	NodeID     int64
	Kinds      []KindContext
	Properties map[string]any
}

// RelationshipContext is the template representation of a relationship and
// its endpoint nodes.
type RelationshipContext struct {
	RelationshipID int64
	Source         NodeContext
	Target         NodeContext
	Kind           KindContext
	Properties     map[string]any
}

// KindContext is the template representation of a kind.
type KindContext struct {
	KindID *int32
	Name   string
}

// String lets template helpers such as join render kinds by name rather than
// exposing the internal KindContext struct representation.
func (s KindContext) String() string {
	return s.Name
}

type kindInfoContent struct {
	Markdown struct {
		Content string `json:"content"`
	} `json:"markdown"`
}

// we don't support these template functions
var unsupportedFns = []string{
	"bcrypt",
	"htpasswd",
	"genPrivateKey",
	"derivePassword",
	"buildCustomCert",
	"genCA",
	"genCAWithKey",
	"genSelfSignedCert",
	"genSignedCert",
	"genSignedCertWithKey",
	"encryptAES",
	"decryptAES",
	"regexMatch",
	"regexFindAll",
	"regexFind",
	"regexReplaceAll",
	"regexReplaceAllLiteral",
	"regexSplit",
	"mustRegexMatch",
	"mustRegexFindAll",
	"mustRegexFind",
	"mustRegexReplaceAll",
	"mustRegexReplaceAllLiteral",
	"mustRegexSplit",
	"urlParse",
	"urlJoin",
	"randInt",
	"ago",
}

func renderKindInfoMarkdown(content json.RawMessage, context any) (string, error) {
	var contentView kindInfoContent

	if err := json.Unmarshal(content, &contentView); err != nil {
		return "", fmt.Errorf("unmarshalling markdown content: %w", err)
	}

	functions := sprig.HermeticTxtFuncMap()
	for _, functionName := range unsupportedFns {
		delete(functions, functionName)
	}

	parsedTemplate, err := template.New("kind-info-markdown").
		Funcs(functions).
		Parse(contentView.Markdown.Content)
	if err != nil {
		return contentView.Markdown.Content, fmt.Errorf("parsing markdown template: %w", err)
	}

	var rendered bytes.Buffer
	if err := parsedTemplate.Execute(&rendered, context); err != nil {
		return contentView.Markdown.Content, fmt.Errorf("executing markdown template: %w", err)
	}

	return rendered.String(), nil
}

func newNodeContext(node Node) NodeContext {
	kinds := make([]KindContext, 0, len(node.Kinds))

	for _, kind := range node.Kinds {
		kinds = append(kinds, KindContext{
			KindID: kind.ID,
			Name:   kind.Name,
		})
	}

	return NodeContext{
		NodeID:     node.ID,
		Kinds:      kinds,
		Properties: node.Properties,
	}
}

func newRelationshipContext(relationship Relationship, source Node, target Node) RelationshipContext {
	return RelationshipContext{
		RelationshipID: relationship.ID,
		Source:         newNodeContext(source),
		Target:         newNodeContext(target),
		Kind: KindContext{
			KindID: relationship.Kind.ID,
			Name:   relationship.Kind.Name,
		},
		Properties: relationship.Properties,
	}
}

func renderRelationshipKindInfos(ctx context.Context, relationship Relationship, source Node, target Node) {
	relationshipContext := newRelationshipContext(relationship, source, target)

	for index := range relationship.KindInfos {
		if len(relationship.KindInfos[index].Content) == 0 {
			continue
		}

		renderedMarkdown, err := renderKindInfoMarkdown(
			relationship.KindInfos[index].Content,
			relationshipContext,
		)
		if err != nil {
			relationship.KindInfos[index].RenderedMarkdown = renderedMarkdown
			relationship.KindInfos[index].TemplateError = err.Error()
			slog.WarnContext(ctx, "Failed to render relationship kind info markdown",
				slog.String("info_key", relationship.KindInfos[index].InfoKey),
				attr.Error(err),
			)
			continue
		}

		relationship.KindInfos[index].RenderedMarkdown = renderedMarkdown
	}
}

func renderNodeKindInfos(ctx context.Context, node *Node) {
	nodeContext := newNodeContext(*node)

	for index := range node.KindInfos {
		if len(node.KindInfos[index].Content) == 0 {
			continue
		}

		renderedMarkdown, err := renderKindInfoMarkdown(
			node.KindInfos[index].Content,
			nodeContext,
		)
		if err != nil {
			node.KindInfos[index].RenderedMarkdown = renderedMarkdown
			node.KindInfos[index].TemplateError = err.Error()
			slog.WarnContext(ctx, "Failed to render node kind info markdown",
				slog.String("info_key", node.KindInfos[index].InfoKey),
				attr.Error(err),
			)
			continue
		}

		node.KindInfos[index].RenderedMarkdown = renderedMarkdown
	}
}
