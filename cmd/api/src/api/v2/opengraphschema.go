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
	"math"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/database"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	"github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/bhlog/attr"
	"github.com/specterops/bloodhound/packages/go/bomenc"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
	"github.com/specterops/dawgs/graph"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemaextensions.go -package=mocks . OpenGraphSchemaService
type OpenGraphSchemaService interface {
	UpsertOpenGraphExtension(ctx context.Context, openGraphExtension model.GraphExtensionInput) (bool, error)
	ListExtensions(ctx context.Context) (model.GraphSchemaExtensions, error)
	DeleteExtension(ctx context.Context, extensionID int32) error
	GetEnvironmentKindsAndSchemaEnvironmentData(ctx context.Context, onlyBuiltin bool) (graph.Kinds, model.EnvironmentKindsToEnvironment, error)
	GetSchemaFindings(ctx context.Context, filters model.Filters, sort model.Sort, skip, limit int) ([]model.SchemaFinding, int, error)
}

type GraphExtensionPayload struct {
	GraphSchemaExtension         GraphSchemaExtensionPayload           `json:"schema"`
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
	Name          string                     `json:"name"`
	Description   string                     `json:"description"`
	IsTraversable bool                       `json:"is_traversable"` // indicates whether the edge-kind is a traversable path
	Info          map[string]KindInfoPayload `json:"info"`
}

type GraphSchemaNodeKindsPayload struct {
	Name          string                     `json:"name"`
	DisplayName   string                     `json:"display_name"`    // can be different from name but usually isn't other than Base/Entity
	Description   string                     `json:"description"`     // human-readable description of the node kind
	IsDisplayKind bool                       `json:"is_display_kind"` // indicates if this kind should supersede others and be displayed
	Icon          string                     `json:"icon"`            // font-awesome icon for the registered node kind
	IconColor     string                     `json:"color"`           // icon hex color
	Info          map[string]KindInfoPayload `json:"info"`
}

type KindInfoPayload struct {
	Title    string          `json:"title"`
	Position int             `json:"position"`
	Markdown json.RawMessage `json:"markdown"` // Raw: {"content": "..."}
}

type EnvironmentPayload struct {
	EnvironmentKind string   `json:"environment_kind"`
	SourceKind      string   `json:"source_kind"`
	PrincipalKinds  []string `json:"principal_kinds"`
}

type RelationshipFindingsPayload struct {
	Name             string             `json:"name"`
	DisplayName      string             `json:"display_name"`
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

	if request.Body == nil {
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

	var graphExtensionInput model.GraphExtensionInput

	if graphExtensionPayload, err = extractExtensionData(request.Body); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	} else if graphExtensionInput, err = convertGraphExtensionPayloadToGraphExtension(graphExtensionPayload); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		return
	} else if updated, err = s.OpenGraphSchemaService.UpsertOpenGraphExtension(ctx, graphExtensionInput); err != nil {
		switch {
		case strings.Contains(err.Error(), model.ErrGraphExtensionValidation.Error()) ||
			strings.Contains(err.Error(), model.ErrGraphExtensionBuiltIn.Error()):
			api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
		case strings.Contains(err.Error(), model.ErrGraphDBRefreshKinds.Error()):
			fallthrough
		default:
			slog.WarnContext(
				ctx,
				"Error updating open graph schema",
				attr.Error(err),
			)
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
	var graphExtension GraphExtensionPayload

	if normFile, err := bomenc.NormalizeToUTF8(payload); err != nil {
		return graphExtension, fmt.Errorf("failed to normalize json payload: %w", err)
	} else if err = json.NewDecoder(normFile).Decode(&graphExtension); err != nil {
		return graphExtension, fmt.Errorf("unable to decode graph extension payload: %w", err)
	}

	return graphExtension, nil
}

// parseInfoPayload converts the typed KindInfoPayload map to model.KindInfoInputs slice
func parseInfoPayload(infoPayload map[string]KindInfoPayload) (model.KindInfoInputs, error) {
	var (
		wrappedContent map[string]json.RawMessage
		contentBytes   []byte
		markdown       json.RawMessage
		err            error
	)

	if len(infoPayload) == 0 {
		return make(model.KindInfoInputs, 0), nil
	}

	result := make(model.KindInfoInputs, 0, len(infoPayload))

	for key, payload := range infoPayload {
		// Guard against int32 overflow before the cast below; a position outside the
		// int32 range would silently wrap and corrupt the stored value.
		if payload.Position > math.MaxInt32 || payload.Position < math.MinInt32 {
			return nil, fmt.Errorf("%w: position %d for info key %q is outside the int32 range", model.ErrInvalidKindInfoPosition, payload.Position, key)
		}

		// Handle empty markdown by providing empty content structure
		if len(payload.Markdown) == 0 {
			markdown = []byte(`{"content":""}`)
		} else {
			markdown = payload.Markdown
		}

		wrappedContent = map[string]json.RawMessage{
			"markdown": markdown,
		}

		if contentBytes, err = json.Marshal(wrappedContent); err != nil {
			return nil, fmt.Errorf("failed to wrap markdown content for info key %s: %w", key, err)
		}

		result = append(result, model.KindInfoInput{
			InfoKey:  key,
			Title:    payload.Title,
			Position: int32(payload.Position),
			Content:  contentBytes,
		})
	}

	return result, nil
}

// convertGraphExtensionPayloadToGraphExtension - converts the GraphExtensionInput view layer model to the service layer model.
func convertGraphExtensionPayloadToGraphExtension(payload GraphExtensionPayload) (model.GraphExtensionInput, error) {
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
			EnvironmentsInput:      make(model.EnvironmentsInput, 0),
		}
		infoInputs model.KindInfoInputs
		err        error
	)

	for _, nodeKindPayload := range payload.GraphSchemaNodeKinds {
		if infoInputs, err = parseInfoPayload(nodeKindPayload.Info); err != nil {
			return model.GraphExtensionInput{}, fmt.Errorf("error parsing node kind %s info: %w", nodeKindPayload.Name, err)
		}

		graphExtension.NodeKindsInput = append(graphExtension.NodeKindsInput,
			model.NodeInput{
				Name:          nodeKindPayload.Name,
				DisplayName:   nodeKindPayload.DisplayName,
				Description:   nodeKindPayload.Description,
				IsDisplayKind: nodeKindPayload.IsDisplayKind,
				Icon:          nodeKindPayload.Icon,
				IconColor:     nodeKindPayload.IconColor,
				Info:          infoInputs,
			})
	}
	for _, edgeKindPayload := range payload.GraphSchemaRelationshipKinds {
		if infoInputs, err = parseInfoPayload(edgeKindPayload.Info); err != nil {
			return model.GraphExtensionInput{}, fmt.Errorf("error parsing relationship kind %s info: %w", edgeKindPayload.Name, err)
		}

		graphExtension.RelationshipKindsInput = append(graphExtension.RelationshipKindsInput,
			model.RelationshipInput{
				Name:          edgeKindPayload.Name,
				Description:   edgeKindPayload.Description,
				IsTraversable: edgeKindPayload.IsTraversable,
				Info:          infoInputs,
			})
	}
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
	return graphExtension, nil
}

type ExtensionsResponse struct {
	Extensions []ExtensionInfo `json:"extensions"`
}

type ExtensionInfo struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Version   string `json:"version"`
	IsBuiltIn bool   `json:"is_builtin"`
}

func (s Resources) ListExtensions(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
	)

	if extensions, err := s.OpenGraphSchemaService.ListExtensions(ctx); err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else {
		var extensionsResponse = make([]ExtensionInfo, len(extensions))
		for i, extension := range extensions {
			extensionsResponse[i] = ExtensionInfo{
				ID:        extension.ID,
				Name:      extension.DisplayName,
				Version:   extension.Version,
				IsBuiltIn: extension.IsBuiltin,
				Namespace: extension.Namespace,
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

	if extID, err := strconv.ParseInt(extensionID, 10, 32); err != nil {
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
