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

package pgtransition

import (
	"bytes"
	"fmt"
	"github.com/specterops/bloodhound/cypher/analyzer"
	"github.com/specterops/bloodhound/cypher/backend/pgsql"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/model/pg"
	"github.com/specterops/bloodhound/dawgs/query"
)

type AllShortestPathsArguments struct {
	RootCriteria      string
	TraversalCriteria string
	TerminalCriteria  string
	MaxDepth          int
}

func RewriteParameters(regularQuery *model.RegularQuery) error {
	return analyzer.Analyze(regularQuery, func(analyzerInst *analyzer.Analyzer) {
		analyzer.WithVisitor(analyzerInst, func(stack *model.WalkStack, node *model.Parameter) error {
			parameterValue := node.Value

			switch typedParameterValue := parameterValue.(type) {
			case string:
				// The cypher AST model expects strings to be contained within a single quote wrapper
				parameterValue = "'" + typedParameterValue + "'"
			}

			switch typedTrunk := stack.Trunk().(type) {
			case model.ExpressionList:
				typedTrunk.Replace(typedTrunk.IndexOf(node), query.Literal(parameterValue))

			case *model.PartialComparison:
				typedTrunk.Right = query.Literal(parameterValue)
			}

			return nil
		})
	})
}

func RemoveEmptyExpressionLists(stack *model.WalkStack, element model.Expression) error {
	var (
		shouldRemove  = false
		shouldReplace = false

		replacementExpression model.Expression
	)

	switch typedElement := element.(type) {
	case model.ExpressionList:
		shouldRemove = typedElement.Len() == 0

	case *model.Parenthetical:
		switch typedParentheticalElement := typedElement.Expression.(type) {
		case model.ExpressionList:
			numExpressions := typedParentheticalElement.Len()

			shouldRemove = numExpressions == 0
			shouldReplace = numExpressions == 1

			if shouldReplace {
				// Dump the parenthetical and the joined expression by grabbing the only element in the joined
				// expression for replacement
				replacementExpression = typedParentheticalElement.Get(0)
			}
		}
	}

	if shouldRemove {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Remove(element)
		}
	} else if shouldReplace {
		switch typedParent := stack.Trunk().(type) {
		case model.ExpressionList:
			typedParent.Replace(typedParent.IndexOf(element), replacementExpression)
		}
	}

	return nil
}

type Ripper struct {
	targetVariableSymbol string
}

func (s *Ripper) Enter(stack *model.WalkStack, expression model.Expression) error {
	if expressionList, isExpressionList := stack.Trunk().(model.ExpressionList); isExpressionList {
		switch typedExpression := expression.(type) {
		case *model.KindMatcher:
			// Look for constraints
			if variable, typeOK := typedExpression.Reference.(*model.Variable); !typeOK {
				return fmt.Errorf("expected variable in all shortests paths kind matcher but saw: %T", typedExpression.Reference)
			} else if variable.Symbol != s.targetVariableSymbol {
				// Rip this expression since it's a comparison that targets a variable we don't care about
				expressionList.Remove(expression)
			} else {
				switch s.targetVariableSymbol {
				case query.EdgeStartSymbol, query.EdgeEndSymbol:
					expressionList.Replace(expressionList.IndexOf(expression), pg.NewAnnotatedKindMatcher(typedExpression, pg.Node))

				case query.EdgeSymbol:
					expressionList.Replace(expressionList.IndexOf(expression), pg.NewAnnotatedKindMatcher(typedExpression, pg.Edge))

				default:
					return fmt.Errorf("unsupported variable symbol: %s", s.targetVariableSymbol)
				}
			}

		case *model.Comparison:
			var leftHandNode = typedExpression.Left

			// Unwrap function invocations that may wrap the left hand expression
			switch typedNode := leftHandNode.(type) {
			case *model.Variable:
			case *model.PropertyLookup:
				leftHandNode = typedNode.Atom

			case *model.FunctionInvocation:
				if typedNode.Name == model.IdentityFunction {
					// Validate the length of the arguments for sanity checking
					if len(typedNode.Arguments) != 1 {
						return fmt.Errorf("expected only 1 argument")
					}

					// If this is an ID lookup of the variable pull the variable reference out of it
					leftHandNode = typedNode.Arguments[0]
				}

			default:
				return fmt.Errorf("unexpected left hand comparison expression: %T", leftHandNode)
			}

			// Look for constraints
			if variable, typeOK := leftHandNode.(*model.Variable); !typeOK {
				return fmt.Errorf("expected *pgsql.AnnotatedVariable in all shortests paths comparison but saw: %T", leftHandNode)
			} else if variable.Symbol != s.targetVariableSymbol {
				// Rip this expression since it's a comparison that targets a variable we don't care about
				expressionList.Remove(expression)
			}
		}
	}

	return nil
}

func (s *Ripper) Exit(stack *model.WalkStack, expression model.Expression) error {
	return nil
}

func TranslateAllShortestPaths(regularQuery *model.RegularQuery, kindMapper pgsql.KindMapper) (AllShortestPathsArguments, error) {
	aspArguments := AllShortestPathsArguments{
		MaxDepth: 12,
	}

	if regularQuery.SingleQuery.MultiPartQuery != nil {
		return aspArguments, fmt.Errorf("multi-part queries not supported")
	}

	if numReadingClauses := len(regularQuery.SingleQuery.SinglePartQuery.ReadingClauses); numReadingClauses != 1 {
		return aspArguments, fmt.Errorf("expected one reading clause but saw %d", numReadingClauses)
	}

	if err := RewriteParameters(regularQuery); err != nil {
		return aspArguments, err
	}

	readingClause := regularQuery.SingleQuery.SinglePartQuery.ReadingClauses[0]

	if readingClause.Match == nil || readingClause.Match.Where == nil {
		return aspArguments, fmt.Errorf("no match or where clause specified")
	}

	if len(readingClause.Match.Where.Expressions) != 1 {
		return aspArguments, fmt.Errorf("expected where clause to have only one top-level and expression")
	}

	if topLevelConjunction, typeOK := readingClause.Match.Where.Expressions[0].(*model.Conjunction); !typeOK {
		return aspArguments, fmt.Errorf("expected where clause to have only one top-level and expression")
	} else {
		var (
			rootNodeCopy     = model.Copy(topLevelConjunction)
			edgeCopy         = model.Copy(topLevelConjunction)
			terminalNodeCopy = model.Copy(topLevelConjunction)
		)

		if err := model.Walk(rootNodeCopy, &Ripper{
			targetVariableSymbol: query.EdgeStartSymbol,
		}); err != nil {
			return aspArguments, err
		}

		if err := model.Walk(edgeCopy, &Ripper{
			targetVariableSymbol: query.EdgeSymbol,
		}); err != nil {
			return aspArguments, err
		}

		if err := model.Walk(terminalNodeCopy, &Ripper{
			targetVariableSymbol: query.EdgeEndSymbol,
		}); err != nil {
			return aspArguments, err
		}

		buffer := &bytes.Buffer{}
		emitter := pgsql.NewEmitter(false, kindMapper)

		if len(rootNodeCopy.Expressions) == 0 {
			return aspArguments, fmt.Errorf("expected start node constraints but found none")
		} else {
			if err := analyzer.Analyze(rootNodeCopy, func(analyzerInst *analyzer.Analyzer) {
				analyzer.WithVisitor(analyzerInst, RemoveEmptyExpressionLists)
			}, pg.CollectPGSQLTypes); err != nil {
				return aspArguments, err
			}

			if err := emitter.WriteExpression(buffer, rootNodeCopy); err != nil {
				return aspArguments, err
			} else {
				aspArguments.RootCriteria = buffer.String()
				buffer.Reset()
			}
		}

		if len(edgeCopy.Expressions) > 0 {
			if err := analyzer.Analyze(edgeCopy, func(analyzerInst *analyzer.Analyzer) {
				analyzer.WithVisitor(analyzerInst, RemoveEmptyExpressionLists)
			}, pg.CollectPGSQLTypes); err != nil {
				return aspArguments, err
			}

			if err := emitter.WriteExpression(buffer, edgeCopy); err != nil {
				return aspArguments, err
			}

			aspArguments.TraversalCriteria = buffer.String()
			buffer.Reset()
		}

		if len(terminalNodeCopy.Expressions) == 0 {
			return aspArguments, fmt.Errorf("expected start node constraints but found none")
		} else {
			if err := analyzer.Analyze(terminalNodeCopy, func(analyzerInst *analyzer.Analyzer) {
				analyzer.WithVisitor(analyzerInst, RemoveEmptyExpressionLists)
			}, pg.CollectPGSQLTypes); err != nil {
				return aspArguments, err
			}

			if err := emitter.WriteExpression(buffer, terminalNodeCopy); err != nil {
				return aspArguments, err
			} else {
				aspArguments.TerminalCriteria = buffer.String()
				buffer.Reset()
			}
		}
	}

	return aspArguments, nil
}
