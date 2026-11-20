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
	adAnalysis "github.com/specterops/bloodhound/packages/go/analysis/ad"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

type metricCounts map[string]int

type envPropertyKeys struct {
	environmentNodeID       string
	memberNodeEnvironmentID string
}

type countResults struct {
	environmentID      string
	countsByMetricName metricCounts
	metricType         string
}

func dataQualityGraphStats(ctx context.Context, db database.Database, graphDB graph.Database) (model.DataQualityStats, model.DataQualityAggregations, error) {
	var (
		runID string

		stats        = make(model.DataQualityStats, 0)
		aggregations = make(model.DataQualityAggregations, 0)
	)

	if newUUID, err := uuid.NewV4(); err != nil {
		return stats, aggregations, fmt.Errorf("could not generate new UUID: %w", err)
	} else {
		runID = newUUID.String()
	}

	extensions, _, err := db.GetGraphSchemaExtensions(ctx, model.Filters{}, model.Sort{}, 0, 0)
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

		relationshipKinds, err := getRelationshipKinds(ctx, db, extension)
		if err != nil {
			return stats, aggregations, err
		}

		// Each extension may define multiple environments with their own environment kind, principal kinds, and source kind.
		for _, environment := range environments {
			environmentSourceKind, err := getEnvironmentSourceKind(ctx, db, environment)
			if err != nil {
				return stats, aggregations, err
			}

			countMetrics, seenKinds, err := countMetricsPerEnvironment(ctx, graphDB, extension, graph.StringKind(environment.EnvironmentKindName), environmentSourceKind.ToKind(), relationshipKinds)
			if err != nil {
				return stats, aggregations, fmt.Errorf("could not count metrics for environment kind %s: %w", environment.EnvironmentKindName, err)
			}

			kindIDs, err := getCountKindIDs(ctx, db, seenKinds)
			if err != nil {
				return stats, aggregations, err
			}

			metricStats, metricAggregations := buildDataQualityStats(runID, extension, environment, countMetrics, kindIDs)

			stats = append(stats, metricStats...)
			aggregations = append(aggregations, metricAggregations...)
		}
	}

	return stats, aggregations, nil
}

func getRelationshipKinds(ctx context.Context, db database.Database, extension model.GraphSchemaExtension) (graph.Kinds, error) {
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

func getEnvironmentSourceKind(ctx context.Context, db database.Database, environment model.SchemaEnvironment) (model.Kind, error) {
	kinds, err := db.GetKindsByIDs(ctx, environment.SourceKindId)
	if err != nil {
		return model.Kind{}, fmt.Errorf("could not get source kind for schema environment %d: %w", environment.ID, err)
	}

	if len(kinds) == 0 {
		return model.Kind{}, fmt.Errorf("could not get source kind for schema environment %d: %w", environment.ID, database.ErrNotFound)
	}

	return kinds[0], nil
}

func countMetricsPerEnvironment(ctx context.Context, graphDB graph.Database, extension model.GraphSchemaExtension, environmentKind graph.Kind, sourceKind graph.Kind, relationshipKinds graph.Kinds) ([]countResults, map[string]struct{}, error) {
	var (
		results              []countResults
		seenCountedKindNames = map[string]struct{}{}
		propertyKeys         = dataQualityEnvProperties(environmentKind)
	)

	err := graphDB.ReadTransaction(ctx, func(tx graph.Transaction) error {
		environmentIDs, err := getEnvironmentIDs(ctx, tx, environmentKind, propertyKeys.environmentNodeID)
		if err != nil {
			return err
		}

		for _, environmentID := range environmentIDs {
			// Node metrics
			if nodeMetrics, err := countNodeMetricsForEnvironment(tx, sourceKind, environmentID, propertyKeys.memberNodeEnvironmentID, seenCountedKindNames); err != nil {
				return err
			} else {
				results = append(results, nodeMetrics)
			}

			// Relationship Metrics
			if relationshipMetrics, err := countRelationshipMetricsForEnvironment(tx, sourceKind, relationshipKinds, environmentID, propertyKeys.memberNodeEnvironmentID, seenCountedKindNames); err != nil {
				return err
			} else {
				results = append(results, relationshipMetrics)
			}

			// Custom metrics for extensions (Hardcoded for now but can be dynamic later)
			switch extension.Name {
			case "AD":
				if customMetrics, err := countADCustomMetricsForEnvironment(tx, extension, environmentID); err != nil {
					return err
				} else {
					results = append(results, customMetrics...)
				}
			}
		}

		return nil
	})

	return results, seenCountedKindNames, err
}

func countADCustomMetricsForEnvironment(tx graph.Transaction, extension model.GraphSchemaExtension, environmentID string) ([]countResults, error) {
	userSessionCompleteness, err := adAnalysis.FetchUserSessionCompleteness(tx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("could not get user session completeness for environment %s: %w", environmentID, err)
	}
	localGroupCompleteness, err := adAnalysis.FetchLocalGroupCompleteness(tx, environmentID)
	if err != nil {
		return nil, fmt.Errorf("could not get local group completeness for environment %s: %w", environmentID, err)
	}

	return []countResults{
		{
			environmentID: environmentID,
			metricType:    fmt.Sprintf("%s_%s", extension.Namespace, model.DataQualityMetricTypeSessionCompleteness),
			countsByMetricName: metricCounts{
				"active_user_count_with_sessions": int(userSessionCompleteness.ActiveUserCountWithSessions),
				"active_user_count":               int(userSessionCompleteness.ActiveUserCount),
			},
		},
		{
			environmentID: environmentID,
			metricType:    fmt.Sprintf("%s_%s", extension.Namespace, model.DataQualityMetricTypeLocalGroupCompleteness),
			countsByMetricName: metricCounts{
				"active_computer_count_with_admins": int(localGroupCompleteness.ActiveComputerCountWithAdmins),
				"active_computer_count":             int(localGroupCompleteness.ActiveComputerCount),
			},
		},
	}, nil
}

func countNodeMetricsForEnvironment(tx graph.Transaction, sourceKind graph.Kind, environmentID string, membershipProperty string, seenCountedKindNames map[string]struct{}) (countResults, error) {
	var (
		results = countResults{
			environmentID:      environmentID,
			metricType:         string(model.DataQualityMetricTypeNode),
			countsByMetricName: make(metricCounts),
		}
	)

	err := tx.Nodes().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Node(), sourceKind),
			query.Equals(query.NodeProperty(membershipProperty), environmentID),
		)
	}).FetchKinds(func(cursor graph.Cursor[graph.KindsResult]) error {
		for result := range cursor.Chan() {
			for _, kind := range result.Kinds {
				if kind.Is(sourceKind) || model.IsExtendedNodeKind(kind) {
					continue
				}

				results.countsByMetricName[kind.String()]++
				seenCountedKindNames[kind.String()] = struct{}{}
			}
		}

		return cursor.Error()
	})

	return results, err
}

func countRelationshipMetricsForEnvironment(tx graph.Transaction, sourceKind graph.Kind, relationshipKinds graph.Kinds, environmentID string, membershipProperty string, seenCountedKindNames map[string]struct{}) (countResults, error) {
	var (
		results = countResults{
			environmentID:      environmentID,
			metricType:         string(model.DataQualityMetricTypeRelationship),
			countsByMetricName: make(metricCounts),
		}
	)

	if len(relationshipKinds) == 0 {
		return results, nil
	}

	err := tx.Relationships().Filterf(func() graph.Criteria {
		return query.And(
			query.Kind(query.Start(), sourceKind),
			query.Equals(query.StartProperty(membershipProperty), environmentID),
			query.KindIn(query.Relationship(), relationshipKinds...),
		)
	}).FetchKinds(func(cursor graph.Cursor[graph.RelationshipKindsResult]) error {
		for result := range cursor.Chan() {
			results.countsByMetricName[result.Kind.String()]++
			seenCountedKindNames[result.Kind.String()] = struct{}{}
		}

		return cursor.Error()
	})

	return results, err
}

// AD and Azure identify environment nodes by objectid, but group their members
// by domainsid and tenantid. Custom environments use environmentid for both.
func dataQualityEnvProperties(environmentKind graph.Kind) envPropertyKeys {
	switch environmentKind {
	case ad.Domain:
		return envPropertyKeys{environmentNodeID: common.ObjectID.String(), memberNodeEnvironmentID: ad.DomainSID.String()}
	case azure.Tenant:
		return envPropertyKeys{environmentNodeID: common.ObjectID.String(), memberNodeEnvironmentID: azure.TenantID.String()}
	default:
		return envPropertyKeys{environmentNodeID: graphschema.EnvironmentIDKey, memberNodeEnvironmentID: graphschema.EnvironmentIDKey}
	}
}

func getEnvironmentIDs(ctx context.Context, tx graph.Transaction, environmentKind graph.Kind, environmentIDProperty string) ([]string, error) {
	var environmentIDs []string

	environments, err := ops.FetchNodes(tx.Nodes().Filterf(func() graph.Criteria {
		return query.Kind(query.Node(), environmentKind)
	}))
	if err != nil {
		return nil, err
	}

	for _, environment := range environments {
		environmentID, err := environment.Properties.Get(environmentIDProperty).String()
		if err != nil {
			slog.WarnContext(ctx, "Environment node does not have a valid ID property",
				slog.Uint64("node_id", uint64(environment.ID)),
				slog.String("environment_kind", environmentKind.String()),
				slog.String("id_property", environmentIDProperty),
				attr.Error(err),
			)
			continue
		}

		environmentIDs = append(environmentIDs, environmentID)
	}

	return environmentIDs, nil
}

func getCountKindIDs(ctx context.Context, db database.Database, seenKinds map[string]struct{}) (map[string]null.Int32, error) {
	if len(seenKinds) == 0 {
		return map[string]null.Int32{}, nil
	}

	var kindNames []string
	for kind := range seenKinds {
		kindNames = append(kindNames, kind)
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

func buildDataQualityStats(runID string, extension model.GraphSchemaExtension, environment model.SchemaEnvironment, countMetrics []countResults, kindIDsByName map[string]null.Int32) (model.DataQualityStats, model.DataQualityAggregations) {
	type metricKey struct {
		metricType model.DataQualityMetricType
		metricName string
	}

	var (
		aggregatedCounts = map[metricKey]int{}
		stats            = make(model.DataQualityStats, 0)
		aggregations     = make(model.DataQualityAggregations, 0)
	)

	for _, counts := range countMetrics {
		for metricName, count := range counts.countsByMetricName {
			stat := model.DataQualityStat{
				RunID:                   runID,
				SchemaExtensionID:       extension.ID,
				SchemaEnvironmentKindID: environment.EnvironmentKindId,
				EnvironmentID:           counts.environmentID,
				MetricType:              model.DataQualityMetricType(counts.metricType),
				MetricName:              metricName,
				MetricValue:             float64(count),
				KindID:                  kindIDsByName[metricName],
			}
			stats = append(stats, stat)

			aggregatedCounts[metricKey{metricType: stat.MetricType, metricName: stat.MetricName}] += count
		}
	}

	for metric, count := range aggregatedCounts {
		aggregations = append(aggregations, model.DataQualityAggregation{
			RunID:                   runID,
			SchemaExtensionID:       extension.ID,
			SchemaEnvironmentKindID: environment.EnvironmentKindId,
			MetricType:              metric.metricType,
			MetricName:              metric.metricName,
			MetricValue:             float64(count),
			KindID:                  kindIDsByName[metric.metricName],
		})
	}

	return stats, aggregations
}
