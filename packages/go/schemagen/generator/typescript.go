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

package generator

import (
	"fmt"

	"github.com/specterops/bloodhound/schemagen/model"
	"github.com/specterops/bloodhound/schemagen/tsgen"
)

func generateToDisplaySwitch(root tsgen.File, typeSymbol string, values []model.StringEnum) {
	root.Export().Function().ID(typeSymbol + "ToDisplay").
		Parameters(
			tsgen.Parameter("value", typeSymbol),
		).
		ID("string").ID("|").ID("undefined").
		Block(func(cursor tsgen.Cursor) {
			cursor.Switch(tsgen.ID("value")).Block(func(cursor tsgen.Cursor) {
				for _, value := range values {
					cursor.Case(tsgen.Qualified(typeSymbol, value.Symbol)).Block(func(cursor tsgen.Cursor) {
						cursor.Returns(tsgen.Literal(value.GetName()))
					})
				}
				cursor.Default(tsgen.EmptyHandler)
				cursor.Returns(tsgen.ID("undefined"))
			})
		})
}

func GenerateTypeScriptStringEnum(root tsgen.File, typeSymbol string, values []model.StringEnum) {
	if len(values) == 0 {
		return
	}

	root.Export().Enum().ID(typeSymbol).Defs(func(cursor tsgen.Cursor) {
		for _, value := range values {
			cursor.ID(value.Symbol).OP("=").Literal(value.GetRepresentation())
		}
	})

	generateToDisplaySwitch(root, typeSymbol, values)
}

func GenerateTypeScriptPathfindingEdgesFn(
	root tsgen.File,
	functionName string,
	typeSymbol string,
	values []model.StringEnum,
) {
	root.Export().Function().
		ID(functionName).
		Parameters().
		ID(fmt.Sprintf("%s[]", typeSymbol)).
		Block(func(cursor tsgen.Cursor) {
			var idArray []tsgen.CursorHandler

			for _, value := range values {
				idArray = append(idArray, tsgen.Qualified(typeSymbol, value.Symbol))
			}

			cursor.Returns(tsgen.List(idArray...))
		})
}

func GenerateTypeScriptCommon(root tsgen.File, schema model.Graph) {
	GenerateTypeScriptStringEnum(root, "CommonNodeKind", schema.NodeKinds)
	GenerateTypeScriptStringEnum(root, "CommonRelationshipKind", schema.RelationshipKinds)
	GenerateTypeScriptStringEnum(root, "CommonKind", append(schema.NodeKinds, schema.RelationshipKinds...))
	GenerateTypeScriptStringEnum(root, "CommonKindProperties", schema.Properties)
}

func GenerateTypeScriptActiveDirectory(root tsgen.File, schema model.ActiveDirectory) {
	GenerateTypeScriptStringEnum(root, "ActiveDirectoryNodeKind", schema.NodeKinds)
	GenerateTypeScriptStringEnum(root, "ActiveDirectoryRelationshipKind", schema.RelationshipKinds)
	GenerateTypeScriptStringEnum(root, "ActiveDirectoryKind", append(schema.NodeKinds, schema.RelationshipKinds...))
	GenerateTypeScriptStringEnum(root, "ActiveDirectoryKindProperties", schema.Properties)

	GenerateTypeScriptPathfindingEdgesFn(root, "ActiveDirectoryPathfindingEdges", "ActiveDirectoryRelationshipKind", schema.PathfindingRelationships)
}

func GenerateTypeScriptAzure(root tsgen.File, schema model.Azure) {
	GenerateTypeScriptStringEnum(root, "AzureNodeKind", schema.NodeKinds)
	GenerateTypeScriptStringEnum(root, "AzureRelationshipKind", schema.RelationshipKinds)
	GenerateTypeScriptStringEnum(root, "AzureKind", append(schema.NodeKinds, schema.RelationshipKinds...))
	GenerateTypeScriptStringEnum(root, "AzureKindProperties", schema.Properties)

	GenerateTypeScriptPathfindingEdgesFn(root, "AzurePathfindingEdges", "AzureRelationshipKind", schema.PathfindingRelationships)
}
