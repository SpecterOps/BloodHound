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

package featureflag

import (
	"fmt"
	"log/slog"

	fromEnv "github.com/open-feature/go-sdk-contrib/providers/from-env/pkg"
	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

// Flag key constants for OpenFeature.
// The from-env provider reads these from environment variables of the same name.
// The environment variable value must be a JSON configuration conforming to the from-env provider format.
//
// Example:
//
//	export FEATURE_ETAC_ENABLED='{"defaultVariant":"disabled","variants":[{"name":"enabled","criteria":[],"value":true},{"name":"disabled","criteria":[],"value":false}]}'
const (
	ETACEnabled         = "FEATURE_ETAC_ENABLED"
	PZMultiTierAnalysis = "FEATURE_PZ_MULTI_TIER_ANALYSIS"
	PZTierLimit         = "FEATURE_PZ_TIER_LIMIT"
	PZLabelLimit        = "FEATURE_PZ_LABEL_LIMIT"
)

// SetupProvider initializes the OpenFeature SDK with the from-env provider.
// This should be called once at application startup before any flag evaluations.
func SetupProvider() {
	provider := fromEnv.NewProvider()
	if err := openfeature.SetProviderAndWait(provider); err != nil {
		slog.Warn("Failed to set OpenFeature from-env provider", slog.String("error", err.Error()))
	}
}

// NewClient creates a new OpenFeature client for flag evaluations.
func NewClient() *openfeature.Client {
	return openfeature.NewClient("bloodhound")
}

// TestFlags holds feature flag overrides for test providers.
type TestFlags struct {
	ETACEnabled         bool
	PZMultiTierAnalysis bool
	PZTierLimit         int64
	PZLabelLimit        int64
}

func boolVariant(enabled bool) string {
	if enabled {
		return "on"
	}
	return "off"
}

func boolFlag(enabled bool) memprovider.InMemoryFlag {
	return memprovider.InMemoryFlag{
		State:          memprovider.Enabled,
		DefaultVariant: boolVariant(enabled),
		Variants: map[string]any{
			"on":  true,
			"off": false,
		},
	}
}

func intFlag(value int64) memprovider.InMemoryFlag {
	return memprovider.InMemoryFlag{
		State:          memprovider.Enabled,
		DefaultVariant: "value",
		Variants: map[string]any{
			"value": value,
		},
	}
}

// NewTestClient creates an OpenFeature client backed by an InMemoryProvider for use in tests.
// It registers a uniquely named provider to avoid conflicts between parallel tests.
func NewTestClient(flags TestFlags) *openfeature.Client {
	provider := memprovider.NewInMemoryProvider(map[string]memprovider.InMemoryFlag{
		ETACEnabled:         boolFlag(flags.ETACEnabled),
		PZMultiTierAnalysis: boolFlag(flags.PZMultiTierAnalysis),
		PZTierLimit:         intFlag(flags.PZTierLimit),
		PZLabelLimit:        intFlag(flags.PZLabelLimit),
	})

	name := fmt.Sprintf("test-%p", &provider)
	if err := openfeature.SetNamedProviderAndWait(name, provider); err != nil {
		slog.Warn("Failed to set test OpenFeature provider", slog.String("error", err.Error()))
	}

	return openfeature.NewClient(name)
}
