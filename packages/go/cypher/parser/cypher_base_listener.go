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

// Code generated from grammar/Cypher.g4 by ANTLR 4.13.0. DO NOT EDIT.

package parser // Cypher

import "github.com/antlr4-go/antlr/v4"

// BaseCypherListener is a complete listener for a parse tree produced by CypherParser.
type BaseCypherListener struct{}

var _ CypherListener = &BaseCypherListener{}

// VisitTerminal is called when a terminal node is visited.
func (s *BaseCypherListener) VisitTerminal(node antlr.TerminalNode) {}

// VisitErrorNode is called when an error node is visited.
func (s *BaseCypherListener) VisitErrorNode(node antlr.ErrorNode) {}

// EnterEveryRule is called when any rule is entered.
func (s *BaseCypherListener) EnterEveryRule(ctx antlr.ParserRuleContext) {}

// ExitEveryRule is called when any rule is exited.
func (s *BaseCypherListener) ExitEveryRule(ctx antlr.ParserRuleContext) {}

// EnterOC_Cypher is called when production oC_Cypher is entered.
func (s *BaseCypherListener) EnterOC_Cypher(ctx *OC_CypherContext) {}

// ExitOC_Cypher is called when production oC_Cypher is exited.
func (s *BaseCypherListener) ExitOC_Cypher(ctx *OC_CypherContext) {}

// EnterOC_QueryOptions is called when production oC_QueryOptions is entered.
func (s *BaseCypherListener) EnterOC_QueryOptions(ctx *OC_QueryOptionsContext) {}

// ExitOC_QueryOptions is called when production oC_QueryOptions is exited.
func (s *BaseCypherListener) ExitOC_QueryOptions(ctx *OC_QueryOptionsContext) {}

// EnterOC_AnyCypherOption is called when production oC_AnyCypherOption is entered.
func (s *BaseCypherListener) EnterOC_AnyCypherOption(ctx *OC_AnyCypherOptionContext) {}

// ExitOC_AnyCypherOption is called when production oC_AnyCypherOption is exited.
func (s *BaseCypherListener) ExitOC_AnyCypherOption(ctx *OC_AnyCypherOptionContext) {}

// EnterOC_CypherOption is called when production oC_CypherOption is entered.
func (s *BaseCypherListener) EnterOC_CypherOption(ctx *OC_CypherOptionContext) {}

// ExitOC_CypherOption is called when production oC_CypherOption is exited.
func (s *BaseCypherListener) ExitOC_CypherOption(ctx *OC_CypherOptionContext) {}

// EnterOC_VersionNumber is called when production oC_VersionNumber is entered.
func (s *BaseCypherListener) EnterOC_VersionNumber(ctx *OC_VersionNumberContext) {}

// ExitOC_VersionNumber is called when production oC_VersionNumber is exited.
func (s *BaseCypherListener) ExitOC_VersionNumber(ctx *OC_VersionNumberContext) {}

// EnterOC_Explain is called when production oC_Explain is entered.
func (s *BaseCypherListener) EnterOC_Explain(ctx *OC_ExplainContext) {}

// ExitOC_Explain is called when production oC_Explain is exited.
func (s *BaseCypherListener) ExitOC_Explain(ctx *OC_ExplainContext) {}

// EnterOC_Profile is called when production oC_Profile is entered.
func (s *BaseCypherListener) EnterOC_Profile(ctx *OC_ProfileContext) {}

// ExitOC_Profile is called when production oC_Profile is exited.
func (s *BaseCypherListener) ExitOC_Profile(ctx *OC_ProfileContext) {}

// EnterOC_ConfigurationOption is called when production oC_ConfigurationOption is entered.
func (s *BaseCypherListener) EnterOC_ConfigurationOption(ctx *OC_ConfigurationOptionContext) {}

// ExitOC_ConfigurationOption is called when production oC_ConfigurationOption is exited.
func (s *BaseCypherListener) ExitOC_ConfigurationOption(ctx *OC_ConfigurationOptionContext) {}

// EnterOC_Statement is called when production oC_Statement is entered.
func (s *BaseCypherListener) EnterOC_Statement(ctx *OC_StatementContext) {}

// ExitOC_Statement is called when production oC_Statement is exited.
func (s *BaseCypherListener) ExitOC_Statement(ctx *OC_StatementContext) {}

// EnterOC_Query is called when production oC_Query is entered.
func (s *BaseCypherListener) EnterOC_Query(ctx *OC_QueryContext) {}

// ExitOC_Query is called when production oC_Query is exited.
func (s *BaseCypherListener) ExitOC_Query(ctx *OC_QueryContext) {}

// EnterOC_RegularQuery is called when production oC_RegularQuery is entered.
func (s *BaseCypherListener) EnterOC_RegularQuery(ctx *OC_RegularQueryContext) {}

// ExitOC_RegularQuery is called when production oC_RegularQuery is exited.
func (s *BaseCypherListener) ExitOC_RegularQuery(ctx *OC_RegularQueryContext) {}

// EnterOC_BulkImportQuery is called when production oC_BulkImportQuery is entered.
func (s *BaseCypherListener) EnterOC_BulkImportQuery(ctx *OC_BulkImportQueryContext) {}

// ExitOC_BulkImportQuery is called when production oC_BulkImportQuery is exited.
func (s *BaseCypherListener) ExitOC_BulkImportQuery(ctx *OC_BulkImportQueryContext) {}

// EnterOC_PeriodicCommitHint is called when production oC_PeriodicCommitHint is entered.
func (s *BaseCypherListener) EnterOC_PeriodicCommitHint(ctx *OC_PeriodicCommitHintContext) {}

// ExitOC_PeriodicCommitHint is called when production oC_PeriodicCommitHint is exited.
func (s *BaseCypherListener) ExitOC_PeriodicCommitHint(ctx *OC_PeriodicCommitHintContext) {}

// EnterOC_LoadCSVQuery is called when production oC_LoadCSVQuery is entered.
func (s *BaseCypherListener) EnterOC_LoadCSVQuery(ctx *OC_LoadCSVQueryContext) {}

// ExitOC_LoadCSVQuery is called when production oC_LoadCSVQuery is exited.
func (s *BaseCypherListener) ExitOC_LoadCSVQuery(ctx *OC_LoadCSVQueryContext) {}

// EnterOC_Union is called when production oC_Union is entered.
func (s *BaseCypherListener) EnterOC_Union(ctx *OC_UnionContext) {}

// ExitOC_Union is called when production oC_Union is exited.
func (s *BaseCypherListener) ExitOC_Union(ctx *OC_UnionContext) {}

// EnterOC_SingleQuery is called when production oC_SingleQuery is entered.
func (s *BaseCypherListener) EnterOC_SingleQuery(ctx *OC_SingleQueryContext) {}

// ExitOC_SingleQuery is called when production oC_SingleQuery is exited.
func (s *BaseCypherListener) ExitOC_SingleQuery(ctx *OC_SingleQueryContext) {}

// EnterOC_SinglePartQuery is called when production oC_SinglePartQuery is entered.
func (s *BaseCypherListener) EnterOC_SinglePartQuery(ctx *OC_SinglePartQueryContext) {}

// ExitOC_SinglePartQuery is called when production oC_SinglePartQuery is exited.
func (s *BaseCypherListener) ExitOC_SinglePartQuery(ctx *OC_SinglePartQueryContext) {}

// EnterOC_MultiPartQuery is called when production oC_MultiPartQuery is entered.
func (s *BaseCypherListener) EnterOC_MultiPartQuery(ctx *OC_MultiPartQueryContext) {}

// ExitOC_MultiPartQuery is called when production oC_MultiPartQuery is exited.
func (s *BaseCypherListener) ExitOC_MultiPartQuery(ctx *OC_MultiPartQueryContext) {}

// EnterOC_UpdatingClause is called when production oC_UpdatingClause is entered.
func (s *BaseCypherListener) EnterOC_UpdatingClause(ctx *OC_UpdatingClauseContext) {}

// ExitOC_UpdatingClause is called when production oC_UpdatingClause is exited.
func (s *BaseCypherListener) ExitOC_UpdatingClause(ctx *OC_UpdatingClauseContext) {}

// EnterOC_ReadingClause is called when production oC_ReadingClause is entered.
func (s *BaseCypherListener) EnterOC_ReadingClause(ctx *OC_ReadingClauseContext) {}

// ExitOC_ReadingClause is called when production oC_ReadingClause is exited.
func (s *BaseCypherListener) ExitOC_ReadingClause(ctx *OC_ReadingClauseContext) {}

// EnterOC_Command is called when production oC_Command is entered.
func (s *BaseCypherListener) EnterOC_Command(ctx *OC_CommandContext) {}

// ExitOC_Command is called when production oC_Command is exited.
func (s *BaseCypherListener) ExitOC_Command(ctx *OC_CommandContext) {}

// EnterOC_CreateUniqueConstraint is called when production oC_CreateUniqueConstraint is entered.
func (s *BaseCypherListener) EnterOC_CreateUniqueConstraint(ctx *OC_CreateUniqueConstraintContext) {}

// ExitOC_CreateUniqueConstraint is called when production oC_CreateUniqueConstraint is exited.
func (s *BaseCypherListener) ExitOC_CreateUniqueConstraint(ctx *OC_CreateUniqueConstraintContext) {}

// EnterOC_CreateNodePropertyExistenceConstraint is called when production oC_CreateNodePropertyExistenceConstraint is entered.
func (s *BaseCypherListener) EnterOC_CreateNodePropertyExistenceConstraint(ctx *OC_CreateNodePropertyExistenceConstraintContext) {
}

// ExitOC_CreateNodePropertyExistenceConstraint is called when production oC_CreateNodePropertyExistenceConstraint is exited.
func (s *BaseCypherListener) ExitOC_CreateNodePropertyExistenceConstraint(ctx *OC_CreateNodePropertyExistenceConstraintContext) {
}

// EnterOC_CreateRelationshipPropertyExistenceConstraint is called when production oC_CreateRelationshipPropertyExistenceConstraint is entered.
func (s *BaseCypherListener) EnterOC_CreateRelationshipPropertyExistenceConstraint(ctx *OC_CreateRelationshipPropertyExistenceConstraintContext) {
}

// ExitOC_CreateRelationshipPropertyExistenceConstraint is called when production oC_CreateRelationshipPropertyExistenceConstraint is exited.
func (s *BaseCypherListener) ExitOC_CreateRelationshipPropertyExistenceConstraint(ctx *OC_CreateRelationshipPropertyExistenceConstraintContext) {
}

// EnterOC_CreateIndex is called when production oC_CreateIndex is entered.
func (s *BaseCypherListener) EnterOC_CreateIndex(ctx *OC_CreateIndexContext) {}

// ExitOC_CreateIndex is called when production oC_CreateIndex is exited.
func (s *BaseCypherListener) ExitOC_CreateIndex(ctx *OC_CreateIndexContext) {}

// EnterOC_DropUniqueConstraint is called when production oC_DropUniqueConstraint is entered.
func (s *BaseCypherListener) EnterOC_DropUniqueConstraint(ctx *OC_DropUniqueConstraintContext) {}

// ExitOC_DropUniqueConstraint is called when production oC_DropUniqueConstraint is exited.
func (s *BaseCypherListener) ExitOC_DropUniqueConstraint(ctx *OC_DropUniqueConstraintContext) {}

// EnterOC_DropNodePropertyExistenceConstraint is called when production oC_DropNodePropertyExistenceConstraint is entered.
func (s *BaseCypherListener) EnterOC_DropNodePropertyExistenceConstraint(ctx *OC_DropNodePropertyExistenceConstraintContext) {
}

// ExitOC_DropNodePropertyExistenceConstraint is called when production oC_DropNodePropertyExistenceConstraint is exited.
func (s *BaseCypherListener) ExitOC_DropNodePropertyExistenceConstraint(ctx *OC_DropNodePropertyExistenceConstraintContext) {
}

// EnterOC_DropRelationshipPropertyExistenceConstraint is called when production oC_DropRelationshipPropertyExistenceConstraint is entered.
func (s *BaseCypherListener) EnterOC_DropRelationshipPropertyExistenceConstraint(ctx *OC_DropRelationshipPropertyExistenceConstraintContext) {
}

// ExitOC_DropRelationshipPropertyExistenceConstraint is called when production oC_DropRelationshipPropertyExistenceConstraint is exited.
func (s *BaseCypherListener) ExitOC_DropRelationshipPropertyExistenceConstraint(ctx *OC_DropRelationshipPropertyExistenceConstraintContext) {
}

// EnterOC_DropIndex is called when production oC_DropIndex is entered.
func (s *BaseCypherListener) EnterOC_DropIndex(ctx *OC_DropIndexContext) {}

// ExitOC_DropIndex is called when production oC_DropIndex is exited.
func (s *BaseCypherListener) ExitOC_DropIndex(ctx *OC_DropIndexContext) {}

// EnterOC_Index is called when production oC_Index is entered.
func (s *BaseCypherListener) EnterOC_Index(ctx *OC_IndexContext) {}

// ExitOC_Index is called when production oC_Index is exited.
func (s *BaseCypherListener) ExitOC_Index(ctx *OC_IndexContext) {}

// EnterOC_UniqueConstraint is called when production oC_UniqueConstraint is entered.
func (s *BaseCypherListener) EnterOC_UniqueConstraint(ctx *OC_UniqueConstraintContext) {}

// ExitOC_UniqueConstraint is called when production oC_UniqueConstraint is exited.
func (s *BaseCypherListener) ExitOC_UniqueConstraint(ctx *OC_UniqueConstraintContext) {}

// EnterOC_NodePropertyExistenceConstraint is called when production oC_NodePropertyExistenceConstraint is entered.
func (s *BaseCypherListener) EnterOC_NodePropertyExistenceConstraint(ctx *OC_NodePropertyExistenceConstraintContext) {
}

// ExitOC_NodePropertyExistenceConstraint is called when production oC_NodePropertyExistenceConstraint is exited.
func (s *BaseCypherListener) ExitOC_NodePropertyExistenceConstraint(ctx *OC_NodePropertyExistenceConstraintContext) {
}

// EnterOC_RelationshipPropertyExistenceConstraint is called when production oC_RelationshipPropertyExistenceConstraint is entered.
func (s *BaseCypherListener) EnterOC_RelationshipPropertyExistenceConstraint(ctx *OC_RelationshipPropertyExistenceConstraintContext) {
}

// ExitOC_RelationshipPropertyExistenceConstraint is called when production oC_RelationshipPropertyExistenceConstraint is exited.
func (s *BaseCypherListener) ExitOC_RelationshipPropertyExistenceConstraint(ctx *OC_RelationshipPropertyExistenceConstraintContext) {
}

// EnterOC_RelationshipPatternSyntax is called when production oC_RelationshipPatternSyntax is entered.
func (s *BaseCypherListener) EnterOC_RelationshipPatternSyntax(ctx *OC_RelationshipPatternSyntaxContext) {
}

// ExitOC_RelationshipPatternSyntax is called when production oC_RelationshipPatternSyntax is exited.
func (s *BaseCypherListener) ExitOC_RelationshipPatternSyntax(ctx *OC_RelationshipPatternSyntaxContext) {
}

// EnterOC_LoadCSV is called when production oC_LoadCSV is entered.
func (s *BaseCypherListener) EnterOC_LoadCSV(ctx *OC_LoadCSVContext) {}

// ExitOC_LoadCSV is called when production oC_LoadCSV is exited.
func (s *BaseCypherListener) ExitOC_LoadCSV(ctx *OC_LoadCSVContext) {}

// EnterOC_Match is called when production oC_Match is entered.
func (s *BaseCypherListener) EnterOC_Match(ctx *OC_MatchContext) {}

// ExitOC_Match is called when production oC_Match is exited.
func (s *BaseCypherListener) ExitOC_Match(ctx *OC_MatchContext) {}

// EnterOC_Unwind is called when production oC_Unwind is entered.
func (s *BaseCypherListener) EnterOC_Unwind(ctx *OC_UnwindContext) {}

// ExitOC_Unwind is called when production oC_Unwind is exited.
func (s *BaseCypherListener) ExitOC_Unwind(ctx *OC_UnwindContext) {}

// EnterOC_Merge is called when production oC_Merge is entered.
func (s *BaseCypherListener) EnterOC_Merge(ctx *OC_MergeContext) {}

// ExitOC_Merge is called when production oC_Merge is exited.
func (s *BaseCypherListener) ExitOC_Merge(ctx *OC_MergeContext) {}

// EnterOC_MergeAction is called when production oC_MergeAction is entered.
func (s *BaseCypherListener) EnterOC_MergeAction(ctx *OC_MergeActionContext) {}

// ExitOC_MergeAction is called when production oC_MergeAction is exited.
func (s *BaseCypherListener) ExitOC_MergeAction(ctx *OC_MergeActionContext) {}

// EnterOC_Create is called when production oC_Create is entered.
func (s *BaseCypherListener) EnterOC_Create(ctx *OC_CreateContext) {}

// ExitOC_Create is called when production oC_Create is exited.
func (s *BaseCypherListener) ExitOC_Create(ctx *OC_CreateContext) {}

// EnterOC_CreateUnique is called when production oC_CreateUnique is entered.
func (s *BaseCypherListener) EnterOC_CreateUnique(ctx *OC_CreateUniqueContext) {}

// ExitOC_CreateUnique is called when production oC_CreateUnique is exited.
func (s *BaseCypherListener) ExitOC_CreateUnique(ctx *OC_CreateUniqueContext) {}

// EnterOC_Set is called when production oC_Set is entered.
func (s *BaseCypherListener) EnterOC_Set(ctx *OC_SetContext) {}

// ExitOC_Set is called when production oC_Set is exited.
func (s *BaseCypherListener) ExitOC_Set(ctx *OC_SetContext) {}

// EnterOC_SetItem is called when production oC_SetItem is entered.
func (s *BaseCypherListener) EnterOC_SetItem(ctx *OC_SetItemContext) {}

// ExitOC_SetItem is called when production oC_SetItem is exited.
func (s *BaseCypherListener) ExitOC_SetItem(ctx *OC_SetItemContext) {}

// EnterOC_Delete is called when production oC_Delete is entered.
func (s *BaseCypherListener) EnterOC_Delete(ctx *OC_DeleteContext) {}

// ExitOC_Delete is called when production oC_Delete is exited.
func (s *BaseCypherListener) ExitOC_Delete(ctx *OC_DeleteContext) {}

// EnterOC_Remove is called when production oC_Remove is entered.
func (s *BaseCypherListener) EnterOC_Remove(ctx *OC_RemoveContext) {}

// ExitOC_Remove is called when production oC_Remove is exited.
func (s *BaseCypherListener) ExitOC_Remove(ctx *OC_RemoveContext) {}

// EnterOC_RemoveItem is called when production oC_RemoveItem is entered.
func (s *BaseCypherListener) EnterOC_RemoveItem(ctx *OC_RemoveItemContext) {}

// ExitOC_RemoveItem is called when production oC_RemoveItem is exited.
func (s *BaseCypherListener) ExitOC_RemoveItem(ctx *OC_RemoveItemContext) {}

// EnterOC_Foreach is called when production oC_Foreach is entered.
func (s *BaseCypherListener) EnterOC_Foreach(ctx *OC_ForeachContext) {}

// ExitOC_Foreach is called when production oC_Foreach is exited.
func (s *BaseCypherListener) ExitOC_Foreach(ctx *OC_ForeachContext) {}

// EnterOC_InQueryCall is called when production oC_InQueryCall is entered.
func (s *BaseCypherListener) EnterOC_InQueryCall(ctx *OC_InQueryCallContext) {}

// ExitOC_InQueryCall is called when production oC_InQueryCall is exited.
func (s *BaseCypherListener) ExitOC_InQueryCall(ctx *OC_InQueryCallContext) {}

// EnterOC_StandaloneCall is called when production oC_StandaloneCall is entered.
func (s *BaseCypherListener) EnterOC_StandaloneCall(ctx *OC_StandaloneCallContext) {}

// ExitOC_StandaloneCall is called when production oC_StandaloneCall is exited.
func (s *BaseCypherListener) ExitOC_StandaloneCall(ctx *OC_StandaloneCallContext) {}

// EnterOC_YieldItems is called when production oC_YieldItems is entered.
func (s *BaseCypherListener) EnterOC_YieldItems(ctx *OC_YieldItemsContext) {}

// ExitOC_YieldItems is called when production oC_YieldItems is exited.
func (s *BaseCypherListener) ExitOC_YieldItems(ctx *OC_YieldItemsContext) {}

// EnterOC_YieldItem is called when production oC_YieldItem is entered.
func (s *BaseCypherListener) EnterOC_YieldItem(ctx *OC_YieldItemContext) {}

// ExitOC_YieldItem is called when production oC_YieldItem is exited.
func (s *BaseCypherListener) ExitOC_YieldItem(ctx *OC_YieldItemContext) {}

// EnterOC_With is called when production oC_With is entered.
func (s *BaseCypherListener) EnterOC_With(ctx *OC_WithContext) {}

// ExitOC_With is called when production oC_With is exited.
func (s *BaseCypherListener) ExitOC_With(ctx *OC_WithContext) {}

// EnterOC_Return is called when production oC_Return is entered.
func (s *BaseCypherListener) EnterOC_Return(ctx *OC_ReturnContext) {}

// ExitOC_Return is called when production oC_Return is exited.
func (s *BaseCypherListener) ExitOC_Return(ctx *OC_ReturnContext) {}

// EnterOC_ProjectionBody is called when production oC_ProjectionBody is entered.
func (s *BaseCypherListener) EnterOC_ProjectionBody(ctx *OC_ProjectionBodyContext) {}

// ExitOC_ProjectionBody is called when production oC_ProjectionBody is exited.
func (s *BaseCypherListener) ExitOC_ProjectionBody(ctx *OC_ProjectionBodyContext) {}

// EnterOC_ProjectionItems is called when production oC_ProjectionItems is entered.
func (s *BaseCypherListener) EnterOC_ProjectionItems(ctx *OC_ProjectionItemsContext) {}

// ExitOC_ProjectionItems is called when production oC_ProjectionItems is exited.
func (s *BaseCypherListener) ExitOC_ProjectionItems(ctx *OC_ProjectionItemsContext) {}

// EnterOC_ProjectionItem is called when production oC_ProjectionItem is entered.
func (s *BaseCypherListener) EnterOC_ProjectionItem(ctx *OC_ProjectionItemContext) {}

// ExitOC_ProjectionItem is called when production oC_ProjectionItem is exited.
func (s *BaseCypherListener) ExitOC_ProjectionItem(ctx *OC_ProjectionItemContext) {}

// EnterOC_Order is called when production oC_Order is entered.
func (s *BaseCypherListener) EnterOC_Order(ctx *OC_OrderContext) {}

// ExitOC_Order is called when production oC_Order is exited.
func (s *BaseCypherListener) ExitOC_Order(ctx *OC_OrderContext) {}

// EnterOC_Skip is called when production oC_Skip is entered.
func (s *BaseCypherListener) EnterOC_Skip(ctx *OC_SkipContext) {}

// ExitOC_Skip is called when production oC_Skip is exited.
func (s *BaseCypherListener) ExitOC_Skip(ctx *OC_SkipContext) {}

// EnterOC_Limit is called when production oC_Limit is entered.
func (s *BaseCypherListener) EnterOC_Limit(ctx *OC_LimitContext) {}

// ExitOC_Limit is called when production oC_Limit is exited.
func (s *BaseCypherListener) ExitOC_Limit(ctx *OC_LimitContext) {}

// EnterOC_SortItem is called when production oC_SortItem is entered.
func (s *BaseCypherListener) EnterOC_SortItem(ctx *OC_SortItemContext) {}

// ExitOC_SortItem is called when production oC_SortItem is exited.
func (s *BaseCypherListener) ExitOC_SortItem(ctx *OC_SortItemContext) {}

// EnterOC_Hint is called when production oC_Hint is entered.
func (s *BaseCypherListener) EnterOC_Hint(ctx *OC_HintContext) {}

// ExitOC_Hint is called when production oC_Hint is exited.
func (s *BaseCypherListener) ExitOC_Hint(ctx *OC_HintContext) {}

// EnterOC_Start is called when production oC_Start is entered.
func (s *BaseCypherListener) EnterOC_Start(ctx *OC_StartContext) {}

// ExitOC_Start is called when production oC_Start is exited.
func (s *BaseCypherListener) ExitOC_Start(ctx *OC_StartContext) {}

// EnterOC_StartPoint is called when production oC_StartPoint is entered.
func (s *BaseCypherListener) EnterOC_StartPoint(ctx *OC_StartPointContext) {}

// ExitOC_StartPoint is called when production oC_StartPoint is exited.
func (s *BaseCypherListener) ExitOC_StartPoint(ctx *OC_StartPointContext) {}

// EnterOC_Lookup is called when production oC_Lookup is entered.
func (s *BaseCypherListener) EnterOC_Lookup(ctx *OC_LookupContext) {}

// ExitOC_Lookup is called when production oC_Lookup is exited.
func (s *BaseCypherListener) ExitOC_Lookup(ctx *OC_LookupContext) {}

// EnterOC_NodeLookup is called when production oC_NodeLookup is entered.
func (s *BaseCypherListener) EnterOC_NodeLookup(ctx *OC_NodeLookupContext) {}

// ExitOC_NodeLookup is called when production oC_NodeLookup is exited.
func (s *BaseCypherListener) ExitOC_NodeLookup(ctx *OC_NodeLookupContext) {}

// EnterOC_RelationshipLookup is called when production oC_RelationshipLookup is entered.
func (s *BaseCypherListener) EnterOC_RelationshipLookup(ctx *OC_RelationshipLookupContext) {}

// ExitOC_RelationshipLookup is called when production oC_RelationshipLookup is exited.
func (s *BaseCypherListener) ExitOC_RelationshipLookup(ctx *OC_RelationshipLookupContext) {}

// EnterOC_IdentifiedIndexLookup is called when production oC_IdentifiedIndexLookup is entered.
func (s *BaseCypherListener) EnterOC_IdentifiedIndexLookup(ctx *OC_IdentifiedIndexLookupContext) {}

// ExitOC_IdentifiedIndexLookup is called when production oC_IdentifiedIndexLookup is exited.
func (s *BaseCypherListener) ExitOC_IdentifiedIndexLookup(ctx *OC_IdentifiedIndexLookupContext) {}

// EnterOC_IndexQuery is called when production oC_IndexQuery is entered.
func (s *BaseCypherListener) EnterOC_IndexQuery(ctx *OC_IndexQueryContext) {}

// ExitOC_IndexQuery is called when production oC_IndexQuery is exited.
func (s *BaseCypherListener) ExitOC_IndexQuery(ctx *OC_IndexQueryContext) {}

// EnterOC_IdLookup is called when production oC_IdLookup is entered.
func (s *BaseCypherListener) EnterOC_IdLookup(ctx *OC_IdLookupContext) {}

// ExitOC_IdLookup is called when production oC_IdLookup is exited.
func (s *BaseCypherListener) ExitOC_IdLookup(ctx *OC_IdLookupContext) {}

// EnterOC_LiteralIds is called when production oC_LiteralIds is entered.
func (s *BaseCypherListener) EnterOC_LiteralIds(ctx *OC_LiteralIdsContext) {}

// ExitOC_LiteralIds is called when production oC_LiteralIds is exited.
func (s *BaseCypherListener) ExitOC_LiteralIds(ctx *OC_LiteralIdsContext) {}

// EnterOC_Where is called when production oC_Where is entered.
func (s *BaseCypherListener) EnterOC_Where(ctx *OC_WhereContext) {}

// ExitOC_Where is called when production oC_Where is exited.
func (s *BaseCypherListener) ExitOC_Where(ctx *OC_WhereContext) {}

// EnterOC_Pattern is called when production oC_Pattern is entered.
func (s *BaseCypherListener) EnterOC_Pattern(ctx *OC_PatternContext) {}

// ExitOC_Pattern is called when production oC_Pattern is exited.
func (s *BaseCypherListener) ExitOC_Pattern(ctx *OC_PatternContext) {}

// EnterOC_PatternPart is called when production oC_PatternPart is entered.
func (s *BaseCypherListener) EnterOC_PatternPart(ctx *OC_PatternPartContext) {}

// ExitOC_PatternPart is called when production oC_PatternPart is exited.
func (s *BaseCypherListener) ExitOC_PatternPart(ctx *OC_PatternPartContext) {}

// EnterOC_AnonymousPatternPart is called when production oC_AnonymousPatternPart is entered.
func (s *BaseCypherListener) EnterOC_AnonymousPatternPart(ctx *OC_AnonymousPatternPartContext) {}

// ExitOC_AnonymousPatternPart is called when production oC_AnonymousPatternPart is exited.
func (s *BaseCypherListener) ExitOC_AnonymousPatternPart(ctx *OC_AnonymousPatternPartContext) {}

// EnterOC_ShortestPathPattern is called when production oC_ShortestPathPattern is entered.
func (s *BaseCypherListener) EnterOC_ShortestPathPattern(ctx *OC_ShortestPathPatternContext) {}

// ExitOC_ShortestPathPattern is called when production oC_ShortestPathPattern is exited.
func (s *BaseCypherListener) ExitOC_ShortestPathPattern(ctx *OC_ShortestPathPatternContext) {}

// EnterOC_PatternElement is called when production oC_PatternElement is entered.
func (s *BaseCypherListener) EnterOC_PatternElement(ctx *OC_PatternElementContext) {}

// ExitOC_PatternElement is called when production oC_PatternElement is exited.
func (s *BaseCypherListener) ExitOC_PatternElement(ctx *OC_PatternElementContext) {}

// EnterOC_RelationshipsPattern is called when production oC_RelationshipsPattern is entered.
func (s *BaseCypherListener) EnterOC_RelationshipsPattern(ctx *OC_RelationshipsPatternContext) {}

// ExitOC_RelationshipsPattern is called when production oC_RelationshipsPattern is exited.
func (s *BaseCypherListener) ExitOC_RelationshipsPattern(ctx *OC_RelationshipsPatternContext) {}

// EnterOC_NodePattern is called when production oC_NodePattern is entered.
func (s *BaseCypherListener) EnterOC_NodePattern(ctx *OC_NodePatternContext) {}

// ExitOC_NodePattern is called when production oC_NodePattern is exited.
func (s *BaseCypherListener) ExitOC_NodePattern(ctx *OC_NodePatternContext) {}

// EnterOC_PatternElementChain is called when production oC_PatternElementChain is entered.
func (s *BaseCypherListener) EnterOC_PatternElementChain(ctx *OC_PatternElementChainContext) {}

// ExitOC_PatternElementChain is called when production oC_PatternElementChain is exited.
func (s *BaseCypherListener) ExitOC_PatternElementChain(ctx *OC_PatternElementChainContext) {}

// EnterOC_RelationshipPattern is called when production oC_RelationshipPattern is entered.
func (s *BaseCypherListener) EnterOC_RelationshipPattern(ctx *OC_RelationshipPatternContext) {}

// ExitOC_RelationshipPattern is called when production oC_RelationshipPattern is exited.
func (s *BaseCypherListener) ExitOC_RelationshipPattern(ctx *OC_RelationshipPatternContext) {}

// EnterOC_RelationshipDetail is called when production oC_RelationshipDetail is entered.
func (s *BaseCypherListener) EnterOC_RelationshipDetail(ctx *OC_RelationshipDetailContext) {}

// ExitOC_RelationshipDetail is called when production oC_RelationshipDetail is exited.
func (s *BaseCypherListener) ExitOC_RelationshipDetail(ctx *OC_RelationshipDetailContext) {}

// EnterOC_Properties is called when production oC_Properties is entered.
func (s *BaseCypherListener) EnterOC_Properties(ctx *OC_PropertiesContext) {}

// ExitOC_Properties is called when production oC_Properties is exited.
func (s *BaseCypherListener) ExitOC_Properties(ctx *OC_PropertiesContext) {}

// EnterOC_RelType is called when production oC_RelType is entered.
func (s *BaseCypherListener) EnterOC_RelType(ctx *OC_RelTypeContext) {}

// ExitOC_RelType is called when production oC_RelType is exited.
func (s *BaseCypherListener) ExitOC_RelType(ctx *OC_RelTypeContext) {}

// EnterOC_RelationshipTypes is called when production oC_RelationshipTypes is entered.
func (s *BaseCypherListener) EnterOC_RelationshipTypes(ctx *OC_RelationshipTypesContext) {}

// ExitOC_RelationshipTypes is called when production oC_RelationshipTypes is exited.
func (s *BaseCypherListener) ExitOC_RelationshipTypes(ctx *OC_RelationshipTypesContext) {}

// EnterOC_NodeLabels is called when production oC_NodeLabels is entered.
func (s *BaseCypherListener) EnterOC_NodeLabels(ctx *OC_NodeLabelsContext) {}

// ExitOC_NodeLabels is called when production oC_NodeLabels is exited.
func (s *BaseCypherListener) ExitOC_NodeLabels(ctx *OC_NodeLabelsContext) {}

// EnterOC_NodeLabel is called when production oC_NodeLabel is entered.
func (s *BaseCypherListener) EnterOC_NodeLabel(ctx *OC_NodeLabelContext) {}

// ExitOC_NodeLabel is called when production oC_NodeLabel is exited.
func (s *BaseCypherListener) ExitOC_NodeLabel(ctx *OC_NodeLabelContext) {}

// EnterOC_RangeLiteral is called when production oC_RangeLiteral is entered.
func (s *BaseCypherListener) EnterOC_RangeLiteral(ctx *OC_RangeLiteralContext) {}

// ExitOC_RangeLiteral is called when production oC_RangeLiteral is exited.
func (s *BaseCypherListener) ExitOC_RangeLiteral(ctx *OC_RangeLiteralContext) {}

// EnterOC_LabelName is called when production oC_LabelName is entered.
func (s *BaseCypherListener) EnterOC_LabelName(ctx *OC_LabelNameContext) {}

// ExitOC_LabelName is called when production oC_LabelName is exited.
func (s *BaseCypherListener) ExitOC_LabelName(ctx *OC_LabelNameContext) {}

// EnterOC_RelTypeName is called when production oC_RelTypeName is entered.
func (s *BaseCypherListener) EnterOC_RelTypeName(ctx *OC_RelTypeNameContext) {}

// ExitOC_RelTypeName is called when production oC_RelTypeName is exited.
func (s *BaseCypherListener) ExitOC_RelTypeName(ctx *OC_RelTypeNameContext) {}

// EnterOC_PropertyExpression is called when production oC_PropertyExpression is entered.
func (s *BaseCypherListener) EnterOC_PropertyExpression(ctx *OC_PropertyExpressionContext) {}

// ExitOC_PropertyExpression is called when production oC_PropertyExpression is exited.
func (s *BaseCypherListener) ExitOC_PropertyExpression(ctx *OC_PropertyExpressionContext) {}

// EnterOC_Expression is called when production oC_Expression is entered.
func (s *BaseCypherListener) EnterOC_Expression(ctx *OC_ExpressionContext) {}

// ExitOC_Expression is called when production oC_Expression is exited.
func (s *BaseCypherListener) ExitOC_Expression(ctx *OC_ExpressionContext) {}

// EnterOC_OrExpression is called when production oC_OrExpression is entered.
func (s *BaseCypherListener) EnterOC_OrExpression(ctx *OC_OrExpressionContext) {}

// ExitOC_OrExpression is called when production oC_OrExpression is exited.
func (s *BaseCypherListener) ExitOC_OrExpression(ctx *OC_OrExpressionContext) {}

// EnterOC_XorExpression is called when production oC_XorExpression is entered.
func (s *BaseCypherListener) EnterOC_XorExpression(ctx *OC_XorExpressionContext) {}

// ExitOC_XorExpression is called when production oC_XorExpression is exited.
func (s *BaseCypherListener) ExitOC_XorExpression(ctx *OC_XorExpressionContext) {}

// EnterOC_AndExpression is called when production oC_AndExpression is entered.
func (s *BaseCypherListener) EnterOC_AndExpression(ctx *OC_AndExpressionContext) {}

// ExitOC_AndExpression is called when production oC_AndExpression is exited.
func (s *BaseCypherListener) ExitOC_AndExpression(ctx *OC_AndExpressionContext) {}

// EnterOC_NotExpression is called when production oC_NotExpression is entered.
func (s *BaseCypherListener) EnterOC_NotExpression(ctx *OC_NotExpressionContext) {}

// ExitOC_NotExpression is called when production oC_NotExpression is exited.
func (s *BaseCypherListener) ExitOC_NotExpression(ctx *OC_NotExpressionContext) {}

// EnterOC_ComparisonExpression is called when production oC_ComparisonExpression is entered.
func (s *BaseCypherListener) EnterOC_ComparisonExpression(ctx *OC_ComparisonExpressionContext) {}

// ExitOC_ComparisonExpression is called when production oC_ComparisonExpression is exited.
func (s *BaseCypherListener) ExitOC_ComparisonExpression(ctx *OC_ComparisonExpressionContext) {}

// EnterOC_PartialComparisonExpression is called when production oC_PartialComparisonExpression is entered.
func (s *BaseCypherListener) EnterOC_PartialComparisonExpression(ctx *OC_PartialComparisonExpressionContext) {
}

// ExitOC_PartialComparisonExpression is called when production oC_PartialComparisonExpression is exited.
func (s *BaseCypherListener) ExitOC_PartialComparisonExpression(ctx *OC_PartialComparisonExpressionContext) {
}

// EnterOC_StringListNullPredicateExpression is called when production oC_StringListNullPredicateExpression is entered.
func (s *BaseCypherListener) EnterOC_StringListNullPredicateExpression(ctx *OC_StringListNullPredicateExpressionContext) {
}

// ExitOC_StringListNullPredicateExpression is called when production oC_StringListNullPredicateExpression is exited.
func (s *BaseCypherListener) ExitOC_StringListNullPredicateExpression(ctx *OC_StringListNullPredicateExpressionContext) {
}

// EnterOC_StringPredicateExpression is called when production oC_StringPredicateExpression is entered.
func (s *BaseCypherListener) EnterOC_StringPredicateExpression(ctx *OC_StringPredicateExpressionContext) {
}

// ExitOC_StringPredicateExpression is called when production oC_StringPredicateExpression is exited.
func (s *BaseCypherListener) ExitOC_StringPredicateExpression(ctx *OC_StringPredicateExpressionContext) {
}

// EnterOC_ListPredicateExpression is called when production oC_ListPredicateExpression is entered.
func (s *BaseCypherListener) EnterOC_ListPredicateExpression(ctx *OC_ListPredicateExpressionContext) {
}

// ExitOC_ListPredicateExpression is called when production oC_ListPredicateExpression is exited.
func (s *BaseCypherListener) ExitOC_ListPredicateExpression(ctx *OC_ListPredicateExpressionContext) {}

// EnterOC_NullPredicateExpression is called when production oC_NullPredicateExpression is entered.
func (s *BaseCypherListener) EnterOC_NullPredicateExpression(ctx *OC_NullPredicateExpressionContext) {
}

// ExitOC_NullPredicateExpression is called when production oC_NullPredicateExpression is exited.
func (s *BaseCypherListener) ExitOC_NullPredicateExpression(ctx *OC_NullPredicateExpressionContext) {}

// EnterOC_RegularExpression is called when production oC_RegularExpression is entered.
func (s *BaseCypherListener) EnterOC_RegularExpression(ctx *OC_RegularExpressionContext) {}

// ExitOC_RegularExpression is called when production oC_RegularExpression is exited.
func (s *BaseCypherListener) ExitOC_RegularExpression(ctx *OC_RegularExpressionContext) {}

// EnterOC_AddOrSubtractExpression is called when production oC_AddOrSubtractExpression is entered.
func (s *BaseCypherListener) EnterOC_AddOrSubtractExpression(ctx *OC_AddOrSubtractExpressionContext) {
}

// ExitOC_AddOrSubtractExpression is called when production oC_AddOrSubtractExpression is exited.
func (s *BaseCypherListener) ExitOC_AddOrSubtractExpression(ctx *OC_AddOrSubtractExpressionContext) {}

// EnterOC_MultiplyDivideModuloExpression is called when production oC_MultiplyDivideModuloExpression is entered.
func (s *BaseCypherListener) EnterOC_MultiplyDivideModuloExpression(ctx *OC_MultiplyDivideModuloExpressionContext) {
}

// ExitOC_MultiplyDivideModuloExpression is called when production oC_MultiplyDivideModuloExpression is exited.
func (s *BaseCypherListener) ExitOC_MultiplyDivideModuloExpression(ctx *OC_MultiplyDivideModuloExpressionContext) {
}

// EnterOC_PowerOfExpression is called when production oC_PowerOfExpression is entered.
func (s *BaseCypherListener) EnterOC_PowerOfExpression(ctx *OC_PowerOfExpressionContext) {}

// ExitOC_PowerOfExpression is called when production oC_PowerOfExpression is exited.
func (s *BaseCypherListener) ExitOC_PowerOfExpression(ctx *OC_PowerOfExpressionContext) {}

// EnterOC_UnaryAddOrSubtractExpression is called when production oC_UnaryAddOrSubtractExpression is entered.
func (s *BaseCypherListener) EnterOC_UnaryAddOrSubtractExpression(ctx *OC_UnaryAddOrSubtractExpressionContext) {
}

// ExitOC_UnaryAddOrSubtractExpression is called when production oC_UnaryAddOrSubtractExpression is exited.
func (s *BaseCypherListener) ExitOC_UnaryAddOrSubtractExpression(ctx *OC_UnaryAddOrSubtractExpressionContext) {
}

// EnterOC_NonArithmeticOperatorExpression is called when production oC_NonArithmeticOperatorExpression is entered.
func (s *BaseCypherListener) EnterOC_NonArithmeticOperatorExpression(ctx *OC_NonArithmeticOperatorExpressionContext) {
}

// ExitOC_NonArithmeticOperatorExpression is called when production oC_NonArithmeticOperatorExpression is exited.
func (s *BaseCypherListener) ExitOC_NonArithmeticOperatorExpression(ctx *OC_NonArithmeticOperatorExpressionContext) {
}

// EnterOC_ListOperatorExpression is called when production oC_ListOperatorExpression is entered.
func (s *BaseCypherListener) EnterOC_ListOperatorExpression(ctx *OC_ListOperatorExpressionContext) {}

// ExitOC_ListOperatorExpression is called when production oC_ListOperatorExpression is exited.
func (s *BaseCypherListener) ExitOC_ListOperatorExpression(ctx *OC_ListOperatorExpressionContext) {}

// EnterOC_PropertyLookup is called when production oC_PropertyLookup is entered.
func (s *BaseCypherListener) EnterOC_PropertyLookup(ctx *OC_PropertyLookupContext) {}

// ExitOC_PropertyLookup is called when production oC_PropertyLookup is exited.
func (s *BaseCypherListener) ExitOC_PropertyLookup(ctx *OC_PropertyLookupContext) {}

// EnterOC_Atom is called when production oC_Atom is entered.
func (s *BaseCypherListener) EnterOC_Atom(ctx *OC_AtomContext) {}

// ExitOC_Atom is called when production oC_Atom is exited.
func (s *BaseCypherListener) ExitOC_Atom(ctx *OC_AtomContext) {}

// EnterOC_CaseExpression is called when production oC_CaseExpression is entered.
func (s *BaseCypherListener) EnterOC_CaseExpression(ctx *OC_CaseExpressionContext) {}

// ExitOC_CaseExpression is called when production oC_CaseExpression is exited.
func (s *BaseCypherListener) ExitOC_CaseExpression(ctx *OC_CaseExpressionContext) {}

// EnterOC_CaseAlternative is called when production oC_CaseAlternative is entered.
func (s *BaseCypherListener) EnterOC_CaseAlternative(ctx *OC_CaseAlternativeContext) {}

// ExitOC_CaseAlternative is called when production oC_CaseAlternative is exited.
func (s *BaseCypherListener) ExitOC_CaseAlternative(ctx *OC_CaseAlternativeContext) {}

// EnterOC_ListComprehension is called when production oC_ListComprehension is entered.
func (s *BaseCypherListener) EnterOC_ListComprehension(ctx *OC_ListComprehensionContext) {}

// ExitOC_ListComprehension is called when production oC_ListComprehension is exited.
func (s *BaseCypherListener) ExitOC_ListComprehension(ctx *OC_ListComprehensionContext) {}

// EnterOC_PatternComprehension is called when production oC_PatternComprehension is entered.
func (s *BaseCypherListener) EnterOC_PatternComprehension(ctx *OC_PatternComprehensionContext) {}

// ExitOC_PatternComprehension is called when production oC_PatternComprehension is exited.
func (s *BaseCypherListener) ExitOC_PatternComprehension(ctx *OC_PatternComprehensionContext) {}

// EnterOC_LegacyListExpression is called when production oC_LegacyListExpression is entered.
func (s *BaseCypherListener) EnterOC_LegacyListExpression(ctx *OC_LegacyListExpressionContext) {}

// ExitOC_LegacyListExpression is called when production oC_LegacyListExpression is exited.
func (s *BaseCypherListener) ExitOC_LegacyListExpression(ctx *OC_LegacyListExpressionContext) {}

// EnterOC_Reduce is called when production oC_Reduce is entered.
func (s *BaseCypherListener) EnterOC_Reduce(ctx *OC_ReduceContext) {}

// ExitOC_Reduce is called when production oC_Reduce is exited.
func (s *BaseCypherListener) ExitOC_Reduce(ctx *OC_ReduceContext) {}

// EnterOC_Quantifier is called when production oC_Quantifier is entered.
func (s *BaseCypherListener) EnterOC_Quantifier(ctx *OC_QuantifierContext) {}

// ExitOC_Quantifier is called when production oC_Quantifier is exited.
func (s *BaseCypherListener) ExitOC_Quantifier(ctx *OC_QuantifierContext) {}

// EnterOC_FilterExpression is called when production oC_FilterExpression is entered.
func (s *BaseCypherListener) EnterOC_FilterExpression(ctx *OC_FilterExpressionContext) {}

// ExitOC_FilterExpression is called when production oC_FilterExpression is exited.
func (s *BaseCypherListener) ExitOC_FilterExpression(ctx *OC_FilterExpressionContext) {}

// EnterOC_PatternPredicate is called when production oC_PatternPredicate is entered.
func (s *BaseCypherListener) EnterOC_PatternPredicate(ctx *OC_PatternPredicateContext) {}

// ExitOC_PatternPredicate is called when production oC_PatternPredicate is exited.
func (s *BaseCypherListener) ExitOC_PatternPredicate(ctx *OC_PatternPredicateContext) {}

// EnterOC_ParenthesizedExpression is called when production oC_ParenthesizedExpression is entered.
func (s *BaseCypherListener) EnterOC_ParenthesizedExpression(ctx *OC_ParenthesizedExpressionContext) {
}

// ExitOC_ParenthesizedExpression is called when production oC_ParenthesizedExpression is exited.
func (s *BaseCypherListener) ExitOC_ParenthesizedExpression(ctx *OC_ParenthesizedExpressionContext) {}

// EnterOC_IdInColl is called when production oC_IdInColl is entered.
func (s *BaseCypherListener) EnterOC_IdInColl(ctx *OC_IdInCollContext) {}

// ExitOC_IdInColl is called when production oC_IdInColl is exited.
func (s *BaseCypherListener) ExitOC_IdInColl(ctx *OC_IdInCollContext) {}

// EnterOC_FunctionInvocation is called when production oC_FunctionInvocation is entered.
func (s *BaseCypherListener) EnterOC_FunctionInvocation(ctx *OC_FunctionInvocationContext) {}

// ExitOC_FunctionInvocation is called when production oC_FunctionInvocation is exited.
func (s *BaseCypherListener) ExitOC_FunctionInvocation(ctx *OC_FunctionInvocationContext) {}

// EnterOC_FunctionName is called when production oC_FunctionName is entered.
func (s *BaseCypherListener) EnterOC_FunctionName(ctx *OC_FunctionNameContext) {}

// ExitOC_FunctionName is called when production oC_FunctionName is exited.
func (s *BaseCypherListener) ExitOC_FunctionName(ctx *OC_FunctionNameContext) {}

// EnterOC_ExistentialSubquery is called when production oC_ExistentialSubquery is entered.
func (s *BaseCypherListener) EnterOC_ExistentialSubquery(ctx *OC_ExistentialSubqueryContext) {}

// ExitOC_ExistentialSubquery is called when production oC_ExistentialSubquery is exited.
func (s *BaseCypherListener) ExitOC_ExistentialSubquery(ctx *OC_ExistentialSubqueryContext) {}

// EnterOC_ExplicitProcedureInvocation is called when production oC_ExplicitProcedureInvocation is entered.
func (s *BaseCypherListener) EnterOC_ExplicitProcedureInvocation(ctx *OC_ExplicitProcedureInvocationContext) {
}

// ExitOC_ExplicitProcedureInvocation is called when production oC_ExplicitProcedureInvocation is exited.
func (s *BaseCypherListener) ExitOC_ExplicitProcedureInvocation(ctx *OC_ExplicitProcedureInvocationContext) {
}

// EnterOC_ImplicitProcedureInvocation is called when production oC_ImplicitProcedureInvocation is entered.
func (s *BaseCypherListener) EnterOC_ImplicitProcedureInvocation(ctx *OC_ImplicitProcedureInvocationContext) {
}

// ExitOC_ImplicitProcedureInvocation is called when production oC_ImplicitProcedureInvocation is exited.
func (s *BaseCypherListener) ExitOC_ImplicitProcedureInvocation(ctx *OC_ImplicitProcedureInvocationContext) {
}

// EnterOC_ProcedureResultField is called when production oC_ProcedureResultField is entered.
func (s *BaseCypherListener) EnterOC_ProcedureResultField(ctx *OC_ProcedureResultFieldContext) {}

// ExitOC_ProcedureResultField is called when production oC_ProcedureResultField is exited.
func (s *BaseCypherListener) ExitOC_ProcedureResultField(ctx *OC_ProcedureResultFieldContext) {}

// EnterOC_ProcedureName is called when production oC_ProcedureName is entered.
func (s *BaseCypherListener) EnterOC_ProcedureName(ctx *OC_ProcedureNameContext) {}

// ExitOC_ProcedureName is called when production oC_ProcedureName is exited.
func (s *BaseCypherListener) ExitOC_ProcedureName(ctx *OC_ProcedureNameContext) {}

// EnterOC_Namespace is called when production oC_Namespace is entered.
func (s *BaseCypherListener) EnterOC_Namespace(ctx *OC_NamespaceContext) {}

// ExitOC_Namespace is called when production oC_Namespace is exited.
func (s *BaseCypherListener) ExitOC_Namespace(ctx *OC_NamespaceContext) {}

// EnterOC_Variable is called when production oC_Variable is entered.
func (s *BaseCypherListener) EnterOC_Variable(ctx *OC_VariableContext) {}

// ExitOC_Variable is called when production oC_Variable is exited.
func (s *BaseCypherListener) ExitOC_Variable(ctx *OC_VariableContext) {}

// EnterOC_Literal is called when production oC_Literal is entered.
func (s *BaseCypherListener) EnterOC_Literal(ctx *OC_LiteralContext) {}

// ExitOC_Literal is called when production oC_Literal is exited.
func (s *BaseCypherListener) ExitOC_Literal(ctx *OC_LiteralContext) {}

// EnterOC_BooleanLiteral is called when production oC_BooleanLiteral is entered.
func (s *BaseCypherListener) EnterOC_BooleanLiteral(ctx *OC_BooleanLiteralContext) {}

// ExitOC_BooleanLiteral is called when production oC_BooleanLiteral is exited.
func (s *BaseCypherListener) ExitOC_BooleanLiteral(ctx *OC_BooleanLiteralContext) {}

// EnterOC_NumberLiteral is called when production oC_NumberLiteral is entered.
func (s *BaseCypherListener) EnterOC_NumberLiteral(ctx *OC_NumberLiteralContext) {}

// ExitOC_NumberLiteral is called when production oC_NumberLiteral is exited.
func (s *BaseCypherListener) ExitOC_NumberLiteral(ctx *OC_NumberLiteralContext) {}

// EnterOC_IntegerLiteral is called when production oC_IntegerLiteral is entered.
func (s *BaseCypherListener) EnterOC_IntegerLiteral(ctx *OC_IntegerLiteralContext) {}

// ExitOC_IntegerLiteral is called when production oC_IntegerLiteral is exited.
func (s *BaseCypherListener) ExitOC_IntegerLiteral(ctx *OC_IntegerLiteralContext) {}

// EnterOC_DoubleLiteral is called when production oC_DoubleLiteral is entered.
func (s *BaseCypherListener) EnterOC_DoubleLiteral(ctx *OC_DoubleLiteralContext) {}

// ExitOC_DoubleLiteral is called when production oC_DoubleLiteral is exited.
func (s *BaseCypherListener) ExitOC_DoubleLiteral(ctx *OC_DoubleLiteralContext) {}

// EnterOC_ListLiteral is called when production oC_ListLiteral is entered.
func (s *BaseCypherListener) EnterOC_ListLiteral(ctx *OC_ListLiteralContext) {}

// ExitOC_ListLiteral is called when production oC_ListLiteral is exited.
func (s *BaseCypherListener) ExitOC_ListLiteral(ctx *OC_ListLiteralContext) {}

// EnterOC_MapLiteral is called when production oC_MapLiteral is entered.
func (s *BaseCypherListener) EnterOC_MapLiteral(ctx *OC_MapLiteralContext) {}

// ExitOC_MapLiteral is called when production oC_MapLiteral is exited.
func (s *BaseCypherListener) ExitOC_MapLiteral(ctx *OC_MapLiteralContext) {}

// EnterOC_PropertyKeyName is called when production oC_PropertyKeyName is entered.
func (s *BaseCypherListener) EnterOC_PropertyKeyName(ctx *OC_PropertyKeyNameContext) {}

// ExitOC_PropertyKeyName is called when production oC_PropertyKeyName is exited.
func (s *BaseCypherListener) ExitOC_PropertyKeyName(ctx *OC_PropertyKeyNameContext) {}

// EnterOC_LegacyParameter is called when production oC_LegacyParameter is entered.
func (s *BaseCypherListener) EnterOC_LegacyParameter(ctx *OC_LegacyParameterContext) {}

// ExitOC_LegacyParameter is called when production oC_LegacyParameter is exited.
func (s *BaseCypherListener) ExitOC_LegacyParameter(ctx *OC_LegacyParameterContext) {}

// EnterOC_Parameter is called when production oC_Parameter is entered.
func (s *BaseCypherListener) EnterOC_Parameter(ctx *OC_ParameterContext) {}

// ExitOC_Parameter is called when production oC_Parameter is exited.
func (s *BaseCypherListener) ExitOC_Parameter(ctx *OC_ParameterContext) {}

// EnterOC_SchemaName is called when production oC_SchemaName is entered.
func (s *BaseCypherListener) EnterOC_SchemaName(ctx *OC_SchemaNameContext) {}

// ExitOC_SchemaName is called when production oC_SchemaName is exited.
func (s *BaseCypherListener) ExitOC_SchemaName(ctx *OC_SchemaNameContext) {}

// EnterOC_ReservedWord is called when production oC_ReservedWord is entered.
func (s *BaseCypherListener) EnterOC_ReservedWord(ctx *OC_ReservedWordContext) {}

// ExitOC_ReservedWord is called when production oC_ReservedWord is exited.
func (s *BaseCypherListener) ExitOC_ReservedWord(ctx *OC_ReservedWordContext) {}

// EnterOC_SymbolicName is called when production oC_SymbolicName is entered.
func (s *BaseCypherListener) EnterOC_SymbolicName(ctx *OC_SymbolicNameContext) {}

// ExitOC_SymbolicName is called when production oC_SymbolicName is exited.
func (s *BaseCypherListener) ExitOC_SymbolicName(ctx *OC_SymbolicNameContext) {}

// EnterOC_LeftArrowHead is called when production oC_LeftArrowHead is entered.
func (s *BaseCypherListener) EnterOC_LeftArrowHead(ctx *OC_LeftArrowHeadContext) {}

// ExitOC_LeftArrowHead is called when production oC_LeftArrowHead is exited.
func (s *BaseCypherListener) ExitOC_LeftArrowHead(ctx *OC_LeftArrowHeadContext) {}

// EnterOC_RightArrowHead is called when production oC_RightArrowHead is entered.
func (s *BaseCypherListener) EnterOC_RightArrowHead(ctx *OC_RightArrowHeadContext) {}

// ExitOC_RightArrowHead is called when production oC_RightArrowHead is exited.
func (s *BaseCypherListener) ExitOC_RightArrowHead(ctx *OC_RightArrowHeadContext) {}

// EnterOC_Dash is called when production oC_Dash is entered.
func (s *BaseCypherListener) EnterOC_Dash(ctx *OC_DashContext) {}

// ExitOC_Dash is called when production oC_Dash is exited.
func (s *BaseCypherListener) ExitOC_Dash(ctx *OC_DashContext) {}
