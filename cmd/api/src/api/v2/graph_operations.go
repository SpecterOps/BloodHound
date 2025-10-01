// Copyright 2025 Specter Ops, Inc.
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
)

type CreateNodeRequest struct {
	ObjectID   string                 `json:"object_id"`
	Labels     []string               `json:"labels"`
	Properties map[string]interface{} `json:"properties"`
}

type CreateEdgeRequest struct {
	SourceObjectID string                 `json:"source_object_id"`
	TargetObjectID string                 `json:"target_object_id"`
	EdgeKind       string                 `json:"edge_kind"`
	Properties     map[string]interface{} `json:"properties,omitempty"`
}

type DeleteEdgeRequest struct {
	SourceObjectID string `json:"source_object_id"`
	TargetObjectID string `json:"target_object_id"`
	EdgeKind       string `json:"edge_kind"`
}

// CreateNode handles HTTP requests to create a new graph node.
// The request is routed through the GraphOpsLog service which logs the operation
// before executing it against the graph database.
func (s Resources) CreateNode(response http.ResponseWriter, request *http.Request) {
	var req CreateNodeRequest
	defer request.Body.Close()

	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	if req.ObjectID == "" || len(req.Labels) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "object_id and labels required", request), response)
		return
	}

	// Use the graph operations service to create the node (logs then executes)
	if err := s.GraphOpsLog.CreateNode(request.Context(), req.ObjectID, req.Labels, req.Properties); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(map[string]string{"object_id": req.ObjectID})
}

// CreateEdge handles HTTP requests to create a new graph edge.
// The request is routed through the GraphOpsLog service which logs the operation
// before executing it against the graph database.
func (s Resources) CreateEdge(response http.ResponseWriter, request *http.Request) {
	var req CreateEdgeRequest
	defer request.Body.Close()

	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	if req.SourceObjectID == "" || req.TargetObjectID == "" || req.EdgeKind == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "source_object_id, target_object_id, and edge_kind required", request), response)
		return
	}

	// Use the graph operations service to create the edge (logs then executes)
	if err := s.GraphOpsLog.CreateEdge(request.Context(), req.SourceObjectID, req.TargetObjectID, req.EdgeKind, req.Properties); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(map[string]string{"edge_kind": req.EdgeKind})
}

// DeleteNode handles HTTP requests to delete a graph node.
// The request is routed through the GraphOpsLog service which logs the operation
// before executing it against the graph database.
func (s Resources) DeleteNode(response http.ResponseWriter, request *http.Request) {
	objectID := mux.Vars(request)["object_id"]
	if objectID == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "object_id required", request), response)
		return
	}

	// Use the graph operations service to delete the node (logs then executes)
	if err := s.GraphOpsLog.DeleteNode(request.Context(), objectID); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusOK)
}

// DeleteEdge handles HTTP requests to delete a graph edge.
// The request is routed through the GraphOpsLog service which logs the operation
// before executing it against the graph database.
func (s Resources) DeleteEdge(response http.ResponseWriter, request *http.Request) {
	var req DeleteEdgeRequest
	defer request.Body.Close()

	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	if req.SourceObjectID == "" || req.TargetObjectID == "" || req.EdgeKind == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "source_object_id, target_object_id, and edge_kind required", request), response)
		return
	}

	// Use the graph operations service to delete the edge (logs then executes)
	if err := s.GraphOpsLog.DeleteEdge(request.Context(), req.SourceObjectID, req.TargetObjectID, req.EdgeKind); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusOK)
}

// GetReplayLog retrieves the most recent graph operation replay log entries.
// Returns the last 100 operations by default.
func (s Resources) GetReplayLog(response http.ResponseWriter, request *http.Request) {
	// Retrieve the last 100 changes (can be made configurable via query param if needed)
	entries, err := s.GraphOpsLog.GetRecentChanges(request.Context(), 100)
	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusOK)
	json.NewEncoder(response).Encode(map[string]interface{}{
		"count":   len(entries),
		"entries": entries,
	})
}
