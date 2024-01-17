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
	"fmt"
	"testing"

	"github.com/specterops/bloodhound/lab"
	"github.com/stretchr/testify/require"
)

type FizzBizz struct {
	Fizz string
	Bizz string
}

func NewFizzBizzFixture(fizz, bizz string) *lab.Fixture[FizzBizz] {
	return lab.NewFixture(func(harness *lab.Harness) (FizzBizz, error) {
		return FizzBizz{
			Fizz: fizz,
			Bizz: bizz,
		}, nil
	}, nil)
}

var BasicFizzBizzFixture = NewFizzBizzFixture("fizz", "bizz")
var FancyFizzBizzFixture = NewFizzBizzFixture("fizzy", "bizzy")

func NewFizzBizzHarness() *lab.Harness {
	harness := lab.NewHarness()
	lab.Pack(harness, BasicFuzzBuzzFixture) // Depends on BasicFooBarFixture
	lab.Pack(harness, BasicFizzBizzFixture)
	lab.Pack(harness, FancyFizzBizzFixture)
	return harness
}

func Test_UseHarness(t *testing.T) {
	assert, harness := lab.UseHarness(t, NewFizzBizzHarness())
	foobar, ok := lab.Unpack(harness, BasicFooBarFixture)
	assert.True(ok, "unable to unpack foobar fixture")
	assert.Equal("foo", foobar.Foo)
	assert.Equal("bar", foobar.Bar)

	fuzzbuzz, ok := lab.Unpack(harness, BasicFuzzBuzzFixture)
	assert.True(ok, "unable to unpack fuzzbuzz fixture")
	assert.Equal("foo", fuzzbuzz.Fuzz)
	assert.Equal("buzz", fuzzbuzz.Buzz)

	basicfizzbizz, ok := lab.Unpack(harness, BasicFizzBizzFixture)
	assert.True(ok, "unable to unpack basic fizzbizz fixture")
	assert.Equal("fizz", basicfizzbizz.Fizz)
	assert.Equal("bizz", basicfizzbizz.Bizz)

	fancyfizzbizz, ok := lab.Unpack(harness, FancyFizzBizzFixture)
	assert.True(ok, "unable to unpack fancy fizzbizz fixture")
	assert.Equal("fizzy", fancyfizzbizz.Fizz)
	assert.Equal("bizzy", fancyfizzbizz.Bizz)

	emptyFuzzBuzz, ok := lab.Unpack(harness, NewFuzzBuzzFixture())
	assert.False(ok, "should not be able to unpack fixtures that were not initialized by the harness")
	assert.Empty(emptyFuzzBuzz)
}

func Test_HarnessSetup(t *testing.T) {
	assert := require.New(t)

	emptyHarness := lab.NewHarness()
	_, err := emptyHarness.Setup()
	assert.NoError(err)

	badHarness := lab.NewHarness()
	badFixture := &lab.Fixture[any]{}
	lab.Pack(badHarness, badFixture)
	_, err = badHarness.Setup()
	assert.Error(err)
}

func Test_HarnessTeardown(t *testing.T) {
	assert := require.New(t)

	emptyHarness := lab.NewHarness()
	assert.NoError(emptyHarness.Teardown(nil))

	badHarness := lab.NewHarness()
	badFixture := lab.NewFixture(nil, func(*lab.Harness, any) error {
		return fmt.Errorf("fake error")
	})
	lab.Pack(badHarness, badFixture)
	assert.Error(badHarness.Teardown(badHarness))
}
