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
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/dawgs/graph"
	"github.com/specterops/bloodhound/src/api"
	"gorm.io/gorm"
	// "github.com/specterops/bloodhound/src/model"
)

type CreateEnvironmentConfigurationRequest struct {
	Name string          `json:"name"`
	Data json.RawMessage `json:"data"`
}


// upload environment description to postgres
func (s Resources) CreateEnvironmentConfiguration(response http.ResponseWriter, request *http.Request) {
	var createRequest CreateEnvironmentConfigurationRequest

	if err := api.ReadJSONRequestPayloadLimited(&createRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	if createRequest.Name == "" || len(createRequest.Data) == 0 {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "name and data fields are required", request), response)
		return
	}

	// Validate the data structure
	var dataMap map[string]interface{}
	if err := json.Unmarshal(createRequest.Data, &dataMap); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "invalid data format", request), response)
		return
	}

	if _, hasMeta := dataMap["meta"]; !hasMeta {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "missing meta field in data", request), response)
		return
	}

	if _, hasNodeTypes := dataMap["nodeTypes"]; !hasNodeTypes {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "missing nodeTypes field in data", request), response)
		return
	}

	if _, hasRelationshipTypes := dataMap["relationshipTypes"]; !hasRelationshipTypes {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "missing relationshipTypes field in data", request), response)
		return
	}

	envConfig, err := s.DB.CreateEnvironmentConfiguration(request.Context(), createRequest.Name, createRequest.Data)
	if err != nil {
		api.HandleDatabaseError(request, response, err)
		return
	}

	api.WriteBasicResponse(request.Context(), envConfig, http.StatusCreated, response)
}

func (s Resources) GetEnvironmentConfiguration(response http.ResponseWriter, request *http.Request) {
    vars := mux.Vars(request)
    name := vars["name"]

    if name == "" {
        api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "Environment name is required", request), response)
        return
    }

    envConfig, err := s.DB.GetEnvironmentConfiguration(request.Context(), name)
    if err != nil {
        if errors.Is(err, gorm.ErrRecordNotFound) {
            api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "Environment configuration not found", request), response)
        } else {
            api.HandleDatabaseError(request, response, err)
        }
        return
    }

    api.WriteBasicResponse(request.Context(), envConfig, http.StatusOK, response)
}

// upload environment data
type UploadEnvironmentDataRequest struct {
	NodeData         []NodeData         `json:"nodeData"`
	RelationshipData []RelationshipData `json:"relationshipData"`
}

type NodeData struct {
	ID         string                  `json:"id"`
	Type       string                 `json:"type"`
	Properties map[string]interface{} `json:"properties"`
}

type RelationshipData struct {
	Type        string `json:"type"`
	StartNodeID string  `json:"startNodeID"`
	EndNodeID   string  `json:"endNodeID"`
}

func (s Resources) UploadEnvironmentData(response http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	name := vars["name"]


	var uploadRequest UploadEnvironmentDataRequest
	if err := api.ReadJSONRequestPayloadLimited(&uploadRequest, request); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	}

	// fetch the environment configuration
	envConfig, err := s.DB.GetEnvironmentConfiguration(request.Context(), name)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusNotFound, "Environment configuration not found", request), response)
		} else {
			api.HandleDatabaseError(request, response, err)
		}
		return
	}

	// parse the config data
	var configData struct {
		Meta struct {
			Prefix string `json:"prefix"`
		} `json:"meta"`
		NodeTypes         []map[string]interface{} `json:"nodeTypes"`
		RelationshipTypes []map[string]interface{} `json:"relationshipTypes"`
	}
	if err := json.Unmarshal(envConfig.Data, &configData); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, "Failed to parse environment configuration", request), response)
		return
	}

	// create nodes
	for _, nodeData := range uploadRequest.NodeData {
		if err := CreateNode(s.Graph, request.Context(), nodeData, configData); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			return
		}
	}

	// create relationships
	for _, relData := range uploadRequest.RelationshipData {
		if err := CreateRelationship(s.Graph, request.Context(), relData, configData); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			return
		}
	}
	api.WriteBasicResponse(request.Context(), "Data uploaded successfully", http.StatusOK, response)
}

func CreateNode(graph graph.Database, ctx context.Context, nodeData NodeData, configData struct {
	Meta struct {
		Prefix string `json:"prefix"`
	} `json:"meta"`


	NodeTypes         []map[string]interface{} `json:"nodeTypes"`
	RelationshipTypes []map[string]interface{} `json:"relationshipTypes"`
}) error {

	// Create node in Neo4j
	var pairs []string
	params := map[string]interface{}{
		"id": nodeData.ID,
	}

	for k, v := range nodeData.Properties {
		pairs = append(pairs, fmt.Sprintf("%s: '%s'", k, v))
	}

	propertyString := strings.Join(pairs, ", ")

	query := fmt.Sprintf("CREATE (n:%s:%s {id: $id, %s})", configData.Meta.Prefix, nodeData.Type, propertyString)
	err := graph.Run(ctx, query, params)
	return err
}

func CreateRelationship(graph graph.Database, ctx context.Context, relData RelationshipData, configData struct {
	Meta struct {
		Prefix string `json:"prefix"`
	} `json:"meta"`
	NodeTypes         []map[string]interface{} `json:"nodeTypes"`
	RelationshipTypes []map[string]interface{} `json:"relationshipTypes"`
}) error {

	// Create relationship in Neo4j
	query := fmt.Sprintf(`
		MATCH (start {id: $startID})
		MATCH (end {id: $endID})
		CREATE (start)-[r:%s]->(end)
		RETURN r
	`, relData.Type)
	params := map[string]interface{}{
		"startID": relData.StartNodeID,
		"endID":   relData.EndNodeID,
	}
	err := graph.Run(ctx, query, params)
	return err
}