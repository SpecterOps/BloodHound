// Copyright 2023 Specter Ops, Inc.
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

package frontend

import (
	"errors"
	"fmt"

	"github.com/antlr4-go/antlr/v4"
	"github.com/specterops/bloodhound/cypher/parser"
)

type descentEntry struct {
	visitor Visitor
	depth   int
}

// Context satisfies the antlr.ParseTreeListener interface needed for antlr's tree walker.
type Context struct {
	visitorStack []*descentEntry
	filters      []Visitor

	Errors []error
}

func NewContext(filters ...Visitor) *Context {
	ctx := &Context{
		filters: filters,
	}

	for _, filter := range filters {
		filter.SetContext(ctx)
	}

	return ctx
}

func (s *Context) SyntaxError(recognizer antlr.Recognizer, offendingSymbol any, line, column int, msg string, e antlr.RecognitionException) {
	s.AddErrors(&SyntaxError{
		Line:            line,
		Column:          column,
		OffendingSymbol: offendingSymbol,
		Message:         msg,
	})
}

func (s *Context) ReportAmbiguity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, exact bool, ambigAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
}

func (s *Context) ReportAttemptingFullContext(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex int, conflictingAlts *antlr.BitSet, configs *antlr.ATNConfigSet) {
}

func (s *Context) ReportContextSensitivity(recognizer antlr.Parser, dfa *antlr.DFA, startIndex, stopIndex, prediction int, configs *antlr.ATNConfigSet) {
}

func (s *Context) GetErrors() error {
	return errors.Join(s.Errors...)
}

func (s *Context) AddErrors(errs ...error) {
	for _, err := range errs {
		if err != nil {
			s.Errors = append(s.Errors, err)
		}
	}
}

func (s *Context) Enter(visitor Visitor) {
	s.visitorStack = append(s.visitorStack, &descentEntry{
		visitor: visitor,
	})

	visitor.SetContext(s)
}

func (s *Context) Exit() Visitor {
	var (
		idx             = len(s.visitorStack) - 1
		previousVisitor = s.visitorStack[idx]
	)

	if previousVisitor.depth != 0 {
		panic(fmt.Sprintf("Depth of visitor is %d but expected 0.", previousVisitor.depth))
	}

	s.visitorStack = s.visitorStack[:idx]
	return previousVisitor.visitor
}

func (s *Context) EnterEveryRule(ctx antlr.ParserRuleContext) {
	// Filter entry for this rule
	for _, filter := range s.filters {
		ctx.EnterRule(filter)
	}

	// Lookup the current visitor to increment its depth. This is done without checking to see if the
	// depth of the current descendant is equal to zero since entry is always allowed
	currentVisitorEntry := s.visitorStack[len(s.visitorStack)-1]
	currentVisitorEntry.depth++

	// Finally delegate to the visitor
	ctx.EnterRule(currentVisitorEntry.visitor)
}

func (s *Context) ExitEveryRule(ctx antlr.ParserRuleContext) {
	// Lookup the current visitor to decrement its depth.
	currentVisitorEntry := s.visitorStack[len(s.visitorStack)-1]

	// If the depth of the next descent entry is 0 then we've reached the point where
	// this visitor is expected to be popped off by its ancestor visitor
	if currentVisitorEntry.depth == 0 {
		currentVisitorEntry = s.visitorStack[len(s.visitorStack)-2]
	}

	currentVisitorEntry.depth--

	// Finally delegate to the visitor
	ctx.ExitRule(currentVisitorEntry.visitor)
}

func (s *Context) VisitTerminal(node antlr.TerminalNode) {
	s.visitorStack[len(s.visitorStack)-1].visitor.VisitTerminal(node)
}

func (s *Context) VisitErrorNode(node antlr.ErrorNode) {
	s.visitorStack[len(s.visitorStack)-1].visitor.VisitErrorNode(node)
}

type ContextAware interface {
	SetContext(ctx *Context)
}

type Visitor interface {
	parser.CypherListener
	ContextAware
}

type BaseVisitor struct {
	ctx *Context
}

func (s *BaseVisitor) newUnsupportedRuleError(c antlr.ParserRuleContext) {
	s.ctx.AddErrors(
		SyntaxError{
			Line:            c.GetStart().GetLine(),
			Column:          c.GetStart().GetColumn(),
			OffendingSymbol: c.GetText(),
			Message:         fmt.Sprintf("%s rule is not supported", parser.CypherParserStaticData.RuleNames[c.GetRuleIndex()]),
		},
	)
}

func (s *BaseVisitor) SetContext(ctx *Context) {
	s.ctx = ctx
}

/**************** UNSUPPORTED RULES IN GRAMMAR */
func (s *BaseVisitor) EnterOC_Profile(c *parser.OC_ProfileContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_BulkImportQuery(c *parser.OC_BulkImportQueryContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_PeriodicCommitHint(c *parser.OC_PeriodicCommitHintContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_Union(c *parser.OC_UnionContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_Command(c *parser.OC_CommandContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_Foreach(c *parser.OC_ForeachContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_Start(c *parser.OC_StartContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_CaseExpression(c *parser.OC_CaseExpressionContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_LegacyListExpression(c *parser.OC_LegacyListExpressionContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_Reduce(c *parser.OC_ReduceContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_ExistentialSubquery(c *parser.OC_ExistentialSubqueryContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_LegacyParameter(c *parser.OC_LegacyParameterContext) {
	s.newUnsupportedRuleError(c)
}

func (s *BaseVisitor) EnterOC_Explain(c *parser.OC_ExplainContext) {
	s.newUnsupportedRuleError(c)
}

/**************** EMPTY STUBS ON BASEVISITOR  */
func (s *BaseVisitor) VisitTerminal(node antlr.TerminalNode) {}

func (s *BaseVisitor) VisitErrorNode(node antlr.ErrorNode) {}

func (s *BaseVisitor) EnterEveryRule(ctx antlr.ParserRuleContext) {}

func (s *BaseVisitor) ExitEveryRule(ctx antlr.ParserRuleContext) {}

func (s *BaseVisitor) EnterOC_Cypher(c *parser.OC_CypherContext) {}

func (s *BaseVisitor) EnterOC_QueryOptions(c *parser.OC_QueryOptionsContext) {}

func (s *BaseVisitor) EnterOC_AnyCypherOption(c *parser.OC_AnyCypherOptionContext) {}

func (s *BaseVisitor) EnterOC_CypherOption(c *parser.OC_CypherOptionContext) {}

func (s *BaseVisitor) EnterOC_VersionNumber(c *parser.OC_VersionNumberContext) {}

func (s *BaseVisitor) EnterOC_ConfigurationOption(c *parser.OC_ConfigurationOptionContext) {}

func (s *BaseVisitor) EnterOC_Statement(c *parser.OC_StatementContext) {}

func (s *BaseVisitor) EnterOC_Query(c *parser.OC_QueryContext) {}

func (s *BaseVisitor) EnterOC_RegularQuery(c *parser.OC_RegularQueryContext) {}

func (s *BaseVisitor) EnterOC_LoadCSVQuery(c *parser.OC_LoadCSVQueryContext) {}

func (s *BaseVisitor) EnterOC_SingleQuery(c *parser.OC_SingleQueryContext) {}

func (s *BaseVisitor) EnterOC_SinglePartQuery(c *parser.OC_SinglePartQueryContext) {}

func (s *BaseVisitor) EnterOC_MultiPartQuery(c *parser.OC_MultiPartQueryContext) {}

func (s *BaseVisitor) EnterOC_UpdatingClause(c *parser.OC_UpdatingClauseContext) {}

func (s *BaseVisitor) EnterOC_ReadingClause(c *parser.OC_ReadingClauseContext) {}

func (s *BaseVisitor) EnterOC_CreateUniqueConstraint(c *parser.OC_CreateUniqueConstraintContext) {}

func (s *BaseVisitor) EnterOC_CreateNodePropertyExistenceConstraint(c *parser.OC_CreateNodePropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) EnterOC_CreateRelationshipPropertyExistenceConstraint(c *parser.OC_CreateRelationshipPropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) EnterOC_CreateIndex(c *parser.OC_CreateIndexContext) {}

func (s *BaseVisitor) EnterOC_DropUniqueConstraint(c *parser.OC_DropUniqueConstraintContext) {}

func (s *BaseVisitor) EnterOC_DropNodePropertyExistenceConstraint(c *parser.OC_DropNodePropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) EnterOC_DropRelationshipPropertyExistenceConstraint(c *parser.OC_DropRelationshipPropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) EnterOC_DropIndex(c *parser.OC_DropIndexContext) {}

func (s *BaseVisitor) EnterOC_Index(c *parser.OC_IndexContext) {}

func (s *BaseVisitor) EnterOC_UniqueConstraint(c *parser.OC_UniqueConstraintContext) {}

func (s *BaseVisitor) EnterOC_NodePropertyExistenceConstraint(c *parser.OC_NodePropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) EnterOC_RelationshipPropertyExistenceConstraint(c *parser.OC_RelationshipPropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) EnterOC_RelationshipPatternSyntax(c *parser.OC_RelationshipPatternSyntaxContext) {
}

func (s *BaseVisitor) EnterOC_LoadCSV(c *parser.OC_LoadCSVContext) {}

func (s *BaseVisitor) EnterOC_Match(c *parser.OC_MatchContext) {}

func (s *BaseVisitor) EnterOC_Unwind(c *parser.OC_UnwindContext) {}

func (s *BaseVisitor) EnterOC_Merge(c *parser.OC_MergeContext) {}

func (s *BaseVisitor) EnterOC_MergeAction(c *parser.OC_MergeActionContext) {}

func (s *BaseVisitor) EnterOC_Create(c *parser.OC_CreateContext) {}

func (s *BaseVisitor) EnterOC_CreateUnique(c *parser.OC_CreateUniqueContext) {}

func (s *BaseVisitor) EnterOC_Set(c *parser.OC_SetContext) {}

func (s *BaseVisitor) EnterOC_SetItem(c *parser.OC_SetItemContext) {}

func (s *BaseVisitor) EnterOC_Delete(c *parser.OC_DeleteContext) {}

func (s *BaseVisitor) EnterOC_Remove(c *parser.OC_RemoveContext) {}

func (s *BaseVisitor) EnterOC_RemoveItem(c *parser.OC_RemoveItemContext) {}

func (s *BaseVisitor) EnterOC_InQueryCall(c *parser.OC_InQueryCallContext) {}

func (s *BaseVisitor) EnterOC_StandaloneCall(c *parser.OC_StandaloneCallContext) {}

func (s *BaseVisitor) EnterOC_YieldItems(c *parser.OC_YieldItemsContext) {}

func (s *BaseVisitor) EnterOC_YieldItem(c *parser.OC_YieldItemContext) {}

func (s *BaseVisitor) EnterOC_With(c *parser.OC_WithContext) {}

func (s *BaseVisitor) EnterOC_Return(c *parser.OC_ReturnContext) {}

func (s *BaseVisitor) EnterOC_ProjectionBody(c *parser.OC_ProjectionBodyContext) {}

func (s *BaseVisitor) EnterOC_ProjectionItems(c *parser.OC_ProjectionItemsContext) {}

func (s *BaseVisitor) EnterOC_ProjectionItem(c *parser.OC_ProjectionItemContext) {}

func (s *BaseVisitor) EnterOC_Order(c *parser.OC_OrderContext) {}

func (s *BaseVisitor) EnterOC_Skip(c *parser.OC_SkipContext) {}

func (s *BaseVisitor) EnterOC_Limit(c *parser.OC_LimitContext) {}

func (s *BaseVisitor) EnterOC_SortItem(c *parser.OC_SortItemContext) {}

func (s *BaseVisitor) EnterOC_Hint(c *parser.OC_HintContext) {}

func (s *BaseVisitor) EnterOC_StartPoint(c *parser.OC_StartPointContext) {}

func (s *BaseVisitor) EnterOC_Lookup(c *parser.OC_LookupContext) {}

func (s *BaseVisitor) EnterOC_NodeLookup(c *parser.OC_NodeLookupContext) {}

func (s *BaseVisitor) EnterOC_RelationshipLookup(c *parser.OC_RelationshipLookupContext) {}

func (s *BaseVisitor) EnterOC_IdentifiedIndexLookup(c *parser.OC_IdentifiedIndexLookupContext) {}

func (s *BaseVisitor) EnterOC_IndexQuery(c *parser.OC_IndexQueryContext) {}

func (s *BaseVisitor) EnterOC_IdLookup(c *parser.OC_IdLookupContext) {}

func (s *BaseVisitor) EnterOC_LiteralIds(c *parser.OC_LiteralIdsContext) {}

func (s *BaseVisitor) EnterOC_Where(c *parser.OC_WhereContext) {}

func (s *BaseVisitor) EnterOC_Pattern(c *parser.OC_PatternContext) {}

func (s *BaseVisitor) EnterOC_PatternPart(c *parser.OC_PatternPartContext) {}

func (s *BaseVisitor) EnterOC_AnonymousPatternPart(c *parser.OC_AnonymousPatternPartContext) {}

func (s *BaseVisitor) EnterOC_ShortestPathPattern(c *parser.OC_ShortestPathPatternContext) {}

func (s *BaseVisitor) EnterOC_PatternElement(c *parser.OC_PatternElementContext) {}

func (s *BaseVisitor) EnterOC_RelationshipsPattern(c *parser.OC_RelationshipsPatternContext) {}

func (s *BaseVisitor) EnterOC_NodePattern(c *parser.OC_NodePatternContext) {}

func (s *BaseVisitor) EnterOC_PatternElementChain(c *parser.OC_PatternElementChainContext) {}

func (s *BaseVisitor) EnterOC_RelationshipPattern(c *parser.OC_RelationshipPatternContext) {}

func (s *BaseVisitor) EnterOC_RelationshipDetail(c *parser.OC_RelationshipDetailContext) {}

func (s *BaseVisitor) EnterOC_Properties(c *parser.OC_PropertiesContext) {}

func (s *BaseVisitor) EnterOC_RelType(c *parser.OC_RelTypeContext) {}

func (s *BaseVisitor) EnterOC_RelationshipTypes(c *parser.OC_RelationshipTypesContext) {}

func (s *BaseVisitor) EnterOC_NodeLabels(c *parser.OC_NodeLabelsContext) {}

func (s *BaseVisitor) EnterOC_NodeLabel(c *parser.OC_NodeLabelContext) {}

func (s *BaseVisitor) EnterOC_RangeLiteral(c *parser.OC_RangeLiteralContext) {}

func (s *BaseVisitor) EnterOC_LabelName(c *parser.OC_LabelNameContext) {}

func (s *BaseVisitor) EnterOC_RelTypeName(c *parser.OC_RelTypeNameContext) {}

func (s *BaseVisitor) EnterOC_PropertyExpression(c *parser.OC_PropertyExpressionContext) {}

func (s *BaseVisitor) EnterOC_Expression(c *parser.OC_ExpressionContext) {}

func (s *BaseVisitor) EnterOC_OrExpression(c *parser.OC_OrExpressionContext) {}

func (s *BaseVisitor) EnterOC_XorExpression(c *parser.OC_XorExpressionContext) {}

func (s *BaseVisitor) EnterOC_AndExpression(c *parser.OC_AndExpressionContext) {}

func (s *BaseVisitor) EnterOC_NotExpression(c *parser.OC_NotExpressionContext) {}

func (s *BaseVisitor) EnterOC_ComparisonExpression(c *parser.OC_ComparisonExpressionContext) {}

func (s *BaseVisitor) EnterOC_PartialComparisonExpression(c *parser.OC_PartialComparisonExpressionContext) {
}

func (s *BaseVisitor) EnterOC_StringListNullPredicateExpression(c *parser.OC_StringListNullPredicateExpressionContext) {
}

func (s *BaseVisitor) EnterOC_StringPredicateExpression(c *parser.OC_StringPredicateExpressionContext) {
}

func (s *BaseVisitor) EnterOC_ListPredicateExpression(c *parser.OC_ListPredicateExpressionContext) {}

func (s *BaseVisitor) EnterOC_NullPredicateExpression(c *parser.OC_NullPredicateExpressionContext) {}

func (s *BaseVisitor) EnterOC_RegularExpression(c *parser.OC_RegularExpressionContext) {}

func (s *BaseVisitor) EnterOC_AddOrSubtractExpression(c *parser.OC_AddOrSubtractExpressionContext) {}

func (s *BaseVisitor) EnterOC_MultiplyDivideModuloExpression(c *parser.OC_MultiplyDivideModuloExpressionContext) {
}

func (s *BaseVisitor) EnterOC_PowerOfExpression(c *parser.OC_PowerOfExpressionContext) {}

func (s *BaseVisitor) EnterOC_UnaryAddOrSubtractExpression(c *parser.OC_UnaryAddOrSubtractExpressionContext) {
}

func (s *BaseVisitor) EnterOC_NonArithmeticOperatorExpression(c *parser.OC_NonArithmeticOperatorExpressionContext) {
}

func (s *BaseVisitor) EnterOC_ListOperatorExpression(c *parser.OC_ListOperatorExpressionContext) {}

func (s *BaseVisitor) EnterOC_PropertyLookup(c *parser.OC_PropertyLookupContext) {}

func (s *BaseVisitor) EnterOC_Atom(c *parser.OC_AtomContext) {}

func (s *BaseVisitor) EnterOC_CaseAlternative(c *parser.OC_CaseAlternativeContext) {}

func (s *BaseVisitor) EnterOC_ListComprehension(c *parser.OC_ListComprehensionContext) {}

func (s *BaseVisitor) EnterOC_PatternComprehension(c *parser.OC_PatternComprehensionContext) {}

func (s *BaseVisitor) EnterOC_Quantifier(c *parser.OC_QuantifierContext) {}

func (s *BaseVisitor) EnterOC_FilterExpression(c *parser.OC_FilterExpressionContext) {}

func (s *BaseVisitor) EnterOC_PatternPredicate(c *parser.OC_PatternPredicateContext) {}

func (s *BaseVisitor) EnterOC_ParenthesizedExpression(c *parser.OC_ParenthesizedExpressionContext) {}

func (s *BaseVisitor) EnterOC_IdInColl(c *parser.OC_IdInCollContext) {}

func (s *BaseVisitor) EnterOC_FunctionInvocation(c *parser.OC_FunctionInvocationContext) {}

func (s *BaseVisitor) EnterOC_FunctionName(c *parser.OC_FunctionNameContext) {}

func (s *BaseVisitor) EnterOC_ExplicitProcedureInvocation(c *parser.OC_ExplicitProcedureInvocationContext) {
}

func (s *BaseVisitor) EnterOC_ImplicitProcedureInvocation(c *parser.OC_ImplicitProcedureInvocationContext) {
}

func (s *BaseVisitor) EnterOC_ProcedureResultField(c *parser.OC_ProcedureResultFieldContext) {}

func (s *BaseVisitor) EnterOC_ProcedureName(c *parser.OC_ProcedureNameContext) {}

func (s *BaseVisitor) EnterOC_Namespace(c *parser.OC_NamespaceContext) {}

func (s *BaseVisitor) EnterOC_Variable(c *parser.OC_VariableContext) {}

func (s *BaseVisitor) EnterOC_Literal(c *parser.OC_LiteralContext) {}

func (s *BaseVisitor) EnterOC_BooleanLiteral(c *parser.OC_BooleanLiteralContext) {}

func (s *BaseVisitor) EnterOC_NumberLiteral(c *parser.OC_NumberLiteralContext) {}

func (s *BaseVisitor) EnterOC_IntegerLiteral(c *parser.OC_IntegerLiteralContext) {}

func (s *BaseVisitor) EnterOC_DoubleLiteral(c *parser.OC_DoubleLiteralContext) {}

func (s *BaseVisitor) EnterOC_ListLiteral(c *parser.OC_ListLiteralContext) {}

func (s *BaseVisitor) EnterOC_MapLiteral(c *parser.OC_MapLiteralContext) {}

func (s *BaseVisitor) EnterOC_PropertyKeyName(c *parser.OC_PropertyKeyNameContext) {}

func (s *BaseVisitor) EnterOC_Parameter(c *parser.OC_ParameterContext) {}

func (s *BaseVisitor) EnterOC_SchemaName(c *parser.OC_SchemaNameContext) {}

func (s *BaseVisitor) EnterOC_ReservedWord(c *parser.OC_ReservedWordContext) {}

func (s *BaseVisitor) EnterOC_SymbolicName(c *parser.OC_SymbolicNameContext) {}

func (s *BaseVisitor) EnterOC_LeftArrowHead(c *parser.OC_LeftArrowHeadContext) {}

func (s *BaseVisitor) EnterOC_RightArrowHead(c *parser.OC_RightArrowHeadContext) {}

func (s *BaseVisitor) EnterOC_Dash(c *parser.OC_DashContext) {}

func (s *BaseVisitor) ExitOC_Cypher(c *parser.OC_CypherContext) {}

func (s *BaseVisitor) ExitOC_QueryOptions(c *parser.OC_QueryOptionsContext) {}

func (s *BaseVisitor) ExitOC_AnyCypherOption(c *parser.OC_AnyCypherOptionContext) {}

func (s *BaseVisitor) ExitOC_CypherOption(c *parser.OC_CypherOptionContext) {}

func (s *BaseVisitor) ExitOC_VersionNumber(c *parser.OC_VersionNumberContext) {}

func (s *BaseVisitor) ExitOC_Explain(c *parser.OC_ExplainContext) {}

func (s *BaseVisitor) ExitOC_Profile(c *parser.OC_ProfileContext) {}

func (s *BaseVisitor) ExitOC_ConfigurationOption(c *parser.OC_ConfigurationOptionContext) {}

func (s *BaseVisitor) ExitOC_Statement(c *parser.OC_StatementContext) {}

func (s *BaseVisitor) ExitOC_Query(c *parser.OC_QueryContext) {}

func (s *BaseVisitor) ExitOC_RegularQuery(c *parser.OC_RegularQueryContext) {}

func (s *BaseVisitor) ExitOC_BulkImportQuery(c *parser.OC_BulkImportQueryContext) {}

func (s *BaseVisitor) ExitOC_PeriodicCommitHint(c *parser.OC_PeriodicCommitHintContext) {}

func (s *BaseVisitor) ExitOC_LoadCSVQuery(c *parser.OC_LoadCSVQueryContext) {}

func (s *BaseVisitor) ExitOC_Union(c *parser.OC_UnionContext) {}

func (s *BaseVisitor) ExitOC_SingleQuery(c *parser.OC_SingleQueryContext) {}

func (s *BaseVisitor) ExitOC_SinglePartQuery(c *parser.OC_SinglePartQueryContext) {}

func (s *BaseVisitor) ExitOC_MultiPartQuery(c *parser.OC_MultiPartQueryContext) {}

func (s *BaseVisitor) ExitOC_UpdatingClause(c *parser.OC_UpdatingClauseContext) {}

func (s *BaseVisitor) ExitOC_ReadingClause(c *parser.OC_ReadingClauseContext) {}

func (s *BaseVisitor) ExitOC_Command(c *parser.OC_CommandContext) {}

func (s *BaseVisitor) ExitOC_CreateUniqueConstraint(c *parser.OC_CreateUniqueConstraintContext) {}

func (s *BaseVisitor) ExitOC_CreateNodePropertyExistenceConstraint(c *parser.OC_CreateNodePropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) ExitOC_CreateRelationshipPropertyExistenceConstraint(c *parser.OC_CreateRelationshipPropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) ExitOC_CreateIndex(c *parser.OC_CreateIndexContext) {}

func (s *BaseVisitor) ExitOC_DropUniqueConstraint(c *parser.OC_DropUniqueConstraintContext) {}

func (s *BaseVisitor) ExitOC_DropNodePropertyExistenceConstraint(c *parser.OC_DropNodePropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) ExitOC_DropRelationshipPropertyExistenceConstraint(c *parser.OC_DropRelationshipPropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) ExitOC_DropIndex(c *parser.OC_DropIndexContext) {}

func (s *BaseVisitor) ExitOC_Index(c *parser.OC_IndexContext) {}

func (s *BaseVisitor) ExitOC_UniqueConstraint(c *parser.OC_UniqueConstraintContext) {}

func (s *BaseVisitor) ExitOC_NodePropertyExistenceConstraint(c *parser.OC_NodePropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) ExitOC_RelationshipPropertyExistenceConstraint(c *parser.OC_RelationshipPropertyExistenceConstraintContext) {
}

func (s *BaseVisitor) ExitOC_RelationshipPatternSyntax(c *parser.OC_RelationshipPatternSyntaxContext) {
}

func (s *BaseVisitor) ExitOC_LoadCSV(c *parser.OC_LoadCSVContext) {}

func (s *BaseVisitor) ExitOC_Match(c *parser.OC_MatchContext) {}

func (s *BaseVisitor) ExitOC_Unwind(c *parser.OC_UnwindContext) {}

func (s *BaseVisitor) ExitOC_Merge(c *parser.OC_MergeContext) {}

func (s *BaseVisitor) ExitOC_MergeAction(c *parser.OC_MergeActionContext) {}

func (s *BaseVisitor) ExitOC_Create(c *parser.OC_CreateContext) {}

func (s *BaseVisitor) ExitOC_CreateUnique(c *parser.OC_CreateUniqueContext) {}

func (s *BaseVisitor) ExitOC_Set(c *parser.OC_SetContext) {}

func (s *BaseVisitor) ExitOC_SetItem(c *parser.OC_SetItemContext) {}

func (s *BaseVisitor) ExitOC_Delete(c *parser.OC_DeleteContext) {}

func (s *BaseVisitor) ExitOC_Remove(c *parser.OC_RemoveContext) {}

func (s *BaseVisitor) ExitOC_RemoveItem(c *parser.OC_RemoveItemContext) {}

func (s *BaseVisitor) ExitOC_Foreach(c *parser.OC_ForeachContext) {}

func (s *BaseVisitor) ExitOC_InQueryCall(c *parser.OC_InQueryCallContext) {}

func (s *BaseVisitor) ExitOC_StandaloneCall(c *parser.OC_StandaloneCallContext) {}

func (s *BaseVisitor) ExitOC_YieldItems(c *parser.OC_YieldItemsContext) {}

func (s *BaseVisitor) ExitOC_YieldItem(c *parser.OC_YieldItemContext) {}

func (s *BaseVisitor) ExitOC_With(c *parser.OC_WithContext) {}

func (s *BaseVisitor) ExitOC_Return(c *parser.OC_ReturnContext) {}

func (s *BaseVisitor) ExitOC_ProjectionBody(c *parser.OC_ProjectionBodyContext) {}

func (s *BaseVisitor) ExitOC_ProjectionItems(c *parser.OC_ProjectionItemsContext) {}

func (s *BaseVisitor) ExitOC_ProjectionItem(c *parser.OC_ProjectionItemContext) {}

func (s *BaseVisitor) ExitOC_Order(c *parser.OC_OrderContext) {}

func (s *BaseVisitor) ExitOC_Skip(c *parser.OC_SkipContext) {}

func (s *BaseVisitor) ExitOC_Limit(c *parser.OC_LimitContext) {}

func (s *BaseVisitor) ExitOC_SortItem(c *parser.OC_SortItemContext) {}

func (s *BaseVisitor) ExitOC_Hint(c *parser.OC_HintContext) {}

func (s *BaseVisitor) ExitOC_Start(c *parser.OC_StartContext) {}

func (s *BaseVisitor) ExitOC_StartPoint(c *parser.OC_StartPointContext) {}

func (s *BaseVisitor) ExitOC_Lookup(c *parser.OC_LookupContext) {}

func (s *BaseVisitor) ExitOC_NodeLookup(c *parser.OC_NodeLookupContext) {}

func (s *BaseVisitor) ExitOC_RelationshipLookup(c *parser.OC_RelationshipLookupContext) {}

func (s *BaseVisitor) ExitOC_IdentifiedIndexLookup(c *parser.OC_IdentifiedIndexLookupContext) {}

func (s *BaseVisitor) ExitOC_IndexQuery(c *parser.OC_IndexQueryContext) {}

func (s *BaseVisitor) ExitOC_IdLookup(c *parser.OC_IdLookupContext) {}

func (s *BaseVisitor) ExitOC_LiteralIds(c *parser.OC_LiteralIdsContext) {}

func (s *BaseVisitor) ExitOC_Where(c *parser.OC_WhereContext) {}

func (s *BaseVisitor) ExitOC_Pattern(c *parser.OC_PatternContext) {}

func (s *BaseVisitor) ExitOC_PatternPart(c *parser.OC_PatternPartContext) {}

func (s *BaseVisitor) ExitOC_AnonymousPatternPart(c *parser.OC_AnonymousPatternPartContext) {}

func (s *BaseVisitor) ExitOC_ShortestPathPattern(c *parser.OC_ShortestPathPatternContext) {}

func (s *BaseVisitor) ExitOC_PatternElement(c *parser.OC_PatternElementContext) {}

func (s *BaseVisitor) ExitOC_RelationshipsPattern(c *parser.OC_RelationshipsPatternContext) {}

func (s *BaseVisitor) ExitOC_NodePattern(c *parser.OC_NodePatternContext) {}

func (s *BaseVisitor) ExitOC_PatternElementChain(c *parser.OC_PatternElementChainContext) {}

func (s *BaseVisitor) ExitOC_RelationshipPattern(c *parser.OC_RelationshipPatternContext) {}

func (s *BaseVisitor) ExitOC_RelationshipDetail(c *parser.OC_RelationshipDetailContext) {}

func (s *BaseVisitor) ExitOC_Properties(c *parser.OC_PropertiesContext) {}

func (s *BaseVisitor) ExitOC_RelType(c *parser.OC_RelTypeContext) {}

func (s *BaseVisitor) ExitOC_RelationshipTypes(c *parser.OC_RelationshipTypesContext) {}

func (s *BaseVisitor) ExitOC_NodeLabels(c *parser.OC_NodeLabelsContext) {}

func (s *BaseVisitor) ExitOC_NodeLabel(c *parser.OC_NodeLabelContext) {}

func (s *BaseVisitor) ExitOC_RangeLiteral(c *parser.OC_RangeLiteralContext) {}

func (s *BaseVisitor) ExitOC_LabelName(c *parser.OC_LabelNameContext) {}

func (s *BaseVisitor) ExitOC_RelTypeName(c *parser.OC_RelTypeNameContext) {}

func (s *BaseVisitor) ExitOC_PropertyExpression(c *parser.OC_PropertyExpressionContext) {}

func (s *BaseVisitor) ExitOC_Expression(c *parser.OC_ExpressionContext) {}

func (s *BaseVisitor) ExitOC_OrExpression(c *parser.OC_OrExpressionContext) {}

func (s *BaseVisitor) ExitOC_XorExpression(c *parser.OC_XorExpressionContext) {}

func (s *BaseVisitor) ExitOC_AndExpression(c *parser.OC_AndExpressionContext) {}

func (s *BaseVisitor) ExitOC_NotExpression(c *parser.OC_NotExpressionContext) {}

func (s *BaseVisitor) ExitOC_ComparisonExpression(c *parser.OC_ComparisonExpressionContext) {}

func (s *BaseVisitor) ExitOC_PartialComparisonExpression(c *parser.OC_PartialComparisonExpressionContext) {
}

func (s *BaseVisitor) ExitOC_StringListNullPredicateExpression(c *parser.OC_StringListNullPredicateExpressionContext) {
}

func (s *BaseVisitor) ExitOC_StringPredicateExpression(c *parser.OC_StringPredicateExpressionContext) {
}

func (s *BaseVisitor) ExitOC_ListPredicateExpression(c *parser.OC_ListPredicateExpressionContext) {}

func (s *BaseVisitor) ExitOC_NullPredicateExpression(c *parser.OC_NullPredicateExpressionContext) {}

func (s *BaseVisitor) ExitOC_RegularExpression(c *parser.OC_RegularExpressionContext) {}

func (s *BaseVisitor) ExitOC_AddOrSubtractExpression(c *parser.OC_AddOrSubtractExpressionContext) {}

func (s *BaseVisitor) ExitOC_MultiplyDivideModuloExpression(c *parser.OC_MultiplyDivideModuloExpressionContext) {
}

func (s *BaseVisitor) ExitOC_PowerOfExpression(c *parser.OC_PowerOfExpressionContext) {}

func (s *BaseVisitor) ExitOC_UnaryAddOrSubtractExpression(c *parser.OC_UnaryAddOrSubtractExpressionContext) {
}

func (s *BaseVisitor) ExitOC_NonArithmeticOperatorExpression(c *parser.OC_NonArithmeticOperatorExpressionContext) {
}

func (s *BaseVisitor) ExitOC_ListOperatorExpression(c *parser.OC_ListOperatorExpressionContext) {}

func (s *BaseVisitor) ExitOC_PropertyLookup(c *parser.OC_PropertyLookupContext) {}

func (s *BaseVisitor) ExitOC_Atom(c *parser.OC_AtomContext) {}

func (s *BaseVisitor) ExitOC_CaseExpression(c *parser.OC_CaseExpressionContext) {}

func (s *BaseVisitor) ExitOC_CaseAlternative(c *parser.OC_CaseAlternativeContext) {}

func (s *BaseVisitor) ExitOC_ListComprehension(c *parser.OC_ListComprehensionContext) {}

func (s *BaseVisitor) ExitOC_PatternComprehension(c *parser.OC_PatternComprehensionContext) {}

func (s *BaseVisitor) ExitOC_LegacyListExpression(c *parser.OC_LegacyListExpressionContext) {}

func (s *BaseVisitor) ExitOC_Reduce(c *parser.OC_ReduceContext) {}

func (s *BaseVisitor) ExitOC_Quantifier(c *parser.OC_QuantifierContext) {}

func (s *BaseVisitor) ExitOC_FilterExpression(c *parser.OC_FilterExpressionContext) {}

func (s *BaseVisitor) ExitOC_PatternPredicate(c *parser.OC_PatternPredicateContext) {}

func (s *BaseVisitor) ExitOC_ParenthesizedExpression(c *parser.OC_ParenthesizedExpressionContext) {}

func (s *BaseVisitor) ExitOC_IdInColl(c *parser.OC_IdInCollContext) {}

func (s *BaseVisitor) ExitOC_FunctionInvocation(c *parser.OC_FunctionInvocationContext) {}

func (s *BaseVisitor) ExitOC_FunctionName(c *parser.OC_FunctionNameContext) {}

func (s *BaseVisitor) ExitOC_ExistentialSubquery(c *parser.OC_ExistentialSubqueryContext) {}

func (s *BaseVisitor) ExitOC_ExplicitProcedureInvocation(c *parser.OC_ExplicitProcedureInvocationContext) {
}

func (s *BaseVisitor) ExitOC_ImplicitProcedureInvocation(c *parser.OC_ImplicitProcedureInvocationContext) {
}

func (s *BaseVisitor) ExitOC_ProcedureResultField(c *parser.OC_ProcedureResultFieldContext) {}

func (s *BaseVisitor) ExitOC_ProcedureName(c *parser.OC_ProcedureNameContext) {}

func (s *BaseVisitor) ExitOC_Namespace(c *parser.OC_NamespaceContext) {}

func (s *BaseVisitor) ExitOC_Variable(c *parser.OC_VariableContext) {}

func (s *BaseVisitor) ExitOC_Literal(c *parser.OC_LiteralContext) {}

func (s *BaseVisitor) ExitOC_BooleanLiteral(c *parser.OC_BooleanLiteralContext) {}

func (s *BaseVisitor) ExitOC_NumberLiteral(c *parser.OC_NumberLiteralContext) {}

func (s *BaseVisitor) ExitOC_IntegerLiteral(c *parser.OC_IntegerLiteralContext) {}

func (s *BaseVisitor) ExitOC_DoubleLiteral(c *parser.OC_DoubleLiteralContext) {}

func (s *BaseVisitor) ExitOC_ListLiteral(c *parser.OC_ListLiteralContext) {}

func (s *BaseVisitor) ExitOC_MapLiteral(c *parser.OC_MapLiteralContext) {}

func (s *BaseVisitor) ExitOC_PropertyKeyName(c *parser.OC_PropertyKeyNameContext) {}

func (s *BaseVisitor) ExitOC_LegacyParameter(c *parser.OC_LegacyParameterContext) {}

func (s *BaseVisitor) ExitOC_Parameter(c *parser.OC_ParameterContext) {}

func (s *BaseVisitor) ExitOC_SchemaName(c *parser.OC_SchemaNameContext) {}

func (s *BaseVisitor) ExitOC_ReservedWord(c *parser.OC_ReservedWordContext) {}

func (s *BaseVisitor) ExitOC_SymbolicName(c *parser.OC_SymbolicNameContext) {}

func (s *BaseVisitor) ExitOC_LeftArrowHead(c *parser.OC_LeftArrowHeadContext) {}

func (s *BaseVisitor) ExitOC_RightArrowHead(c *parser.OC_RightArrowHeadContext) {}

func (s *BaseVisitor) ExitOC_Dash(c *parser.OC_DashContext) {}
