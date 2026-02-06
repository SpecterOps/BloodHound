// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
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
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	authctx "github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemaextensions.go -package=mocks . OpenGraphSchemaService
type OpenGraphSchemaService interface {
	UpsertOpenGraphExtension(ctx context.Context, openGraphExtension model.GraphExtensionInput) (bool, error)
	ListExtensions(ctx context.Context) (model.GraphSchemaExtensions, error)
	DeleteExtension(ctx context.Context, extensionID int32) error
}

type GraphExtensionPayload struct {
	GraphSchemaExtension GraphSchemaExtensionPayload `json:"schema"`
	// GraphSchemaProperties        []GraphSchemaPropertiesPayload        `json:"properties"`
	GraphSchemaRelationshipKinds []GraphSchemaRelationshipKindsPayload `json:"relationship_kinds"`
	GraphSchemaNodeKinds         []GraphSchemaNodeKindsPayload         `json:"node_kinds"`
	GraphEnvironments            []EnvironmentPayload                  `json:"environments"`
	GraphRelationshipFindings    []RelationshipFindingsPayload         `json:"relationship_findings"`
}

type GraphSchemaExtensionPayload struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Version     string `json:"version"`
	Namespace   string `json:"namespace"`
}

type GraphSchemaRelationshipKindsPayload struct {
	Name          string `json:"name"`
	Description   string `json:"description"`
	IsTraversable bool   `json:"is_traversable"` // indicates whether the edge-kind is a traversable path
}

type GraphSchemaNodeKindsPayload struct {
	Name          string `json:"name"`
	DisplayName   string `json:"display_name"`    // can be different from name but usually isn't other than Base/Entity
	Description   string `json:"description"`     // human-readable description of the node kind
	IsDisplayKind bool   `json:"is_display_kind"` // indicates if this kind should supersede others and be displayed
	Icon          string `json:"icon"`            // font-awesome icon for the registered node kind
	IconColor     string `json:"color"`           // icon hex color
}
type GraphSchemaPropertiesPayload struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
}

type EnvironmentPayload struct {
	EnvironmentKind string   `json:"environment_kind"`
	SourceKind      string   `json:"source_kind"`
	PrincipalKinds  []string `json:"principal_kinds"`
}

type RelationshipFindingsPayload struct {
	Name             string             `json:"name"`
	DisplayName      string             `json:"display_name"`
	SourceKind       string             `json:"source_kind"`
	RelationshipKind string             `json:"relationship_kind"`
	EnvironmentKind  string             `json:"environment_kind"`
	Remediation      RemediationPayload `json:"remediation"`
}

type RemediationPayload struct {
	ShortDescription string `json:"short_description"`
	LongDescription  string `json:"long_description"`
	ShortRemediation string `json:"short_remediation"`
	LongRemediation  string `json:"long_remediation"`
}

// OpenGraphSchemaIngest - handles incoming graph extension upsert requests
func (s Resources) OpenGraphSchemaIngest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
		err error

		updated bool

		extractExtensionData  func(file io.Reader) (GraphExtensionPayload, error)
		graphExtensionPayload GraphExtensionPayload
	)

	// feature flag is checked as part of middleware
	if user, isUser := auth.GetUserFromAuthCtx(authctx.FromRequest(request).AuthCtx); !isUser {
		var errMessage = "No associated user found"
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusUnauthorized, errMessage, request), response)
		return
	} else if !user.Roles.Has(model.Role{Name: auth.RoleAdministrator}) {
		var errMessage = "user does not have sufficient permissions to create or update an extension"
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusForbidden, errMessage, request), response)
		return
	} else if request.Body == nil {
		var errMessage = "open graph extension payload cannot be empty"
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, errMessage, request), response)
		return
	}
	request.Body = http.MaxBytesReader(response, request.Body, api.DefaultAPIPayloadReadLimitBytes)
	defer request.Body.Close()
	switch {
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
		extractExtensionData = extractExtensionDataFromJSON
	// func can be created if ZIP file support ends up being needed
	case utils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
		fallthrough
	//	extractExtensionData = extractExtensionDataFromZipFile
	default:
		var errMessage = fmt.Sprintf("%s; Content type must be application/json",
			fmt.Sprintf("invalid content-type: %s", request.Header[headers.ContentType.String()]))
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusUnsupportedMediaType, errMessage, request), response)
		return
	}

	if graphExtensionPayload, err = extractExtensionData(request.Body); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	} else if updated, err = s.OpenGraphSchemaService.UpsertOpenGraphExtension(ctx,
		convertGraphExtensionPayloadToGraphExtension(graphExtensionPayload)); err != nil {
		switch {
		case strings.Contains(err.Error(), model.ErrGraphExtensionValidation.Error()) ||
			strings.Contains(err.Error(), model.ErrGraphExtensionBuiltIn.Error()):
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		case strings.Contains(err.Error(), model.ErrGraphDBRefreshKinds.Error()):
			fallthrough
		default:
			slog.WarnContext(ctx, err.Error())
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		}
	} else if updated {
		response.WriteHeader(http.StatusOK)
	} else {
		response.WriteHeader(http.StatusCreated)
	}
}

// extractExtensionDataFromJSON - extracts a GraphExtensionPayload from the incoming payload. Will return an error
// if the decoder fails to decode the payload.
func extractExtensionDataFromJSON(payload io.Reader) (GraphExtensionPayload, error) {
	var (
		err            error
		decoder        = json.NewDecoder(payload)
		graphExtension GraphExtensionPayload
	)

	if err = decoder.Decode(&graphExtension); err != nil {
		return graphExtension, fmt.Errorf("unable to decode graph extension payload: %w", err)
	}
	return graphExtension, nil
}

// convertGraphExtensionPayloadToGraphExtension - converts the GraphExtensionInput view layer model to the service layer model.
func convertGraphExtensionPayloadToGraphExtension(payload GraphExtensionPayload) model.GraphExtensionInput {
	var (
		graphExtension = model.GraphExtensionInput{
			ExtensionInput: model.ExtensionInput{
				Name:        payload.GraphSchemaExtension.Name,
				DisplayName: payload.GraphSchemaExtension.DisplayName,
				Version:     payload.GraphSchemaExtension.Version,
				Namespace:   payload.GraphSchemaExtension.Namespace,
			},
			NodeKindsInput:         make(model.NodesInput, 0),
			RelationshipKindsInput: make(model.RelationshipsInput, 0),
			PropertiesInput:        make(model.PropertiesInput, 0),
			EnvironmentsInput:      make(model.EnvironmentsInput, 0),
		}
	)

	for _, nodeKindPayload := range payload.GraphSchemaNodeKinds {
		graphExtension.NodeKindsInput = append(graphExtension.NodeKindsInput,
			model.NodeInput{
				Name:          nodeKindPayload.Name,
				DisplayName:   nodeKindPayload.DisplayName,
				Description:   nodeKindPayload.Description,
				IsDisplayKind: nodeKindPayload.IsDisplayKind,
				Icon:          nodeKindPayload.Icon,
				IconColor:     nodeKindPayload.IconColor,
			})
	}
	for _, edgeKindPayload := range payload.GraphSchemaRelationshipKinds {
		graphExtension.RelationshipKindsInput = append(graphExtension.RelationshipKindsInput,
			model.RelationshipInput{
				Name:          edgeKindPayload.Name,
				Description:   edgeKindPayload.Description,
				IsTraversable: edgeKindPayload.IsTraversable,
			})
	}
	/*
		for _, propertyPayload := range payload.GraphSchemaProperties {
			graphExtension.PropertiesInput = append(graphExtension.PropertiesInput,
				model.PropertyInput{
					Name:        propertyPayload.Name,
					DisplayName: propertyPayload.DisplayName,
					DataType:    propertyPayload.DataType,
					Description: propertyPayload.Description,
				})
		}

	*/
	for _, environmentPayload := range payload.GraphEnvironments {
		graphExtension.EnvironmentsInput = append(graphExtension.EnvironmentsInput,
			model.EnvironmentInput{
				EnvironmentKindName: environmentPayload.EnvironmentKind,
				SourceKindName:      environmentPayload.SourceKind,
				PrincipalKinds:      environmentPayload.PrincipalKinds,
			})
	}
	for _, findingPayload := range payload.GraphRelationshipFindings {
		graphExtension.RelationshipFindingsInput = append(graphExtension.RelationshipFindingsInput, model.RelationshipFindingInput{
			Name:                 findingPayload.Name,
			DisplayName:          findingPayload.DisplayName,
			SourceKindName:       findingPayload.SourceKind,
			RelationshipKindName: findingPayload.RelationshipKind,
			EnvironmentKindName:  findingPayload.EnvironmentKind,
			RemediationInput: model.RemediationInput{
				ShortDescription: findingPayload.Remediation.ShortDescription,
				LongDescription:  findingPayload.Remediation.LongDescription,
				ShortRemediation: findingPayload.Remediation.ShortRemediation,
				LongRemediation:  findingPayload.Remediation.LongRemediation,
			},
		})
	}
	return graphExtension
}

type ExtensionsResponse struct {
	Extensions []ExtensionInfo `json:"extensions"`
}

type ExtensionInfo struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func (s Resources) ListExtensions(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
	)

	if extensions, err := s.OpenGraphSchemaService.ListExtensions(ctx); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error listing graph schema extensions: %v", err), request), response)
		return
	} else {
		var extensionsResponse = make([]ExtensionInfo, len(extensions))
		for i, extension := range extensions {
			extensionsResponse[i] = ExtensionInfo{
				ID:      strconv.Itoa(int(extension.ID)),
				Name:    extension.DisplayName,
				Version: extension.Version,
			}
		}

		api.WriteBasicResponse(ctx, ExtensionsResponse{Extensions: extensionsResponse}, http.StatusOK, response)
	}
}

func (s Resources) DeleteExtension(response http.ResponseWriter, request *http.Request) {
	var (
		ctx         = request.Context()
		extensionID = mux.Vars(request)[api.URIPathVariableExtensionID]
	)

	// feature flag is checked as part of middleware
	if user, isUser := auth.GetUserFromAuthCtx(authctx.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusUnauthorized, "No associated user found", request), response)
	} else if !user.Roles.Has(model.Role{Name: auth.RoleAdministrator}) {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusForbidden, "user does not have sufficient permissions to delete an extension", request), response)
	} else if extID, err := strconv.ParseInt(extensionID, 10, 32); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponseDetailsIDMalformed, request), response)
	} else if err := s.OpenGraphSchemaService.DeleteExtension(ctx, int32(extID)); err != nil {
		if errors.Is(err, database.ErrNotFound) {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusNotFound, fmt.Sprintf("no extension found matching extension id: %s", extensionID), request), response)
		} else if errors.Is(err, model.ErrGraphExtensionBuiltIn) {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, "built-in extensions cannot be deleted", request), response)
		} else {
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		}
	} else {
		response.WriteHeader(http.StatusNoContent)
	}
}
