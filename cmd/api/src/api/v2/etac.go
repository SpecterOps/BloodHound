// Copyright 2025 Specter Ops, Inc.
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
	"slices"
	"time"

	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/services/dogtags"
	"github.com/specterops/dawgs/graph"
)

type UpdateEnvironmentRequest struct {
	EnvironmentID string `json:"environment_id"`
}

type UpdateUserETACRequest struct {
	Environments []UpdateEnvironmentRequest `json:"environments"`
}

func CheckUserAccessToEnvironments(ctx context.Context, db database.EnvironmentTargetedAccessControlData, user model.User, environments ...string) (bool, error) {
	if user.AllEnvironments {
		return true, nil
	}

	allowedList, err := db.GetEnvironmentTargetedAccessControlForUser(ctx, user)

	if err != nil {
		return false, err
	}

	allowedMap := make(map[string]struct{}, len(allowedList))
	for _, envAccess := range allowedList {
		allowedMap[envAccess.EnvironmentID] = struct{}{}
	}

	for _, env := range environments {
		_, ok := allowedMap[env]

		if !ok {
			return false, nil
		}
	}

	return true, nil
}

// ExtractEnvironmentIDsFromUser is a helper function
// to extract a user's environments from their model as a list of strings
func ExtractEnvironmentIDsFromUser(user *model.User) []string {
	list := make([]string, 0, len(user.EnvironmentTargetedAccessControl))

	for _, envAccess := range user.EnvironmentTargetedAccessControl {
		list = append(list, envAccess.EnvironmentID)
	}

	return list
}

// ShouldFilterForETAC determines whether ETAC filtering should be applied
// based on the feature flag and user's environment access.
func ShouldFilterForETAC(dogTagsService dogtags.Service, user model.User) bool {
	if etacEnabled := dogTagsService.GetFlagAsBool(dogtags.ETAC_ENABLED); !etacEnabled {
		return false
	} else if user.AllEnvironments {
		// no filtering required if user has all environments
		return false
	}

	return true
}

// filterETACGraph applies ETAC(Environment-based Access Control) filtering for the CypherQuery endpoint.
// Nodes that the user does not have access to are replaced with hidden placeholder nodes,
// and edges connected to hidden nodes are marked as hidden.
func filterETACGraph(graphResponse model.UnifiedGraph, user model.User) (model.UnifiedGraph, error) {
	accessList := ExtractEnvironmentIDsFromUser(&user)

	filteredResponse := model.UnifiedGraph{}
	filteredNodes := make(map[string]model.UnifiedNode)

	environmentKeys := []string{"domainsid", "tenantid"}

	// filter nodes based on environment access
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
			// user has access, we keep original node
			filteredNodes[id] = node
		} else {
			// extract node source kind for display in hidden label
			var kind string
			if len(node.Kinds) > 0 && node.Kinds[0] != "" {
				kind = node.Kinds[0]
			} else {
				kind = "Unknown"
			}

			label := fmt.Sprintf("** Hidden %s Object **", kind)
			filteredNodes[id] = model.UnifiedNode{
				Label:         label,
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
	}

	filteredResponse.Nodes = filteredNodes
	filteredEdges := make([]model.UnifiedEdge, 0, len(graphResponse.Edges))

	// mark edges as hidden if attached to a hidden node
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
			// nodes on both ends of edge are accessible, we keep original edge
			filteredEdges = append(filteredEdges, edge)
		}
	}
	filteredResponse.Edges = filteredEdges

	// ensure literals are filtered out of etac filtered responses
	filteredResponse.Literals = graph.Literals{}

	return filteredResponse, nil
}
