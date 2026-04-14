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

// ParseClientVersion extracts the client type from a user agent string and delegates to
// the appropriate version parser. Supported formats:
//   - azurehound/vX.Y.Z[-rcN][+docker]
//   - openhound/vX.Y.Z[-rcN]
//   - sharphound/X.Y.Z.W[-rcN]
func ParseClientVersion(userAgent string) (ClientVersion, error) {
	clientType, rawVersion, err := parseClientUserAgent(userAgent)
	if err != nil {
		return ClientVersion{}, err
	}

	switch clientType {
	case ClientTypeAzureHound, ClientTypeOpenHound:
		return parseCollectorVersion(clientType, rawVersion)
	case ClientTypeSharpHound:
		return parseSharpHoundVersion(rawVersion)
	default:
		return ClientVersion{}, ErrInvalidClientType
	}
}

// clientVersionFromSemver maps a parsed semver.Version into a ClientVersion struct.
func clientVersionFromSemver(clientType ClientType, parsedVersion *semver.Version, extra int) ClientVersion {
	return ClientVersion{
		ClientType:    clientType,
		Major:         int(parsedVersion.Major()),
		Minor:         int(parsedVersion.Minor()),
		Patch:         int(parsedVersion.Patch()),
		Extra:         extra,
		Prerelease:    parsedVersion.Prerelease(),
		BuildMetadata: parsedVersion.Metadata(),
	}
}

// parseClientUserAgent splits a user agent string into its client type and raw version component.
// Returns ErrInvalidClientType if the user agent does not match a known client prefix.
func parseClientUserAgent(userAgent string) (ClientType, string, error) {
	switch {
	case strings.HasPrefix(userAgent, "azurehound/"):
		rawVersion, _ := strings.CutPrefix(userAgent, "azurehound/")
		return ClientTypeAzureHound, rawVersion, nil
	case strings.HasPrefix(userAgent, "openhound/"):
		rawVersion, _ := strings.CutPrefix(userAgent, "openhound/")
		return ClientTypeOpenHound, rawVersion, nil
	case strings.HasPrefix(userAgent, "sharphound/"):
		rawVersion, _ := strings.CutPrefix(userAgent, "sharphound/")
		return ClientTypeSharpHound, rawVersion, nil
	default:
		return 0, "", ErrInvalidClientType
	}
}

// parseCollectorVersion parses a raw version string for AzureHound or OpenHound collectors
// using semver, then validates that the prerelease and build metadata conform to allowed values.
func parseCollectorVersion(clientType ClientType, rawVersion string) (ClientVersion, error) {
	version := ClientVersion{ClientType: clientType}

	if rawVersion == "" {
		return version, ErrInvalidCollectorVersion
	}

	parsedVersion, err := parseSemverVersion(rawVersion)
	if err != nil {
		return version, ErrInvalidCollectorVersion
	}

	if !isValidCollectorSemver(parsedVersion, clientType) {
		return version, ErrInvalidCollectorVersion
	}

	return clientVersionFromSemver(clientType, parsedVersion, 0), nil
}

// parseSharpHoundVersion parses a raw version string in SharpHound's X.Y.Z.W[-rcN] format.
// The fourth numeric component is stored as Extra, and the first three are parsed via semver.
func parseSharpHoundVersion(rawVersion string) (ClientVersion, error) {
	version := ClientVersion{ClientType: ClientTypeSharpHound}

	if rawVersion == "" {
		return version, ErrInvalidSharpHoundVersion
	}

	normalizedVersion, extra, err := splitSharpHoundVersion(rawVersion)
	if err != nil {
		return version, ErrInvalidSharpHoundVersion
	}

	parsedVersion, err := parseSemverVersion(normalizedVersion)
	if err != nil {
		return version, ErrInvalidSharpHoundVersion
	}

	if !isValidSharpHoundSemver(parsedVersion) {
		return version, ErrInvalidSharpHoundVersion
	}

	return clientVersionFromSemver(ClientTypeSharpHound, parsedVersion, extra), nil
}

// splitSharpHoundVersion splits a SharpHound version string into a semver-compatible version
// and the fourth numeric component (extra). For example, "2.1.0.3-rc1" returns ("2.1.0-rc1", 3, nil).
func splitSharpHoundVersion(rawVersion string) (string, int, error) {
	sharpHoundParts := strings.Split(rawVersion, ".")

	if len(sharpHoundParts) != 4 {
		return "", 0, ErrInvalidSharpHoundVersion
	}

	extraPart, prereleasePart, _ := strings.Cut(sharpHoundParts[3], "-")
	extra, err := strconv.Atoi(extraPart)
	if err != nil {
		return "", 0, ErrInvalidSharpHoundVersion
	}

	return strings.Join(sharpHoundParts[:3], ".") + buildPrereleaseSuffix(prereleasePart), extra, nil
}

// buildPrereleaseSuffix returns a prerelease suffix prefixed with "-" for use in semver strings,
// or an empty string if no prerelease is present.
func buildPrereleaseSuffix(prerelease string) string {
	if prerelease == "" {
		return ""
	}

	return "-" + prerelease
}

// isValidCollectorSemver validates that a parsed semver version conforms to collector-specific rules:
//   - Prerelease must be empty or match the format "rcN" where N >= 1.
//   - AzureHound build metadata must be empty or "docker".
//   - OpenHound build metadata must be empty.
func isValidCollectorSemver(parsedVersion *semver.Version, clientType ClientType) bool {
	if !isValidRCPrerelease(parsedVersion.Prerelease()) {
		return false
	}

	if clientType == ClientTypeAzureHound {
		return parsedVersion.Metadata() == "" || parsedVersion.Metadata() == "docker"
	}

	return parsedVersion.Metadata() == ""
}

// isValidRCPrerelease validates that a prerelease string is either empty or matches the
// format "rcN" where N is an integer >= 1.
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

	num, err := strconv.Atoi(prereleaseNumber)
	if err != nil || num < 1 {
		return false
	}

	return true
}

// isValidSharpHoundSemver validates that a parsed semver version conforms to SharpHound rules:
// prerelease must be empty or "rcN" where N >= 1, and build metadata must be empty.
func isValidSharpHoundSemver(parsedVersion *semver.Version) bool {
	return isValidRCPrerelease(parsedVersion.Prerelease()) && parsedVersion.Metadata() == ""
}

// parseSemverVersion strips an optional leading "v" and parses the version using semver.StrictNewVersion.
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
