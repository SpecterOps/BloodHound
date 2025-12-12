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
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
	"github.com/specterops/bloodhound/cmd/api/src/model"
)

func (s Resources) OpenGraphSchemaIngest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx         = request.Context()
		err         error
		graphSchema model.GraphSchema
	)

	err = json.NewDecoder(request.Body).Decode(&graphSchema)
	if err != nil {
		// return 400
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("unable to parse opengraph schema: %v", err), request), response)
		return
	}

	/*
		1. payload hits endpoint and assume valid auth and method
		2. determine schema extension format based on content-type header
		3. either parse file or parse json schema into schema_extension api model
		    - can be json file, form data or zip with json file
		    - cant decode = 400
		4. validate schema extension api model
		   - ensure it has a non-empty extension
		   - ensure node/edge kind and property slices arent empty
		5. Pass extension schema to service layer
	*/

	err = s.openGraphSchemaService.UpsertGraphSchemaExtension(ctx, graphSchema)
	if err != nil {
		api.WriteErrorResponse(ctx, api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("unable to update graph schema: %v", err), request), response)
		return
	}

	response.WriteHeader(http.StatusCreated)
}
