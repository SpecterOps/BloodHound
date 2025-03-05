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
	"encoding/json"
	"fmt"
	"github.com/specterops/bloodhound/src/version"
	"golang.org/x/exp/slices"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
)

const (
	// TODO: Paths should probably be a configuration option
	CollectorZipFileTemplate    = "%s-%s.zip"
	CollectorSHA256FileTemplate = "%s-%s.zip.sha256"

	CollectorTypePathParameterName       = "collector_type"
	CollectorReleaseTagPathParameterName = "release_tag"

	osCaptureGroup   = 2
	archCaptureGroup = 3
	shaCaptureGroup  = 4
)

var (
	releaseParsingRegex = regexp.MustCompile(`^(\w+)-(\w+)-(\w+)\.zip(\.sha256)?$`)
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
		slog.ErrorContext(request.Context(), fmt.Sprintf("Manifest doesn't exist for %s collector", collectorType))
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
			slog.ErrorContext(request.Context(), fmt.Sprintf("Manifest doesn't exist for %s collector", collectorType))
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			return
		} else {
			fileName = fmt.Sprintf(CollectorZipFileTemplate, collectorType, collectorManifest.Latest)
		}
	} else {
		fileName = fmt.Sprintf(CollectorZipFileTemplate, collectorType, releaseTag)
	}

	if data, err := os.ReadFile(filepath.Join(s.Config.CollectorsDirectory(), collectorType, fileName)); err != nil {
		slog.ErrorContext(request.Context(), fmt.Sprintf("Could not open collector file for download: %v", err))
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
			slog.ErrorContext(request.Context(), fmt.Sprintf("Manifest doesn't exist for %s collector", collectorType))
			api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
			return
		} else {
			fileName = fmt.Sprintf(CollectorSHA256FileTemplate, collectorType, collectorManifest.Latest)
		}
	} else {
		fileName = fmt.Sprintf(CollectorSHA256FileTemplate, collectorType, releaseTag)
	}

	if data, err := os.ReadFile(filepath.Join(s.Config.CollectorsDirectory(), collectorType, fileName)); err != nil {
		slog.ErrorContext(request.Context(), fmt.Sprintf("Could not open collector file for download: %v", err))
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else {
		api.WriteBinaryResponse(request.Context(), data, fileName, http.StatusOK, response)
	}
}

func (s *Resources) GetKennelManifest(response http.ResponseWriter, request *http.Request) {
	var (
		sharphoundUrl    = "https://api.github.com/repos/SpecterOps/SharpHound/releases"
		sharphoundResult []GitHubRelease

		azurehoundUrl    = "https://api.github.com/repos/SpecterOps/AzureHound/releases"
		azurehoundResult []GitHubRelease

		manifest = Manifest{}
	)

	if req, err := http.NewRequest("GET", sharphoundUrl, nil); err != nil {
		slog.ErrorContext(request.Context(), "Failed creating new request", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else if res, err := http.DefaultClient.Do(req); err != nil {
		slog.ErrorContext(request.Context(), "Failed completing http request", "destination_url", sharphoundUrl, "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else if err := json.NewDecoder(res.Body).Decode(&sharphoundResult); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else if collecterInfo, err := parseResults(sharphoundResult); err != nil {
		slog.ErrorContext(request.Context(), "Failed parsing sharphound manifest", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else {
		slices.SortFunc(collecterInfo, func(a, b Release) bool {
			return a.Version.LessThan(b.Version)
		})
		manifest.Sharphound = collecterInfo[:min(5, len(collecterInfo))]
	}

	if req, err := http.NewRequest("GET", azurehoundUrl, nil); err != nil {
		slog.ErrorContext(request.Context(), "Failed creating new request", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else if res, err := http.DefaultClient.Do(req); err != nil {
		slog.ErrorContext(request.Context(), "Failed completing http request", "destination_url", azurehoundUrl, "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if err := json.NewDecoder(res.Body).Decode(&azurehoundResult); err != nil {
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
	} else if collecterInfo, err := parseResults(azurehoundResult); err != nil {
		slog.ErrorContext(request.Context(), "Failed parsing azurehound manifest", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else {
		slices.SortFunc(collecterInfo, func(a, b Release) bool {
			return a.Version.GreaterThan(b.Version)
		})
		manifest.Azurehound = collecterInfo[:min(5, len(collecterInfo))]
	}

	api.WriteJSONResponse(request.Context(), manifest, http.StatusOK, response)
}

func parseResults(githubReleases []GitHubRelease) ([]Release, error) {
	var (
		releases = make([]Release, 0)
	)

	for _, githubRelease := range githubReleases {
		releaseAssets := make(map[string]*ReleaseAsset)

		releaseVersion, err := version.Parse(githubRelease.TagName)
		if err != nil {
			continue
		}

		for _, asset := range githubRelease.Assets {
			if strings.Contains(asset.Name, "debug") {
				continue
			}

			if matches := releaseParsingRegex.FindAllStringSubmatch(asset.Name, 1); len(matches) != 1 {
				continue
			} else {
				releaseParts := matches[0]

				if releaseParts[shaCaptureGroup] == ".sha256" {
					if releaseAsset, found := releaseAssets[strings.TrimSuffix(asset.Name, ".sha256")]; found {
						releaseAsset.ChecksumDownload = asset.BrowserDownloadURL
					} else {
						releaseAssets[strings.TrimSuffix(asset.Name, ".sha256")] = &ReleaseAsset{
							ChecksumDownload: asset.BrowserDownloadURL,
						}
					}
				} else {
					if releaseAsset, found := releaseAssets[asset.Name]; found {
						releaseAsset.Name = asset.Name
						releaseAsset.DownloadUrl = asset.BrowserDownloadURL
						releaseAsset.Os = releaseParts[osCaptureGroup]
						releaseAsset.Arch = releaseParts[archCaptureGroup]
					} else {
						releaseAssets[asset.Name] = &ReleaseAsset{
							Name:        asset.Name,
							DownloadUrl: asset.BrowserDownloadURL,
							Os:          releaseParts[osCaptureGroup],
							Arch:        releaseParts[archCaptureGroup],
						}
					}
				}
			}
		}

		assetsList := make([]*ReleaseAsset, 0)
		for _, asset := range releaseAssets {
			assetsList = append(assetsList, asset)
		}

		if len(assetsList) > 0 {
			releases = append(releases, Release{
				Version:       releaseVersion,
				ReleaseDate:   githubRelease.PublishedAt,
				ReleaseAssets: assetsList,
			})
		}
	}

	return releases, nil
}

type Contents struct {
	Key          string    `xml:"Key"`
	LastModified time.Time `xml:"LastModified"`
}

type Release struct {
	Version       version.Version `json:"version"`
	ReleaseDate   time.Time       `json:"release_date"`
	ReleaseAssets []*ReleaseAsset `json:"release_assets"`
}

type ReleaseAsset struct {
	Name             string `json:"name"`
	DownloadUrl      string `json:"download_url"`
	ChecksumDownload string `json:"checksum_download"`
	Os               string `json:"os"`
	Arch             string `json:"arch"`
}

type Manifest struct {
	Sharphound []Release `json:"sharphound"`
	Azurehound []Release `json:"azurehound"`
}

type GitHubRelease struct {
	TagName     string    `json:"tag_name"`
	Assets      []Asset   `json:"assets"`
	PublishedAt time.Time `json:"published_at"`
}

type Asset struct {
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	BrowserDownloadURL string `json:"browser_download_url"`
}
