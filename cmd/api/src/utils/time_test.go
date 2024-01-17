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
	"testing"
	"time"

	"github.com/specterops/bloodhound/src/utils"

	"github.com/stretchr/testify/require"
)

func TestReadRFC3339(t *testing.T) {
	var (
		now  = time.Now()
		then = now.Add(-time.Hour)
	)

	parsed, err := utils.ReadRFC3339(now.Format(time.RFC3339Nano), then)

	require.Nil(t, err)
	require.True(t, now.Equal(parsed))

	parsed, err = utils.ReadRFC3339("", then)

	require.Nil(t, err)
	require.True(t, then.Equal(parsed))

	_, err = utils.ReadRFC3339("This is not a valid timestamp.", then)

	require.NotNil(t, err)
}
