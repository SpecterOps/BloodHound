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

package lab_test

import (
	"testing"
	"time"

	"github.com/specterops/bloodhound/lab"
	"github.com/stretchr/testify/require"
)

func Test_UseHooks(t *testing.T) {
	t.Run("should return a time fixture that is initialized before the test", func(t *testing.T) {
		setup := func() (time.Time, error) {
			return time.Now(), nil
		}
		assert, timeFixture := lab.UseHooks(t, setup, nil)
		assert.Greater(time.Now(), timeFixture)
	})

	t.Run("should call setup before teardown", func(t *testing.T) {
		var (
			setupTimestamp    time.Time
			teardownTimestamp time.Time
			assert            = require.New(t)
		)

		t.Run("no-op", func(t *testing.T) {
			lab.UseHooks(t, func() (string, error) {
				setupTimestamp = time.Now()
				return "foo", nil
			}, func(string) error {
				teardownTimestamp = time.Now()
				return nil
			})
		})
		assert.Greater(teardownTimestamp, setupTimestamp)
	})
}
