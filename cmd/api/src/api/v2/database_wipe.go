// Copyright 2024 Specter Ops, Inc.
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
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"slices"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/ops"
	"github.com/specterops/dawgs/query"
)

type DatabaseWipe struct {
	DeleteCollectedGraphData bool     `json:"deleteCollectedGraphData"`
	DeleteSourceKinds        []int    `json:"deleteSourceKinds"` // an id of 0 represents "sourceless" data
	DeleteRelationships      []string `json:"deleteRelationships"`

	DeleteFileIngestHistory   bool  `json:"deleteFileIngestHistory"`
	DeleteDataQualityHistory  bool  `json:"deleteDataQualityHistory"`
	DeleteAssetGroupSelectors []int `json:"deleteAssetGroupSelectors"`
}

func (s Resources) HandleDatabaseWipe(response http.ResponseWriter, request *http.Request) {

	var (
		payload DatabaseWipe
		err     error
		// use this struct to flag any fields that failed to delete
		errors []string
		// deleting collected graph data OR high value selectors starts analsyis
		kickoffAnalysis bool
		auditEntry      model.AuditEntry
	)

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed.", request),
			response,
		)
		return
	}

	// return `BadRequest` if request is empty
	isEmptyRequest := !payload.DeleteCollectedGraphData && !payload.DeleteDataQualityHistory && !payload.DeleteFileIngestHistory && len(payload.DeleteRelationships) == 0 && len(payload.DeleteAssetGroupSelectors) == 0 && len(payload.DeleteSourceKinds) == 0
	if isEmptyRequest {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusBadRequest, "please select something to delete", request),
			response,
		)
		return
	}

	isMixedDeleteRequest := payload.DeleteCollectedGraphData && len(payload.DeleteSourceKinds) > 0
	if isMixedDeleteRequest {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusBadRequest, "request may only specify either deleteCollectedGraphData or deleteSourceKinds, not both", request),
			response,
		)
		return
	}

	if auditEntry, err = model.NewAuditEntry(
		model.AuditLogActionDeleteBloodhoundData,
		model.AuditLogStatusIntent,
		model.AuditData{
			"options": payload,
		},
	); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request),
			response,
		)
		return
	}

	// create an intent audit log
	if err := s.DB.AppendAuditLog(request.Context(), auditEntry); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusInternalServerError, "failure creating an intent audit log", request),
			response,
		)
		return
	}

	deleteGraph := payload.DeleteCollectedGraphData || len(payload.DeleteSourceKinds) > 0
	if deleteGraph {
		if clearGraphDataFlag, err := s.DB.GetFlagByKey(request.Context(), appcfg.FeatureClearGraphData); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, "unable to inspect the feature flag for clearing graph data", request),
				response,
			)
			return
		} else if !clearGraphDataFlag.Enabled {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusBadRequest, "deleting graph data is currently disabled", request),
				response,
			)
			return
		} else {
			var userId string
			if user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
				slog.WarnContext(request.Context(), "Encountered request analysis for unknown user, this shouldn't happen")
				userId = "unknown-user-database-wipe"
			} else {
				userId = user.ID.String()
			}

			if deleteRequest, err := s.BuildDeleteRequest(request.Context(), userId, payload); err != nil {
				api.WriteErrorResponse(
					request.Context(),
					api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("failure building delete request: %s", err.Error()), request),
					response,
				)
			} else if err := s.DB.RequestCollectedGraphDataDeletion(request.Context(), deleteRequest); err != nil {
				api.HandleDatabaseError(request, response, err)
				return
			}

			s.handleAuditLogForDatabaseWipe(request.Context(), &auditEntry, true, "collected graph data")
		}

	}

	// delete asset group selectors
	if len(payload.DeleteAssetGroupSelectors) > 0 {
		if failed := s.deleteHighValueSelectors(request.Context(), &auditEntry, payload.DeleteAssetGroupSelectors); failed {
			errors = append(errors, "custom high value selectors")
		} else {
			kickoffAnalysis = true
		}
	}

	// if deleting `nodes` or deleting `asset group selectors` is successful, kickoff an analysis
	if kickoffAnalysis {
		var userId string
		if user, isUser := auth.GetUserFromAuthCtx(ctx.FromRequest(request).AuthCtx); !isUser {
			slog.WarnContext(request.Context(), "Encountered request analysis for unknown user, this shouldn't happen")
			userId = "unknown-user-database-wipe"
		} else {
			userId = user.ID.String()
		}

		if err := s.DB.RequestAnalysis(request.Context(), userId); err != nil {
			api.HandleDatabaseError(request, response, err)
			return
		}
	}

	// delete file ingest history
	if payload.DeleteFileIngestHistory {
		if failure := s.deleteFileIngestHistory(request.Context(), &auditEntry); failure {
			errors = append(errors, "file ingest history")
		}
	}

	// delete data quality history
	if payload.DeleteDataQualityHistory {
		if failure := s.deleteDataQualityHistory(request.Context(), &auditEntry); failure {
			errors = append(errors, "data quality history")
		}
	}

	// delete requested graph edges by name
	if len(payload.DeleteRelationships) > 0 {
		if failure := s.deleteEdges(request.Context(), &auditEntry, payload.DeleteRelationships); failure {
			errors = append(errors, "graph edges")
		}
	}

	// return a user-friendly error message indicating what operations failed
	if len(errors) > 0 {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("We encountered an error while deleting %s.  Please submit your request again.", strings.Join(errors, ", ")), request),
			response,
		)
		return
	} else {
		response.WriteHeader(http.StatusNoContent)
	}

}

func (s Resources) deleteHighValueSelectors(ctx context.Context, auditEntry *model.AuditEntry, assetGroupIDs []int) (failure bool) {

	if err := s.DB.DeleteAssetGroupSelectorsForAssetGroups(ctx, assetGroupIDs); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("%s: %s", "there was an error deleting asset group selectors ", err.Error()))
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "high value selectors")
		return true
	} else {
		// if succesful, handle audit log and kick off analysis
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, "high value selectors")
		return false
	}
}

func (s Resources) deleteFileIngestHistory(ctx context.Context, auditEntry *model.AuditEntry) (failure bool) {
	if err := s.DB.DeleteAllIngestJobs(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("%s: %s", "there was an error deleting file ingest history", err.Error()))
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "file ingest history")
		return true
	} else {
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, "file ingest history")
		return false
	}
}

func (s Resources) deleteDataQualityHistory(ctx context.Context, auditEntry *model.AuditEntry) (failure bool) {
	if err := s.DB.DeleteAllDataQuality(ctx); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("%s: %s", "there was an error deleting data quality history", err.Error()))
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "data quality history")
		return true
	} else {
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, "data quality history")
		return false
	}
}

func (s Resources) deleteEdges(ctx context.Context, auditEntry *model.AuditEntry, edgeNames []string) (failure bool) {
	// Use the graph batch API to find and delete relationships matching the provided edge names
	if err := s.Graph.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, name := range edgeNames {
			targetCriteria := query.Kind(query.Relationship(), graph.StringKind(name))

			rels, err := ops.FetchRelationships(batch.Relationships().Filter(targetCriteria))
			if err != nil {
				return err
			}

			for _, rel := range rels {
				if err := batch.DeleteRelationship(rel.ID); err != nil {
					return err
				}
			}
		}

		return nil
	}); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("%s: %s", "there was an error deleting graph edges", err.Error()))
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, strings.Join(edgeNames, ", "))
		return true
	} else {
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, strings.Join(edgeNames, ", "))
		return false
	}
}

func (s Resources) handleAuditLogForDatabaseWipe(ctx context.Context, auditEntry *model.AuditEntry, success bool, msg string) {
	if success {
		auditEntry.Status = model.AuditLogStatusSuccess
		auditEntry.Model = model.AuditData{
			"delete_request_successful": msg,
		}
	} else {
		auditEntry.Status = model.AuditLogStatusFailure
		auditEntry.Model = model.AuditData{
			"delete_failed": msg,
		}
	}

	if err := s.DB.AppendAuditLog(ctx, *auditEntry); err != nil {
		slog.ErrorContext(ctx, fmt.Sprintf("%s: %s", "error writing to audit log", err.Error()))
	}
}

func (s Resources) BuildDeleteRequest(ctx context.Context, userID string, payload DatabaseWipe) (model.AnalysisRequest, error) {
	deleteRequest := model.AnalysisRequest{
		RequestedBy:    userID,
		RequestType:    model.AnalysisRequestDeletion,
		DeleteAllGraph: payload.DeleteCollectedGraphData,
	}

	if slices.Contains(payload.DeleteSourceKinds, 0) {
		deleteRequest.DeleteSourcelessGraph = true
	}

	if len(payload.DeleteSourceKinds) > 0 {
		// Load source kind definitions from DB
		sourceKinds, err := s.DB.GetSourceKinds(ctx)
		if err != nil {
			return deleteRequest, fmt.Errorf("failed to get source kinds: %w", err)
		}

		// Recover the source kind names from the provided IDs
		requestedKinds := make(graph.Kinds, 0, len(payload.DeleteSourceKinds))
		for _, id := range payload.DeleteSourceKinds {
			found := false
			for _, sk := range sourceKinds {
				if sk.ID == id {
					requestedKinds = append(requestedKinds, sk.Name)
					found = true
					break
				}
			}
			if !found && id != 0 { // id of 0 is our internal convention meaning "sourceless". this is not an error case
				return deleteRequest, fmt.Errorf("requested source kind id %d not found", id)
			}
		}

		// Validate that all requested kinds are legitimate
		validKinds, err := s.Graph.FetchKinds(ctx)
		if err != nil {
			return deleteRequest, fmt.Errorf("failed to fetch valid kinds: %w", err)
		}

		// Create a fast lookup map for validation
		validSet := make(map[graph.Kind]struct{}, len(validKinds))
		for _, k := range validKinds {
			validSet[k] = struct{}{}
		}

		for _, rk := range requestedKinds {
			if _, ok := validSet[rk]; !ok {
				return deleteRequest, fmt.Errorf("requested source kind %q is not a valid kind", rk)
			}
		}

		// All kinds are valid
		deleteRequest.DeleteSourceKinds = requestedKinds.Strings()
	}

	return deleteRequest, nil
}
