/**
 * Copyright (c) 2015-2023 "Neo Technology,"
 * Network Engine for Objects in Lund AB [http://neotechnology.com]
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 * Attribution Notice under the terms of the Apache License 2.0
 *
 * This work was created by the collective efforts of the openCypher community.
 * Without limiting the terms of Section 6, any Derivative Work that is not
 * approved by the public consensus process of the openCypher Implementers Group
 * should not be described as “Cypher” (and Cypher® is a registered trademark of
 * Neo4j Inc.) or as "openCypher". Extensions by implementers or prototypes or
 * proposals for change that have been documented or implemented should only be
 * described as "implementation extensions to Cypher" or as "proposed changes to
 * Cypher that are not yet approved by the openCypher community".
 */
grammar Cypher;

oC_Cypher
      :  SP? oC_QueryOptions oC_Statement ( SP? ';' )? SP? EOF ;

oC_QueryOptions
            :  ( oC_AnyCypherOption SP? )* ;

oC_AnyCypherOption
               :  oC_CypherOption
                   | oC_Explain
                   | oC_Profile
                   ;

oC_CypherOption
            :  CYPHER ( SP oC_VersionNumber )? ( SP oC_ConfigurationOption )* ;

CYPHER : ( 'C' | 'c' ) ( 'Y' | 'y' ) ( 'P' | 'p' ) ( 'H' | 'h' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ;

oC_VersionNumber
             :  RegularDecimalReal ;

oC_Explain
       :  EXPLAIN ;

EXPLAIN : ( 'E' | 'e' ) ( 'X' | 'x' ) ( 'P' | 'p' ) ( 'L' | 'l' ) ( 'A' | 'a' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ;

oC_Profile
       :  PROFILE ;

PROFILE : ( 'P' | 'p' ) ( 'R' | 'r' ) ( 'O' | 'o' ) ( 'F' | 'f' ) ( 'I' | 'i' ) ( 'L' | 'l' ) ( 'E' | 'e' ) ;

oC_ConfigurationOption
                   :  oC_SymbolicName SP? '=' SP? oC_SymbolicName ;

oC_Statement
         :  oC_Command
             | oC_Query
             ;

oC_Query
     :  oC_RegularQuery
         | oC_StandaloneCall
         | oC_BulkImportQuery
         ;

oC_RegularQuery
            :  oC_SingleQuery ( SP? oC_Union )* ;

oC_BulkImportQuery
               :  oC_PeriodicCommitHint SP? oC_LoadCSVQuery ;

oC_PeriodicCommitHint
                  :  USING SP PERIODIC SP COMMIT ( SP oC_IntegerLiteral )? ;

USING : ( 'U' | 'u' ) ( 'S' | 's' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'G' | 'g' ) ;

PERIODIC : ( 'P' | 'p' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ( 'I' | 'i' ) ( 'O' | 'o' ) ( 'D' | 'd' ) ( 'I' | 'i' ) ( 'C' | 'c' ) ;

COMMIT : ( 'C' | 'c' ) ( 'O' | 'o' ) ( 'M' | 'm' ) ( 'M' | 'm' ) ( 'I' | 'i' ) ( 'T' | 't' ) ;

oC_LoadCSVQuery
            :  oC_LoadCSV oC_SingleQuery ;

oC_Union
     :  ( UNION SP ALL SP? oC_SingleQuery )
         | ( UNION SP? oC_SingleQuery )
         ;

UNION : ( 'U' | 'u' ) ( 'N' | 'n' ) ( 'I' | 'i' ) ( 'O' | 'o' ) ( 'N' | 'n' ) ;

ALL : ( 'A' | 'a' ) ( 'L' | 'l' ) ( 'L' | 'l' ) ;

oC_SingleQuery
           :  oC_SinglePartQuery
               | oC_MultiPartQuery
               ;

oC_SinglePartQuery
               :  ( ( oC_ReadingClause SP? )* oC_Return )
                   | ( ( oC_ReadingClause SP? )* oC_UpdatingClause ( SP? oC_UpdatingClause )* ( SP? oC_Return )? )
                   ;

oC_MultiPartQuery
              :  ( ( oC_ReadingClause SP? )* ( oC_UpdatingClause SP? )* oC_With SP? )+ oC_SinglePartQuery ;

oC_UpdatingClause
              :  oC_Create
                  | oC_Merge
                  | oC_CreateUnique
                  | oC_Foreach
                  | oC_Delete
                  | oC_Set
                  | oC_Remove
                  ;

oC_ReadingClause
             :  oC_LoadCSV
                 | oC_Start
                 | oC_Match
                 | oC_Unwind
                 | oC_InQueryCall
                 ;

oC_Command
       :  oC_CreateIndex
           | oC_DropIndex
           | oC_CreateUniqueConstraint
           | oC_DropUniqueConstraint
           | oC_CreateNodePropertyExistenceConstraint
           | oC_DropNodePropertyExistenceConstraint
           | oC_CreateRelationshipPropertyExistenceConstraint
           | oC_DropRelationshipPropertyExistenceConstraint
           ;

oC_CreateUniqueConstraint
                      :  CREATE SP oC_UniqueConstraint ;

CREATE : ( 'C' | 'c' ) ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'A' | 'a' ) ( 'T' | 't' ) ( 'E' | 'e' ) ;

oC_CreateNodePropertyExistenceConstraint
                                     :  CREATE SP oC_NodePropertyExistenceConstraint ;

oC_CreateRelationshipPropertyExistenceConstraint
                                             :  CREATE SP oC_RelationshipPropertyExistenceConstraint ;

oC_CreateIndex
           :  CREATE SP oC_Index ;

oC_DropUniqueConstraint
                    :  DROP SP oC_UniqueConstraint ;

DROP : ( 'D' | 'd' ) ( 'R' | 'r' ) ( 'O' | 'o' ) ( 'P' | 'p' ) ;

oC_DropNodePropertyExistenceConstraint
                                   :  DROP SP oC_NodePropertyExistenceConstraint ;

oC_DropRelationshipPropertyExistenceConstraint
                                           :  DROP SP oC_RelationshipPropertyExistenceConstraint ;

oC_DropIndex
         :  DROP SP oC_Index ;

oC_Index
     :  INDEX SP ON SP? oC_NodeLabel '(' oC_PropertyKeyName ')' ;

INDEX : ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ( 'E' | 'e' ) ( 'X' | 'x' ) ;

ON : ( 'O' | 'o' ) ( 'N' | 'n' ) ;

oC_UniqueConstraint
                :  CONSTRAINT SP ON SP? '(' oC_Variable oC_NodeLabel ')' SP? ASSERT SP oC_PropertyExpression SP IS SP UNIQUE ;

CONSTRAINT : ( 'C' | 'c' ) ( 'O' | 'o' ) ( 'N' | 'n' ) ( 'S' | 's' ) ( 'T' | 't' ) ( 'R' | 'r' ) ( 'A' | 'a' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'T' | 't' ) ;

ASSERT : ( 'A' | 'a' ) ( 'S' | 's' ) ( 'S' | 's' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ( 'T' | 't' ) ;

IS : ( 'I' | 'i' ) ( 'S' | 's' ) ;

UNIQUE : ( 'U' | 'u' ) ( 'N' | 'n' ) ( 'I' | 'i' ) ( 'Q' | 'q' ) ( 'U' | 'u' ) ( 'E' | 'e' ) ;

oC_NodePropertyExistenceConstraint
                               :  CONSTRAINT SP ON SP? '(' oC_Variable oC_NodeLabel ')' SP? ASSERT SP EXISTS SP? '(' oC_PropertyExpression ')' ;

EXISTS : ( 'E' | 'e' ) ( 'X' | 'x' ) ( 'I' | 'i' ) ( 'S' | 's' ) ( 'T' | 't' ) ( 'S' | 's' ) ;

oC_RelationshipPropertyExistenceConstraint
                                       :  CONSTRAINT SP ON SP? oC_RelationshipPatternSyntax SP? ASSERT SP EXISTS SP? '(' oC_PropertyExpression ')' ;

oC_RelationshipPatternSyntax
                         :  ( '(' SP? ')' oC_Dash '[' oC_Variable oC_RelType ']' oC_Dash '(' SP? ')' )
                             | ( '(' SP? ')' oC_Dash '[' oC_Variable oC_RelType ']' oC_Dash oC_RightArrowHead '(' SP? ')' )
                             | ( '(' SP? ')' oC_LeftArrowHead oC_Dash '[' oC_Variable oC_RelType ']' oC_Dash '(' SP? ')' )
                             ;

oC_LoadCSV
       :  LOAD SP CSV SP ( WITH SP HEADERS SP )? FROM SP oC_Expression SP AS SP oC_Variable SP ( FIELDTERMINATOR SP StringLiteral )? ;

LOAD : ( 'L' | 'l' ) ( 'O' | 'o' ) ( 'A' | 'a' ) ( 'D' | 'd' ) ;

CSV : ( 'C' | 'c' ) ( 'S' | 's' ) ( 'V' | 'v' ) ;

WITH : ( 'W' | 'w' ) ( 'I' | 'i' ) ( 'T' | 't' ) ( 'H' | 'h' ) ;

HEADERS : ( 'H' | 'h' ) ( 'E' | 'e' ) ( 'A' | 'a' ) ( 'D' | 'd' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ( 'S' | 's' ) ;

FROM : ( 'F' | 'f' ) ( 'R' | 'r' ) ( 'O' | 'o' ) ( 'M' | 'm' ) ;

AS : ( 'A' | 'a' ) ( 'S' | 's' ) ;

FIELDTERMINATOR : ( 'F' | 'f' ) ( 'I' | 'i' ) ( 'E' | 'e' ) ( 'L' | 'l' ) ( 'D' | 'd' ) ( 'T' | 't' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ( 'M' | 'm' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'A' | 'a' ) ( 'T' | 't' ) ( 'O' | 'o' ) ( 'R' | 'r' ) ;

oC_Match
     :  ( OPTIONAL SP )? MATCH SP? oC_Pattern ( oC_Hint )* ( SP? oC_Where )? ;

OPTIONAL : ( 'O' | 'o' ) ( 'P' | 'p' ) ( 'T' | 't' ) ( 'I' | 'i' ) ( 'O' | 'o' ) ( 'N' | 'n' ) ( 'A' | 'a' ) ( 'L' | 'l' ) ;

MATCH : ( 'M' | 'm' ) ( 'A' | 'a' ) ( 'T' | 't' ) ( 'C' | 'c' ) ( 'H' | 'h' ) ;

oC_Unwind
      :  UNWIND SP? oC_Expression SP AS SP oC_Variable ;

UNWIND : ( 'U' | 'u' ) ( 'N' | 'n' ) ( 'W' | 'w' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ;

oC_Merge
     :  MERGE SP? oC_PatternPart ( SP oC_MergeAction )* ;

MERGE : ( 'M' | 'm' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ( 'G' | 'g' ) ( 'E' | 'e' ) ;

oC_MergeAction
           :  ( ON SP MATCH SP oC_Set )
               | ( ON SP CREATE SP oC_Set )
               ;

oC_Create
      :  CREATE SP? oC_Pattern ;

oC_CreateUnique
            :  CREATE SP UNIQUE SP? oC_Pattern ;

oC_Set
   :  SET SP? oC_SetItem ( SP? ',' SP? oC_SetItem )* ;

SET : ( 'S' | 's' ) ( 'E' | 'e' ) ( 'T' | 't' ) ;

oC_SetItem
       :  ( oC_PropertyExpression SP? '=' SP? oC_Expression )
           | ( oC_Variable SP? '=' SP? oC_Expression )
           | ( oC_Variable SP? '+=' SP? oC_Expression )
           | ( oC_Variable SP? oC_NodeLabels )
           ;

oC_Delete
      :  ( DETACH SP )? DELETE SP? oC_Expression ( SP? ',' SP? oC_Expression )* ;

DETACH : ( 'D' | 'd' ) ( 'E' | 'e' ) ( 'T' | 't' ) ( 'A' | 'a' ) ( 'C' | 'c' ) ( 'H' | 'h' ) ;

DELETE : ( 'D' | 'd' ) ( 'E' | 'e' ) ( 'L' | 'l' ) ( 'E' | 'e' ) ( 'T' | 't' ) ( 'E' | 'e' ) ;

oC_Remove
      :  REMOVE SP oC_RemoveItem ( SP? ',' SP? oC_RemoveItem )* ;

REMOVE : ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'M' | 'm' ) ( 'O' | 'o' ) ( 'V' | 'v' ) ( 'E' | 'e' ) ;

oC_RemoveItem
          :  ( oC_Variable oC_NodeLabels )
              | oC_PropertyExpression
              ;

oC_Foreach
       :  FOREACH SP? '(' SP? oC_Variable SP IN SP oC_Expression SP? '|' ( SP oC_UpdatingClause )+ SP? ')' ;

FOREACH : ( 'F' | 'f' ) ( 'O' | 'o' ) ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'A' | 'a' ) ( 'C' | 'c' ) ( 'H' | 'h' ) ;

IN : ( 'I' | 'i' ) ( 'N' | 'n' ) ;

oC_InQueryCall
           :  CALL SP oC_ExplicitProcedureInvocation ( SP? YIELD SP oC_YieldItems )? ;

CALL : ( 'C' | 'c' ) ( 'A' | 'a' ) ( 'L' | 'l' ) ( 'L' | 'l' ) ;

YIELD : ( 'Y' | 'y' ) ( 'I' | 'i' ) ( 'E' | 'e' ) ( 'L' | 'l' ) ( 'D' | 'd' ) ;

oC_StandaloneCall
              :  CALL SP ( oC_ExplicitProcedureInvocation | oC_ImplicitProcedureInvocation ) ( SP? YIELD SP ( '*' | oC_YieldItems ) )? ;

oC_YieldItems
          :  oC_YieldItem ( SP? ',' SP? oC_YieldItem )* ( SP? oC_Where )? ;

oC_YieldItem
         :  ( oC_ProcedureResultField SP AS SP )? oC_Variable ;

oC_With
    :  WITH oC_ProjectionBody ( SP? oC_Where )? ;

oC_Return
      :  RETURN oC_ProjectionBody ;

RETURN : ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'T' | 't' ) ( 'U' | 'u' ) ( 'R' | 'r' ) ( 'N' | 'n' ) ;

oC_ProjectionBody
              :  ( SP? DISTINCT )? SP oC_ProjectionItems ( SP oC_Order )? ( SP oC_Skip )? ( SP oC_Limit )? ;

DISTINCT : ( 'D' | 'd' ) ( 'I' | 'i' ) ( 'S' | 's' ) ( 'T' | 't' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'C' | 'c' ) ( 'T' | 't' ) ;

oC_ProjectionItems
               :  ( '*' ( SP? ',' SP? oC_ProjectionItem )* )
                   | ( oC_ProjectionItem ( SP? ',' SP? oC_ProjectionItem )* )
                   ;

oC_ProjectionItem
              :  ( oC_Expression SP AS SP oC_Variable )
                  | oC_Expression
                  ;

oC_Order
     :  ORDER SP BY SP oC_SortItem ( ',' SP? oC_SortItem )* ;

ORDER : ( 'O' | 'o' ) ( 'R' | 'r' ) ( 'D' | 'd' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ;

BY : ( 'B' | 'b' ) ( 'Y' | 'y' ) ;

oC_Skip
    :  L_SKIP SP oC_Expression ;

L_SKIP : ( 'S' | 's' ) ( 'K' | 'k' ) ( 'I' | 'i' ) ( 'P' | 'p' ) ;

oC_Limit
     :  LIMIT SP oC_Expression ;

LIMIT : ( 'L' | 'l' ) ( 'I' | 'i' ) ( 'M' | 'm' ) ( 'I' | 'i' ) ( 'T' | 't' ) ;

oC_SortItem
        :  oC_Expression ( SP? ( ASCENDING | ASC | DESCENDING | DESC ) )? ;

ASCENDING : ( 'A' | 'a' ) ( 'S' | 's' ) ( 'C' | 'c' ) ( 'E' | 'e' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'G' | 'g' ) ;

ASC : ( 'A' | 'a' ) ( 'S' | 's' ) ( 'C' | 'c' ) ;

DESCENDING : ( 'D' | 'd' ) ( 'E' | 'e' ) ( 'S' | 's' ) ( 'C' | 'c' ) ( 'E' | 'e' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'G' | 'g' ) ;

DESC : ( 'D' | 'd' ) ( 'E' | 'e' ) ( 'S' | 's' ) ( 'C' | 'c' ) ;

oC_Hint
    :  SP? ( ( USING SP INDEX SP oC_Variable oC_NodeLabel '(' oC_PropertyKeyName ')' ) | ( USING SP JOIN SP ON SP oC_Variable ( SP? ',' SP? oC_Variable )* ) | ( USING SP SCAN SP oC_Variable oC_NodeLabel ) ) ;

JOIN : ( 'J' | 'j' ) ( 'O' | 'o' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ;

SCAN : ( 'S' | 's' ) ( 'C' | 'c' ) ( 'A' | 'a' ) ( 'N' | 'n' ) ;

oC_Start
     :  START SP oC_StartPoint ( SP? ',' SP? oC_StartPoint )* oC_Where? ;

START : ( 'S' | 's' ) ( 'T' | 't' ) ( 'A' | 'a' ) ( 'R' | 'r' ) ( 'T' | 't' ) ;

oC_StartPoint
          :  oC_Variable SP? '=' SP? oC_Lookup ;

oC_Lookup
      :  oC_NodeLookup
          | oC_RelationshipLookup
          ;

oC_NodeLookup
          :  NODE SP? ( oC_IdentifiedIndexLookup | oC_IndexQuery | oC_IdLookup ) ;

NODE : ( 'N' | 'n' ) ( 'O' | 'o' ) ( 'D' | 'd' ) ( 'E' | 'e' ) ;

oC_RelationshipLookup
                  :  ( RELATIONSHIP | REL ) ( oC_IdentifiedIndexLookup | oC_IndexQuery | oC_IdLookup ) ;

RELATIONSHIP : ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'L' | 'l' ) ( 'A' | 'a' ) ( 'T' | 't' ) ( 'I' | 'i' ) ( 'O' | 'o' ) ( 'N' | 'n' ) ( 'S' | 's' ) ( 'H' | 'h' ) ( 'I' | 'i' ) ( 'P' | 'p' ) ;

REL : ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'L' | 'l' ) ;

oC_IdentifiedIndexLookup
                     :  ':' oC_SymbolicName '(' oC_SymbolicName '=' ( StringLiteral | oC_LegacyParameter ) ')' ;

oC_IndexQuery
          :  ':' oC_SymbolicName '(' ( StringLiteral | oC_LegacyParameter ) ')' ;

oC_IdLookup
        :  '(' ( oC_LiteralIds | oC_LegacyParameter | '*' ) ')' ;

oC_LiteralIds
          :  oC_IntegerLiteral ( SP? ',' SP? oC_IntegerLiteral )* ;

oC_Where
     :  WHERE SP oC_Expression ;

WHERE : ( 'W' | 'w' ) ( 'H' | 'h' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ( 'E' | 'e' ) ;

oC_Pattern
       :  oC_PatternPart ( SP? ',' SP? oC_PatternPart )* ;

oC_PatternPart
           :  ( oC_Variable SP? '=' SP? oC_AnonymousPatternPart )
               | oC_AnonymousPatternPart
               ;

oC_AnonymousPatternPart
                    :  oC_ShortestPathPattern
                        | oC_PatternElement
                        ;

oC_ShortestPathPattern
                   :  ( SHORTESTPATH '(' oC_PatternElement ')' )
                       | ( ALLSHORTESTPATHS '(' oC_PatternElement ')' )
                       ;

SHORTESTPATH : ( 'S' | 's' ) ( 'H' | 'h' ) ( 'O' | 'o' ) ( 'R' | 'r' ) ( 'T' | 't' ) ( 'E' | 'e' ) ( 'S' | 's' ) ( 'T' | 't' ) ( 'P' | 'p' ) ( 'A' | 'a' ) ( 'T' | 't' ) ( 'H' | 'h' ) ;

ALLSHORTESTPATHS : ( 'A' | 'a' ) ( 'L' | 'l' ) ( 'L' | 'l' ) ( 'S' | 's' ) ( 'H' | 'h' ) ( 'O' | 'o' ) ( 'R' | 'r' ) ( 'T' | 't' ) ( 'E' | 'e' ) ( 'S' | 's' ) ( 'T' | 't' ) ( 'P' | 'p' ) ( 'A' | 'a' ) ( 'T' | 't' ) ( 'H' | 'h' ) ( 'S' | 's' ) ;

oC_PatternElement
              :  ( oC_NodePattern ( SP? oC_PatternElementChain )* )
                  | ( '(' oC_PatternElement ')' )
                  ;

oC_RelationshipsPattern
                    :  oC_NodePattern ( SP? oC_PatternElementChain )+ ;

oC_NodePattern
           :  '(' SP? ( oC_Variable SP? )? ( oC_NodeLabels SP? )? ( oC_Properties SP? )? ')' ;

oC_PatternElementChain
                   :  oC_RelationshipPattern SP? oC_NodePattern ;

oC_RelationshipPattern
                   :  ( oC_LeftArrowHead SP? oC_Dash SP? oC_RelationshipDetail? SP? oC_Dash SP? oC_RightArrowHead )
                       | ( oC_LeftArrowHead SP? oC_Dash SP? oC_RelationshipDetail? SP? oC_Dash )
                       | ( oC_Dash SP? oC_RelationshipDetail? SP? oC_Dash SP? oC_RightArrowHead )
                       | ( oC_Dash SP? oC_RelationshipDetail? SP? oC_Dash )
                       ;

oC_RelationshipDetail
                  :  '[' SP? ( oC_Variable SP? )? ( oC_RelationshipTypes SP? )? oC_RangeLiteral? ( oC_Properties SP? )? ']' ;

oC_Properties
          :  oC_MapLiteral
              | oC_Parameter
              | oC_LegacyParameter
              ;

oC_RelType
       :  ':' SP? oC_RelTypeName ;

oC_RelationshipTypes
                 :  ':' SP? oC_RelTypeName ( SP? '|' ':'? SP? oC_RelTypeName )* ;

oC_NodeLabels
          :  oC_NodeLabel ( SP? oC_NodeLabel )* ;

oC_NodeLabel
         :  ':' SP? oC_LabelName ;

oC_RangeLiteral
            :  '*' SP? ( oC_IntegerLiteral SP? )? ( '..' SP? ( oC_IntegerLiteral SP? )? )? ;

oC_LabelName
         :  oC_SchemaName ;

oC_RelTypeName
           :  oC_SchemaName ;

oC_PropertyExpression
                  :  oC_Atom ( SP? oC_PropertyLookup )+ ;

oC_Expression
          :  oC_OrExpression ;

oC_OrExpression
            :  oC_XorExpression ( SP OR SP oC_XorExpression )* ;

OR : ( 'O' | 'o' ) ( 'R' | 'r' ) ;

oC_XorExpression
             :  oC_AndExpression ( SP XOR SP oC_AndExpression )* ;

XOR : ( 'X' | 'x' ) ( 'O' | 'o' ) ( 'R' | 'r' ) ;

oC_AndExpression
             :  oC_NotExpression ( SP AND SP oC_NotExpression )* ;

AND : ( 'A' | 'a' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ;

oC_NotExpression
             :  ( NOT SP? )* oC_ComparisonExpression ;

NOT : ( 'N' | 'n' ) ( 'O' | 'o' ) ( 'T' | 't' ) ;

oC_ComparisonExpression
                    :  oC_StringListNullPredicateExpression ( SP? oC_PartialComparisonExpression )* ;

oC_PartialComparisonExpression
                           :  ( '=' SP? oC_StringListNullPredicateExpression )
                               | ( '<>' SP? oC_StringListNullPredicateExpression )
                               | ( '<' SP? oC_StringListNullPredicateExpression )
                               | ( '>' SP? oC_StringListNullPredicateExpression )
                               | ( '<=' SP? oC_StringListNullPredicateExpression )
                               | ( '>=' SP? oC_StringListNullPredicateExpression )
                               ;

oC_StringListNullPredicateExpression
                                 :  oC_AddOrSubtractExpression ( oC_StringPredicateExpression | oC_ListPredicateExpression | oC_NullPredicateExpression )* ;

oC_StringPredicateExpression
                         :  ( oC_RegularExpression | ( SP STARTS SP WITH ) | ( SP ENDS SP WITH ) | ( SP CONTAINS ) ) SP? oC_AddOrSubtractExpression ;

STARTS : ( 'S' | 's' ) ( 'T' | 't' ) ( 'A' | 'a' ) ( 'R' | 'r' ) ( 'T' | 't' ) ( 'S' | 's' ) ;

ENDS : ( 'E' | 'e' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ( 'S' | 's' ) ;

CONTAINS : ( 'C' | 'c' ) ( 'O' | 'o' ) ( 'N' | 'n' ) ( 'T' | 't' ) ( 'A' | 'a' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'S' | 's' ) ;

oC_ListPredicateExpression
                       :  SP IN SP? oC_AddOrSubtractExpression ;

oC_NullPredicateExpression
                       :  ( SP IS SP NULL )
                           | ( SP IS SP NOT SP NULL )
                           ;

NULL : ( 'N' | 'n' ) ( 'U' | 'u' ) ( 'L' | 'l' ) ( 'L' | 'l' ) ;

oC_RegularExpression
                 :  SP? '=~' ;

oC_AddOrSubtractExpression
                       :  oC_MultiplyDivideModuloExpression ( ( SP? '+' SP? oC_MultiplyDivideModuloExpression ) | ( SP? '-' SP? oC_MultiplyDivideModuloExpression ) )* ;

oC_MultiplyDivideModuloExpression
                              :  oC_PowerOfExpression ( ( SP? '*' SP? oC_PowerOfExpression ) | ( SP? '/' SP? oC_PowerOfExpression ) | ( SP? '%' SP? oC_PowerOfExpression ) )* ;

oC_PowerOfExpression
                 :  oC_UnaryAddOrSubtractExpression ( SP? '^' SP? oC_UnaryAddOrSubtractExpression )* ;

oC_UnaryAddOrSubtractExpression
                            :  oC_NonArithmeticOperatorExpression
                                | ( ( '+' | '-' ) SP? oC_NonArithmeticOperatorExpression )
                                ;

oC_NonArithmeticOperatorExpression
                               :  oC_Atom ( ( SP? oC_ListOperatorExpression ) | ( SP? oC_PropertyLookup ) )* ( SP? oC_NodeLabels )? ;

oC_ListOperatorExpression
                      :  ( '[' oC_Expression ']' )
                          | ( '[' oC_Expression? '..' oC_Expression? ']' )
                          ;

oC_PropertyLookup
              :  '.' SP? ( oC_PropertyKeyName ) ;

oC_Atom
    :  oC_Literal
        | oC_Parameter
        | oC_LegacyParameter
        | oC_CaseExpression
        | ( COUNT SP? '(' SP? '*' SP? ')' )
        | oC_ListComprehension
        | oC_PatternComprehension
        | oC_LegacyListExpression
        | oC_Reduce
        | oC_Quantifier
        | oC_ShortestPathPattern
        | oC_PatternPredicate
        | oC_ParenthesizedExpression
        | oC_FunctionInvocation
        | oC_ExistentialSubquery
        | oC_Variable
        ;

COUNT : ( 'C' | 'c' ) ( 'O' | 'o' ) ( 'U' | 'u' ) ( 'N' | 'n' ) ( 'T' | 't' ) ;

oC_CaseExpression
              :  ( ( CASE ( SP? oC_CaseAlternative )+ ) | ( CASE SP? oC_Expression ( SP? oC_CaseAlternative )+ ) ) ( SP? ELSE SP? oC_Expression )? SP? END ;

CASE : ( 'C' | 'c' ) ( 'A' | 'a' ) ( 'S' | 's' ) ( 'E' | 'e' ) ;

ELSE : ( 'E' | 'e' ) ( 'L' | 'l' ) ( 'S' | 's' ) ( 'E' | 'e' ) ;

END : ( 'E' | 'e' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ;

oC_CaseAlternative
               :  WHEN SP? oC_Expression SP? THEN SP? oC_Expression ;

WHEN : ( 'W' | 'w' ) ( 'H' | 'h' ) ( 'E' | 'e' ) ( 'N' | 'n' ) ;

THEN : ( 'T' | 't' ) ( 'H' | 'h' ) ( 'E' | 'e' ) ( 'N' | 'n' ) ;

oC_ListComprehension
                 :  '[' SP? oC_FilterExpression ( SP? '|' SP? oC_Expression )? SP? ']' ;

oC_PatternComprehension
                    :  '[' SP? ( oC_Variable SP? '=' SP? )? oC_RelationshipsPattern SP? ( oC_Where SP? )? '|' SP? oC_Expression SP? ']' ;

oC_LegacyListExpression
                    :  ( FILTER SP? '(' SP? oC_FilterExpression SP? ')' )
                        | ( EXTRACT SP? '(' SP? oC_FilterExpression SP? ( SP? '|' oC_Expression )? ')' )
                        ;

FILTER : ( 'F' | 'f' ) ( 'I' | 'i' ) ( 'L' | 'l' ) ( 'T' | 't' ) ( 'E' | 'e' ) ( 'R' | 'r' ) ;

EXTRACT : ( 'E' | 'e' ) ( 'X' | 'x' ) ( 'T' | 't' ) ( 'R' | 'r' ) ( 'A' | 'a' ) ( 'C' | 'c' ) ( 'T' | 't' ) ;

oC_Reduce
      :  REDUCE SP? '(' oC_Variable '=' oC_Expression ',' oC_IdInColl '|' oC_Expression ')' ;

REDUCE : ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'D' | 'd' ) ( 'U' | 'u' ) ( 'C' | 'c' ) ( 'E' | 'e' ) ;

oC_Quantifier
          :  ( ALL SP? '(' SP? oC_FilterExpression SP? ')' )
              | ( ANY SP? '(' SP? oC_FilterExpression SP? ')' )
              | ( NONE SP? '(' SP? oC_FilterExpression SP? ')' )
              | ( SINGLE SP? '(' SP? oC_FilterExpression SP? ')' )
              ;

ANY : ( 'A' | 'a' ) ( 'N' | 'n' ) ( 'Y' | 'y' ) ;

NONE : ( 'N' | 'n' ) ( 'O' | 'o' ) ( 'N' | 'n' ) ( 'E' | 'e' ) ;

SINGLE : ( 'S' | 's' ) ( 'I' | 'i' ) ( 'N' | 'n' ) ( 'G' | 'g' ) ( 'L' | 'l' ) ( 'E' | 'e' ) ;

oC_FilterExpression
                :  oC_IdInColl ( SP? oC_Where )? ;

oC_PatternPredicate
                :  oC_RelationshipsPattern ;

oC_ParenthesizedExpression
                       :  '(' SP? oC_Expression SP? ')' ;

oC_IdInColl
        :  oC_Variable SP IN SP oC_Expression ;

oC_FunctionInvocation
                  :  oC_FunctionName SP? '(' SP? ( DISTINCT SP? )? ( oC_Expression SP? ( ',' SP? oC_Expression SP? )* )? ')' ;

oC_FunctionName
            :  oC_Namespace oC_SymbolicName ;

oC_ExistentialSubquery
                   :  EXISTS SP? '{' SP? ( oC_RegularQuery | ( oC_Pattern ( SP? oC_Where )? ) ) SP? '}' ;

oC_ExplicitProcedureInvocation
                           :  oC_ProcedureName SP? '(' SP? ( oC_Expression SP? ( ',' SP? oC_Expression SP? )* )? ')' ;

oC_ImplicitProcedureInvocation
                           :  oC_ProcedureName ;

oC_ProcedureResultField
                    :  oC_SymbolicName ;

oC_ProcedureName
             :  oC_Namespace oC_SymbolicName ;

oC_Namespace
         :  ( oC_SymbolicName '.' )* ;

oC_Variable
        :  oC_SymbolicName ;

oC_Literal
       :  oC_BooleanLiteral
           | NULL
           | oC_NumberLiteral
           | StringLiteral
           | oC_ListLiteral
           | oC_MapLiteral
           ;

oC_BooleanLiteral
              :  TRUE
                  | FALSE
                  ;

TRUE : ( 'T' | 't' ) ( 'R' | 'r' ) ( 'U' | 'u' ) ( 'E' | 'e' ) ;

FALSE : ( 'F' | 'f' ) ( 'A' | 'a' ) ( 'L' | 'l' ) ( 'S' | 's' ) ( 'E' | 'e' ) ;

oC_NumberLiteral
             :  oC_DoubleLiteral
                 | oC_IntegerLiteral
                 ;

oC_IntegerLiteral
              :  HexInteger
                  | OctalInteger
                  | DecimalInteger
                  ;

HexInteger
          :  '0x' ( HexDigit )+ ;

DecimalInteger
              :  ZeroDigit
                  | ( NonZeroDigit ( Digit )* )
                  ;

OctalInteger
            :  '0o' ( OctDigit )+ ;

HexLetter
         :  ( ( 'A' | 'a' ) )
             | ( ( 'B' | 'b' ) )
             | ( ( 'C' | 'c' ) )
             | ( ( 'D' | 'd' ) )
             | ( ( 'E' | 'e' ) )
             | ( ( 'F' | 'f' ) )
             ;

HexDigit
        :  Digit
            | HexLetter
            ;

Digit
     :  ZeroDigit
         | NonZeroDigit
         ;

NonZeroDigit
            :  NonZeroOctDigit
                | '8'
                | '9'
                ;

NonZeroOctDigit
               :  '1'
                   | '2'
                   | '3'
                   | '4'
                   | '5'
                   | '6'
                   | '7'
                   ;

OctDigit
        :  ZeroDigit
            | NonZeroOctDigit
            ;

ZeroDigit
         :  '0' ;

oC_DoubleLiteral
             :  ExponentDecimalReal
                 | RegularDecimalReal
                 ;

ExponentDecimalReal
                   :  ( ( Digit )+ | ( ( Digit )+ '.' ( Digit )+ ) | ( '.' ( Digit )+ ) ) ( ( 'E' | 'e' ) ) '-'? ( Digit )+ ;

RegularDecimalReal
                  :  ( Digit )* '.' ( Digit )+ ;

StringLiteral
             :  ( '"' ( StringLiteral_0 | EscapedChar )* '"' )
                 | ( '\'' ( StringLiteral_1 | EscapedChar )* '\'' )
                 ;

EscapedChar
           :  '\\' ( '\\' | '\'' | '"' | ( ( 'B' | 'b' ) ) | ( ( 'F' | 'f' ) ) | ( ( 'N' | 'n' ) ) | ( ( 'R' | 'r' ) ) | ( ( 'T' | 't' ) ) | ( ( ( 'U' | 'u' ) ) ( HexDigit HexDigit HexDigit HexDigit ) ) | ( ( ( 'U' | 'u' ) ) ( HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit HexDigit ) ) ) ;

oC_ListLiteral
           :  '[' SP? ( oC_Expression SP? ( ',' SP? oC_Expression SP? )* )? ']' ;

oC_MapLiteral
          :  '{' SP? ( oC_PropertyKeyName SP? ':' SP? oC_Expression SP? ( ',' SP? oC_PropertyKeyName SP? ':' SP? oC_Expression SP? )* )? '}' ;

oC_PropertyKeyName
               :  oC_SchemaName ;

oC_LegacyParameter
               :  '{' SP? ( oC_SymbolicName | DecimalInteger ) SP? '}' ;

oC_Parameter
         :  '$' ( oC_SymbolicName | DecimalInteger ) ;

oC_SchemaName
          :  oC_SymbolicName
              | oC_ReservedWord
              ;

oC_ReservedWord
            :  ALL
                | ASC
                | ASCENDING
                | BY
                | CREATE
                | DELETE
                | DESC
                | DESCENDING
                | DETACH
                | EXISTS
                | LIMIT
                | MATCH
                | MERGE
                | ON
                | OPTIONAL
                | ORDER
                | REMOVE
                | RETURN
                | SET
                | L_SKIP
                | WHERE
                | WITH
                | UNION
                | UNWIND
                | AND
                | AS
                | CONTAINS
                | DISTINCT
                | ENDS
                | IN
                | IS
                | NOT
                | OR
                | STARTS
                | XOR
                | FALSE
                | TRUE
                | NULL
                | CONSTRAINT
                | DO
                | FOR
                | REQUIRE
                | UNIQUE
                | CASE
                | WHEN
                | THEN
                | ELSE
                | END
                | MANDATORY
                | SCALAR
                | OF
                | ADD
                | DROP
                ;

DO : ( 'D' | 'd' ) ( 'O' | 'o' ) ;

FOR : ( 'F' | 'f' ) ( 'O' | 'o' ) ( 'R' | 'r' ) ;

REQUIRE : ( 'R' | 'r' ) ( 'E' | 'e' ) ( 'Q' | 'q' ) ( 'U' | 'u' ) ( 'I' | 'i' ) ( 'R' | 'r' ) ( 'E' | 'e' ) ;

MANDATORY : ( 'M' | 'm' ) ( 'A' | 'a' ) ( 'N' | 'n' ) ( 'D' | 'd' ) ( 'A' | 'a' ) ( 'T' | 't' ) ( 'O' | 'o' ) ( 'R' | 'r' ) ( 'Y' | 'y' ) ;

SCALAR : ( 'S' | 's' ) ( 'C' | 'c' ) ( 'A' | 'a' ) ( 'L' | 'l' ) ( 'A' | 'a' ) ( 'R' | 'r' ) ;

OF : ( 'O' | 'o' ) ( 'F' | 'f' ) ;

ADD : ( 'A' | 'a' ) ( 'D' | 'd' ) ( 'D' | 'd' ) ;

oC_SymbolicName
            :  UnescapedSymbolicName
                | EscapedSymbolicName
                | HexLetter
                | COUNT
                | FILTER
                | EXTRACT
                | ANY
                | NONE
                | SINGLE
                ;

UnescapedSymbolicName
                     :  IdentifierStart ( IdentifierPart )* ;

/**
 * Based on the unicode identifier and pattern syntax
 *   (http://www.unicode.org/reports/tr31/)
 * And extended with a few characters.
 */
IdentifierStart
               :  ID_Start
                   | Pc
                   ;

/**
 * Based on the unicode identifier and pattern syntax
 *   (http://www.unicode.org/reports/tr31/)
 * And extended with a few characters.
 */
IdentifierPart
              :  ID_Continue
                  | Sc
                  ;

/**
 * Any character except "`", enclosed within `backticks`. Backticks are escaped with double backticks.
 */
EscapedSymbolicName
                   :  ( '`' ( EscapedSymbolicName_0 )* '`' )+ ;

SP
  :  ( WHITESPACE )+ ;

WHITESPACE
          :  SPACE
              | TAB
              | LF
              | VT
              | FF
              | CR
              | FS
              | GS
              | RS
              | US
              | '\u1680'
              | '\u180e'
              | '\u2000'
              | '\u2001'
              | '\u2002'
              | '\u2003'
              | '\u2004'
              | '\u2005'
              | '\u2006'
              | '\u2008'
              | '\u2009'
              | '\u200a'
              | '\u2028'
              | '\u2029'
              | '\u205f'
              | '\u3000'
              | '\u00a0'
              | '\u2007'
              | '\u202f'
              | Comment
              ;

Comment
       :  ( '/*' ( Comment_1 | ( '*' Comment_2 ) )* '*/' )
           | ( '//' ( Comment_3 )* CR? ( LF | EOF ) )
           ;

oC_LeftArrowHead
             :  '<'
                 | '\u27e8'
                 | '\u3008'
                 | '\ufe64'
                 | '\uff1c'
                 ;

oC_RightArrowHead
              :  '>'
                  | '\u27e9'
                  | '\u3009'
                  | '\ufe65'
                  | '\uff1e'
                  ;

oC_Dash
    :  '-'
        | '\u00ad'
        | '\u2010'
        | '\u2011'
        | '\u2012'
        | '\u2013'
        | '\u2014'
        | '\u2015'
        | '\u2212'
        | '\ufe58'
        | '\ufe63'
        | '\uff0d'
        ;

fragment FF : [\f] ;

fragment EscapedSymbolicName_0 : ~[`] ;

fragment RS : [\u001E] ;

fragment ID_Continue : [\p{ID_Continue}] ;

fragment Comment_1 : ~[*] ;

fragment StringLiteral_1 : ~['\\] ;

fragment Comment_3 : ~[\n\r] ;

fragment Comment_2 : ~[/] ;

fragment GS : [\u001D] ;

fragment FS : [\u001C] ;

fragment CR : [\r] ;

fragment Sc : [\p{Sc}] ;

fragment SPACE : [ ] ;

fragment Pc : [\p{Pc}] ;

fragment TAB : [\t] ;

fragment StringLiteral_0 : ~["\\] ;

fragment LF : [\n] ;

fragment VT : [\u000B] ;

fragment US : [\u001F] ;

fragment ID_Start : [\p{ID_Start}] ;
