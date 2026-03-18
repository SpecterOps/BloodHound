// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package endpoint

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/bhlog/measure"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/specterops/dawgs/cypher/models/cypher"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

// CopyMatchExpressionsSorted creates a deep copy of the provided match expressions
// and returns them sorted by Key, then Operator. This ensures deterministic ordering
// for hashing and query generation.
func CopyMatchExpressionsSorted(matchExpressions []ein.MatchExpression) []ein.MatchExpression {
	sorted := make([]ein.MatchExpression, len(matchExpressions))
	copy(sorted, matchExpressions)

	slices.SortFunc(sorted, func(a, b ein.MatchExpression) int {
		if a.Key > b.Key {
			return 1
		}

		if a.Key < b.Key {
			return -1
		}

		if a.Operator > b.Operator {
			return 1
		}

		if a.Operator < b.Operator {
			return -1
		}

		return 0
	})

	return sorted
}

// caseInsensitiveEquals constructs a Cypher comparison that checks if the reference
// property equals the target value in a case-insensitive manner using the `toLower` function.
func caseInsensitiveEquals(reference, target graph.Criteria) *cypher.Comparison {
	return cypher.NewComparison(
		cypher.NewSimpleFunctionInvocation(cypher.ToLowerFunction, reference),
		cypher.OperatorEquals,
		target,
	)
}

// newMatchExpr converts a list of endpoint match expressions into a single composite
// database query criteria. It includes the node Kind filter if provided and combines
// individual property matches using logical AND. It supports Equals and EqualsIgnoreCase operators.
func newMatchExpr(identityKind graph.Kind, matchExpressions []ein.MatchExpression) (graph.Criteria, error) {
	var (
		sortedMatchExpressions = CopyMatchExpressionsSorted(matchExpressions)
		cypherExpressions      []graph.Criteria
	)

	if identityKind != nil && identityKind.String() != "" {
		cypherExpressions = make([]graph.Criteria, 0, len(sortedMatchExpressions)+1)
		cypherExpressions = append(cypherExpressions, query.Kind(query.Node(), identityKind))
	} else {
		cypherExpressions = make([]graph.Criteria, 0, len(sortedMatchExpressions))
	}

	for _, matchExpression := range sortedMatchExpressions {
		switch matchExpression.Operator {
		case ein.OperatorEquals:
			nextExpression := query.Equals(
				query.NodeProperty(matchExpression.Key),
				query.Parameter(matchExpression.Value),
			)

			cypherExpressions = append(cypherExpressions, nextExpression)

		case ein.OperatorEqualsIgnoreCase:
			if strValue, typeOK := matchExpression.Value.(string); !typeOK {
				return nil, fmt.Errorf("case insensitive equals requires value type of string but got %T", matchExpression.Value)
			} else {
				nextExpression := caseInsensitiveEquals(
					query.NodeProperty(matchExpression.Key),
					query.Parameter(strings.ToLower(strValue)),
				)

				cypherExpressions = append(cypherExpressions, nextExpression)
			}

		default:
			return nil, fmt.Errorf("unsupported match expression operator: %s", matchExpression.Operator)
		}
	}

	return query.And(cypherExpressions...), nil
}

// getNodeObjectID executes a database query to find exactly one node matching the given criteria
// and returns its "objectid" property. It validates that exactly one result exists; if zero or
// multiple nodes are found, it returns an appropriate error.
func getNodeObjectID(tx graph.Transaction, criteria graph.Criteria) (string, error) {
	var (
		nodeObjectID string
		nodeQuery    = tx.Nodes().Filter(criteria)
		err          = nodeQuery.Query(func(results graph.Result) error {
			defer results.Close()

			if !results.Next() {
				if err := results.Error(); err != nil {
					return err
				}

				return graph.ErrNoResultsFound
			}

			if err := results.Scan(&nodeObjectID); err != nil {
				return err
			}

			if results.Next() {
				return errors.New("ambigious matcher with more than one node matched")
			}

			return results.Error()
		}, query.Returning(
			query.NodeProperty("objectid")),
		)
	)

	return nodeObjectID, err
}

// ResolutionError is used to return an error related to the resolution of an
// inserted edge's endpoint.
type ResolutionError struct {
	error error
}

func (s ResolutionError) Error() string {
	return "unable to resolve endpoint: " + s.error.Error()
}

// formatMatchers outputs a human-readable string that describes the match expressions. Used
// for debugging and error reporting.
func formatMatchers(matchers []ein.MatchExpression) string {
	builder := &bytes.Buffer{}

	for idx, matcher := range matchers {
		if idx > 0 {
			builder.WriteString(", ")
		}

		builder.WriteString("\"")
		builder.WriteString(matcher.Key)
		builder.WriteString("\" ")
		builder.WriteString(string(matcher.Operator))
		builder.WriteString(" ")

		if jsonBytes, err := json.Marshal(matcher.Value); err != nil {
			builder.WriteString("uanble_to_marshal:\"")
			builder.WriteString(err.Error())
			builder.WriteString("\"")
		} else {
			builder.WriteString(string(jsonBytes))
		}
	}

	return builder.String()
}

// newPropertyMatcherError formats a new ResolutionError with details on the match attempted
// if the given error is of graph.ErrNoResultsFound indicating that the query for the
// endpoint ran but returned no result.
func newPropertyMatcherError(ingestEntry ein.IngestibleEndpoint, err error) ResolutionError {
	formattedMatchers := formatMatchers(ingestEntry.Matchers)

	if errors.Is(err, graph.ErrNoResultsFound) {
		return NewResolutionError(errors.New(formattedMatchers))
	}

	return NewResolutionError(fmt.Errorf("unexpected error: %w on matchers: %s", err, formattedMatchers))
}

// NewResolutionError wraps a generic error into a standardized endpoint resolution error.
func NewResolutionError(err error) ResolutionError {
	return ResolutionError{
		error: err,
	}
}

// rewriteLegacyNameMatchIngestibleEndpoint rewriets an IngestibleEndpoint from a legacy supported
// match_by "name" strategy to a generic property match.
func rewriteLegacyNameMatchIngestibleEndpoint(ingestEntry ein.IngestibleEndpoint) (ein.IngestibleEndpoint, error) {
	switch ingestEntry.MatchBy {
	case ein.MatchByName:
		if ingestEntry.Value == "" {
			return ingestEntry, errors.New("empty value for name match_by strategy")
		}

		return ein.IngestibleEndpoint{
			MatchBy: ein.MatchByProperty,
			Kind:    ingestEntry.Kind,
			Matchers: []ein.MatchExpression{{
				Key:      "name",
				Operator: ein.OperatorEqualsIgnoreCase,
				Value:    ingestEntry.Value,
			}},
		}, nil
	}

	return ingestEntry, nil
}

// resolveIngestibleEndpoint attempts to resolve an IngestibleEndpoint by querying the database
// based on its MatchBy strategy (Property or Name). If successful, it returns a new endpoint
// where MatchBy is set to MatchByID and Value contains the resolved object ID. If the endpoint
// is already resolved (MatchByObjectId), it returns it unchanged.
func resolveIngestibleEndpoint(tx graph.Transaction, ingestEntry ein.IngestibleEndpoint) (ein.IngestibleEndpoint, error) {
	if ingestEntry, err := rewriteLegacyNameMatchIngestibleEndpoint(ingestEntry); err != nil {
		return ingestEntry, err
	} else {
		switch ingestEntry.MatchBy {
		case ein.MatchByProperty:
			if criteria, err := newMatchExpr(ingestEntry.Kind, ingestEntry.Matchers); err != nil {
				return ingestEntry, NewResolutionError(err)
			} else if objectID, err := getNodeObjectID(tx, criteria); err != nil {
				return ingestEntry, newPropertyMatcherError(ingestEntry, err)
			} else {
				return ein.IngestibleEndpoint{
					Kind:    ingestEntry.Kind,
					MatchBy: ein.MatchByID,
					Value:   objectID,
				}, nil
			}

		default:
			return ingestEntry, nil
		}
	}
}

// ResolveAll orchestrates the parallel resolution of a batch of IngestibleRelationships.
// It initializes a Resolver with a pool of database workers, submits all entries for processing,
// collects the resolved relationships in a separate goroutine, and waits for completion.
// It returns the fully resolved list of relationships and any aggregated errors encountered
// during the worker execution. This function logs its duration and operation details to the context logger.
func ResolveAll(ctx context.Context, endpointResolver *Resolver, ingestEntries []ein.IngestibleRelationship) ([]ein.IngestibleRelationship, error) {
	defer measure.ContextLogAndMeasure(ctx, slog.LevelInfo, "ResolveAll")()

	// Start a new parallel resolution
	endpointResolver.Start(ctx, analysis.MaximumDatabaseParallelWorkers)

	// Update ingest entries in-place
	for idx := range ingestEntries {
		if !endpointResolver.Submit(ctx, &ingestEntries[idx]) {
			break
		}
	}

	var (
		resolverErrors  = endpointResolver.Done()
		resolvedEntries = make([]ein.IngestibleRelationship, 0, len(ingestEntries))
	)

	for _, ingestEntry := range ingestEntries {
		if ein.OrIngestMatchStrategyDefault(ingestEntry.Source.MatchBy) == ein.MatchByID && ein.OrIngestMatchStrategyDefault(ingestEntry.Target.MatchBy) == ein.MatchByID {
			resolvedEntries = append(resolvedEntries, ingestEntry)
		}
	}

	return resolvedEntries, resolverErrors
}
