// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0
package endpoint

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/stretchr/testify/require"
)

// TestRewriteLegacyNameMatchIngestibleEndpoint_CasingGatedByUseRawObjectIDs verifies that the
// implicit legacy match_by "name" rewrite uses a case-sensitive operator (OperatorEquals) when
// useRawObjectIDs is true, and a case-insensitive operator (OperatorEqualsIgnoreCase) when false.
func TestRewriteLegacyNameMatchIngestibleEndpoint_CasingGatedByUseRawObjectIDs(t *testing.T) {
	t.Run("useRawObjectIDs=true produces case-sensitive name match", func(t *testing.T) {
		input := ein.IngestibleEndpoint{
			Value:   "SomeName",
			MatchBy: ein.MatchByName,
			Kind:    nil,
		}

		rewritten, err := rewriteLegacyNameMatchIngestibleEndpoint(input, true)
		require.NoError(t, err)
		require.Equal(t, ein.MatchByProperty, rewritten.MatchBy)
		require.Len(t, rewritten.Matchers, 1)
		require.Equal(t, "name", rewritten.Matchers[0].Key)
		require.Equal(t, ein.OperatorEquals, rewritten.Matchers[0].Operator)
		require.Equal(t, "SomeName", rewritten.Matchers[0].Value)
	})

	t.Run("useRawObjectIDs=false produces case-insensitive name match", func(t *testing.T) {
		input := ein.IngestibleEndpoint{
			Value:   "SomeName",
			MatchBy: ein.MatchByName,
			Kind:    nil,
		}

		rewritten, err := rewriteLegacyNameMatchIngestibleEndpoint(input, false)
		require.NoError(t, err)
		require.Equal(t, ein.MatchByProperty, rewritten.MatchBy)
		require.Len(t, rewritten.Matchers, 1)
		require.Equal(t, "name", rewritten.Matchers[0].Key)
		require.Equal(t, ein.OperatorEqualsIgnoreCase, rewritten.Matchers[0].Operator)
		require.Equal(t, "SomeName", rewritten.Matchers[0].Value)
	})

	t.Run("empty value returns error regardless of flag", func(t *testing.T) {
		input := ein.IngestibleEndpoint{
			Value:   "",
			MatchBy: ein.MatchByName,
		}

		_, errTrue := rewriteLegacyNameMatchIngestibleEndpoint(input, true)
		require.Error(t, errTrue)

		_, errFalse := rewriteLegacyNameMatchIngestibleEndpoint(input, false)
		require.Error(t, errFalse)
	})

	t.Run("non-name match_by is passed through unchanged regardless of flag", func(t *testing.T) {
		input := ein.IngestibleEndpoint{
			Value:   "some-id",
			MatchBy: ein.MatchByID,
		}

		rewrittenTrue, errTrue := rewriteLegacyNameMatchIngestibleEndpoint(input, true)
		require.NoError(t, errTrue)
		require.Equal(t, input, rewrittenTrue)

		rewrittenFalse, errFalse := rewriteLegacyNameMatchIngestibleEndpoint(input, false)
		require.NoError(t, errFalse)
		require.Equal(t, input, rewrittenFalse)
	})

	t.Run("explicit property matcher with equals_ignore_case is unaffected by rewrite", func(t *testing.T) {
		input := ein.IngestibleEndpoint{
			MatchBy: ein.MatchByProperty,
			Matchers: []ein.MatchExpression{
				{Key: "email", Operator: ein.OperatorEqualsIgnoreCase, Value: "user@example.com"},
			},
		}

		rewrittenTrue, errTrue := rewriteLegacyNameMatchIngestibleEndpoint(input, true)
		require.NoError(t, errTrue)
		require.Equal(t, input, rewrittenTrue)

		rewrittenFalse, errFalse := rewriteLegacyNameMatchIngestibleEndpoint(input, false)
		require.NoError(t, errFalse)
		require.Equal(t, input, rewrittenFalse)
	})
}
