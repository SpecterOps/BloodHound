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
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/specterops/bloodhound/src/api"
	"github.com/specterops/bloodhound/src/version"
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
		sharphoundUrl = "https://api.github.com/repos/SpecterOps/SharpHound/releases?per_page=10"
		azurehoundUrl = "https://api.github.com/repos/SpecterOps/AzureHound/releases?per_page=10"
		manifest      = Manifest{}
	)

	if sharphoundResult, err := queryGithub(sharphoundUrl); err != nil {
		slog.ErrorContext(request.Context(), "Failed querying github for sharphound assets", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else if releases, err := parseGithubResults(sharphoundResult); err != nil {
		slog.ErrorContext(request.Context(), "Failed parsing sharphound github releases", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else {
		slices.SortFunc(releases, SortReleasesByVersionDesc)
		manifest.Sharphound = releases[:min(5, len(releases))]
	}

	if azurehoundResult, err := queryGithub(azurehoundUrl); err != nil {
		slog.ErrorContext(request.Context(), "Failed querying github for azurehound assets", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else if releases, err := parseGithubResults(azurehoundResult); err != nil {
		slog.ErrorContext(request.Context(), "Failed parsing azurehound github releases", "error", err)
		api.WriteErrorResponse(request.Context(), api.BuildErrorResponse(http.StatusInternalServerError, api.ErrorResponseDetailsInternalServerError, request), response)
		return
	} else {
		slices.SortFunc(releases, SortReleasesByVersionDesc)
		manifest.Azurehound = releases[:min(5, len(releases))]
	}

	api.WriteBasicResponse(request.Context(), manifest, http.StatusOK, response)
}

type Manifest struct {
	Sharphound []Release `json:"sharphound"`
	Azurehound []Release `json:"azurehound"`
}

type Release struct {
	Version       version.Version `json:"version"`
	VersionMeta   VersionMeta     `json:"version_meta"`
	ReleaseDate   time.Time       `json:"release_date"`
	ReleaseAssets []*ReleaseAsset `json:"release_assets"`
}

type VersionMeta struct {
	Major      int    `json:"major"`
	Minor      int    `json:"minor"`
	Patch      int    `json:"patch"`
	Prerelease string `json:"prerelease"`
}

type ReleaseAsset struct {
	Name                string `json:"name"`
	DownloadUrl         string `json:"download_url"`
	ChecksumDownloadUrl string `json:"checksum_download_url"`
	Os                  string `json:"os"`
	Arch                string `json:"arch"`
}

type githubRelease struct {
	TagName     string        `json:"tag_name"`
	Assets      []githubAsset `json:"assets"`
	PublishedAt time.Time     `json:"published_at"`
}

type githubAsset struct {
	Name               string `json:"name"`
	ContentType        string `json:"content_type"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

func SortReleasesByVersionDesc(a, b Release) int {
	if a.Version.Equals(b.Version) {
		return 0
	} else if a.Version.LessThan(b.Version) {
		return 1
	} else {
		return -1
	}
}

func queryGithub(url string) ([]githubRelease, error) {
	var result []githubRelease

	if req, err := http.NewRequest("GET", url, nil); err != nil {
		return nil, err
	} else if res, err := http.DefaultClient.Do(req); err != nil {
		return nil, err
	} else if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func parseGithubResults(githubReleases []githubRelease) ([]Release, error) {
	var (
		releases = make([]Release, 0)
	)

	for _, ghRelease := range githubReleases {
		releaseAssets := make(map[string]*ReleaseAsset)

		releaseVersion, err := version.Parse(ghRelease.TagName)
		if err != nil {
			continue
		}

		for _, ghAsset := range ghRelease.Assets {
			if matches := releaseParsingRegex.FindAllStringSubmatch(ghAsset.Name, 1); len(matches) != 1 {
				continue
			} else {
				releaseParts := matches[0]

				if releaseParts[shaCaptureGroup] == ".sha256" {
					name := strings.TrimSuffix(ghAsset.Name, ".sha256")
					releaseAsset, found := releaseAssets[name]
					if !found {
						releaseAsset = &ReleaseAsset{}
						releaseAssets[name] = releaseAsset
					}

					releaseAsset.ChecksumDownloadUrl = ghAsset.BrowserDownloadURL
				} else {
					releaseAsset, found := releaseAssets[ghAsset.Name]
					if !found {
						releaseAsset = &ReleaseAsset{}
						releaseAssets[ghAsset.Name] = releaseAsset
					}

					releaseAsset.Name = ghAsset.Name
					releaseAsset.DownloadUrl = ghAsset.BrowserDownloadURL
					releaseAsset.Os = releaseParts[osCaptureGroup]
					releaseAsset.Arch = releaseParts[archCaptureGroup]
				}
			}
		}

		if len(releaseAssets) == 0 {
			continue
		}

		assetsList := make([]*ReleaseAsset, 0)
		for _, asset := range releaseAssets {
			assetsList = append(assetsList, asset)
		}

		releases = append(releases, Release{
			Version:       releaseVersion,
			VersionMeta:   VersionMeta(releaseVersion),
			ReleaseDate:   ghRelease.PublishedAt,
			ReleaseAssets: assetsList,
		})
	}

	return releases, nil
}
