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
	"fmt"
	"net/http"
	"os"
	"path/filepath"

	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/log"

	"github.com/gorilla/mux"
)

const (
	// TODO: Paths should probably be a configuration option
	CollectorZipFileTemplate    = "%s-%s.zip"
	CollectorSHA256FileTemplate = "%s-%s.zip.sha256"

	CollectorTypePathParameterName       = "collector_type"
	CollectorReleaseTagPathParameterName = "release_tag"
)

type CollectorType string

const (
	CollectorTypeSharpHound CollectorType = "sharphound"
	CollectorTypeAzurehound CollectorType = "azurehound"
)

func (s CollectorType) String() string {
	switch s {
	case CollectorTypeAzurehound:
		return string(CollectorTypeAzurehound)
	case CollectorTypeSharpHound:
		return string(CollectorTypeSharpHound)
	default:
		return "InvalidCollectorType"
	}
}

// GetCollectorManifest provides a json manifest of versions for a collector {azurehound|sharphound}
func (s *Resources) GetCollectorManifest(response http.ResponseWriter, request *http.Request) {
	var (
		requestVars   = mux.Vars(request)
		collectorType = requestVars[CollectorTypePathParameterName]
	)

	if CollectorType(collectorType).String() == "InvalidCollectorType" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid collector type: %s", collectorType), request), response)
	} else if collectorManifest, ok := s.CollectorManifests[collectorType]; !ok {
		log.Errorf("Manifest doesn't exist for %s collector", collectorType)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		api.WriteBasicResponse(request.Context(), collectorManifest, http.StatusOK, response)
	}
}

// DownloadCollectorByVersion provides collector package by its semver or "latest" tag
func (s *Resources) DownloadCollectorByVersion(response http.ResponseWriter, request *http.Request) {
	var (
		requestVars   = mux.Vars(request)
		collectorType = requestVars[CollectorTypePathParameterName]
		releaseTag    = requestVars[CollectorReleaseTagPathParameterName]
		fileName      string
	)

	if CollectorType(collectorType).String() == "InvalidCollectorType" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid collector type: %s", collectorType), request), response)
	} else if releaseTag == "latest" {
		if collectorManifest, ok := s.CollectorManifests[collectorType]; !ok {
			log.Errorf("Manifest doesn't exist for %s collector", collectorType)
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			return
		} else {
			fileName = fmt.Sprintf(CollectorZipFileTemplate, collectorType, collectorManifest.Latest)
		}
	} else {
		fileName = fmt.Sprintf(CollectorZipFileTemplate, collectorType, releaseTag)
	}

	if data, err := os.ReadFile(filepath.Join(s.Config.CollectorsDirectory(), collectorType, fileName)); err != nil {
		log.Errorf("Could not open collector file for download: %v", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		api.WriteBinaryResponse(request.Context(), data, fileName, http.StatusOK, response)
	}
}

// DownloadCollectorChecksumByVersion provides collector checksum file for a given semver or "latest" tag
func (s *Resources) DownloadCollectorChecksumByVersion(response http.ResponseWriter, request *http.Request) {
	var (
		requestVars   = mux.Vars(request)
		collectorType = requestVars[CollectorTypePathParameterName]
		releaseTag    = requestVars[CollectorReleaseTagPathParameterName]
		fileName      string
	)

	if CollectorType(collectorType).String() == "InvalidCollectorType" {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusBadRequest, fmt.Sprintf("Invalid collector type: %s", collectorType), request), response)
	} else if releaseTag == "latest" {
		if collectorManifest, ok := s.CollectorManifests[collectorType]; !ok {
			log.Errorf("Manifest doesn't exist for %s collector", collectorType)
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			return
		} else {
			fileName = fmt.Sprintf(CollectorSHA256FileTemplate, collectorType, collectorManifest.Latest)
		}
	} else {
		fileName = fmt.Sprintf(CollectorSHA256FileTemplate, collectorType, releaseTag)
	}

	if data, err := os.ReadFile(filepath.Join(s.Config.CollectorsDirectory(), collectorType, fileName)); err != nil {
		log.Errorf("Could not open collector file for download: %v", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		api.WriteBinaryResponse(request.Context(), data, fileName, http.StatusOK, response)
	}
}
