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
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/graphschema"
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
// Uses 8 bytes (64-bit) of entropy for the suffix. Returns an error if
// crypto/rand fails.
func generatePseudonym(prefix string, seen map[string]struct{}) (string, error) {
	for attempts := 0; attempts < 10; attempts++ {
		b := make([]byte, 8)
		if _, err := rand.Read(b); err != nil {
			return "", fmt.Errorf("crypto/rand.Read failed: %w", err)
		}
		name := prefix + "-" + strings.ToUpper(hex.EncodeToString(b))
		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			return name, nil
		}
	}
	// Extremely unlikely: 10 consecutive collisions in 64-bit space
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("crypto/rand.Read failed on fallback: %w", err)
	}
	name := prefix + "-" + strings.ToUpper(hex.EncodeToString(b))
	seen[name] = struct{}{}
	return name, nil
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
	hasEntries, err := s.DB.HasAnonymizeTranslationEntries(request.Context())
	if err != nil {
		slog.ErrorContext(request.Context(), "AnonymizeStatus: failed to check translation entries", attr.Error(err))
		api.WriteErrorResponse(request.Context(),
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to check anonymization status: %v", err), request),
			response)
		return
	}

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
	if hasEntries, err := s.DB.HasAnonymizeTranslationEntries(ctx); err != nil {
		slog.ErrorContext(ctx, "Anonymize: failed to check existing entries", attr.Error(err))
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to check anonymization state: %v", err), request),
			response)
		return
	} else if hasEntries {
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
	seen := make(map[string]struct{})

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

					pseudonym, err := generatePseudonym(strings.ToUpper(key), seen)
					if err != nil {
						return fmt.Errorf("failed to generate pseudonym for node %d property %s: %w", node.ID, key, err)
					}
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
		slog.ErrorContext(ctx, "Anonymize: failed to read graph nodes", attr.Error(err))
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

	// Phase 2: Persist the translation table to PostgreSQL first so we can
	// recover if the graph write fails.
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
		slog.ErrorContext(ctx, "Anonymize: failed to save translation table", attr.Error(err))
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to save translation table: %v", err), request),
			response)
		return
	}

	// Phase 3: Write anonymized values into the graph
	var appliedCount, skippedCount int
	if err := s.Graph.WriteTransaction(ctx, func(tx graph.Transaction) error {
		for _, t := range translations {
			node, err := tx.Nodes().Filter(
				query.Equals(query.NodeID(), t.NodeID),
			).First()
			if err != nil {
				if graph.IsErrNotFound(err) {
					slog.WarnContext(ctx, "Anonymize: node not found, skipping",
						slog.Uint64("node_id", uint64(t.NodeID)))
					skippedCount++
					continue
				}
				return fmt.Errorf("failed to fetch node %d: %w", t.NodeID, err)
			}

			// Compare-and-swap: only update if the property still matches the snapshot
			currentVal, _ := node.Properties.Get(t.PropertyKey).String()
			if currentVal != t.OriginalVal {
				slog.WarnContext(ctx, "Anonymize: property changed since snapshot, skipping",
					slog.Uint64("node_id", uint64(t.NodeID)),
					slog.String("property", t.PropertyKey))
				skippedCount++
				continue
			}

			node.Properties.Set(t.PropertyKey, t.AnonymizedVal)
			if err := tx.UpdateNode(node); err != nil {
				return fmt.Errorf("failed to update node %d property %s: %w", t.NodeID, t.PropertyKey, err)
			}
			appliedCount++
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "Anonymize: failed to write anonymized data", attr.Error(err))
		// Roll back the translation table since graph write failed
		if deleteErr := s.DB.DeleteAnonymizeTranslationEntries(ctx); deleteErr != nil {
			slog.ErrorContext(ctx, "Anonymize: failed to roll back translation table after graph write failure", attr.Error(deleteErr))
		}
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to write anonymized data: %v", err), request),
			response)
		return
	}

	api.WriteBasicResponse(ctx, map[string]any{
		"anonymized":    true,
		"applied_count": appliedCount,
		"skipped_count": skippedCount,
		"total_entries": len(translations),
	}, http.StatusOK, response)
}

// HandleRestoreAnonymizedData reads the translation table and restores
// original property values to every affected node, then clears the table.
func (s Resources) HandleRestoreAnonymizedData(response http.ResponseWriter, request *http.Request) {
	ctx := request.Context()

	entries, err := s.DB.GetAnonymizeTranslationEntries(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Restore: failed to read translation entries", attr.Error(err))
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to read translation table: %v", err), request),
			response)
		return
	}

	if len(entries) == 0 {
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
				if graph.IsErrNotFound(err) {
					slog.WarnContext(ctx, "Restore: node not found, skipping",
						slog.Int64("node_id", entry.NodeGraphID))
					continue
				}
				return fmt.Errorf("failed to fetch node %d: %w", entry.NodeGraphID, err)
			}

			// Compare-and-swap: only restore if the property still holds the anonymized value
			currentVal, _ := node.Properties.Get(entry.PropertyKey).String()
			if currentVal != entry.AnonymizedValue {
				slog.WarnContext(ctx, "Restore: property changed since anonymization, skipping",
					slog.Int64("node_id", entry.NodeGraphID),
					slog.String("property", entry.PropertyKey))
				continue
			}

			node.Properties.Set(entry.PropertyKey, entry.OriginalValue)
			if err := tx.UpdateNode(node); err != nil {
				return fmt.Errorf("failed to restore node %d property %s: %w", entry.NodeGraphID, entry.PropertyKey, err)
			}
		}
		return nil
	}); err != nil {
		slog.ErrorContext(ctx, "Restore: failed to write original data", attr.Error(err))
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("Failed to restore data: %v", err), request),
			response)
		return
	}

	// Clear the translation table
	if err := s.DB.DeleteAnonymizeTranslationEntries(ctx); err != nil {
		slog.ErrorContext(ctx, "Restore: failed to clear translation table", attr.Error(err))
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
	queryParam := strings.TrimSpace(request.URL.Query().Get("query"))

	if queryParam == "" {
		api.WriteErrorResponse(ctx,
			api.BuildErrorResponse(http.StatusBadRequest, "Query parameter 'query' is required.", request),
			response)
		return
	}

	entries, err := s.DB.SearchAnonymizeTranslationEntries(ctx, queryParam)
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

// kindLabel returns a human-friendly label for a node's kind set,
// preferring the most specific kind over generic base kinds.
func kindLabel(kinds graph.Kinds) string {
	return graphschema.PrimaryNodeKind(nil, kinds).String()
}
