// Copyright 2025 Specter Ops, Inc.
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
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/auth"
	ctx2 "github.com/specterops/bloodhound/cmd/api/src/ctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/cmd/api/src/model/ingest"
	bhUtils "github.com/specterops/bloodhound/cmd/api/src/utils"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemaextensions.go -package=mocks . OpenGraphSchemaService

type OpenGraphSchemaService interface {
	UpsertOpenGraphExtension(ctx context.Context, graphSchema model.GraphSchema) (bool, error)
}

func (s Resources) OpenGraphSchemaIngest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx  = request.Context()
		err  error
		flag appcfg.FeatureFlag

		updated bool

		extractExtensionData func(file io.Reader) (model.GraphSchema, error)
		graphSchemaPayload   model.GraphSchema
	)

	// TODO: what to return if feature flag is not enabled
	if flag, err = s.DB.GetFlagByKey(ctx, appcfg.FeatureOpenGraphExtensionManagement); err != nil {
		api.HandleDatabaseError(request, response, err)
	} else if !flag.Enabled {
		response.WriteHeader(http.StatusNotFound)
	} else if user, isUser := auth.GetUserFromAuthCtx(ctx2.FromRequest(request).AuthCtx); !isUser {
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
		case bhUtils.HeaderMatches(request.Header, headers.ContentType.String(), ingest.AllowedZipFileUploadTypes...):
			fallthrough
		//	extractExtensionData = extractExtensionDataFromZipFile - will be needed for a future
		default:
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusUnsupportedMediaType,
				fmt.Sprintf("%s; Content type must be application/json",
					fmt.Errorf("invalid content-type: %s", request.Header[headers.ContentType.String()])), request), response)
			return
		}

		if graphSchemaPayload, err = extractExtensionData(request.Body); err != nil {
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, err.Error(), request), response)
			return
		}

		if updated, err = s.openGraphSchemaService.UpsertOpenGraphExtension(ctx, graphSchemaPayload); err != nil {
			switch {
			// TODO: more error types (ex: validation)
			default:
				api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("unable to update graph schema: %v", err), request), response)
				return
			}
		} else if updated {
			response.WriteHeader(http.StatusOK)
		} else {
			response.WriteHeader(http.StatusCreated)
		}
	}
}

// extractExtensionDataFromJSON - extracts a model.GraphSchema from the incoming payload. Will return an error if there
// are any extra fields or if the decoder fails to decode the payload.
func extractExtensionDataFromJSON(payload io.Reader) (model.GraphSchema, error) {
	var (
		err         error
		decoder     = json.NewDecoder(payload)
		graphSchema model.GraphSchema
	)
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&graphSchema); err != nil {
		return graphSchema, err
	}
	return graphSchema, nil
}
