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

package services

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/packages/go/graphschema"
	"github.com/specterops/bloodhound/packages/go/graphschema/ad"
	"github.com/specterops/bloodhound/packages/go/graphschema/azure"
	"github.com/specterops/dawgs/graph"
)

const (
	DefaultGraphExpansionLimit = 500
	MaxGraphExpansionLimit     = 1000

	GraphExpansionDirectionInbound  = "inbound"
	GraphExpansionDirectionOutbound = "outbound"
)

var (
	ErrInvalidGraphExpansionLimit            = errors.New("invalid graph expansion limit")
	ErrInvalidGraphExpansionDirection        = errors.New("invalid graph expansion direction")
	ErrInvalidGraphExpansionRelationshipKind = errors.New("invalid graph expansion relationship kind")
)

type GraphExpansionRequest struct {
	NodeID            int64
	Direction         string
	Limit             int
	IncludeProperties bool
	User              model.User
}

type GraphExpansion struct {
	Graph     model.UnifiedGraph
	Limit     int
	Truncated bool
}

type GraphExpansionValidationError struct {
	Cause   error
	Message string
}

func (s GraphExpansionValidationError) Error() string {
	return s.Message
}

func (s GraphExpansionValidationError) Unwrap() error {
	return s.Cause
}

func NewGraphExpansionValidationError(cause error, message string) error {
	return GraphExpansionValidationError{Cause: cause, Message: message}
}

func GraphExpansionLimit(requestedLimit int) (int, error) {
	if requestedLimit == 0 {
		return DefaultGraphExpansionLimit, nil
	}

	if requestedLimit < 0 {
		return 0, NewGraphExpansionValidationError(ErrInvalidGraphExpansionLimit, "limit must be greater than 0")
	}

	if requestedLimit > MaxGraphExpansionLimit {
		return 0, NewGraphExpansionValidationError(ErrInvalidGraphExpansionLimit, fmt.Sprintf("limit must be less than or equal to %d", MaxGraphExpansionLimit))
	}

	return requestedLimit, nil
}

func PruneGraphExpansionResponse(graphResponse model.UnifiedGraph, limit int) (model.UnifiedGraph, bool) {
	truncated := len(graphResponse.Edges) > limit
	if truncated {
		graphResponse.Edges = graphResponse.Edges[:limit]
	}

	retainedNodeIDs := map[string]struct{}{}
	for _, edge := range graphResponse.Edges {
		retainedNodeIDs[edge.Source] = struct{}{}
		retainedNodeIDs[edge.Target] = struct{}{}
	}

	retainedNodes := make(map[string]model.UnifiedNode, len(retainedNodeIDs))
	for nodeID, node := range graphResponse.Nodes {
		if _, retain := retainedNodeIDs[nodeID]; retain {
			retainedNodes[nodeID] = node
		}
	}

	graphResponse.Nodes = retainedNodes

	return graphResponse, truncated
}

func (s *Service) ExpandGraph(ctx context.Context, request GraphExpansionRequest) (GraphExpansion, error) {
	limit, err := GraphExpansionLimit(request.Limit)
	if err != nil {
		return GraphExpansion{}, err
	}

	anchorNode, err := s.db.GetNode(ctx, request.NodeID)
	if err != nil {
		return GraphExpansion{}, err
	}
	if !s.nodeAccessChecker.CanAccessNode(ctx, anchorNode) {
		return GraphExpansion{}, ErrNodeAccessDenied
	}

	graphResponse, err := s.db.ExpandGraph(ctx, request.NodeID, request.Direction, limit+1)
	if err != nil {
		return GraphExpansion{}, err
	}

	wasTruncated := len(graphResponse.Edges) > limit
	if s.shouldFilterGraphExpansionForETAC(request.User) {
		graphResponse = filterGraphExpansionForETAC(graphResponse, request.User)
	}

	prunedGraph, truncatedAfterFiltering := PruneGraphExpansionResponse(graphResponse, limit)
	if !request.IncludeProperties {
		prunedGraph = ClearGraphExpansionProperties(prunedGraph)
	}

	return GraphExpansion{
		Graph:     prunedGraph,
		Limit:     limit,
		Truncated: wasTruncated || truncatedAfterFiltering,
	}, nil
}

func ClearGraphExpansionProperties(graphResponse model.UnifiedGraph) model.UnifiedGraph {
	for id, node := range graphResponse.Nodes {
		node.Properties = nil
		graphResponse.Nodes[id] = node
	}

	for i, edge := range graphResponse.Edges {
		edge.Properties = nil
		graphResponse.Edges[i] = edge
	}

	return graphResponse
}

func (s *Service) shouldFilterGraphExpansionForETAC(user model.User) bool {
	if s.accessControl == nil {
		return false
	}

	return s.accessControl.ShouldFilterForETAC(&user)
}

func filterGraphExpansionForETAC(graphResponse model.UnifiedGraph, user model.User) model.UnifiedGraph {
	accessList := make([]string, 0, len(user.EnvironmentTargetedAccessControl))
	for _, envAccess := range user.EnvironmentTargetedAccessControl {
		accessList = append(accessList, envAccess.EnvironmentID)
	}

	filteredNodes := make(map[string]model.UnifiedNode, len(graphResponse.Nodes))
	environmentKeys := []string{ad.DomainSID.String(), azure.TenantID.String(), graphschema.EnvironmentIDKey}

	for id, node := range graphResponse.Nodes {
		include := false
		for _, key := range environmentKeys {
			if val, ok := node.Properties[key]; ok {
				if envStr, ok := val.(string); ok && slices.Contains(accessList, envStr) {
					include = true
					break
				}
			}
		}

		if include {
			filteredNodes[id] = node
			continue
		}

		kind := "Unknown"
		if len(node.Kinds) > 0 && node.Kinds[0] != "" {
			kind = node.Kinds[0]
		}

		filteredNodes[id] = model.UnifiedNode{
			Label:         fmt.Sprintf("** Hidden %s Object **", kind),
			Kind:          "HIDDEN",
			Kinds:         []string{},
			ObjectId:      "HIDDEN",
			IsTierZero:    false,
			IsOwnedObject: false,
			LastSeen:      time.Time{},
			Properties:    nil,
			Hidden:        true,
		}
	}

	filteredEdges := make([]model.UnifiedEdge, 0, len(graphResponse.Edges))
	for _, edge := range graphResponse.Edges {
		if filteredNodes[edge.Target].Hidden || filteredNodes[edge.Source].Hidden {
			filteredEdges = append(filteredEdges, model.UnifiedEdge{
				Source:     edge.Source,
				Target:     edge.Target,
				Label:      "** Hidden Edge **",
				Kind:       "HIDDEN",
				LastSeen:   time.Time{},
				Properties: nil,
			})
		} else {
			filteredEdges = append(filteredEdges, edge)
		}
	}

	return model.UnifiedGraph{
		Nodes:    filteredNodes,
		Edges:    filteredEdges,
		Literals: graph.Literals{},
	}
}
