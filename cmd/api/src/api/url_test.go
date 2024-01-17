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

package api_test

import (
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/src/api"

	"github.com/stretchr/testify/require"
)

func TestURLJoinPathEscapesSlashes(t *testing.T) {
	link, err := url.Parse("www.test.com/")
	require.NoError(t, err)

	result := api.URLJoinPath(*link, "/extension1", "extension2/", "extension3")
	expected := "www.test.com/extension1/extension2/extension3"
	require.Equal(t, expected, result.Path)
}

func TestURLJoinPathNoPrefix(t *testing.T) {
	link, err := url.Parse("www.test.com")
	require.NoError(t, err)

	result := api.URLJoinPath(*link, "extension1", "extension2")
	expected := "www.test.com/extension1/extension2"
	require.Equal(t, expected, result.Path)
}
