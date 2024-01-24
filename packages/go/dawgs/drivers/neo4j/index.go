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
	"github.com/specterops/bloodhound/log"
	"strings"

	"github.com/specterops/bloodhound/dawgs/graph"
)

const (
	nativeBTreeIndexProvider  = "native-btree-1.0"
	nativeLuceneIndexProvider = "lucene+native-3.0"

	dropPropertyIndexStatement        = "drop index $name;"
	dropPropertyConstraintStatement   = "drop constraint $name;"
	createPropertyIndexStatement      = "call db.createIndex($name, $labels, $properties, $provider);"
	createPropertyConstraintStatement = "call db.createUniquePropertyConstraint($name, $labels, $properties, $provider);"
)

type neo4jIndex struct {
	graph.Index

	kind graph.Kind
}

type neo4jConstraint struct {
	graph.Constraint

	kind graph.Kind
}

type neo4jSchema struct {
	Indexes     map[string]neo4jIndex
	Constraints map[string]neo4jConstraint
}

func newNeo4jSchema() neo4jSchema {
	return neo4jSchema{
		Indexes:     map[string]neo4jIndex{},
		Constraints: map[string]neo4jConstraint{},
	}
}

func toNeo4jSchema(dbSchema graph.Schema) neo4jSchema {
	neo4jSchemaInst := newNeo4jSchema()

	for _, graphSchema := range dbSchema.Graphs {
		for _, index := range graphSchema.NodeIndexes {
			for _, kind := range graphSchema.Nodes {
				indexName := strings.ToLower(kind.String()) + "_" + strings.ToLower(index.Field) + "_index"

				neo4jSchemaInst.Indexes[indexName] = neo4jIndex{
					Index: graph.Index{
						Name:  indexName,
						Field: index.Field,
						Type:  index.Type,
					},
					kind: kind,
				}
			}
		}

		for _, constraint := range graphSchema.NodeConstraints {
			for _, kind := range graphSchema.Nodes {
				constraintName := strings.ToLower(kind.String()) + "_" + strings.ToLower(constraint.Field) + "_constraint"

				neo4jSchemaInst.Constraints[constraintName] = neo4jConstraint{
					Constraint: graph.Constraint{
						Name:  constraintName,
						Field: constraint.Field,
						Type:  constraint.Type,
					},
					kind: kind,
				}
			}
		}
	}

	return neo4jSchemaInst
}

func parseProviderType(provider string) graph.IndexType {
	switch provider {
	case nativeBTreeIndexProvider:
		return graph.BTreeIndex
	case nativeLuceneIndexProvider:
		return graph.TextSearchIndex
	default:
		return graph.UnsupportedIndex
	}
}

func indexTypeProvider(indexType graph.IndexType) string {
	switch indexType {
	case graph.BTreeIndex:
		return nativeBTreeIndexProvider
	case graph.TextSearchIndex:
		return nativeLuceneIndexProvider
	default:
		return ""
	}
}

func assertIndexes(ctx context.Context, db graph.Database, indexesToRemove []string, indexesToAdd map[string]neo4jIndex) error {
	if err := db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for _, indexToRemove := range indexesToRemove {
			log.Infof("Removing index %s", indexToRemove)

			result := tx.Raw(strings.Replace(dropPropertyIndexStatement, "$name", indexToRemove, 1), nil)
			result.Close()

			if err := result.Error(); err != nil {
				return err
			}
		}

		return nil
	}); err != nil {
		return err
	}

	return db.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for indexName, indexToAdd := range indexesToAdd {
			log.Infof("Adding index %s to labels %s on properties %s using %s", indexName, indexToAdd.kind.String(), indexToAdd.Field, indexTypeProvider(indexToAdd.Type))

			if err := db.Run(ctx, createPropertyIndexStatement, map[string]interface{}{
				"name":       indexName,
				"labels":     []string{indexToAdd.kind.String()},
				"properties": []string{indexToAdd.Field},
				"provider":   indexTypeProvider(indexToAdd.Type),
			}); err != nil {
				return err
			}
		}

		return nil
	})
}

func assertConstraints(ctx context.Context, db graph.Database, constraintsToRemove []string, constraintsToAdd map[string]neo4jConstraint) error {
	for _, constraintToRemove := range constraintsToRemove {
		if err := db.Run(ctx, strings.Replace(dropPropertyConstraintStatement, "$name", constraintToRemove, 1), nil); err != nil {
			return err
		}
	}

	for constraintName, constraintToAdd := range constraintsToAdd {
		if err := db.Run(ctx, createPropertyConstraintStatement, map[string]interface{}{
			"name":       constraintName,
			"labels":     []string{constraintToAdd.kind.String()},
			"properties": []string{constraintToAdd.Field},
			"provider":   indexTypeProvider(constraintToAdd.Type),
		}); err != nil {
			return err
		}
	}

	return nil
}

func fetchPresentSchema(ctx context.Context, db graph.Database) (neo4jSchema, error) {
	presentSchema := newNeo4jSchema()

	return presentSchema, db.ReadTransaction(ctx, func(tx graph.Transaction) error {
		if result := tx.Raw("call db.indexes() yield name, uniqueness, provider, labelsOrTypes, properties;", nil); result.Error() != nil {
			return result.Error()
		} else {
			defer result.Close()

			var (
				name       string
				uniqueness string
				provider   string
				labels     []string
				properties []string
			)

			for result.Next() {
				if err := result.Scan(&name, &uniqueness, &provider, &labels, &properties); err != nil {
					return err
				}

				// Need this for neo4j 4.4+ which creates a weird index by default
				if len(labels) == 0 {
					continue
				}

				if len(labels) > 1 || len(properties) > 1 {
					return fmt.Errorf("composite index types are currently not supported")
				}

				if uniqueness == "UNIQUE" {
					presentSchema.Constraints[name] = neo4jConstraint{
						Constraint: graph.Constraint{
							Name:  name,
							Field: properties[0],
							Type:  parseProviderType(provider),
						},
						kind: graph.StringKind(labels[0]),
					}
				} else {
					presentSchema.Indexes[name] = neo4jIndex{
						Index: graph.Index{
							Name:  name,
							Field: properties[0],
							Type:  parseProviderType(provider),
						},
						kind: graph.StringKind(labels[0]),
					}
				}
			}

			return result.Error()
		}
	})
}

func assertSchema(ctx context.Context, db graph.Database, required graph.Schema) error {
	requiredNeo4jSchema := toNeo4jSchema(required)

	if presentNeo4jSchema, err := fetchPresentSchema(ctx, db); err != nil {
		return err
	} else {
		var (
			indexesToRemove     []string
			constraintsToRemove []string
			indexesToAdd        = map[string]neo4jIndex{}
			constraintsToAdd    = map[string]neo4jConstraint{}
		)

		for presentIndexName := range presentNeo4jSchema.Indexes {
			if _, hasMatchingDefinition := requiredNeo4jSchema.Indexes[presentIndexName]; !hasMatchingDefinition {
				indexesToRemove = append(indexesToRemove, presentIndexName)
			}
		}

		for presentConstraintName := range presentNeo4jSchema.Constraints {
			if _, hasMatchingDefinition := requiredNeo4jSchema.Constraints[presentConstraintName]; !hasMatchingDefinition {
				constraintsToRemove = append(constraintsToRemove, presentConstraintName)
			}
		}

		for requiredIndexName, requiredIndex := range requiredNeo4jSchema.Indexes {
			if presentIndex, hasMatchingDefinition := presentNeo4jSchema.Indexes[requiredIndexName]; !hasMatchingDefinition {
				indexesToAdd[requiredIndexName] = requiredIndex
			} else if requiredIndex.Type != presentIndex.Type {
				indexesToRemove = append(indexesToRemove, requiredIndexName)
				indexesToAdd[requiredIndexName] = requiredIndex
			}
		}

		for requiredConstraintName, requiredConstraint := range requiredNeo4jSchema.Constraints {
			if presentConstraint, hasMatchingDefinition := presentNeo4jSchema.Constraints[requiredConstraintName]; !hasMatchingDefinition {
				constraintsToAdd[requiredConstraintName] = requiredConstraint
			} else if requiredConstraint.Type != presentConstraint.Type {
				constraintsToRemove = append(constraintsToRemove, requiredConstraintName)
				constraintsToAdd[requiredConstraintName] = requiredConstraint
			}
		}

		if err := assertConstraints(ctx, db, constraintsToRemove, constraintsToAdd); err != nil {
			return err
		}

		return assertIndexes(ctx, db, indexesToRemove, indexesToAdd)
	}
}
