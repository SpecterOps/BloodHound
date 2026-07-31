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
	"encoding/json"
	"fmt"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// NodeContext is the template representation of a node.
// It intentionally exposes only the node data supported by the template
// contract rather than the complete Node domain object.
type NodeContext struct {
	NodeID     int64
	Kinds      []KindContext
	Properties map[string]any
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
	"randint",
	"ago",
}

func renderNodeKindInfoMarkdown(content json.RawMessage, context NodeContext) (string, error) {
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
		return "", fmt.Errorf("parsing markdown template: %w", err)
	}

	var rendered bytes.Buffer
	if err := parsedTemplate.Execute(&rendered, context); err != nil {
		return "", fmt.Errorf("executing markdown template: %w", err)
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
