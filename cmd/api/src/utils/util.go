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
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/specterops/bloodhound/errors"
	"github.com/specterops/bloodhound/headers"
	"github.com/specterops/bloodhound/mediatypes"
)

var ErrInvalidSharpHoundVersion = errors.New("invalid sharphound version string")
var ErrInvalidAzureHoundVersion = errors.New("invalid azurehound version string")
var ErrRecommendSharphoundVersion = errors.New("please upgrade to sharphound v2.0.3 or above")
var ErrInvalidClientType = errors.New("invalid client type")

type ClientType int

const (
	ClientTypeSharpHound ClientType = 0
	ClientTypeAzureHound ClientType = 1
)

type ClientVersion struct {
	ClientType ClientType
	Major      int
	Minor      int
	Patch      int
	Extra      int
}

// IsValidClientVersion checks the version from a user agent to ensure it's a valid UserAgent and that
// the version of the client is not EOL (currently SHS v1.x and SHS < v2.0.3). Returns an error if invalid
// and nil if valid
func IsValidClientVersion(userAgent string) error {
	if version, err := ParseClientVersion(userAgent); err != nil {
		return fmt.Errorf("error parsing client version: %w", err)
	} else if version.ClientType == ClientTypeAzureHound {
		return nil
	} else if version.ClientType == ClientTypeSharpHound {
		if version.Major < 2 {
			return fmt.Errorf("sharphound v1.x detected: %w", ErrRecommendSharphoundVersion)
		} else if version.Major == 2 && version.Minor == 0 && version.Patch < 3 {
			return fmt.Errorf("sharphound v2.0.2 or lower detected: %w", ErrRecommendSharphoundVersion)
		} else {
			return nil
		}
	} else { // unknown client type
		return ErrInvalidClientType
	}
}

func ParseClientVersion(userAgent string) (ClientVersion, error) {
	if strings.HasPrefix(userAgent, "azurehound") {
		return ParseAzurehoundVersion(userAgent)
	} else if strings.HasPrefix(userAgent, "sharphound") {
		return ParseSharpHoundVersion(userAgent)
	} else {
		return ClientVersion{}, ErrInvalidClientType
	}
}

func ParseAzurehoundVersion(userAgent string) (ClientVersion, error) {
	version := ClientVersion{
		ClientType: ClientTypeAzureHound,
		Major:      0,
		Minor:      0,
		Patch:      0,
		Extra:      0,
	}
	rgx := regexp.MustCompile("azurehound/v?([0-9]+).([0-9]+).([0-9]+)")
	if match := rgx.MatchString(userAgent); !match {
		return version, ErrInvalidAzureHoundVersion
	} else {
		rs := rgx.FindStringSubmatch(userAgent)
		if major, err := strconv.Atoi(rs[1]); err != nil {
			return version, err
		} else if minor, err := strconv.Atoi(rs[2]); err != nil {
			return version, err
		} else if patch, err := strconv.Atoi(rs[3]); err != nil {
			return version, err
		} else {
			version.Major = major
			version.Minor = minor
			version.Patch = patch
			version.Extra = 0
			return version, nil
		}
	}
}

func ParseSharpHoundVersion(userAgent string) (ClientVersion, error) {
	version := ClientVersion{
		ClientType: ClientTypeSharpHound,
		Major:      0,
		Minor:      0,
		Patch:      0,
		Extra:      0,
	}
	rgx := regexp.MustCompile("sharphound/([0-9]+).([0-9]+).([0-9]+).([0-9]+)")
	if match := rgx.MatchString(userAgent); !match {
		return version, ErrInvalidSharpHoundVersion
	} else {
		rs := rgx.FindStringSubmatch(userAgent)
		if major, err := strconv.Atoi(rs[1]); err != nil {
			return version, err
		} else if minor, err := strconv.Atoi(rs[2]); err != nil {
			return version, err
		} else if patch, err := strconv.Atoi(rs[3]); err != nil {
			return version, err
		} else if extra, err := strconv.Atoi(rs[4]); err != nil {
			return version, err
		} else {
			version.Major = major
			version.Minor = minor
			version.Patch = patch
			version.Extra = extra
			return version, nil
		}
	}
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
