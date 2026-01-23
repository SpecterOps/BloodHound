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
	"net/http"

	"github.com/specterops/bloodhound/cmd/api/src/api"
)

//go:generate go run go.uber.org/mock/mockgen -copyright_file ../../../../../LICENSE.header -destination=./mocks/graphschemaextensions.go -package=mocks . OpenGraphSchemaService
type OpenGraphSchemaService interface {
	UpsertGraphSchemaExtension(ctx context.Context, req GraphSchemaExtension) error
}

type GraphSchemaExtension struct {
	Environments []Environment `json:"environments"`
	Findings     []Finding     `json:"findings"`
}

type Environment struct {
	EnvironmentKind string   `json:"environmentKind"`
	SourceKind      string   `json:"sourceKind"`
	PrincipalKinds  []string `json:"principalKinds"`
}

type Finding struct {
	Name             string      `json:"name"`
	DisplayName      string      `json:"displayName"`
	SourceKind       string      `json:"sourceKind"`
	RelationshipKind string      `json:"relationshipKind"`
	EnvironmentKind  string      `json:"environmentKind"`
	Remediation      Remediation `json:"remediation"`
}

type Remediation struct {
	ShortDescription string `json:"shortDescription"`
	LongDescription  string `json:"longDescription"`
	ShortRemediation string `json:"shortRemediation"`
	LongRemediation  string `json:"longRemediation"`
}

// TODO: Implement this - skeleton endpoint to simply test the handler.
func (s Resources) OpenGraphSchemaIngest(response http.ResponseWriter, request *http.Request) {
	var (
		ctx = request.Context()
	)

	var req GraphSchemaExtension
	if err := json.NewDecoder(request.Body).Decode(&req); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, api.ErrorResponsePayloadUnmarshalError, request), response)
		return
	}

	if err := s.openGraphSchemaService.UpsertGraphSchemaExtension(ctx, req); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, fmt.Sprintf("error upserting graph schema extension: %v", err), request), response)
		return
	}

	response.WriteHeader(http.StatusCreated)
}
