// Copyright 2023 Specter Ops, Inc.
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

package tools

import (
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/log"
)

type LoggingError struct {
	Error string `json:"error"`
}

type LoggingLevel struct {
	Level string `json:"level"`
}

func GetLoggingDetails(response http.ResponseWriter, request *http.Request) {
	api.WriteJSONResponse(request.Context(), LoggingLevel{
		Level: log.GlobalLevel().String(),
	}, http.StatusOK, response)
}

func PutLoggingDetails(response http.ResponseWriter, request *http.Request) {
	var level LoggingLevel

	if err := api.ReadJSONRequestPayloadLimited(&level, request); err != nil {
		api.WriteJSONResponse(request.Context(), LoggingError{
			Error: err.Error(),
		}, http.StatusBadRequest, response)
	} else if level, err := log.ParseLevel(level.Level); err != nil {
		api.WriteJSONResponse(request.Context(), LoggingError{
			Error: err.Error(),
		}, http.StatusBadRequest, response)
	} else {
		log.SetGlobalLevel(level)
		response.WriteHeader(http.StatusOK)
	}
}
