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
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/dawgs/ops"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/model"
)

type DatabaseManagement struct {
	DeleteCollectedGraphData bool `json:"deleteCollectedGraphData"`
	DeleteHighValueSelectors bool `json:"deleteHighValueSelectors"`
	DeleteFileIngestHistory  bool `json:"deleteFileIngestHistory"`
	DeleteDataQualityHistory bool `json:"deleteDataQualityHistory"`
	AssetGroupId             int  `json:"assetGroupId"`
}

func (s Resources) HandleDatabaseWipe(response http.ResponseWriter, request *http.Request) {

	var (
		payload DatabaseManagement
		nodeIDs []graph.ID
		options []string
	)

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed.", request),
			response,
		)
	}

	// delete graph
	if payload.DeleteCollectedGraphData {
		options = append(options, "collected graph data")

		if err := s.Graph.ReadTransaction(request.Context(), func(tx graph.Transaction) error {
			fetchedNodeIDs, err := ops.FetchNodeIDs(tx.Nodes())

			nodeIDs = append(nodeIDs, fetchedNodeIDs...)
			return err
		}); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s: %s", "error fetching all nodes", err.Error()), request),
				response,
			)
			return
		}

		if err := s.Graph.BatchOperation(request.Context(), func(batch graph.Batch) error {
			for _, nodeId := range nodeIDs {
				// deleting a node also deletes all of its edges due to a sql trigger
				if err := batch.DeleteNode(nodeId); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s: %s", "error deleting all nodes", err.Error()), request),
				response,
			)
			return
		}

		// if succesful, kick off analysis
		s.TaskNotifier.RequestAnalysis()
	}

	// delete custom high value selectors
	if payload.DeleteHighValueSelectors {
		options = append(options, "custom high value selectors")

		if payload.AssetGroupId == 0 {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusBadRequest, "please provide an assetGroupId to delete", request),
				response,
			)
			return
		}
		if err := s.DB.DeleteAssetGroupSelectorsForAssetGroup(request.Context(), payload.AssetGroupId); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s %d: %s", "there was an error deleting asset group with id = ", payload.AssetGroupId, err.Error()), request),
				response,
			)
			return
		}

		// if succesful, kick off analysis
		s.TaskNotifier.RequestAnalysis()
	}

	// delete file ingest history
	if payload.DeleteFileIngestHistory {
		options = append(options, "file ingest history")

		if err := s.DB.DeleteAllFileUploads(); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s: %s", "there was an error deleting file ingest history", err.Error()), request),
				response,
			)
			return
		}
	}

	// delete data quality history
	if payload.DeleteDataQualityHistory {
		options = append(options, "data quality history")

		if err := s.DB.DeleteAllDataQuality(); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s: %s", "there was an error deleting data quality history", err.Error()), request),
				response,
			)
			return
		}
	}

	if err := s.DB.AppendAuditLog(request.Context(), model.AuditEntry{
		Action: "DeleteBloodhoundData",
		Model: &model.AuditData{
			"options": options,
		},
		Status: model.AuditStatusSuccess,
	}); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusInternalServerError, "there was an error creating audit log for deleting Bloodhound data", request),
			response,
		)
		return
	}

	response.WriteHeader(http.StatusNoContent)
}
