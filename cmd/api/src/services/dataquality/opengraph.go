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

package dataquality

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/database/types/null"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

type kindCountsByEnvironment map[string]map[string]int

type openGraphCountResults struct {
	nodeKindCountsByEnvironmentID         kindCountsByEnvironment
	relationshipKindCountsByEnvironmentID kindCountsByEnvironment
	countedKindNames                      []string
}

func openGraphStats(ctx context.Context, db database.Database, graphDB graph.Database) (model.DataQualityStats, model.DataQualityAggregations, error) {
	var (
		aggregations = make(model.DataQualityAggregations, 0)
		runID        string
		stats        = make(model.DataQualityStats, 0)

		// We only are concerned with non-builtin types since AD and AZ have their own stat collection
		// When we move AD and AZ stat collection to this common table, this will need to be changed
		builtInFilter = model.Filters{"is_builtin": []model.Filter{{Operator: model.Equals, Value: "false"}}}
	)

	if newUUID, err := uuid.NewV4(); err != nil {
		return stats, aggregations, fmt.Errorf("could not generate new UUID: %w", err)
	} else {
		runID = newUUID.String()
	}

	extensions, _, err := db.GetGraphSchemaExtensions(ctx, builtInFilter, model.Sort{}, 0, 0)
	if err != nil {
		return stats, aggregations, fmt.Errorf("could not get graph schema extensions: %w", err)
	}

	for _, extension := range extensions {
		environments, err := db.GetEnvironmentsByExtensionId(ctx, extension.ID)
		if err != nil {
			return stats, aggregations, fmt.Errorf("could not get environments for extension %d: %w", extension.ID, err)
		}

		if len(environments) == 0 {
			continue
		}

		relationshipKinds, err := getOpenGraphRelationshipKinds(ctx, db, extension)
		if err != nil {
			return stats, aggregations, err
		}

		// Each extension may define multiple environments with their own environment kind, principal kinds, and source kind.
		for _, environment := range environments {
			environmentSourceKind, err := getOpenGraphEnvironmentSourceKind(ctx, db, environment)
			if err != nil {
				return stats, aggregations, err
			}

			countResults, err := countOpenGraphMetricsByEnvironment(ctx, graphDB, graph.StringKind(environment.EnvironmentKindName), environmentSourceKind.ToKind(), relationshipKinds)
			if err != nil {
				return stats, aggregations, fmt.Errorf("could not count open graph nodes and relationships for environment kind %s: %w", environment.EnvironmentKindName, err)
			}

			kindIDs, err := getOpenGraphCountKindIDs(ctx, db, countResults.countedKindNames)
			if err != nil {
				return stats, aggregations, err
			}

			for _, metricCountGroup := range []struct {
				metricType            model.DataQualityMetricType
				countsByEnvironmentID kindCountsByEnvironment
			}{
				{
					metricType:            model.DataQualityMetricTypeNode,
					countsByEnvironmentID: countResults.nodeKindCountsByEnvironmentID,
				},
				{
					metricType:            model.DataQualityMetricTypeRelationship,
					countsByEnvironmentID: countResults.relationshipKindCountsByEnvironmentID,
				},
			} {
				metricStats, metricAggregations := buildOpenGraphCountStats(runID, extension, environment, metricCountGroup.metricType, metricCountGroup.countsByEnvironmentID, kindIDs)

				stats = append(stats, metricStats...)
				aggregations = append(aggregations, metricAggregations...)
			}
		}
	}

	return stats, aggregations, nil
}

func getOpenGraphRelationshipKinds(ctx context.Context, db database.Database, extension model.GraphSchemaExtension) (graph.Kinds, error) {
	relationshipKindFilters := model.Filters{"schema_extension_id": []model.Filter{{
		Operator:    model.Equals,
		Value:       fmt.Sprintf("%d", extension.ID),
		SetOperator: model.FilterAnd,
	}}}

	schemaRelationshipKinds, _, err := db.GetGraphSchemaRelationshipKinds(ctx, relationshipKindFilters, model.Sort{}, 0, 0)
	if err != nil {
		return nil, fmt.Errorf("could not get relationship kinds for extension %d: %w", extension.ID, err)
	}

	relationshipKinds := make(graph.Kinds, 0, len(schemaRelationshipKinds))
	for _, schemaRelationshipKind := range schemaRelationshipKinds {
		relationshipKinds = append(relationshipKinds, schemaRelationshipKind.ToKind())
	}

	return relationshipKinds, nil
}

func getOpenGraphEnvironmentSourceKind(ctx context.Context, db database.Database, environment model.SchemaEnvironment) (model.Kind, error) {
	kinds, err := db.GetKindsByIDs(ctx, environment.SourceKindId)
	if err != nil {
		return model.Kind{}, fmt.Errorf("could not get source kind for schema environment %d: %w", environment.ID, err)
	}

	if len(kinds) == 0 {
		return model.Kind{}, fmt.Errorf("could not get source kind for schema environment %d: %w", environment.ID, database.ErrNotFound)
	}

	return kinds[0], nil
}

func countOpenGraphMetricsByEnvironment(ctx context.Context, graphDB graph.Database, environmentKind graph.Kind, sourceKind graph.Kind, relationshipKinds graph.Kinds) (openGraphCountResults, error) {
	var (
		countResults = openGraphCountResults{
			nodeKindCountsByEnvironmentID:         kindCountsByEnvironment{},
			relationshipKindCountsByEnvironmentID: kindCountsByEnvironment{},
			countedKindNames:                      make([]string, 0),
		}
		seenCountedKindNames = map[string]struct{}{}
	)

	err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		environmentIDs, err := getOpenGraphEnvironmentIDs(ctx, tx, environmentKind)
		if err != nil {
			return err
		}

		for _, environmentID := range environmentIDs {
			if err := countOpenGraphNodeMetricsForEnvironment(tx, &countResults, sourceKind, environmentID, seenCountedKindNames); err != nil {
				return err
			}

			if err := countOpenGraphRelationshipMetricsForEnvironment(tx, &countResults, sourceKind, relationshipKinds, environmentID, seenCountedKindNames); err != nil {
				return err
			}
		}

		return nil
	})

	return countResults, err
}

func countOpenGraphNodeMetricsForEnvironment(tx graph.Transaction, countResults *openGraphCountResults, sourceKind graph.Kind, environmentID string, seenCountedKindNames map[string]struct{}) error {
	nodeKindCounts := map[string]int{}

	if err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), sourceKind),
			query.Equals(query.NodeProperty(graphschema.EnvironmentIDKey), environmentID),
		)
	}).FetchKinds(func(cursor graph.Cursor[graph.KindsResult]) error {
		for result := range cursor.Chan() {
			countResults.countedKindNames = countOpenGraphNodeKinds(result.Kinds, sourceKind, nodeKindCounts, countResults.countedKindNames, seenCountedKindNames)
		}

		return cursor.Error()
	}); err != nil {
		return err
	}

	if len(nodeKindCounts) > 0 {
		countResults.nodeKindCountsByEnvironmentID[environmentID] = nodeKindCounts
	}

	return nil
}

func countOpenGraphRelationshipMetricsForEnvironment(tx graph.Transaction, countResults *openGraphCountResults, sourceKind graph.Kind, relationshipKinds graph.Kinds, environmentID string, seenCountedKindNames map[string]struct{}) error {
	if len(relationshipKinds) == 0 {
		return nil
	}

	relationshipKindCounts := map[string]int{}

	if err := tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), sourceKind),
			query.Equals(query.StartProperty(graphschema.EnvironmentIDKey), environmentID),
			query.KindIn(query.Relationship(), relationshipKinds...),
		)
	}).FetchKinds(func(cursor graph.Cursor[graph.RelationshipKindsResult]) error {
		for result := range cursor.Chan() {
			countResults.countedKindNames = countOpenGraphKind(result.Kind.String(), relationshipKindCounts, countResults.countedKindNames, seenCountedKindNames)
		}

		return cursor.Error()
	}); err != nil {
		return err
	}

	if len(relationshipKindCounts) > 0 {
		countResults.relationshipKindCountsByEnvironmentID[environmentID] = relationshipKindCounts
	}

	return nil
}

func countOpenGraphNodeKinds(kinds graph.Kinds, sourceKind graph.Kind, nodeKindCounts map[string]int, countedKindNames []string, seenCountedKindNames map[string]struct{}) []string {
	for _, kind := range kinds {
		if kind.Is(sourceKind) || model.IsExtendedNodeKind(kind) {
			continue
		}

		countedKindNames = countOpenGraphKind(kind.String(), nodeKindCounts, countedKindNames, seenCountedKindNames)
	}

	return countedKindNames
}

func countOpenGraphKind(kindName string, kindCounts map[string]int, countedKindNames []string, seenCountedKindNames map[string]struct{}) []string {
	kindCounts[kindName]++

	if _, seen := seenCountedKindNames[kindName]; !seen {
		countedKindNames = append(countedKindNames, kindName)
		seenCountedKindNames[kindName] = struct{}{}
	}

	return countedKindNames
}

func getOpenGraphEnvironmentIDs(ctx context.Context, tx graph.Transaction, environmentKind graph.Kind) ([]string, error) {
	var environmentIDs []string

	environments, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.Kind(query.Node(), environmentKind)
	}))
	if err != nil {
		return nil, err
	}

	for _, environment := range environments {
		environmentID, err := environment.Properties.Get(graphschema.EnvironmentIDKey).String()
		if err != nil {
			slog.WarnContext(
				ctx,
				"OpenGraph environment node does not have a valid environment ID property",
				slog.Uint64("node_id", uint64(environment.ID)),
				slog.String("environment_kind", environmentKind.String()),
				attr.Error(err),
			)
			continue
		}

		environmentIDs = append(environmentIDs, environmentID)
	}

	return environmentIDs, nil
}

func getOpenGraphCountKindIDs(ctx context.Context, db database.Database, kindNames []string) (map[string]null.Int32, error) {
	if len(kindNames) == 0 {
		return map[string]null.Int32{}, nil
	}

	kinds, err := db.GetKindsByNames(ctx, kindNames...)
	if err != nil && !errors.Is(err, database.ErrNotFound) {
		return nil, fmt.Errorf("could not get data quality metric kinds: %w", err)
	}

	kindIDsByName := make(map[string]null.Int32, len(kinds))
	for _, kind := range kinds {
		kindIDsByName[kind.Name] = null.Int32From(kind.ID)
	}

	return kindIDsByName, nil
}

func buildOpenGraphCountStats(runID string, extension model.GraphSchemaExtension, environment model.SchemaEnvironment, metricType model.DataQualityMetricType, countsByEnvironmentID kindCountsByEnvironment, kindIDsByName map[string]null.Int32) (model.DataQualityStats, model.DataQualityAggregations) {
	var (
		aggregatedCounts = map[string]int{}
		stats            = make(model.DataQualityStats, 0)
		aggregations     = make(model.DataQualityAggregations, 0)
	)

	for environmentID, kindCounts := range countsByEnvironmentID {
		for kindName, count := range kindCounts {
			stats = append(stats, model.DataQualityStat{
				RunID:                   runID,
				SchemaExtensionID:       extension.ID,
				SchemaEnvironmentKindID: environment.EnvironmentKindId,
				EnvironmentID:           environmentID,
				MetricType:              metricType,
				MetricName:              kindName,
				MetricValue:             float64(count),
				KindID:                  kindIDsByName[kindName],
			})

			aggregatedCounts[kindName] += count
		}
	}

	for kindName, count := range aggregatedCounts {
		aggregations = append(aggregations, model.DataQualityAggregation{
			RunID:                   runID,
			SchemaExtensionID:       extension.ID,
			SchemaEnvironmentKindID: environment.EnvironmentKindId,
			MetricType:              metricType,
			MetricName:              kindName,
			MetricValue:             float64(count),
			KindID:                  kindIDsByName[kindName],
		})
	}

	return stats, aggregations
}
