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

// TestGetUseRawObjectIDEnabled_DefaultsToDisabledOnLookupError verifies
// that a lookup failure from the underlying flag store surfaces as legacy behavior
// (returns false) so ingest keeps uppercasing object identifiers by default.
func TestGetUseRawObjectIDEnabled_DefaultsToDisabledOnLookupError(t *testing.T) {
	service := stubFlagByKeyer{
		errsByKey: map[string]error{
			appcfg.FeatureUseRawObjectID: errors.New("db unavailable"),
		},
	}

	require.False(t, appcfg.GetUseRawObjectIDEnabled(context.Background(), service))
}

// TestGetUseRawObjectIDEnabled_DefaultsToDisabledWhenFlagMissing verifies
// that a missing flag record (interpreted by the store as an error) also yields the
// safe default of false, matching the "legacy normalization behavior" contract.
func TestGetUseRawObjectIDEnabled_DefaultsToDisabledWhenFlagMissing(t *testing.T) {
	service := stubFlagByKeyer{}

	require.False(t, appcfg.GetUseRawObjectIDEnabled(context.Background(), service))
}

// TestGetUseRawObjectIDEnabled_ReturnsStoredEnabledValue verifies both
// enabled and disabled flag states are propagated verbatim from the flag store.
func TestGetUseRawObjectIDEnabled_ReturnsStoredEnabledValue(t *testing.T) {
	t.Run("returns true when the flag is enabled", func(t *testing.T) {
		service := stubFlagByKeyer{
			flagsByKey: map[string]appcfg.FeatureFlag{
				appcfg.FeatureUseRawObjectID: {
					Key:     appcfg.FeatureUseRawObjectID,
					Enabled: true,
				},
			},
		}

		require.True(t, appcfg.GetUseRawObjectIDEnabled(context.Background(), service))
	})

	t.Run("returns false when the flag is disabled", func(t *testing.T) {
		service := stubFlagByKeyer{
			flagsByKey: map[string]appcfg.FeatureFlag{
				appcfg.FeatureUseRawObjectID: {
					Key:     appcfg.FeatureUseRawObjectID,
					Enabled: false,
				},
			},
		}

		require.False(t, appcfg.GetUseRawObjectIDEnabled(context.Background(), service))
	})
}
