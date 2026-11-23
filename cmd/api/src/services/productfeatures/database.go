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

package productfeatures

import (
	"context"
	"errors"
	"sync"

	"github.com/open-feature/go-sdk/openfeature"
	"github.com/open-feature/go-sdk/openfeature/isolated"
)

// DatabaseFlags evaluates deployment-local flags through an isolated OpenFeature
// client so the provider used for product entitlements can be configured independently.
type DatabaseFlags struct {
	client func() (*openfeature.Client, error)
}

// NewDatabaseFlags binds the evaluator to a store. Each evaluation reads the
// current value so changes made through the feature flag API take effect immediately.
func NewDatabaseFlags(lookup func(context.Context, string) (bool, error), notFound error) DatabaseFlags {
	return DatabaseFlags{
		client: sync.OnceValues(func() (*openfeature.Client, error) {
			api := isolated.NewAPI()
			provider := databaseProvider{source: &databaseFlagSource{lookup: lookup, notFound: notFound}}
			if err := api.SetProviderAndWait(context.Background(), provider); err != nil {
				return nil, err
			}
			return api.NewClient(), nil
		}),
	}
}

func (s DatabaseFlags) IsEnabled(ctx context.Context, key string) (bool, error) {
	client, err := s.client()
	if err != nil {
		return false, err
	}

	// Keep the original store error for callers that rely on errors.Is. The
	// SDK returns the OpenFeature error code and message from the provider.
	lookup := &databaseLookup{ctx: ctx}
	value, err := client.BooleanValue(context.WithValue(ctx, databaseLookupKey{}, lookup), key, false, openfeature.EvaluationContext{})
	if lookup.err != nil {
		return value, lookup.err
	}
	return value, err
}

type databaseLookupKey struct{}

type databaseLookup struct {
	ctx context.Context
	err error
}

type databaseProvider struct {
	openfeature.NoopProvider
	// The SDK compares provider values when applying lifecycle events.
	source *databaseFlagSource
}

type databaseFlagSource struct {
	lookup   func(context.Context, string) (bool, error)
	notFound error
}

func (s databaseProvider) Metadata() openfeature.Metadata {
	return openfeature.Metadata{Name: "PostgreSQL"}
}

func (s databaseProvider) BooleanEvaluation(ctx context.Context, key string, defaultValue bool, _ openfeature.FlattenedContext) openfeature.BoolResolutionDetail {
	lookup, _ := ctx.Value(databaseLookupKey{}).(*databaseLookup)
	if lookup != nil {
		ctx = lookup.ctx
	}
	value, err := s.source.lookup(ctx, key)
	if err == nil {
		return openfeature.BoolResolutionDetail{
			Value:                    value,
			ProviderResolutionDetail: openfeature.ProviderResolutionDetail{Reason: openfeature.StaticReason},
		}
	}

	if lookup != nil {
		lookup.err = err
	}
	resolutionError := openfeature.NewGeneralResolutionError("database lookup failed", err)
	if errors.Is(err, s.source.notFound) {
		resolutionError = openfeature.NewFlagNotFoundResolutionError("flag is missing from the database")
	}
	return openfeature.BoolResolutionDetail{
		Value: defaultValue,
		ProviderResolutionDetail: openfeature.ProviderResolutionDetail{
			Reason:          openfeature.ErrorReason,
			ResolutionError: resolutionError,
		},
	}
}
