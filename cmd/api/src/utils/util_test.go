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

package utils_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/src/utils"
)

func TestIsValidClientVersion(t *testing.T) {
	var (
		err error
	)

	err = utils.IsValidClientVersion("azurehound/0.0.0")
	require.Nil(t, err)

	err = utils.IsValidClientVersion("sharphound/2.0.3.0")
	require.Nil(t, err)

	err = utils.IsValidClientVersion("sharphound/2.0.2.0")
	require.NotNil(t, err)
	require.ErrorIs(t, err, utils.ErrRecommendSharphoundVersion)

	err = utils.IsValidClientVersion("sharphound/1.9.3.0")
	require.NotNil(t, err)
	require.ErrorIs(t, err, utils.ErrRecommendSharphoundVersion)

	// Unknown client type
	err = utils.IsValidClientVersion("unknown/0.0.0")
	require.NotNil(t, err)
	require.ErrorIs(t, err, utils.ErrInvalidClientType)

	// Valid client type, no version
	err = utils.IsValidClientVersion("azurehound")
	require.NotNil(t, err)
	require.ErrorIs(t, err, utils.ErrInvalidAzureHoundVersion)

	// Invalid UA
	err = utils.IsValidClientVersion("garbage")
	require.NotNil(t, err)
	require.ErrorIs(t, err, utils.ErrInvalidClientType)
}

func TestParseClientVersion(t *testing.T) {
	version, err := utils.ParseClientVersion("sharphound/2.0.6.0")

	require.Nil(t, err)
	require.Equal(t, utils.ClientTypeSharpHound, version.ClientType)
	require.Equal(t, 2, version.Major)
	require.Equal(t, 0, version.Minor)
	require.Equal(t, 6, version.Patch)
	require.Equal(t, 0, version.Extra)

	version, err = utils.ParseClientVersion("sharphound/1.0.25.0")
	require.Nil(t, err)
	require.Equal(t, utils.ClientTypeSharpHound, version.ClientType)
	require.Equal(t, 1, version.Major)
	require.Equal(t, 0, version.Minor)
	require.Equal(t, 25, version.Patch)
	require.Equal(t, 0, version.Extra)

	version, err = utils.ParseClientVersion("azurehound/1.0.1")
	require.Nil(t, err)
	require.Equal(t, utils.ClientTypeAzureHound, version.ClientType)
	require.Equal(t, 1, version.Major)
	require.Equal(t, 0, version.Minor)
	require.Equal(t, 1, version.Patch)
	require.Equal(t, 0, version.Extra)

	version, err = utils.ParseClientVersion("azurehound/v1.0.1")
	require.Nil(t, err)
	require.Equal(t, utils.ClientTypeAzureHound, version.ClientType)
	require.Equal(t, 1, version.Major)
	require.Equal(t, 0, version.Minor)
	require.Equal(t, 1, version.Patch)
	require.Equal(t, 0, version.Extra)

	version, err = utils.ParseClientVersion("teststring")

	require.Equal(t, utils.ErrInvalidClientType, err)

	version, err = utils.ParseClientVersion("sharphound/abc")

	require.Equal(t, utils.ErrInvalidSharpHoundVersion, err)

	version, err = utils.ParseClientVersion("azurehound/abc")

	require.Equal(t, utils.ErrInvalidAzureHoundVersion, err)

	//This is the Eli test
	version, err = utils.ParseClientVersion("v2.-5.:biohazard_sign:")

	require.Equal(t, utils.ErrInvalidClientType, err)
}

func TestWriteResultJson(t *testing.T) {
	rr := httptest.NewRecorder()
	expectedResult := make(map[string]any)
	expectedResult["foo"] = "bar"

	utils.WriteResultJson(rr, expectedResult)

	resp := rr.Result()
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("returned the wrong status code: got %v want %v", status, http.StatusOK)
	}

	bytes, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(bytes, &data); err != nil {
		t.Errorf("could not unmarshal json")
	}

	if data["foo"] != expectedResult["foo"] {
		t.Errorf("returned the wrong body: got %v want %v", data, expectedResult)
	}
}

func TestWriteResultRawJson(t *testing.T) {
	rr := httptest.NewRecorder()
	expectedResult := make(map[string]any)
	expectedResult["foo"] = "bar"
	raw, _ := json.Marshal(expectedResult)

	utils.WriteResultRawJson(rr, raw)

	resp := rr.Result()
	defer resp.Body.Close()

	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("returned the wrong status code: got %v want %v", status, http.StatusOK)
	}

	bytes, _ := io.ReadAll(resp.Body)
	var data map[string]any
	if err := json.Unmarshal(bytes, &data); err != nil {
		t.Errorf("could not unmarshal json")
	}

	if data["foo"] != expectedResult["foo"] {
		t.Errorf("returned the wrong body: got %v want %v", data, expectedResult)
	}
}

func TestGetPageParamsForGraphQuery(t *testing.T) {
	expectedSkip := 0
	expectedLimit := 10
	expectedOrder := "n.foo, n.bar DESC"
	params := url.Values{"sort_by": []string{"foo", "-bar"}}

	if skip, limit, order, err := utils.GetPageParamsForGraphQuery(context.Background(), params); err != nil {
		t.Errorf("failed getting page params: %s", err)
	} else if skip != expectedSkip {
		t.Errorf("returned the wrong skip: got %v want %v", skip, expectedSkip)
	} else if limit != expectedLimit {
		t.Errorf("returned the wrong limit: got %v want %v", limit, expectedLimit)
	} else if order != expectedOrder {
		t.Errorf("returned the wrong order: got %v want %v", order, expectedOrder)
	}
}

func TestGetSkipParamForGraphQuery(t *testing.T) {
	params := url.Values{}
	params.Add("skip", "foo")

	_, err := utils.GetSkipParamForGraphQuery(params)
	require.Error(t, err)

	expectedSkip := "10"
	expectedSkipInt := 10

	params = url.Values{}
	params.Add("skip", expectedSkip)

	actualSkip, err := utils.GetSkipParamForGraphQuery(params)
	require.Nil(t, err)
	require.Equal(t, expectedSkipInt, actualSkip)
}

func TestGetLimitParamForGraphQuery(t *testing.T) {
	params := url.Values{}
	params.Add("limit", "foo")

	_, err := utils.GetLimitParamForGraphQuery(params)
	require.Error(t, err)

	expectedLimit := "10"
	expectedLimitInt := 10

	params = url.Values{}
	params.Add("skip", expectedLimit)

	actualSkip, err := utils.GetSkipParamForGraphQuery(params)
	require.Nil(t, err)
	require.Equal(t, expectedLimitInt, actualSkip)
}

func TestGetOrderForNeo4jQuery(t *testing.T) {
	expectedResult := "n.someColumn, n.anotherColumn DESC"

	params := url.Values{}
	params.Add("sort_by", "someColumn")
	params.Add("sort_by", "-anotherColumn")

	result := utils.GetOrderForNeo4jQuery(params)
	require.Equal(t, expectedResult, result)
}
