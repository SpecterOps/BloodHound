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

package neo4j

import (
	"context"
	"fmt"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
)

const (
	nativeBTreeIndexProvider  = "native-btree-1.0"
	nativeLuceneIndexProvider = "lucene+native-3.0"

	createPropertyIndexStatement      = "CALL db.createIndex($name, $labels, $properties, $provider);"
	createPropertyConstraintStatement = "CALL db.createUniquePropertyConstraint($name, $labels, $properties, $provider)"
)

func parseProviderType(provider string) graph.IndexType {
	switch provider {
	case nativeBTreeIndexProvider:
		return graph.BTreeIndex
	case nativeLuceneIndexProvider:
		return graph.FullTextSearchIndex
	default:
		return graph.UnsupportedIndex
	}
}

func indexTypeProvider(indexType graph.IndexType) string {
	switch indexType {
	case graph.BTreeIndex:
		return nativeBTreeIndexProvider
	case graph.FullTextSearchIndex:
		return nativeLuceneIndexProvider
	default:
		return ""
	}
}

func AssertNodePropertyIndex(db graph.Database, kind graph.Kind, propertyName string, indexType graph.IndexType) error {
	return db.WriteTransaction(context.Background(), func(tx graph.Transaction) error {
		statement := strings.Builder{}

		if indexType != graph.BTreeIndex {
			statement.WriteString("create ")
			statement.WriteString(indexTypeProvider(indexType))
			statement.WriteString(" index ")
		} else {
			statement.WriteString("create index ")
		}

		statement.WriteString(strings.ToLower(kind.String()))
		statement.WriteString("_")
		statement.WriteString(strings.ToLower(propertyName))
		statement.WriteString("_")
		statement.WriteString(indexType.String())
		statement.WriteString(" if not exists for (n:")
		statement.WriteString(kind.String())
		statement.WriteString(") on (n.")
		statement.WriteString(propertyName)
		statement.WriteString(");")

		if result := tx.Run(statement.String(), nil); result.Error() != nil {
			return result.Error()
		} else {
			result.Close()
		}

		return nil
	})
}

func formatDropSchemaCypherStmts(indexSchemas map[string]graph.IndexSchema, constraintSchemas map[string]graph.ConstraintSchema) []string {
	var (
		cypherStatements []string
		builder          strings.Builder
	)

	for _, propertyIndexSchema := range indexSchemas {
		builder.WriteString("drop index ")
		builder.WriteString(propertyIndexSchema.Name)
		builder.WriteString(";")

		cypherStatements = append(cypherStatements, builder.String())
		builder.Reset()
	}

	for _, propertyConstraintSchema := range constraintSchemas {
		builder.WriteString("drop constraint ")
		builder.WriteString(propertyConstraintSchema.Name)
		builder.WriteString(";")

		cypherStatements = append(cypherStatements, builder.String())
		builder.Reset()
	}

	return cypherStatements
}

func assertAgainst(ctx context.Context, requiredSchema, existingSchema *graph.Schema, db graph.Database) error {
	var (
		createConstraints = func(requiredKindSchema *graph.KindSchema, constraints map[string]graph.ConstraintSchema) error {
			for property, constraintToCreate := range constraints {
				if err := db.Run(ctx, createPropertyConstraintStatement, map[string]interface{}{
					"name":       constraintToCreate.Name,
					"labels":     []string{requiredKindSchema.Name()},
					"properties": []string{property},
					"provider":   indexTypeProvider(constraintToCreate.IndexType),
				}); err != nil {
					return err
				}
			}

			return nil
		}

		createIndices = func(requiredKindSchema *graph.KindSchema, indices map[string]graph.IndexSchema) error {
			for property, indexToCreate := range indices {
				if err := db.Run(ctx, createPropertyIndexStatement, map[string]interface{}{
					"name":       indexToCreate.Name,
					"labels":     []string{requiredKindSchema.Name()},
					"properties": []string{property},
					"provider":   indexTypeProvider(indexToCreate.IndexType),
				}); err != nil {
					return err
				}
			}

			return nil
		}
	)

	for _, kindSchema := range existingSchema.Kinds {
		if requiredKindSchema, hasMatchingDefinition := requiredSchema.Kinds[kindSchema.Kind]; !hasMatchingDefinition {
			// Remove all schematic definitions for the kind since there's no matching requirement
			for _, dropStmt := range formatDropSchemaCypherStmts(kindSchema.PropertyIndices, kindSchema.PropertyConstraints) {
				if err := db.Run(ctx, dropStmt, nil); err != nil {
					return err
				}
			}
		} else {
			var (
				indicesToAdd        = map[string]graph.IndexSchema{}
				indicesToRemove     = map[string]graph.IndexSchema{}
				constraintsToAdd    = map[string]graph.ConstraintSchema{}
				constraintsToRemove = map[string]graph.ConstraintSchema{}
			)

			// Match existing schematics to the definitions first
			for property, indexSchema := range kindSchema.PropertyIndices {
				if requiredIndexSchema, hasMatchingDefinition := requiredKindSchema.PropertyIndices[property]; !hasMatchingDefinition {
					// If there's no matching index for this property defined, remove it from the database
					indicesToRemove[property] = indexSchema
				} else if !indexSchema.Equals(requiredIndexSchema) {
					// The existing index does not match the requirement properties, recreate it
					indicesToRemove[property] = indexSchema
					indicesToAdd[property] = requiredIndexSchema
				}
			}

			// Sweep required schematics to ensure that missing entries are created
			for property, requiredIndexSchema := range requiredKindSchema.PropertyIndices {
				if _, hasMatchingDefinition := kindSchema.PropertyIndices[property]; !hasMatchingDefinition {
					// If there's no matching index for this property defined, create it
					indicesToAdd[property] = requiredIndexSchema
				}
			}

			for property, constraintSchema := range kindSchema.PropertyConstraints {
				if requiredConstraintSchema, hasMatchingDefinition := requiredKindSchema.PropertyConstraints[property]; !hasMatchingDefinition {
					// If there's no matching constraint for this property defined, remove it from the database
					constraintsToRemove[property] = constraintSchema
				} else if !constraintSchema.Equals(requiredConstraintSchema) {
					// The existing constraint does not match the requirement properties, recreate it
					constraintsToRemove[property] = constraintSchema
					constraintsToAdd[property] = requiredConstraintSchema
				}
			}

			for property, constraintSchema := range requiredKindSchema.PropertyConstraints {
				if _, hasMatchingDefinition := kindSchema.PropertyConstraints[property]; !hasMatchingDefinition {
					// If there's no matching constraint for this property defined, create it
					constraintsToAdd[property] = constraintSchema
				}
			}

			// Drop all indices and constraints first
			for _, dropStmt := range formatDropSchemaCypherStmts(indicesToRemove, constraintsToRemove) {
				if err := db.Run(ctx, dropStmt, nil); err != nil {
					return err
				}
			}

			if err := createIndices(requiredKindSchema, indicesToAdd); err != nil {
				return err
			}

			if err := createConstraints(requiredKindSchema, constraintsToAdd); err != nil {
				return err
			}
		}
	}

	for _, requiredKindSchema := range requiredSchema.Kinds {
		if _, hasMatchingDefinition := existingSchema.Kinds[requiredKindSchema.Kind]; !hasMatchingDefinition {
			// There's no matching definitions for indices or constraints for the required kind. Create them.
			if err := createIndices(requiredKindSchema, requiredKindSchema.PropertyIndices); err != nil {
				return err
			}

			if err := createConstraints(requiredKindSchema, requiredKindSchema.PropertyConstraints); err != nil {
				return err
			}
		}
	}

	return nil
}

func AssertSchema(ctx context.Context, db graph.Database, desiredSchema *graph.Schema) error {
	if existingSchema, err := db.FetchSchema(ctx); err != nil {
		return fmt.Errorf("could not load schema: %w", err)
	} else {
		return assertAgainst(ctx, desiredSchema, existingSchema, db)
	}
}
