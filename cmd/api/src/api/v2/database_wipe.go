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
	"net/http"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/log"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
)

type DatabaseWipe struct {
	DeleteCollectedGraphData  bool  `json:"deleteCollectedGraphData"`
	DeleteFileIngestHistory   bool  `json:"deleteFileIngestHistory"`
	DeleteDataQualityHistory  bool  `json:"deleteDataQualityHistory"`
	DeleteAssetGroupSelectors []int `json:"deleteAssetGroupSelectors"`
}

func (s Resources) HandleDatabaseWipe(response http.ResponseWriter, request *http.Request) {

	var (
		payload DatabaseWipe
		// use this struct to flag any fields that failed to delete
		errors []string
		// deleting collected graph data OR high value selectors starts analsyis
		kickoffAnalysis bool
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
	if !payload.DeleteCollectedGraphData && !payload.DeleteDataQualityHistory && !payload.DeleteFileIngestHistory && len(payload.DeleteAssetGroupSelectors) == 0 {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusBadRequest, "please select something to delete", request),
			response,
		)
		return
	}

	commitID, err := uuid.NewV4()
	if err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("failure generating uuid: %v", err.Error()), request),
			response,
		)
		return
	}

	auditEntry := &model.AuditEntry{
		Action: "DeleteBloodhoundData",
		Model: &model.AuditData{
			"options": payload,
		},
		Status:   model.AuditStatusIntent,
		CommitID: commitID,
	}

	// create an intent audit log
	if err := s.DB.AppendAuditLog(request.Context(), *auditEntry); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusInternalServerError, "failure creating an intent audit log", request),
			response,
		)
		return
	}

	// delete graph
	if payload.DeleteCollectedGraphData {
		if failed := s.deleteCollectedGraphData(request.Context(), auditEntry); failed {
			errors = append(errors, "collected graph data")
		} else {
			kickoffAnalysis = true
		}
	}

	// delete asset group selectors
	if len(payload.DeleteAssetGroupSelectors) > 0 {
		if failed := s.deleteHighValueSelectors(request.Context(), auditEntry, payload.DeleteAssetGroupSelectors); failed {
			errors = append(errors, "custom high value selectors")
		} else {
			kickoffAnalysis = true
		}
	}

	// if deleting `nodes` or deleting `asset group selectors` is successful, kickoff an analysis
	if kickoffAnalysis {
		s.TaskNotifier.RequestAnalysis()
	}

	// delete file ingest history
	if payload.DeleteFileIngestHistory {
		if failure := s.deleteFileIngestHistory(request.Context(), auditEntry); failure {
			errors = append(errors, "file ingest history")
		}
	}

	// delete data quality history
	if payload.DeleteDataQualityHistory {
		if failure := s.deleteDataQualityHistory(request.Context(), auditEntry); failure {
			errors = append(errors, "data quality history")
		}
	}

	// return a user friendly error message indicating what operations failed
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

func (s Resources) deleteCollectedGraphData(ctx context.Context, auditEntry *model.AuditEntry) (failure bool) {
	var nodeIDs []graph.ID

	if err := s.Graph.ReadTransaction(ctx,
		func(tx graph.Transaction) error {
			fetchedNodeIDs, err := ops.FetchNodeIDs(tx.Nodes())

			nodeIDs = append(nodeIDs, fetchedNodeIDs...)
			return err
		},
	); err != nil {
		log.Errorf("%s: %s", "error fetching all nodes", err.Error())
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "collected graph data")
		return true
	} else if err := s.Graph.BatchOperation(ctx, func(batch graph.Batch) error {
		for _, nodeId := range nodeIDs {
			// deleting a node also deletes all of its edges due to a sql trigger
			if err := batch.DeleteNode(nodeId); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		log.Errorf("%s: %s", "error deleting all nodes", err.Error())
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "collected graph data")
		return true
	} else {
		// if successful, handle audit log and kick off analysis
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, "collected graph data")
		return false
	}
}

func (s Resources) deleteHighValueSelectors(ctx context.Context, auditEntry *model.AuditEntry, assetGroupIDs []int) (failure bool) {

	if err := s.DB.DeleteAssetGroupSelectorsForAssetGroups(ctx, assetGroupIDs); err != nil {
		log.Errorf("%s: %s", "there was an error deleting asset group selectors ", err.Error())
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "high value selectors")
		return true
	} else {
		// if succesful, handle audit log and kick off analysis
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, "high value selectors")
		return false
	}
}

func (s Resources) deleteFileIngestHistory(ctx context.Context, auditEntry *model.AuditEntry) (failure bool) {
	if err := s.DB.DeleteAllFileUploads(ctx); err != nil {
		log.Errorf("%s: %s", "there was an error deleting file ingest history", err.Error())
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "file ingest history")
		return true
	} else {
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, "file ingest history")
		return false
	}
}

func (s Resources) deleteDataQualityHistory(ctx context.Context, auditEntry *model.AuditEntry) (failure bool) {
	if err := s.DB.DeleteAllDataQuality(ctx); err != nil {
		log.Errorf("%s: %s", "there was an error deleting data quality history", err.Error())
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, false, "data quality history")
		return true
	} else {
		s.handleAuditLogForDatabaseWipe(ctx, auditEntry, true, "data quality history")
		return false
	}
}

func (s Resources) handleAuditLogForDatabaseWipe(ctx context.Context, auditEntry *model.AuditEntry, success bool, msg string) {
	if success {
		auditEntry.Status = model.AuditStatusSuccess
		auditEntry.Model = model.AuditData{
			"delete_successful": msg,
		}
	} else {
		auditEntry.Status = model.AuditStatusFailure
		auditEntry.Model = model.AuditData{
			"delete_failed": msg,
		}
	}

	if err := s.DB.AppendAuditLog(ctx, *auditEntry); err != nil {
		log.Errorf("%s: %s", "error writing to audit log", err.Error())
	}
}
