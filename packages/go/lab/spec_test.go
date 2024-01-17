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

func Test_NewSpecWithHooks(t *testing.T) {
	var (
		setupTimestamp    *time.Time
		teardownTimestamp *time.Time
		assert            = require.New(t)
	)
	t.Run("should call teardown after setup", func(t *testing.T) {
		lab.NewSpecWithHooks(t, func() (time.Time, error) {
			now := time.Now()
			setupTimestamp = &now
			return now, nil
		}, func(fixture time.Time) error {
			now := time.Now()
			teardownTimestamp = &now
			return nil
		}).Run(
			lab.TestCase("should not call teardown before test run", func(assert lab.Assertions, fixture time.Time) {
				assert.Nil(teardownTimestamp)
				assert.Greater(time.Now(), fixture)
			}),
		)
	})

	assert.NotNil(setupTimestamp)
	assert.NotNil(teardownTimestamp)
	assert.Greater(*teardownTimestamp, *setupTimestamp)
}

func Test_NewSpec(t *testing.T) {
	lab.NewSpec(t, NewFizzBizzHarness()).Run(
		lab.TestCase("should have BasicFooBarFixture", func(assert lab.Assertions, harness *lab.Harness) {
			foobar, ok := lab.Unpack(harness, BasicFooBarFixture)
			assert.True(ok)
			assert.Equal("foo", foobar.Foo)
			assert.Equal("bar", foobar.Bar)
		}),
		lab.TestCase("should have BasicFuzzBuzzFixture", func(assert lab.Assertions, harness *lab.Harness) {
			fuzzbuzz, ok := lab.Unpack(harness, BasicFuzzBuzzFixture)
			assert.True(ok)
			assert.Equal("foo", fuzzbuzz.Fuzz)
			assert.Equal("buzz", fuzzbuzz.Buzz)
		}),
		lab.TestCase("should have BasicFizzBizzFixture", func(assert lab.Assertions, harness *lab.Harness) {
			basicfizzbizz, ok := lab.Unpack(harness, BasicFizzBizzFixture)
			assert.True(ok)
			assert.Equal("fizz", basicfizzbizz.Fizz)
			assert.Equal("bizz", basicfizzbizz.Bizz)
		}),
		lab.TestCase("should have FancyFizzBizzFixture", func(assert lab.Assertions, harness *lab.Harness) {
			fancyfizzbizz, ok := lab.Unpack(harness, FancyFizzBizzFixture)
			assert.True(ok)
			assert.Equal("fizzy", fancyfizzbizz.Fizz)
			assert.Equal("bizzy", fancyfizzbizz.Bizz)
		}),
		lab.TestCase("should not unpack fixtures not initialized by the harness", func(assert lab.Assertions, harness *lab.Harness) {
			emptyFuzzBuzz, ok := lab.Unpack(harness, NewFuzzBuzzFixture())
			assert.False(ok)
			assert.Empty(emptyFuzzBuzz)
		}),
	)
}
