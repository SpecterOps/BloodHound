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

package lab

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCase[Fixture any](description string, testFn func(assert Assertions, fixture Fixture)) Case[Fixture] {
	return Case[Fixture]{
		Description: description,
		Test:        testFn,
	}
}

type Case[Fixture any] struct {
	Description string
	Test        func(assert Assertions, fixture Fixture)
}

func NewSpec[Ctx TBConstraint](ctx Ctx, harness *Harness) Spec[Ctx, *Harness] {
	return NewSpecWithHooks(ctx, harness.Setup, harness.Teardown)
}

func NewSpecWithHooks[Ctx TBConstraint, Fixture any](ctx Ctx, setup func() (Fixture, error), teardown func(Fixture) error) Spec[Ctx, Fixture] {
	return Spec[Ctx, Fixture]{
		ctx:      ctx,
		setup:    setup,
		teardown: teardown,
	}
}

type Spec[Ctx TBConstraint, Fixture any] struct {
	ctx      Ctx
	setup    func() (Fixture, error)
	teardown func(Fixture) error
}

func (s Spec[TestType, Fixture]) Run(cases ...Case[Fixture]) {
	_, fixture := UseHooks(s.ctx, s.setup, s.teardown)
	ctx := any(s.ctx).(testRunner[TestType])
	for _, testCase := range cases {
		ctx.Run(testCase.Description, func(t TestType) {
			defer func() {
				if recovery := recover(); recovery != nil {
					testing.TB(t).Fatalf("Panic during test execution: %v", recovery)
				}
			}()
			testCase.Test(require.New(testing.TB(t)), fixture)
		})
	}
}
