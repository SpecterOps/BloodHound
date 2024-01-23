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

package model_test

import (
	"testing"

	"github.com/specterops/bloodhound/cypher/frontend"
	"github.com/specterops/bloodhound/cypher/model"
	"github.com/specterops/bloodhound/cypher/test"
	"github.com/stretchr/testify/require"
)

type walker struct{}

func (w walker) Enter(stack *model.WalkStack, expression model.Expression) error {
	return nil
}

func (w walker) Exit(stack *model.WalkStack, expression model.Expression) error {
	return nil
}

func TestWalk(t *testing.T) {
	// Walk through all positive test cases to ensure that the walker can visit the involved types
	for _, testCase := range test.LoadFixture(t, test.PositiveTestCases).RunnableCases() {
		// Only bother with the string match tests
		if testCase.Type == test.TypeStringMatch {
			var (
				details              = test.UnmarshallTestCaseDetails[test.StringMatchTest](t, testCase)
				parseContext         = frontend.NewContext()
				queryModel, parseErr = frontend.ParseCypher(parseContext, details.Query)
			)

			if parseErr != nil {
				t.Fatalf("Parser errors: %s", parseErr.Error())
			}

			require.Nil(t, model.Walk(queryModel, &walker{}))
		}
	}
}
