// Copyright 2026 Specter Ops, Inc.
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

package appcfg_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
)

// stubFlagByKeyer is a minimal in-test implementation of appcfg.GetFlagByKeyer
// that returns a pre-programmed FeatureFlag/error pair for the requested key.
type stubFlagByKeyer struct {
	flagsByKey map[string]appcfg.FeatureFlag
	errsByKey  map[string]error
}

func (s stubFlagByKeyer) GetFlagByKey(_ context.Context, key string) (appcfg.FeatureFlag, error) {
	if err, ok := s.errsByKey[key]; ok {
		return appcfg.FeatureFlag{}, err
	}
	if flag, ok := s.flagsByKey[key]; ok {
		return flag, nil
	}
	return appcfg.FeatureFlag{}, errors.New("flag not found")
}

// TestGetRawIngestObjectIdentifiersEnabled_DefaultsToDisabledOnLookupError verifies
// that a lookup failure from the underlying flag store surfaces as legacy behavior
// (returns false) so ingest keeps uppercasing object identifiers by default.
func TestGetRawIngestObjectIdentifiersEnabled_DefaultsToDisabledOnLookupError(t *testing.T) {
	service := stubFlagByKeyer{
		errsByKey: map[string]error{
			appcfg.FeatureRawIngestObjectIdentifiers: errors.New("db unavailable"),
		},
	}

	require.False(t, appcfg.GetRawIngestObjectIdentifiersEnabled(context.Background(), service))
}

// TestGetRawIngestObjectIdentifiersEnabled_DefaultsToDisabledWhenFlagMissing verifies
// that a missing flag record (interpreted by the store as an error) also yields the
// safe default of false, matching the "legacy normalization behavior" contract.
func TestGetRawIngestObjectIdentifiersEnabled_DefaultsToDisabledWhenFlagMissing(t *testing.T) {
	service := stubFlagByKeyer{}

	require.False(t, appcfg.GetRawIngestObjectIdentifiersEnabled(context.Background(), service))
}

// TestGetRawIngestObjectIdentifiersEnabled_ReturnsStoredEnabledValue verifies both
// enabled and disabled flag states are propagated verbatim from the flag store.
func TestGetRawIngestObjectIdentifiersEnabled_ReturnsStoredEnabledValue(t *testing.T) {
	t.Run("returns true when the flag is enabled", func(t *testing.T) {
		service := stubFlagByKeyer{
			flagsByKey: map[string]appcfg.FeatureFlag{
				appcfg.FeatureRawIngestObjectIdentifiers: {
					Key:     appcfg.FeatureRawIngestObjectIdentifiers,
					Enabled: true,
				},
			},
		}

		require.True(t, appcfg.GetRawIngestObjectIdentifiersEnabled(context.Background(), service))
	})

	t.Run("returns false when the flag is disabled", func(t *testing.T) {
		service := stubFlagByKeyer{
			flagsByKey: map[string]appcfg.FeatureFlag{
				appcfg.FeatureRawIngestObjectIdentifiers: {
					Key:     appcfg.FeatureRawIngestObjectIdentifiers,
					Enabled: false,
				},
			},
		}

		require.False(t, appcfg.GetRawIngestObjectIdentifiersEnabled(context.Background(), service))
	})
}
