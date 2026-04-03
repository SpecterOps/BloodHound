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

package v2

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
)

// anonymizableProperties are the node properties that contain identifiable information
// and should be replaced during anonymization.
var anonymizableProperties = []common.Property{
	common.Name,
	common.DisplayName,
	common.Description,
	common.Email,
	common.OperatingSystem,
	common.Title,
}

// generatePseudonym creates a random hex pseudonym with the given prefix.
func generatePseudonym(prefix string) string {
	b := make([]byte, 4)
	if _, err := rand.Read(b); err != nil {
		return prefix + "-0000"
	}
	return prefix + "-" + strings.ToUpper(hex.EncodeToString(b))
}

// AnonymizeStatusResponse is the JSON shape returned by the status endpoint.
type AnonymizeStatusResponse struct {
	Anonymized      bool `json:"anonymized"`
	BackupAvailable bool `json:"backup_available"`
}

// AnonymizeLookupResult is a single row in a lookup response.
type AnonymizeLookupResult struct {
	OriginalName   string `json:"original_name"`
	AnonymizedName string `json:"anonymized_name"`
	ObjectType     string `json:"object_type"`
}

// AnonymizeLookupResponse wraps the lookup results.
type AnonymizeLookupResponse struct {
	Results []AnonymizeLookupResult `json:"results"`
}

// HandleAnonymizeStatus returns whether the data is currently anonymized
// and whether a backup exists for restoration.
func (s Resources) HandleAnonymizeStatus(response http.ResponseWriter, request *http.Request) {
	entries, err := s.DB.GetAnonymizeTranslationEntries(request.Context())
	if err != nil {
		// Table may not exist yet - treat as non-anonymized
		api.WriteBasicResponse(request.Context(), AnonymizeStatusResponse{
			Anonymized:      false,
			BackupAvailable: false,
		}, http.StatusOK, response)
		return
	}

	hasEntries := len(entries) > 0

	api.WriteBasicResponse(request.Context(), AnonymizeStatusResponse{
		Anonymized:      hasEntries,
		BackupAvailable: hasEntries,
	}, http.StatusOK, response)
}

// HandleAnonymizeData walks every node in the graph, replaces identifiable
// properties with pseudonyms, and stores a translation table so the
// operation can be reversed.
func (s Resources) HandleAnonymizeData(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	// Prevent double-anonymization
	if existing, err := s.DB.GetAnonymizeTranslationEntries(ctx); err == nil && len(existing) > 0 {
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusConflict, "Data is already anonymized. Restore first before re-anonymizing.", request),
			response)
		return
	}

	type translationRow struct {
		NodeID        graph.ID
		PropertyKey   string
		OriginalVal   string
		AnonymizedVal string
		ObjectType    string
	}

	var translations []translationRow

	// Phase 1: Read all nodes and build translation table
	if err := s.Graph.ReadTransaction(ctx, func(tx graph.Transaction) error {
		return tx.Nodes().Fetch(func(cursor graph.Cursor[*graph.Node]) error {
			for node := range cursor.Chan() {
				objectType := kindLabel(node.Kinds)

				for _, prop := range anonymizableProperties {
					key := string(prop)
					val, err := node.Properties.Get(key).String()
					if err != nil || val == "" {
						continue
					}

					pseudonym := generatePseudonym(strings.ToUpper(key))
					translations = append(translations, translationRow{
						NodeID:        node.ID,
						PropertyKey:   key,
						OriginalVal:   val,
						AnonymizedVal: pseudonym,
						ObjectType:    objectType,
					})
				}
			}
			return cursor.Error()
		})
	}); err != nil {
		slog.ErrorContext(ctx, "Anonymize: failed to read graph nodes", "error", err)
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to read graph data: %v", err), request),
			response)
		return
	}

	if len(translations) == 0 {
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusBadRequest, "No identifiable data found to anonymize.", request),
			response)
		return
	}

	// Phase 2: Write anonymized values into the graph
	if err := s.Graph.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for _, t := range translations {
			node, err := tx.Nodes().Filter(
				query.Equals(query.NodeID(), t.NodeID),
			).First()
			if err != nil {
				slog.WarnContext(ctx, "Anonymize: could not fetch node for update", "node_id", t.NodeID, "error", err)
				continue
			}

			node.Properties.Set(t.PropertyKey, t.AnonymizedVal)
			if err := tx.UpdateNode(node); err != nil {
				return fmt.Errorf("failed to update node %d property %s: %w", t.NodeID, t.PropertyKey, err)
			}
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "Anonymize: failed to write anonymized data", "error", err)
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to write anonymized data: %v", err), request),
			response)
		return
	}

	// Phase 3: Persist the translation table to PostgreSQL
	dbEntries := make([]model.AnonymizeTranslationEntry, len(translations))
	for i, t := range translations {
		dbEntries[i] = model.AnonymizeTranslationEntry{
			NodeGraphID:     int64(t.NodeID),
			PropertyKey:     t.PropertyKey,
			OriginalValue:   t.OriginalVal,
			AnonymizedValue: t.AnonymizedVal,
			ObjectType:      t.ObjectType,
		}
	}

	if err := s.DB.SaveAnonymizeTranslationEntries(ctx, dbEntries); err != nil {
		slog.ErrorContext(ctx, "Anonymize: failed to save translation table", "error", err)
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, "Anonymization applied to graph but failed to save translation table. Manual intervention may be required.", request),
			response)
		return
	}

	api.WriteBasicResponse(ctx, map[string]any{
		"anonymized":    true,
		"entries_count": len(translations),
	}, http.StatusOK, response)
}

// HandleRestoreAnonymizedData reads the translation table and restores
// original property values to every affected node, then clears the table.
func (s Resources) HandleRestoreAnonymizedData(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	entries, err := s.DB.GetAnonymizeTranslationEntries(ctx)
	if err != nil || len(entries) == 0 {
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusBadRequest, "No anonymization backup found to restore.", request),
			response)
		return
	}

	// Restore original values
	if err := s.Graph.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for _, entry := range entries {
			node, err := tx.Nodes().Filter(
				query.Equals(query.NodeID(), graph.ID(entry.NodeGraphID)),
			).First()
			if err != nil {
				slog.WarnContext(ctx, "Restore: could not fetch node", "node_id", entry.NodeGraphID, "error", err)
				continue
			}

			node.Properties.Set(entry.PropertyKey, entry.OriginalValue)
			if err := tx.UpdateNode(node); err != nil {
				return fmt.Errorf("failed to restore node %d property %s: %w", entry.NodeGraphID, entry.PropertyKey, err)
			}
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "Restore: failed to write original data", "error", err)
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to restore data: %v", err), request),
			response)
		return
	}

	// Clear the translation table
	if err := s.DB.DeleteAnonymizeTranslationEntries(ctx); err != nil {
		slog.ErrorContext(ctx, "Restore: failed to clear translation table", "error", err)
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, "Data restored but failed to clear translation table.", request),
			response)
		return
	}

	api.WriteBasicResponse(ctx, map[string]string{"status": "restored"}, http.StatusOK, response)
}

// HandleAnonymizeLookup searches the translation table for a given query string,
// matching against both original and anonymized names.
func (s Resources) HandleAnonymizeLookup(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()
	query := request.URL.Query().Get("query")

	if query == "" {
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusBadRequest, "Query parameter 'query' is required.", request),
			response)
		return
	}

	entries, err := s.DB.SearchAnonymizeTranslationEntries(ctx, query)
	if err != nil {
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Lookup failed: %v", err), request),
			response)
		return
	}

	results := make([]AnonymizeLookupResult, len(entries))
	for i, e := range entries {
		results[i] = AnonymizeLookupResult{
			OriginalName:   e.OriginalValue,
			AnonymizedName: e.AnonymizedValue,
			ObjectType:     e.ObjectType,
		}
	}

	api.WriteBasicResponse(ctx, AnonymizeLookupResponse{Results: results}, http.StatusOK, response)
}

// kindLabel returns a human-friendly label for a node's kind set.
func kindLabel(kinds graph.Kinds) string {
	if len(kinds) == 0 {
		return "Unknown"
	}
	for _, k := range kinds {
		s := k.String()
		if s != "" {
			return s
		}
	}
	return "Unknown"
}
