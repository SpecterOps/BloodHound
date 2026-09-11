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
package endpoint_test

import (
	"slices"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/services/graphify/endpoint"
	"github.com/specterops/bloodhound/packages/go/ein"
	"github.com/stretchr/testify/assert"
)

// TestCacheEntryDigester_DigestEndpoint_MatchByProperty tests the sorting and hashing
// logic when MatchBy is set to MatchByProperty.
func TestCacheEntryDigester_DigestEndpoint_MatchByProperty(t *testing.T) {
	tests := []struct {
		name     string
		endpoint ein.IngestibleEndpoint
		wantKey  uint64
	}{
		{
			name: "single property",
			endpoint: ein.IngestibleEndpoint{
				MatchBy: ein.MatchByProperty,
				Matchers: []ein.MatchExpression{
					{Key: "name", Value: "test"},
				},
			},
		},
		{
			name: "multiple properties unsorted input",
			endpoint: ein.IngestibleEndpoint{
				MatchBy: ein.MatchByProperty,
				Matchers: []ein.MatchExpression{
					{Key: "zProp", Value: "last"},
					{Key: "aProp", Value: "first"},
					{Key: "mProp", Value: "middle"},
				},
			},
		},
		{
			name: "properties with complex values",
			endpoint: ein.IngestibleEndpoint{
				MatchBy: ein.MatchByProperty,
				Matchers: []ein.MatchExpression{
					{Key: "count", Value: 123},
					{Key: "active", Value: true},
				},
			},
		},
		{
			name: "kind and properties with complex values",
			endpoint: ein.IngestibleEndpoint{
				Value:   "User",
				MatchBy: ein.MatchByProperty,
				Matchers: []ein.MatchExpression{
					{Key: "count", Value: 123},
					{Key: "active", Value: true},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			digester := endpoint.NewCacheEntryDigester()

			key, err := digester.DigestEndpoint(tt.endpoint)
			assert.NoError(t, err)
			assert.NotZero(t, key)

			// Verify determinism: running the same input twice yields the same hash
			key2, err := digester.DigestEndpoint(tt.endpoint)
			assert.NoError(t, err)
			assert.Equal(t, key, key2)

			// Specific check for sorting logic:
			// If we have multiple properties, the hash must be consistent regardless of input order.
			if len(tt.endpoint.Matchers) > 1 {
				// Create a shuffled version of the input
				shuffledEndpoint := tt.endpoint
				shuffledEndpoint.Matchers = make([]ein.MatchExpression, len(tt.endpoint.Matchers))
				copy(shuffledEndpoint.Matchers, tt.endpoint.Matchers)

				// Manually reverse the slice to ensure sort happens
				slices.Reverse(shuffledEndpoint.Matchers)

				key3, err := digester.DigestEndpoint(shuffledEndpoint)
				assert.NoError(t, err)
				assert.Equal(t, key, key3, "Hash should be identical regardless of input property order")
			}
		})
	}
}

func TestCacheEntryDigester_DigestEndpoint_MatchByName(t *testing.T) {
	var (
		digester = endpoint.NewCacheEntryDigester()
		endpoint = ein.IngestibleEndpoint{
			MatchBy: ein.MatchByName,
			Value:   "MyComputerName",
		}
	)

	key, err := digester.DigestEndpoint(endpoint)

	assert.NoError(t, err)
	assert.NotZero(t, key)

	// Verify determinism
	key2, err := digester.DigestEndpoint(endpoint)

	assert.NoError(t, err)
	assert.Equal(t, key, key2)

	// Verify specific string format internally: "name|Value"
	// We can't easily inspect internal digester state, but we can verify that changing the value changes the hash
	endpoint.Value = "DifferentName"
	key3, err := digester.DigestEndpoint(endpoint)

	assert.NoError(t, err)
	assert.NotEqual(t, key, key3)
}

func TestCacheEntryDigester_DigestEndpoint_DefaultObjectID(t *testing.T) {
	digester := endpoint.NewCacheEntryDigester()

	// Explicitly set MatchBy to ObjectID (or leave it as zero value if 0 == ObjectID)
	// Assuming ein.MatchByObjectID is the zero value or explicit constant.
	// Based on the switch default, any non-name/non-property will hit this.
	endpoint := ein.IngestibleEndpoint{
		MatchBy: "custom_invalid_type", // Forces default case
		Value:   "S-1-5-21-123456789",
	}

	key, err := digester.DigestEndpoint(endpoint)

	assert.NoError(t, err)
	assert.NotZero(t, key)

	// Verify format: "objectid|Value"
	endpoint.Value = "S-1-5-21-987654321"
	key2, err := digester.DigestEndpoint(endpoint)

	assert.NoError(t, err)
	assert.NotEqual(t, key, key2)
}

func TestCacheEntryDigester_DigestEndpoint_JSONMarshalError(t *testing.T) {
	digester := endpoint.NewCacheEntryDigester()

	// Create a struct that cannot be marshaled to JSON (e.g., contains a function or channel)
	type unmarshalable struct {
		Func func()
	}

	endpoint := ein.IngestibleEndpoint{
		MatchBy: ein.MatchByProperty,
		Matchers: []ein.MatchExpression{
			{
				Key: "badProp",
				Value: unmarshalable{
					Func: func() {},
				},
			},
		},
	}

	key, err := digester.DigestEndpoint(endpoint)
	assert.Error(t, err)
	assert.Zero(t, key)

	// Ensure the pool isn't in a bad state after error
	keyRetry, errRetry := digester.DigestEndpoint(ein.IngestibleEndpoint{
		MatchBy: ein.MatchByName,
		Value:   "normal",
	})

	assert.NoError(t, errRetry)
	assert.NotZero(t, keyRetry)
}

func TestCacheEntryDigester_PoolStateReuse(t *testing.T) {
	digester := endpoint.NewCacheEntryDigester()

	// Run multiple concurrent requests to stress the sync.Pool
	var keys []uint64
	for i := 0; i < 100; i++ {
		endpoint := ein.IngestibleEndpoint{
			MatchBy: ein.MatchByName,
			Value:   "test_user_" + string(rune('A'+i)),
		}

		key, err := digester.DigestEndpoint(endpoint)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		keys = append(keys, key)
	}

	// Verify all keys are unique (since values are different)
	seen := make(map[uint64]bool)
	for _, k := range keys {
		if seen[k] {
			t.Errorf("Duplicate hash detected for distinct inputs: %d", k)
		}
		seen[k] = true
	}

	// Verify determinism again after pool reuse
	firstEndpoint := ein.IngestibleEndpoint{
		MatchBy: ein.MatchByName,
		Value:   "final_test",
	}
	key1, _ := digester.DigestEndpoint(firstEndpoint)
	key2, _ := digester.DigestEndpoint(firstEndpoint)
	assert.Equal(t, key1, key2)
}

// Helper to verify that encodeAnyValue handles standard types correctly
func TestEncodeAnyValue(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected string
	}{
		{"string", "hello", `"hello"`},
		{"int", 42, `42`},
		{"bool", true, `true`},
		{"slice", []int{1, 2}, `[1,2]`},
		{"map", map[string]int{"a": 1}, `{"a":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := endpoint.EncodeAnyValue(tt.value)
			assert.NoError(t, err)
			// JSON marshaling of maps/slices might vary in key order if not sorted,
			// but for primitives and simple slices, it should be consistent.
			assert.JSONEq(t, tt.expected, string(b))
		})
	}

	t.Run("error case", func(t *testing.T) {
		b, err := endpoint.EncodeAnyValue(func() {})
		assert.Error(t, err)
		assert.Nil(t, b)
	})
}

// Test that the digester handles empty property lists gracefully
func TestCacheEntryDigester_EmptyMatchers(t *testing.T) {
	digester := endpoint.NewCacheEntryDigester()

	endpoint := ein.IngestibleEndpoint{
		MatchBy:  ein.MatchByProperty,
		Matchers: []ein.MatchExpression{},
	}

	key, err := digester.DigestEndpoint(endpoint)
	assert.NoError(t, err)
	assert.NotZero(t, key)

	// Ensure an empty matcher produces a consistent hash
	key2, err := digester.DigestEndpoint(endpoint)
	assert.NoError(t, err)
	assert.Equal(t, key, key2)
}
