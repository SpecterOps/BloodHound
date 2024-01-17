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

package v2

import (
	"net/http"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/version"
)

// VersionResponse holds data returned in a version query
type VersionResponse struct {
	API    APIVersions `json:"API"`
	Server string      `json:"server_version"`
}

// APIVersions holds the 2 supported API versions
type APIVersions struct {
	CurrentVersion    string `json:"current_version"`
	DeprecatedVersion string `json:"deprecated_version"`
}

// GetVersion returns the supported API versions
func GetVersion(response http.ResponseWriter, request *http.Request) {
	api.WriteBasicResponse(request.Context(), VersionResponse{
		API: APIVersions{
			CurrentVersion:    "v2",
			DeprecatedVersion: "none",
		},
		Server: version.GetVersion().String(),
	}, http.StatusOK, response)
}
