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

// OpenGraphSchemaIngest - handles incoming graph extension upsert requests
func (s Resources) OpenGraphSchemaIngest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
		err error

		updated bool

		extractExtensionData  func(file io.Reader) (model.GraphExtensionPayload, error)
		graphExtensionPayload model.GraphExtensionPayload
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
	} else if graphExtensionInput, err = graphExtensionPayload.ToGraphExtensionInput(); err != nil {
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

// extractExtensionDataFromJSON - extracts a model.GraphExtensionPayload from the incoming payload. Will return an error
// if the decoder fails to decode the payload.
func extractExtensionDataFromJSON(payload io.Reader) (model.GraphExtensionPayload, error) {
	var graphExtension model.GraphExtensionPayload

	if normFile, err := bomenc.NormalizeToUTF8(payload); err != nil {
		return graphExtension, fmt.Errorf("failed to normalize json payload: %w", err)
	} else if err = json.NewDecoder(normFile).Decode(&graphExtension); err != nil {
		return graphExtension, fmt.Errorf("unable to decode graph extension payload: %w", err)
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
