// Copyright 2024 Specter Ops, Inc.
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

package git

import (
	"testing"

	"github.com/Masterminds/semver/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_parseLatestVersion(t *testing.T) {
	latestRelease := *semver.MustParse("v5.5.0")
	latestRC := *semver.MustParse("v5.5.0-rc1")

	t.Run("empty list of versions", func(t *testing.T) {
		_, err := parseLatestVersion([]string{})
		require.ErrorIs(t, err, ErrNoValidSemverFound)
	})

	t.Run("nil list of versions", func(t *testing.T) {
		_, err := parseLatestVersion(nil)
		assert.ErrorIs(t, err, ErrNoValidSemverFound)
	})

	t.Run("list of one invalid version", func(t *testing.T) {
		_, err := parseLatestVersion([]string{"invalid"})
		assert.ErrorIs(t, err, ErrNoValidSemverFound)
	})

	t.Run("list of valid versions with one invalid version", func(t *testing.T) {
		latest, err := parseLatestVersion([]string{"v5.3.0-rc1", "v5.5.0-rc1", "v5.4.0", "invalid"})
		require.Nil(t, err)
		assert.Equal(t, latestRC, latest)
	})

	t.Run("list of valid versions with latest non-RC", func(t *testing.T) {
		latest, err := parseLatestVersion([]string{"v5.3.0-rc1", "v5.5.0", "v5.4.0", "v5.4.0-rc1"})
		require.Nil(t, err)
		assert.Equal(t, latestRelease, latest)
	})
}
