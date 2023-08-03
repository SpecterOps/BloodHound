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

type FooBar struct {
	Foo string
	Bar string
}

func NewFooBarFixture(foo, bar string) *lab.Fixture[FooBar] {
	return lab.NewFixture(func(harness *lab.Harness) (FooBar, error) {
		return FooBar{
			Foo: foo,
			Bar: bar,
		}, nil
	}, nil)
}

type FuzzBuzz struct {
	Fuzz string
	Buzz string
}

func NewFuzzBuzzFixture() *lab.Fixture[FuzzBuzz] {
	fuzzbuzzFixture := lab.NewFixture(func(harness *lab.Harness) (FuzzBuzz, error) {
		if foobar, ok := lab.Unpack(harness, BasicFooBarFixture); !ok {
			return FuzzBuzz{}, fmt.Errorf("missing %T", BasicFooBarFixture)
		} else {
			return FuzzBuzz{
				Fuzz: foobar.Foo,
				Buzz: "buzz",
			}, nil
		}
	}, nil)
	if err := lab.SetDependency(fuzzbuzzFixture, BasicFooBarFixture); err != nil {
		panic(err)
	} else {
		return fuzzbuzzFixture
	}
}

var BasicFooBarFixture = NewFooBarFixture("foo", "bar")
var BasicFuzzBuzzFixture = NewFuzzBuzzFixture()

func Test_UseFixture(t *testing.T) {
	assert, foobar := lab.UseFixture(t, NewFooBarFixture("foo", "bar"))
	assert.Equal("foo", foobar.Foo)
	assert.Equal("bar", foobar.Bar)
}

func Test_FixtureSetup(t *testing.T) {
	assert := require.New(t)
	badFixture1 := lab.NewFixture[any](nil, nil)
	_, err := badFixture1.Setup(nil)
	assert.Error(err)

	badFixture2 := lab.NewFixture(func(h *lab.Harness) (any, error) {
		return nil, fmt.Errorf("fake error")
	}, nil)
	_, err = badFixture2.Setup(nil)
	assert.Error(err)

	goodFixture := lab.NewFixture(func(*lab.Harness) (struct{}, error) {
		return struct{}{}, nil
	}, nil)
	value, err := goodFixture.Setup(nil)
	assert.NoError(err)
	assert.Equal(struct{}{}, value)
}

func Test_FixtureTeardown(t *testing.T) {
	assert := require.New(t)
	fixture1 := lab.NewFixture(nil, func(*lab.Harness, any) error {
		return nil
	})
	fixture2 := lab.NewFixture(nil, func(*lab.Harness, any) error {
		return fmt.Errorf("fake error")
	})
	assert.NoError(fixture1.Teardown(nil, nil))
	assert.Error(fixture2.Teardown(nil, nil))
}

func Test_SetDependency(t *testing.T) {
	assert := require.New(t)
	noOpSetup := func(*lab.Harness) (any, error) {
		return nil, nil
	}
	fixture1 := lab.NewFixture(noOpSetup, nil)
	fixture2 := lab.NewFixture(noOpSetup, nil)
	fixture3 := lab.NewFixture(noOpSetup, nil)
	assert.Error(lab.SetDependency(fixture1, fixture1))
	assert.NoError(lab.SetDependency(fixture1, fixture2))
	assert.NoError(lab.SetDependency(fixture2, fixture3))
	assert.Error(lab.SetDependency(fixture3, fixture1))
	assert, fuzzbuzz := lab.UseFixture(t, NewFuzzBuzzFixture())
	assert.Equal("foo", fuzzbuzz.Fuzz)
	assert.Equal("buzz", fuzzbuzz.Buzz)
}
