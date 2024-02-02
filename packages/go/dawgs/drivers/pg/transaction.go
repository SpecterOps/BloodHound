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

package pg

import (
	"bytes"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/specterops/bloodhound/cypher/backend/pgsql"
	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/dawgs/drivers/pg/model"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/query"
	"github.com/specterops/bloodhound/dawgs/util/size"
)

type driver interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type inspectingDriver struct {
	upstreamDriver driver
}

func (s inspectingDriver) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	inspector().Inspect(sql, arguments)
	return s.upstreamDriver.Exec(ctx, sql, arguments...)
}

func (s inspectingDriver) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	inspector().Inspect(sql, arguments)
	return s.upstreamDriver.Query(ctx, sql, arguments...)
}

func (s inspectingDriver) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	inspector().Inspect(sql, arguments)
	return s.upstreamDriver.QueryRow(ctx, sql, arguments...)
}

type transaction struct {
	schemaManager      *SchemaManager
	queryExecMode      pgx.QueryExecMode
	queryResultsFormat pgx.QueryResultFormats
	ctx                context.Context
	conn               *pgxpool.Conn
	tx                 pgx.Tx
	targetSchema       graph.Graph
	targetSchemaSet    bool
}

func newTransaction(ctx context.Context, conn *pgxpool.Conn, schemaManager *SchemaManager, cfg *Config) (*transaction, error) {
	if pgxTx, err := conn.BeginTx(ctx, cfg.Options); err != nil {
		return nil, err
	} else {
		return &transaction{
			schemaManager:      schemaManager,
			queryExecMode:      cfg.QueryExecMode,
			queryResultsFormat: cfg.QueryResultFormats,
			ctx:                ctx,
			conn:               conn,
			tx:                 pgxTx,
			targetSchemaSet:    false,
		}, nil
	}
}

func (s *transaction) driver() driver {
	if s.tx != nil {
		return inspectingDriver{
			upstreamDriver: s.tx,
		}
	}

	return inspectingDriver{
		upstreamDriver: s.conn,
	}
}

func (s *transaction) TraversalMemoryLimit() size.Size {
	return size.Gibibyte
}

func (s *transaction) WithGraph(schema graph.Graph) graph.Transaction {
	s.targetSchema = schema
	s.targetSchemaSet = true

	return s
}

func (s *transaction) Close() {
	if s.tx != nil {
		s.tx.Rollback(s.ctx)
		s.tx = nil
	}
}

func (s *transaction) getTargetGraph() (model.Graph, error) {
	if !s.targetSchemaSet {
		// Look for a default graph target
		if defaultGraph, hasDefaultGraph := s.schemaManager.DefaultGraph(); !hasDefaultGraph {
			return model.Graph{}, fmt.Errorf("driver operation requires a graph target to be set")
		} else {
			return defaultGraph, nil
		}
	}

	return s.schemaManager.AssertGraph(s, s.targetSchema)
}

func (s *transaction) CreateNode(properties *graph.Properties, kinds ...graph.Kind) (*graph.Node, error) {
	if graphTarget, err := s.getTargetGraph(); err != nil {
		return nil, err
	} else if kindIDSlice, err := s.schemaManager.AssertKinds(s, kinds); err != nil {
		return nil, err
	} else if propertiesJSONB, err := pgsql.PropertiesToJSONB(properties); err != nil {
		return nil, err
	} else {
		var (
			nodeID int32
			result = s.queryRow(createNodeStatement, s.queryExecMode, graphTarget.ID, kindIDSlice, propertiesJSONB)
		)

		if err := result.Scan(&nodeID); err != nil {
			return nil, err
		}

		return graph.NewNode(graph.ID(nodeID), properties, kinds...), nil
	}
}

func (s *transaction) UpdateNode(node *graph.Node) error {
	var (
		properties       = node.Properties
		updateStatements []graph.Criteria
	)

	if addedKinds := node.AddedKinds; len(addedKinds) > 0 {
		updateStatements = append(updateStatements, query.AddKinds(query.Node(), addedKinds))
	}

	if deletedKinds := node.DeletedKinds; len(deletedKinds) > 0 {
		updateStatements = append(updateStatements, query.DeleteKinds(query.Node(), deletedKinds))
	}

	if modifiedProperties := properties.ModifiedProperties(); len(modifiedProperties) > 0 {
		updateStatements = append(updateStatements, query.SetProperties(query.Node(), modifiedProperties))
	}

	if deletedProperties := properties.DeletedProperties(); len(deletedProperties) > 0 {
		updateStatements = append(updateStatements, query.DeleteProperties(query.Node(), deletedProperties...))
	}

	return s.Nodes().Filter(query.Equals(query.NodeID(), node.ID)).Query(func(results graph.Result) error {
		// We don't need to exhaust the result set as the defered close with discard it for us
		return results.Error()
	}, updateStatements...)
}

func (s *transaction) Nodes() graph.NodeQuery {
	return &nodeQuery{
		liveQuery: newLiveQuery(s.ctx, s, s.schemaManager),
	}
}

func (s *transaction) CreateRelationshipByIDs(startNodeID, endNodeID graph.ID, kind graph.Kind, properties *graph.Properties) (*graph.Relationship, error) {
	if graphTarget, err := s.getTargetGraph(); err != nil {
		return nil, err
	} else if kindIDSlice, err := s.schemaManager.AssertKinds(s, graph.Kinds{kind}); err != nil {
		return nil, err
	} else if propertiesJSONB, err := pgsql.PropertiesToJSONB(properties); err != nil {
		return nil, err
	} else {
		var (
			edgeID int32
			result = s.queryRow(createEdgeStatement, s.queryExecMode, graphTarget.ID, startNodeID, endNodeID, kindIDSlice[0], propertiesJSONB)
		)

		if err := result.Scan(&edgeID); err != nil {
			return nil, err
		}

		return graph.NewRelationship(graph.ID(edgeID), startNodeID, endNodeID, properties, kind), nil
	}
}

func (s *transaction) UpdateRelationship(relationship *graph.Relationship) error {
	var (
		modifiedProperties    = relationship.Properties.ModifiedProperties()
		deletedProperties     = relationship.Properties.DeletedProperties()
		numModifiedProperties = len(modifiedProperties)
		numDeletedProperties  = len(deletedProperties)

		statement string
		arguments []any
	)

	if numModifiedProperties > 0 {
		if jsonbArgument, err := pgsql.ValueToJSONB(modifiedProperties); err != nil {
			return err
		} else {
			arguments = append(arguments, jsonbArgument)
		}

		if numDeletedProperties > 0 {
			if textArrayArgument, err := pgsql.StringSliceToTextArray(deletedProperties); err != nil {
				return err
			} else {
				arguments = append(arguments, textArrayArgument)
			}

			statement = edgePropertySetAndDeleteStatement
		} else {
			statement = edgePropertySetOnlyStatement
		}
	} else if numDeletedProperties > 0 {
		if textArrayArgument, err := pgsql.StringSliceToTextArray(deletedProperties); err != nil {
			return err
		} else {
			arguments = append(arguments, textArrayArgument)
		}

		statement = edgePropertyDeleteOnlyStatement
	}

	_, err := s.driver().Exec(s.ctx, statement, append(arguments, relationship.ID)...)
	return err
}

func (s *transaction) Relationships() graph.RelationshipQuery {
	return &relationshipQuery{
		liveQuery: newLiveQuery(s.ctx, s, s.schemaManager),
	}
}

func (s *transaction) queryRow(query string, parameters ...any) pgx.Row {
	queryArgs := []any{s.queryExecMode, s.queryResultsFormat}
	queryArgs = append(queryArgs, parameters...)

	return s.driver().QueryRow(s.ctx, query, queryArgs...)
}

func (s *transaction) query(query string, parameters map[string]any) (pgx.Rows, error) {
	queryArgs := []any{s.queryExecMode, s.queryResultsFormat}

	if parameters != nil || len(parameters) > 0 {
		queryArgs = append(queryArgs, pgx.NamedArgs(parameters))
	}

	return s.driver().Query(s.ctx, query, queryArgs...)
}

func (s *transaction) Query(query string, parameters map[string]any) graph.Result {
	if parsedQuery, err := frontend.ParseCypher(frontend.NewContext(), query); err != nil {
		return graph.NewErrorResult(err)
	} else if translatedParams, err := pgsql.Translate(parsedQuery, s.schemaManager); err != nil {
		return graph.NewErrorResult(err)
	} else {
		var (
			buffer  = &bytes.Buffer{}
			emitter = pgsql.NewEmitter(false, s.schemaManager)
		)

		for key, value := range parameters {
			if _, hasKey := translatedParams[key]; hasKey {
				return graph.NewErrorResult(fmt.Errorf("Query specifies a parameter value that is overwritten by translation: %s", key))
			}

			translatedParams[key] = value
		}

		if err := emitter.Write(parsedQuery, buffer); err != nil {
			return graph.NewErrorResult(err)
		}

		return s.Raw(buffer.String(), parameters)
	}
}

func (s *transaction) Raw(query string, parameters map[string]any) graph.Result {
	if rows, err := s.query(query, parameters); err != nil {
		return graph.NewErrorResult(err)
	} else {
		return &queryResult{
			rows:       rows,
			kindMapper: s.schemaManager,
		}
	}
}

func (s *transaction) Commit() error {
	if s.tx != nil {
		return s.tx.Commit(s.ctx)
	}

	return nil
}
