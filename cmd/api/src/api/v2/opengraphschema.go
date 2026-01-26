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
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	ctx2 "github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	bhUtils "github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/opengraphschemaservice.go -package=mocks . OpenGraphSchemaService

type OpenGraphSchemaService interface {
	UpsertOpenGraphExtension(ctx context.Context, graphExtension model.GraphExtension) (bool, error)
}

type GraphExtensionPayload struct {
	GraphSchemaExtension  GraphSchemaExtensionPayload    `json:"extension"`
	GraphSchemaProperties []GraphSchemaPropertiesPayload `json:"properties"`
	GraphSchemaEdgeKinds  []GraphSchemaEdgeKindsPayload  `json:"relationship_kinds"` // TODO: Rename Edge table to conform with schema
	GraphSchemaNodeKinds  []GraphSchemaNodeKindsPayload  `json:"node_kinds"`
	GraphEnvironments     []EnvironmentPayload           `json:"environments"`
	GraphFinding          []FindingsPayload              `json:"findings"`
}

type GraphSchemaExtensionPayload struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Version     string `json:"version"`
	Namespace   string `json:"namespace"`
}

type EnvironmentPayload struct {
	EnvironmentKind string   `json:"environment_kind"`
	SourceKind      string   `json:"source_kind"`
	PrincipalKinds  []string `json:"principal_kinds"`
}

type GraphSchemaEdgeKindsPayload struct {
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
	IconColor     string `json:"icon_color"`      // icon hex color
}

type GraphSchemaPropertiesPayload struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	DataType    string `json:"data_type"`
	Description string `json:"description"`
}

type FindingsPayload struct {
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
	if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "No associated "+
			"user found", request), response)
	} else if !user.Roles.Has(model.Role{Name: auth.RoleAdministrator}) {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusForbidden, "user does not "+
			"have sufficient permissions to create or update an extension", request), response)
	} else if request.Body == nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, "open graph "+
			"extension payload cannot be empty", request), response)
	} else {
		request.Body = http.MaxBytesReader(response, request.Body, api.DefaultAPIPayloadReadLimitBytes)
		defer request.Body.Close()
		switch {
		case bhUtils.HeaderMatches(request.Header, headers.ContentType.String(), mediatypes.ApplicationJson.String()):
			extractExtensionData = extractExtensionDataFromJSON
		// func can be created if ZIP file support ends up being needed
		case bhUtils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
			fallthrough
		//	extractExtensionData = extractExtensionDataFromZipFile
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnsupportedMediaType,
				fmt.Sprintf("%s; Content type must be application/json",
					fmt.Sprintf("invalid content-type: %s", request.Header[headers.ContentType.String()])), request), response)
			return
		}

		if graphExtensionPayload, err = extractExtensionData(request.Body); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			return
		}

		if updated, err = s.OpenGraphSchemaService.UpsertOpenGraphExtension(ctx,
			ConvertGraphExtensionPayloadToGraphExtension(graphExtensionPayload)); err != nil {
			switch {
			case strings.Contains(err.Error(), model.GraphExtensionValidationError.Error()) || strings.Contains(err.Error(), model.GraphExtensionBuiltInError.Error()):
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			case strings.Contains(err.Error(), model.GraphDBRefreshKindsError.Error()): // TODO: Do we want to return an error or let it succeed?
				fallthrough
			default:
				slog.WarnContext(ctx, err.Error())
				api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
				return
			}
		} else if updated {
			response.WriteHeader(http.StatusOK)
		} else {
			response.WriteHeader(http.StatusCreated)
		}
	}
}

// extractExtensionDataFromJSON - extracts a GraphExtensionPayload from the incoming payload. Will return an error if there
// are any extra fields or if the decoder fails to decode the payload.
func extractExtensionDataFromJSON(payload io.Reader) (GraphExtensionPayload, error) {
	var (
		err            error
		decoder        = json.NewDecoder(payload)
		graphExtension GraphExtensionPayload
	)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&graphExtension); err != nil {
		return graphExtension, fmt.Errorf("unable to decode graph extension payload: %w", err)
	}
	return graphExtension, nil
}

// ConvertGraphExtensionPayloadToGraphExtension - converts the GraphExtension view layer model to the service layer model.
// Exported so it can be tested in the same package as the other functions.
func ConvertGraphExtensionPayloadToGraphExtension(payload GraphExtensionPayload) model.GraphExtension {
	var (
		graphExtension = model.GraphExtension{
			GraphSchemaExtension: model.GraphSchemaExtension{
				Name:        payload.GraphSchemaExtension.Name,
				DisplayName: payload.GraphSchemaExtension.DisplayName,
				Version:     payload.GraphSchemaExtension.Version,
				Namespace:   payload.GraphSchemaExtension.Namespace,
			},
			GraphSchemaNodeKinds:  make(model.GraphSchemaNodeKinds, 0),
			GraphSchemaEdgeKinds:  make(model.GraphSchemaEdgeKinds, 0),
			GraphSchemaProperties: make(model.GraphSchemaProperties, 0),
			GraphEnvironments:     make([]model.GraphEnvironment, 0),
		}
	)

	for _, nodeKindPayload := range payload.GraphSchemaNodeKinds {
		graphExtension.GraphSchemaNodeKinds = append(graphExtension.GraphSchemaNodeKinds,
			model.GraphSchemaNodeKind{
				Name:          nodeKindPayload.Name,
				DisplayName:   nodeKindPayload.DisplayName,
				Description:   nodeKindPayload.Description,
				IsDisplayKind: nodeKindPayload.IsDisplayKind,
				Icon:          nodeKindPayload.Icon,
				IconColor:     nodeKindPayload.IconColor,
			})
	}
	for _, edgeKindPayload := range payload.GraphSchemaEdgeKinds {
		graphExtension.GraphSchemaEdgeKinds = append(graphExtension.GraphSchemaEdgeKinds,
			model.GraphSchemaEdgeKind{
				Name:          edgeKindPayload.Name,
				Description:   edgeKindPayload.Description,
				IsTraversable: edgeKindPayload.IsTraversable,
			})
	}
	for _, propertyPayload := range payload.GraphSchemaProperties {
		graphExtension.GraphSchemaProperties = append(graphExtension.GraphSchemaProperties,
			model.GraphSchemaProperty{
				Name:        propertyPayload.Name,
				DisplayName: propertyPayload.DisplayName,
				DataType:    propertyPayload.DataType,
				Description: propertyPayload.Description,
			})
	}
	for _, environmentPayload := range payload.GraphEnvironments {
		graphExtension.GraphEnvironments = append(graphExtension.GraphEnvironments,
			model.GraphEnvironment{
				EnvironmentKind: environmentPayload.EnvironmentKind,
				SourceKind:      environmentPayload.SourceKind,
				PrincipalKinds:  environmentPayload.PrincipalKinds,
			})
	}
	for _, findingPayload := range payload.GraphFinding {
		graphExtension.GraphFindings = append(graphExtension.GraphFindings, model.GraphFinding{
			Name:             findingPayload.Name,
			DisplayName:      findingPayload.DisplayName,
			SourceKind:       findingPayload.SourceKind,
			RelationshipKind: findingPayload.RelationshipKind,
			EnvironmentKind:  findingPayload.EnvironmentKind,
			Remediation: model.Remediation{
				ShortDescription: findingPayload.Remediation.ShortDescription,
				LongDescription:  findingPayload.Remediation.LongDescription,
				ShortRemediation: findingPayload.Remediation.ShortRemediation,
				LongRemediation:  findingPayload.Remediation.LongRemediation,
			},
		})
	}
	return graphExtension
}
