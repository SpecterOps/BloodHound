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

package config_test

import (
	"encoding/base64"
	"testing"

	"github.com/specterops/bloodhound/src/config"
	"github.com/stretchr/testify/require"
)

func TestGenerateRandomBase64String(t *testing.T) {
	// Using 32 bytes here as that's the number of bits we need for JWT keys
	str, err := config.GenerateRandomBase64String(32)
	require.Nilf(t, err, "Error generating random string: %s", err)
	require.NotEmpty(t, str, "Generated random string is empty")

	bytes, err := base64.StdEncoding.DecodeString(str)
	require.Nilf(t, err, "Error decoding random string: %s", err)
	require.Lenf(t, bytes, 32, "Generated random byte slice is not 32 bytes long")
}
