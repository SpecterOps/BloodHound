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

package utils

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	"github.com/gofrs/uuid"
	"github.com/specterops/bloodhound/packages/go/headers"
	"github.com/specterops/bloodhound/packages/go/mediatypes"
)

var (
	ErrInvalidSharpHoundVersion   = errors.New("invalid sharphound version string")
	ErrInvalidCollectorVersion    = errors.New("invalid collector version string")
	ErrRecommendSharphoundVersion = errors.New("please upgrade to sharphound v2.0.3 or above")
	ErrInvalidClientType          = errors.New("invalid client type")
	ErrInvalidUUID                = errors.New("invalid UUID")
)

type ClientType int

const (
	ClientTypeSharpHound ClientType = 1
	ClientTypeAzureHound ClientType = 2
	ClientTypeOpenHound  ClientType = 3
)

type ClientVersion struct {
	ClientType    ClientType
	Major         int
	Minor         int
	Patch         int
	Extra         int
	Prerelease    string
	BuildMetadata string
}

// IsValidClientVersion checks the version from a user agent to ensure it's a valid UserAgent and that
// the version of the client is not EOL (currently SHS v1.x and SHS < v2.0.3).
// Returns the parsed ClientVersion and an error when invalid.
func IsValidClientVersion(userAgent string) (ClientVersion, error) {
	if version, err := ParseClientVersion(userAgent); err != nil {
		return version, fmt.Errorf("error parsing client version: %w", err)
	} else if version.ClientType == ClientTypeAzureHound {
		return version, nil
	} else if version.ClientType == ClientTypeOpenHound {
		return version, nil
	} else if version.ClientType == ClientTypeSharpHound {
		if version.Major < 2 {
			return version, fmt.Errorf("sharphound v1.x detected: %w", ErrRecommendSharphoundVersion)
		} else if version.Major == 2 && version.Minor == 0 && version.Patch < 3 {
			return version, fmt.Errorf("sharphound v2.0.2 or lower detected: %w", ErrRecommendSharphoundVersion)
		} else {
			return version, nil
		}
	} else { // unknown client type
		return version, ErrInvalidClientType
	}
}

func ParseClientVersion(userAgent string) (ClientVersion, error) {
	if strings.HasPrefix(userAgent, "azurehound") || strings.HasPrefix(userAgent, "openhound") {
		return ParseCollectorVersion(userAgent)
	} else if strings.HasPrefix(userAgent, "sharphound") {
		return ParseSharpHoundVersion(userAgent)
	} else {
		return ClientVersion{}, ErrInvalidClientType
	}
}

func ParseCollectorVersion(userAgent string) (ClientVersion, error) {
	var (
		err           error
		parsedVersion *semver.Version
		rawVersion    string
		version       = ClientVersion{Major: 0, Minor: 0, Patch: 0, Extra: 0, Prerelease: "", BuildMetadata: ""}
	)

	if strings.HasPrefix(userAgent, "azurehound") {
		version.ClientType = ClientTypeAzureHound
		rawVersion = strings.TrimPrefix(userAgent, "azurehound/")
	} else if strings.HasPrefix(userAgent, "openhound") {
		version.ClientType = ClientTypeOpenHound
		rawVersion = strings.TrimPrefix(userAgent, "openhound/")
	} else {
		return ClientVersion{}, ErrInvalidClientType
	}

	if rawVersion == userAgent || rawVersion == "" {
		return version, ErrInvalidCollectorVersion
	} else if parsedVersion, err = parseSemverVersion(rawVersion); err != nil {
		return version, ErrInvalidCollectorVersion
	} else if !isValidCollectorSemver(parsedVersion, version.ClientType) {
		return version, ErrInvalidCollectorVersion
	} else {
		version.Major = int(parsedVersion.Major())
		version.Minor = int(parsedVersion.Minor())
		version.Patch = int(parsedVersion.Patch())
		version.Extra = 0
		version.Prerelease = parsedVersion.Prerelease()
		version.BuildMetadata = parsedVersion.Metadata()
		return version, nil
	}
}

func ParseSharpHoundVersion(userAgent string) (ClientVersion, error) {
	var (
		err             error
		extra           int
		parsedVersion   *semver.Version
		prereleasePart  string
		rawVersion      string
		sharpHoundParts []string
		version         = ClientVersion{
			ClientType:    ClientTypeSharpHound,
			Major:         0,
			Minor:         0,
			Patch:         0,
			Extra:         0,
			Prerelease:    "",
			BuildMetadata: "",
		}
	)

	rawVersion = strings.TrimPrefix(userAgent, "sharphound/")
	if rawVersion == userAgent || rawVersion == "" {
		return version, ErrInvalidSharpHoundVersion
	}

	sharpHoundParts = strings.Split(rawVersion, ".")
	if len(sharpHoundParts) != 4 {
		return version, ErrInvalidSharpHoundVersion
	}

	sharpHoundParts[3], prereleasePart, _ = strings.Cut(sharpHoundParts[3], "-")

	if extra, err = strconv.Atoi(sharpHoundParts[3]); err != nil {
		return version, ErrInvalidSharpHoundVersion
	}

	if parsedVersion, err = parseSemverVersion(strings.Join(sharpHoundParts[:3], ".") + buildPrereleaseSuffix(prereleasePart)); err != nil {
		return version, ErrInvalidSharpHoundVersion
	} else if !isValidSharpHoundSemver(parsedVersion) {
		return version, ErrInvalidSharpHoundVersion
	} else {
		version.Major = int(parsedVersion.Major())
		version.Minor = int(parsedVersion.Minor())
		version.Patch = int(parsedVersion.Patch())
		version.Extra = extra
		version.Prerelease = parsedVersion.Prerelease()
		return version, nil
	}
}

func buildPrereleaseSuffix(prerelease string) string {
	if prerelease == "" {
		return ""
	}

	return "-" + prerelease
}

func isValidCollectorSemver(parsedVersion *semver.Version, clientType ClientType) bool {
	if !isValidRCPrerelease(parsedVersion.Prerelease()) {
		return false
	}

	if clientType == ClientTypeAzureHound {
		if parsedVersion.Prerelease() != "" && parsedVersion.Metadata() != "" {
			return false
		}

		return parsedVersion.Metadata() == "" || parsedVersion.Metadata() == "docker"
	}

	return parsedVersion.Metadata() == ""
}

func isValidRCPrerelease(prerelease string) bool {
	if prerelease == "" {
		return true
	}

	if !strings.HasPrefix(prerelease, "rc") {
		return false
	}

	prereleaseNumber := strings.TrimPrefix(prerelease, "rc")
	if prereleaseNumber == "" {
		return false
	}

	if _, err := strconv.Atoi(prereleaseNumber); err != nil {
		return false
	}

	return true
}

func isValidSharpHoundSemver(parsedVersion *semver.Version) bool {
	return isValidRCPrerelease(parsedVersion.Prerelease()) && parsedVersion.Metadata() == ""
}

func parseSemverVersion(rawVersion string) (*semver.Version, error) {
	return semver.StrictNewVersion(strings.TrimPrefix(rawVersion, "v"))
}

type JsonResult struct {
	Detail string `json:"detail,omitempty"`
	Msg    string `json:"msg,omitempty"`
}

func WriteResultJson(w http.ResponseWriter, r any) {
	result, _ := json.Marshal(r)
	w.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	w.WriteHeader(http.StatusOK)
	w.Write(result)
}

func WriteResultRawJson(w http.ResponseWriter, r []byte) {
	w.Header().Set(headers.ContentType.String(), mediatypes.ApplicationJson.String())
	w.WriteHeader(http.StatusOK)
	w.Write(r)
}

func GetPageParamsForGraphQuery(_ context.Context, params url.Values) (int, int, string, error) {
	if skip, err := GetSkipParamForGraphQuery(params); err != nil {
		return 0, 0, "", err
	} else if limit, err := GetLimitParamForGraphQuery(params); err != nil {
		return 0, 0, "", err
	} else {
		order := GetOrderForNeo4jQuery(params)
		return skip, limit, order, nil
	}
}

func GetSkipParamForGraphQuery(params url.Values) (int, error) {
	skip := 0

	if skipStr := params.Get("skip"); skipStr == "" {
		return 0, nil
	} else if parsedSkip, err := strconv.Atoi(skipStr); err != nil {
		return skip, err
	} else if parsedSkip < 0 {
		return 0, fmt.Errorf(ErrorInvalidSkip, parsedSkip)
	} else {
		return parsedSkip, nil
	}
}

func GetLimitParamForGraphQuery(params url.Values) (int, error) {
	limit := 0

	if limitStr := params.Get("limit"); limitStr == "" {
		return 10, nil
	} else if parsedLimit, err := strconv.Atoi(limitStr); err != nil {
		return limit, err
	} else if parsedLimit < 0 {
		return 0, fmt.Errorf(ErrorInvalidLimit, parsedLimit)
	} else {
		return parsedLimit, nil
	}
}

func GetOrderForNeo4jQuery(params url.Values) string {
	var (
		sortByColumns = params["sort_by"]
		order         []string
	)
	for _, column := range sortByColumns {
		var descending bool
		if string(column[0]) == "-" {
			descending = true
			column = column[1:]
		}

		// the cypher column name is objectid, not id
		if column == "id" {
			column = "objectid"
		}

		column = "n." + column

		if descending {
			order = append(order, column+" DESC")
		} else {
			order = append(order, column)
		}
	}
	return strings.Join(order, ", ")
}

const (
	HSTSSetting = "max-age=31536000; includeSubDomains; preload"

	// Header Templates
	ContentDispositionAttachmentTemplate = "attachment; filename=\"%s\""

	HeaderValueWait = "wait"

	ErrorInvalidSkip  string = "invalid skip: %v"
	ErrorInvalidLimit string = "invalid limit: %v"
)

func HeaderMatches(headers http.Header, key string, target ...string) bool {
	value := strings.ToLower(headers.Get(key))
	if value == "" {
		return false
	}
	for _, t := range target {
		if strings.Contains(value, strings.ToLower(t)) {
			return true
		}
	}
	return false
}

// ParseUUID parses a string into a uuid.UUID and returns an ErrInvalidUUID
// error if the string is not a valid UUID.
func ParseUUID(s string) (uuid.UUID, error) {
	parsedUUID, err := uuid.FromString(s)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%w: %w", ErrInvalidUUID, err)
	}
	return parsedUUID, nil
}

func IsValidEmail(maybeEmail string) bool {
	_, err := mail.ParseAddress(maybeEmail)
	return err == nil
}

// ReplaceFieldValueInJsonString replaces a field value at the root of a JSON object
// If the field does not exist in jsonString, it will effectively do nothing
func ReplaceFieldValueInJsonString(jsonString string, field string, value any) (string, error) {
	var unmarshaled map[string]any
	err := json.Unmarshal([]byte(jsonString), &unmarshaled)
	if err != nil {
		return "", err
	}

	if _, exists := unmarshaled[field]; exists {
		unmarshaled[field] = value
	}

	modifiedJson, err := json.Marshal(unmarshaled)
	if err != nil {
		return "", err
	}

	return string(modifiedJson), nil
}
