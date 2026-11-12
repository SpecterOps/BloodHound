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

// Package featuretest supplies isolated OpenFeature clients for tests.
package featuretest

import (
	"context"
	"testing"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
	"github.com/open-feature/go-sdk/openfeature/memprovider"
)

type Overrides struct {
	Bools   map[string]bool
	Ints    map[string]int64
	Strings map[string]string
}

// NewClient uses the SDK's in-memory provider and releases its resources when
// the test completes. Each client has its own API so parallel tests cannot
// change each other's flags or the application's global provider.
func NewClient(t testing.TB, overrides Overrides) *openfeature.Client {
	var (
		flags         = make(map[string]memprovider.InMemoryFlag)
		evaluationAPI = isolated.NewAPI()
	)
	t.Helper()
	t.Cleanup(func() {
		if err := evaluationAPI.Shutdown(context.Background()); err != nil {
			t.Errorf("shutting down OpenFeature: %v", err)
		}
	})
	for key, value := range overrides.Bools {
		flags[key] = flag(value)
	}
	for key, value := range overrides.Ints {
		flags[key] = flag(value)
	}
	for key, value := range overrides.Strings {
		flags[key] = flag(value)
	}
	if err := evaluationAPI.SetProviderAndWait(t.Context(), memprovider.NewInMemoryProvider(flags)); err != nil {
		t.Fatalf("initializing OpenFeature: %v", err)
	}
	return evaluationAPI.NewClient()
}

func flag(value any) memprovider.InMemoryFlag {
	return memprovider.InMemoryFlag{
		State:          memprovider.Enabled,
		DefaultVariant: "test",
		Variants:       map[string]any{"test": value},
	}
}
