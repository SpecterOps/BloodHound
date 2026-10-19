// Copyright 2024 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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

package integrationparallel

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestIsIntegrationTestFile(t *testing.T) {
	testCases := []struct {
		name            string
		buildConstraint string
		expected        bool
	}{
		{
			name:            "integration",
			buildConstraint: "//go:build integration",
			expected:        true,
		},
		{
			name:            "slow integration",
			buildConstraint: "//go:build slow_integration",
			expected:        true,
		},
		{
			name:            "legacy integration",
			buildConstraint: "// +build integration",
			expected:        true,
		},
		{
			name:            "combined integration constraint",
			buildConstraint: "//go:build integration && linux",
			expected:        true,
		},
		{
			name:            "serial integration",
			buildConstraint: "//go:build serial_integration",
			expected:        false,
		},
		{
			name:            "negated integration",
			buildConstraint: "//go:build !integration",
			expected:        false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			fileSet := token.NewFileSet()
			file, err := parser.ParseFile(fileSet, "example_test.go", testCase.buildConstraint+"\n\npackage example", parser.ParseComments)
			if err != nil {
				t.Fatalf("parsing source: %v", err)
			}

			pass := &analysis.Pass{Fset: fileSet}
			if actual := isIntegrationTestFile(pass, file); actual != testCase.expected {
				t.Errorf("isIntegrationTestFile() = %t, want %t", actual, testCase.expected)
			}
		})
	}
}

func TestRun(t *testing.T) {
	testCases := []struct {
		name                string
		source              string
		filename            string
		expectedDiagnostics int
		expectedFixes       int
	}{
		{
			name: "testing T in integration test",
			source: `//go:build integration

package example

import "testing"

type parallelizer struct{}

func (parallelizer) Parallel() {}

func TestExample(testContext *testing.T) {
	testContext.Parallel()
	parallelizer{}.Parallel()
}
`,
			filename:            "example_integration_test.go",
			expectedDiagnostics: 1,
			expectedFixes:       1,
		},
		{
			name: "testing T in slow integration test",
			source: `//go:build slow_integration

package example

import "testing"

func TestExample(testContext *testing.T) {
	testContext.Parallel()
}
`,
			filename:            "example_slow_integration_test.go",
			expectedDiagnostics: 1,
			expectedFixes:       1,
		},
		{
			name: "testing T in serial integration test",
			source: `//go:build serial_integration

package example

import "testing"

func TestExample(testContext *testing.T) {
	testContext.Parallel()
}
`,
			filename:            "example_serial_integration_test.go",
			expectedDiagnostics: 0,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			analyzerResult := runAnalyzer(t, testCase.filename, testCase.source)
			diagnostics := analyzerResult.diagnostics
			if len(diagnostics) != testCase.expectedDiagnostics {
				t.Errorf("diagnostics = %d, want %d", len(diagnostics), testCase.expectedDiagnostics)
				return
			}

			if len(diagnostics) == 0 {
				return
			}

			suggestedFixes := diagnostics[0].SuggestedFixes
			if len(suggestedFixes) != testCase.expectedFixes {
				t.Errorf("suggested fixes = %d, want %d", len(suggestedFixes), testCase.expectedFixes)
				return
			}

			if testCase.expectedFixes == 0 {
				return
			}

			if suggestedFixes[0].Message != suggestedFixMessage {
				t.Errorf("suggested fix message = %q, want %q", suggestedFixes[0].Message, suggestedFixMessage)
			}

			if len(suggestedFixes[0].TextEdits) != 1 {
				t.Errorf("text edits = %d, want 1", len(suggestedFixes[0].TextEdits))
				return
			}

			fixedSource := applyTextEdit(t, analyzerResult.fileSet, testCase.source, suggestedFixes[0].TextEdits[0])
			if strings.Contains(fixedSource, "testContext.Parallel()") {
				t.Error("suggested fix did not remove t.Parallel()")
			}
		})
	}
}

type analyzerRunResult struct {
	diagnostics []analysis.Diagnostic
	fileSet     *token.FileSet
}

func runAnalyzer(t *testing.T, filename string, source string) analyzerRunResult {
	t.Helper()

	var (
		fileSet  = token.NewFileSet()
		typeInfo = &types.Info{
			Defs:       make(map[*ast.Ident]types.Object),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Uses:       make(map[*ast.Ident]types.Object),
		}
	)

	file, err := parser.ParseFile(fileSet, filename, source, parser.ParseComments)
	if err != nil {
		t.Fatalf("parsing source: %v", err)
	}

	packageInfo, err := (&types.Config{Importer: importer.Default()}).Check("example", fileSet, []*ast.File{file}, typeInfo)
	if err != nil {
		t.Fatalf("type checking source: %v", err)
	}

	diagnostics := make([]analysis.Diagnostic, 0)
	pass := &analysis.Pass{
		Analyzer:  Analyzer,
		Fset:      fileSet,
		Files:     []*ast.File{file},
		Pkg:       packageInfo,
		TypesInfo: typeInfo,
		Report: func(diagnostic analysis.Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	}

	if _, err := run(pass); err != nil {
		t.Fatalf("running analyzer: %v", err)
	}

	return analyzerRunResult{
		diagnostics: diagnostics,
		fileSet:     fileSet,
	}
}

func applyTextEdit(t *testing.T, fileSet *token.FileSet, source string, textEdit analysis.TextEdit) string {
	t.Helper()

	file := fileSet.File(textEdit.Pos)
	if file == nil {
		t.Fatal("finding source file for text edit")
	}

	startOffset := file.Offset(textEdit.Pos)
	endOffset := file.Offset(textEdit.End)
	return source[:startOffset] + string(textEdit.NewText) + source[endOffset:]
}
