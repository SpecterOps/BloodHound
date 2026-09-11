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
	"go/build/constraint"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

const (
	diagnosticMessage   = "t.Parallel is not permitted in integration tests"
	suggestedFixMessage = "Remove t.Parallel()"
)

var Analyzer = &analysis.Analyzer{
	Name: "integrationparallel",
	Doc:  "disallow t.Parallel in integration tests",
	Run:  run,
}

type plugin struct{}

func init() {
	register.Plugin("integrationparallel", New)
}

func New(_ any) (register.LinterPlugin, error) {
	return plugin{}, nil
}

func (s plugin) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{Analyzer}, nil
}

func (s plugin) GetLoadMode() string {
	return register.LoadModeTypesInfo
}

func run(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		if !isIntegrationTestFile(pass, file) {
			continue
		}

		var nodeStack []ast.Node

		ast.Inspect(file, func(node ast.Node) bool {
			if node == nil {
				nodeStack = nodeStack[:len(nodeStack)-1]
				return true
			}

			var parentNode ast.Node
			if len(nodeStack) > 0 {
				parentNode = nodeStack[len(nodeStack)-1]
			}

			nodeStack = append(nodeStack, node)

			callExpression, ok := node.(*ast.CallExpr)
			if !ok {
				return true
			}

			selectorExpression, ok := callExpression.Fun.(*ast.SelectorExpr)
			if !ok || !isTestingTParallel(pass, selectorExpression) {
				return true
			}

			diagnostic := analysis.Diagnostic{
				Pos:     selectorExpression.Sel.Pos(),
				End:     selectorExpression.Sel.End(),
				Message: diagnosticMessage,
			}

			if statement := statementForCall(parentNode, callExpression); statement != nil {
				diagnostic.SuggestedFixes = []analysis.SuggestedFix{{
					Message: suggestedFixMessage,
					TextEdits: []analysis.TextEdit{{
						Pos: statement.Pos(),
						End: statement.End(),
					}},
				}}
			}

			pass.Report(diagnostic)

			return true
		})
	}

	return nil, nil
}

func statementForCall(parentNode ast.Node, callExpression *ast.CallExpr) ast.Stmt {
	switch statement := parentNode.(type) {
	case *ast.ExprStmt:
		if statement.X == callExpression {
			return statement
		}
	case *ast.DeferStmt:
		if statement.Call == callExpression {
			return statement
		}
	case *ast.GoStmt:
		if statement.Call == callExpression {
			return statement
		}
	}

	return nil
}

func isIntegrationTestFile(pass *analysis.Pass, file *ast.File) bool {
	filename := pass.Fset.File(file.Pos()).Name()
	if !strings.HasSuffix(filepath.Base(filename), "_test.go") {
		return false
	}

	for _, commentGroup := range file.Comments {
		for _, comment := range commentGroup.List {
			if !constraint.IsGoBuild(comment.Text) && !constraint.IsPlusBuild(comment.Text) {
				continue
			}

			buildConstraint, err := constraint.Parse(comment.Text)
			if err != nil {
				continue
			}

			return containsIntegrationTag(buildConstraint, false)
		}
	}

	return false
}

func containsIntegrationTag(buildConstraint constraint.Expr, negated bool) bool {
	switch expression := buildConstraint.(type) {
	case *constraint.TagExpr:
		return !negated && (expression.Tag == "integration" || expression.Tag == "slow_integration")
	case *constraint.NotExpr:
		return containsIntegrationTag(expression.X, !negated)
	case *constraint.AndExpr:
		return containsIntegrationTag(expression.X, negated) || containsIntegrationTag(expression.Y, negated)
	case *constraint.OrExpr:
		return containsIntegrationTag(expression.X, negated) || containsIntegrationTag(expression.Y, negated)
	default:
		return false
	}
}

func isTestingTParallel(pass *analysis.Pass, selectorExpression *ast.SelectorExpr) bool {
	selection := pass.TypesInfo.Selections[selectorExpression]
	if selection == nil || selection.Obj().Name() != "Parallel" {
		return false
	}

	method, ok := selection.Obj().(*types.Func)
	if !ok {
		return false
	}

	methodSignature, ok := method.Type().(*types.Signature)
	if !ok || methodSignature.Recv() == nil {
		return false
	}

	receiverType, ok := types.Unalias(methodSignature.Recv().Type()).(*types.Pointer)
	if !ok {
		return false
	}

	namedReceiverType, ok := types.Unalias(receiverType.Elem()).(*types.Named)
	if !ok {
		return false
	}

	typeObject := namedReceiverType.Obj()
	return typeObject.Name() == "T" && typeObject.Pkg() != nil && typeObject.Pkg().Path() == "testing"
}
