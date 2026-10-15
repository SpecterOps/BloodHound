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

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/services/graphify"
	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/packages/go/chow/ingest"
	"github.com/specterops/bloodhound/packages/go/chow/payload"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
)

type nodeKey struct {
	Source string
	ID     string
}
type edgeKey struct {
	Start nodeKey
	Kind  string
	End   nodeKey
}

// memoryGraph reduces production graph updates into an initially empty graph.
// IDs are scoped by their identity kind, and repeated upserts merge properties.
type memoryGraph struct {
	graphify.BatchUpdater
	NodesByID map[nodeKey]*graph.Node
	EdgesByID map[edgeKey]*graph.Properties
}

func newMemoryGraph() *memoryGraph {
	return &memoryGraph{NodesByID: map[nodeKey]*graph.Node{}, EdgesByID: map[edgeKey]*graph.Properties{}}
}

func identity(node *graph.Node, kind graph.Kind, properties []string) (nodeKey, error) {
	if len(properties) != 1 || properties[0] != common.ObjectID.String() {
		return nodeKey{}, fmt.Errorf("unsupported identity properties: %v", properties)
	}
	var (
		objectID, err = node.Properties.Get(common.ObjectID.String()).String()
		source        string
	)
	if err != nil {
		return nodeKey{}, err
	}
	if kind != nil {
		source = kind.String()
	}
	return nodeKey{Source: source, ID: objectID}, nil
}

func (s *memoryGraph) UpdateNodeBy(update graph.NodeUpdate) error {
	var key, err = identity(update.Node, update.IdentityKind, update.IdentityProperties)
	if err != nil {
		return err
	}
	if existing, found := s.NodesByID[key]; found {
		existing.Merge(update.Node)
	} else {
		s.NodesByID[key] = graph.PrepareNode(update.Node.Properties.Clone(), update.Node.Kinds...)
	}
	return nil
}

func (s *memoryGraph) UpdateRelationshipBy(update graph.RelationshipUpdate) error {
	var (
		start, startError = identity(update.Start, update.StartIdentityKind, update.StartIdentityProperties)
		end, endError     = identity(update.End, update.EndIdentityKind, update.EndIdentityProperties)
		key               = edgeKey{Start: start, Kind: update.Relationship.Kind.String(), End: end}
	)
	if startError != nil || endError != nil {
		return errors.Join(startError, endError)
	}
	if len(update.IdentityProperties) != 0 {
		return fmt.Errorf("relationship property identities are unsupported")
	}
	// Relationship upserts also create or enrich their endpoint nodes.
	if err := s.UpdateNodeBy(graph.NodeUpdate{Node: update.Start, IdentityKind: update.StartIdentityKind, IdentityProperties: update.StartIdentityProperties}); err != nil {
		return err
	}
	if err := s.UpdateNodeBy(graph.NodeUpdate{Node: update.End, IdentityKind: update.EndIdentityKind, IdentityProperties: update.EndIdentityProperties}); err != nil {
		return err
	}
	if existing, found := s.EdgesByID[key]; found {
		existing.Merge(update.Relationship.Properties)
	} else {
		s.EdgesByID[key] = update.Relationship.Properties.Clone()
	}
	return nil
}

// The production resolver opens read transactions even for ID-only endpoints.
// This adapter serves those transactions without a database and rejects lookups
// that depend on external graph state instead of silently guessing a match.
type offlineDatabase struct{ graph.Database }
type offlineTransaction struct{ graph.Transaction }
type offlineQuery struct{ graph.NodeQuery }

func (s offlineDatabase) ReadTransaction(ctx context.Context, delegate graph.TransactionDelegate, options ...graph.TransactionOption) error {
	return delegate(offlineTransaction{})
}
func (s offlineTransaction) Nodes() graph.NodeQuery                   { return offlineQuery{} }
func (s offlineQuery) Filter(criteria graph.Criteria) graph.NodeQuery { return s }
func (s offlineQuery) Query(delegate func(graph.Result) error, criteria ...graph.Criteria) error {
	return fmt.Errorf("offline comparison supports only ID-based endpoints; name/property matching requires graph state")
}

func readState(content []byte, expectedType ingest.DataType, useRawIDs bool) (*memoryGraph, error) {
	if !json.Valid(content) {
		return nil, fmt.Errorf("invalid JSON in %s payload", expectedType)
	}
	var (
		schema, schemaError = payload.LoadSchema()
		reader              = bytes.NewReader(content)
		validator           = payload.NewValidator(reader, schema)
		parsed              payload.ParsedData
		state               = newMemoryGraph()
		err                 error
	)
	if schemaError != nil {
		return nil, schemaError
	}
	if expectedType == ingest.DataTypeOpenGraph {
		parsed, _, err = validator.ParseAndValidate()
	} else {
		parsed, err = validator.ParseMetadata()
	}
	if err != nil {
		return nil, err
	}
	if parsed.PayloadType != expectedType {
		return nil, fmt.Errorf("expected %s payload, got %s", expectedType, parsed.PayloadType)
	}
	if _, err := reader.Seek(0, 0); err != nil {
		return nil, err
	}
	var ingestContext = graphify.NewIngestContext(context.Background(),
		graphify.WithIngestTime(time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC)),
		graphify.WithUseRawObjectIDs(useRawIDs),
		graphify.WithBatchUpdater(state),
		graphify.WithEndpointResolver(endpoint.NewResolver(offlineDatabase{})),
	)
	if err := graphify.IngestWrapper(ingestContext, reader, parsed, graphify.ReadOptions{
		RegisterSourceKind: func(kind graph.Kind) error { return nil },
	}); err != nil {
		return nil, err
	}
	return state, nil
}

type difference struct {
	Entity           string `json:"entity"`
	Identity         any    `json:"identity"`
	Field            string `json:"field"`
	AzurePresent     bool   `json:"azure_present"`
	OpenGraphPresent bool   `json:"opengraph_present"`
	Azure            any    `json:"azure,omitempty"`
	OpenGraph        any    `json:"opengraph,omitempty"`
}

type report struct {
	Verdict           string            `json:"verdict"`
	Equivalent        bool              `json:"equivalent"`
	IgnoredProperties []string          `json:"ignored_properties"`
	AzureNodes        int               `json:"azure_nodes"`
	OpenGraphNodes    int               `json:"opengraph_nodes"`
	AzureEdges        int               `json:"azure_edges"`
	OpenGraphEdges    int               `json:"opengraph_edges"`
	Differences       []difference      `json:"differences"`
	Diagnostics       []json.RawMessage `json:"diagnostics"`
}

// Normalize through JSON to compare persisted JSON-compatible values rather
// than Go implementation types (e.g. time.Time versus an RFC3339 string).
func canonicalProperties(properties *graph.Properties, ignored map[string]bool) (map[string]json.RawMessage, error) {
	var (
		result = map[string]json.RawMessage{}
		err    error
	)
	for key, value := range properties.Map {
		if ignored[key] {
			continue
		}
		if result[key], err = json.Marshal(value); err != nil {
			return nil, err
		}
	}
	return result, nil
}

func compareProperties(result *report, entity string, key any, left, right *graph.Properties, ignored map[string]bool) error {
	var (
		leftValues, leftError   = canonicalProperties(left, ignored)
		rightValues, rightError = canonicalProperties(right, ignored)
		keys                    = map[string]bool{}
	)
	if leftError != nil || rightError != nil {
		return errors.Join(leftError, rightError)
	}
	for field := range leftValues {
		keys[field] = true
	}
	for field := range rightValues {
		keys[field] = true
	}
	for field := range keys {
		var (
			leftValue, hasLeft   = leftValues[field]
			rightValue, hasRight = rightValues[field]
		)
		if hasLeft != hasRight || !bytes.Equal(leftValue, rightValue) {
			result.Differences = append(result.Differences, difference{Entity: entity, Identity: key, Field: field, AzurePresent: hasLeft, OpenGraphPresent: hasRight, Azure: leftValue, OpenGraph: rightValue})
		}
	}
	return nil
}

func compareStates(left, right *memoryGraph, ignored map[string]bool) (report, error) {
	var result = report{AzureNodes: len(left.NodesByID), OpenGraphNodes: len(right.NodesByID), AzureEdges: len(left.EdgesByID), OpenGraphEdges: len(right.EdgesByID), Differences: []difference{}, Diagnostics: []json.RawMessage{}}
	for key := range ignored {
		result.IgnoredProperties = append(result.IgnoredProperties, key)
	}
	slices.Sort(result.IgnoredProperties)
	for key, node := range left.NodesByID {
		var other, found = right.NodesByID[key]
		if !found {
			result.Differences = append(result.Differences, difference{Entity: "node", Identity: key, Field: "existence", AzurePresent: true})
			continue
		}
		var leftKinds, rightKinds = node.Kinds.Strings(), other.Kinds.Strings()
		slices.Sort(leftKinds)
		slices.Sort(rightKinds)
		if !slices.Equal(leftKinds, rightKinds) {
			result.Differences = append(result.Differences, difference{Entity: "node", Identity: key, Field: "kinds", AzurePresent: true, OpenGraphPresent: true, Azure: leftKinds, OpenGraph: rightKinds})
		}
		if err := compareProperties(&result, "node", key, node.Properties, other.Properties, ignored); err != nil {
			return report{}, err
		}
	}
	for key := range right.NodesByID {
		if _, found := left.NodesByID[key]; !found {
			result.Differences = append(result.Differences, difference{Entity: "node", Identity: key, Field: "existence", OpenGraphPresent: true})
		}
	}
	for key, properties := range left.EdgesByID {
		if other, found := right.EdgesByID[key]; !found {
			result.Differences = append(result.Differences, difference{Entity: "edge", Identity: key, Field: "existence", AzurePresent: true})
		} else if err := compareProperties(&result, "edge", key, properties, other, ignored); err != nil {
			return report{}, err
		}
	}
	for key := range right.EdgesByID {
		if _, found := left.EdgesByID[key]; !found {
			result.Differences = append(result.Differences, difference{Entity: "edge", Identity: key, Field: "existence", OpenGraphPresent: true})
		}
	}
	slices.SortFunc(result.Differences, func(left, right difference) int {
		leftJSON, _ := json.Marshal(left)
		rightJSON, _ := json.Marshal(right)
		return bytes.Compare(leftJSON, rightJSON)
	})
	result.Equivalent = len(result.Differences) == 0
	result.Verdict = "different"
	if result.Equivalent {
		result.Verdict = "equivalent"
	}
	return result, nil
}

func main() {
	var (
		azurePath         = flag.String("azure", "azurehound_example.json", "AzureHound input JSON")
		openGraphPath     = flag.String("opengraph", "baseline.json", "OpenGraph input JSON")
		reportPath        = flag.String("report", "equivalence-report.json", "JSON comparison report")
		ignoredProperties = flag.String("ignore-properties", "firstseen,lastseen,lastcollected,primarykind", "Comma-separated property exclusions; use an empty string for strict comparison")
		useRawIDs         = flag.Bool("use-raw-object-ids", false, "Match the server's use_raw_object_ids setting")
	)
	flag.Parse()
	if err := run(*azurePath, *openGraphPath, *reportPath, *ignoredProperties, *useRawIDs); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(azurePath, openGraphPath, reportPath, ignoredProperties string, useRawIDs bool) error {
	var (
		azureContent, azureError         = os.ReadFile(azurePath)
		openGraphContent, openGraphError = os.ReadFile(openGraphPath)
		ignored                          = map[string]bool{}
		diagnostics                      bytes.Buffer
		previousLogger                   = slog.Default()
	)
	if azureError != nil || openGraphError != nil {
		return errors.Join(azureError, openGraphError)
	}
	for _, property := range strings.Split(ignoredProperties, ",") {
		if property = strings.TrimSpace(property); property != "" {
			ignored[property] = true
		}
	}
	// Converter warnings/errors must be visible in the report, even when the
	// production converter logs an error without returning it.
	slog.SetDefault(slog.New(slog.NewJSONHandler(&diagnostics, &slog.HandlerOptions{Level: slog.LevelWarn})))
	defer slog.SetDefault(previousLogger)
	var (
		left, leftError   = readState(azureContent, ingest.DataTypeAzure, useRawIDs)
		right, rightError = readState(openGraphContent, ingest.DataTypeOpenGraph, useRawIDs)
	)
	if leftError != nil || rightError != nil {
		return errors.Join(leftError, rightError)
	}
	var result, err = compareStates(left, right, ignored)
	if err != nil {
		return err
	}
	for _, line := range bytes.Split(bytes.TrimSpace(diagnostics.Bytes()), []byte{'\n'}) {
		if len(line) != 0 {
			result.Diagnostics = append(result.Diagnostics, json.RawMessage(bytes.Clone(line)))
		}
	}
	if len(result.Diagnostics) != 0 {
		result.Verdict = "inconclusive"
		result.Equivalent = false
	}
	var content, marshalError = json.MarshalIndent(result, "", "  ")
	if marshalError != nil {
		return marshalError
	}
	if err := os.WriteFile(reportPath, append(content, '\n'), 0o600); err != nil {
		return err
	}
	fmt.Printf("AzureHound: %d nodes, %d edges\nOpenGraph: %d nodes, %d edges\nDifferences: %d; converter diagnostics: %d\nReport: %s\n", result.AzureNodes, result.AzureEdges, result.OpenGraphNodes, result.OpenGraphEdges, len(result.Differences), len(result.Diagnostics), reportPath)
	if len(result.Diagnostics) != 0 {
		return fmt.Errorf("inconclusive: production converters reported warnings/errors; inspect the report")
	}
	if !result.Equivalent {
		return fmt.Errorf("ingest states differ")
	}
	fmt.Println("Equivalent under the listed property exclusions.")
	return nil
}
