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
)

type DatabaseManagement struct {
	CollectedGraphData bool `json:"collectedGraphData"`
	HighValueSelectors bool `json:"highValueSelectors"`
	FileIngestHistory  bool `json:"fileIngestHistory"`
	DataQualityHistory bool `json:"dataQualityHistory"`
	AssetGroupId       int  `json:"assetGroupId"`
}

func (s Resources) HandleDatabaseManagement(response http.ResponseWriter, request *http.Request) {

	var (
		payload DatabaseManagement
		nodeIDs []graph.ID
	)

	if err := api.ReadJSONRequestPayloadLimited(&payload, request); err != nil {
		api.WriteErrorResponse(
			request.Context(),
			api.BuildErrorResponse(http.StatusBadRequest, "JSON malformed.", request),
			response,
		)
	} else if payload.CollectedGraphData {

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
		}

		if err := s.Graph.BatchOperation(request.Context(), func(batch graph.Batch) error {
			for _, nodeId := range nodeIDs {
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
		}

	} else if payload.HighValueSelectors {
		if payload.AssetGroupId == 0 {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusBadRequest, "please provide an assetGroupId to delete", request),
				response,
			)
		}
		if err := s.DB.DeleteAssetGroupSelectors(request.Context(), payload.AssetGroupId); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s %d: %s", "there was an error deleting asset group with id = ", payload.AssetGroupId, err.Error()), request),
				response,
			)
		}
	} else if payload.FileIngestHistory {
		if err := s.DB.DeleteAllFileUploads(); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s: %s", "there was an error deleting file ingest history", err.Error()), request),
				response,
			)
		}
	} else if payload.DataQualityHistory {
		if err := s.DB.DeleteAllDataQuality(); err != nil {
			api.WriteErrorResponse(
				request.Context(),
				api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("%s: %s", "there was an error deleting data quality history", err.Error()), request),
				response,
			)
		}
	} else {
		response.WriteHeader(http.StatusNoContent)
	}
}
