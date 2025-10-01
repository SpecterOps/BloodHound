// Copyright 2025 Specter Ops, Inc.
// SPDX-License-Identifier: Apache-2.0

package v2

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/packages/go/graphschema/common"
	"github.com/specterops/dawgs/graph"
	"github.com/specterops/dawgs/query"
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

	kinds := make(graph.Kinds, len(req.Labels))
	for i, label := range req.Labels {
		kinds[i] = graph.StringKind(label)
	}

	props := req.Properties
	if props == nil {
		props = make(map[string]interface{})
	}
	props[common.ObjectID.String()] = strings.ToUpper(req.ObjectID)
	props[common.LastSeen.String()] = time.Now().UTC()

	err := s.Graph.BatchOperation(request.Context(), func(batch graph.Batch) error {
		return batch.UpdateNodeBy(graph.NodeUpdate{
			Node:               graph.PrepareNode(graph.AsProperties(props), kinds...),
			IdentityKind:       kinds[0],
			IdentityProperties: []string{common.ObjectID.String()},
		})
	})

	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(map[string]string{"object_id": req.ObjectID})
}

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

	props := req.Properties
	if props == nil {
		props = make(map[string]interface{})
	}
	props[common.LastSeen.String()] = time.Now().UTC()

	err := s.Graph.BatchOperation(request.Context(), func(batch graph.Batch) error {
		var src, tgt *graph.Node

		findNode := func(objectID string) (*graph.Node, error) {
			var node *graph.Node
			err := batch.Nodes().Filter(
				query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
			).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for n := range cursor.Chan() {
					node = n
					break
				}
				return cursor.Error()
			})
			return node, err
		}

		var err error
		if src, err = findNode(req.SourceObjectID); err != nil || src == nil {
			return fmt.Errorf("source node not found")
		}
		if tgt, err = findNode(req.TargetObjectID); err != nil || tgt == nil {
			return fmt.Errorf("target node not found")
		}

		return batch.UpdateRelationshipBy(graph.RelationshipUpdate{
			Relationship:            graph.PrepareRelationship(graph.AsProperties(props), graph.StringKind(req.EdgeKind)),
			Start:                   graph.PrepareNode(graph.AsProperties(graph.PropertyMap{common.ObjectID: strings.ToUpper(req.SourceObjectID), common.LastSeen: time.Now().UTC()}), src.Kinds...),
			StartIdentityKind:       src.Kinds[0],
			StartIdentityProperties: []string{common.ObjectID.String()},
			End:                     graph.PrepareNode(graph.AsProperties(graph.PropertyMap{common.ObjectID: strings.ToUpper(req.TargetObjectID), common.LastSeen: time.Now().UTC()}), tgt.Kinds...),
			EndIdentityKind:         tgt.Kinds[0],
			EndIdentityProperties:   []string{common.ObjectID.String()},
		})
	})

	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusCreated)
	json.NewEncoder(response).Encode(map[string]string{"edge_kind": req.EdgeKind})
}

func (s Resources) DeleteNode(response http.ResponseWriter, request *http.Request) {
	objectID := mux.Vars(request)["object_id"]
	if objectID == "" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "object_id required", request), response)
		return
	}

	err := s.Graph.ReadTransaction(request.Context(), func(tx graph.Transaction) error {
		return tx.Nodes().Filter(
			query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
		).Delete()
	})

	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusOK)
}

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

	err := s.Graph.ReadTransaction(request.Context(), func(tx graph.Transaction) error {
		var src, tgt *graph.Node

		findNode := func(objectID string) (*graph.Node, error) {
			var node *graph.Node
			err := tx.Nodes().Filter(
				query.Equals(query.NodeProperty(common.ObjectID.String()), strings.ToUpper(objectID)),
			).Fetch(func(cursor graph.Cursor[*graph.Node]) error {
				for n := range cursor.Chan() {
					node = n
					break
				}
				return cursor.Error()
			})
			return node, err
		}

		var err error
		if src, err = findNode(req.SourceObjectID); err != nil || src == nil {
			return fmt.Errorf("source node not found")
		}
		if tgt, err = findNode(req.TargetObjectID); err != nil || tgt == nil {
			return fmt.Errorf("target node not found")
		}

		return tx.Relationships().Filter(
			query.And(
				query.Equals(query.StartID(), src.ID),
				query.Equals(query.EndID(), tgt.ID),
				query.KindIn(query.Relationship(), graph.StringKind(req.EdgeKind)),
			),
		).Delete()
	})

	if err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, err.Error(), request), response)
		return
	}

	response.WriteHeader(http.StatusOK)
}
