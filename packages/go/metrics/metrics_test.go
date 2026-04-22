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
package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/specterops/bloodhound/packages/go/metrics"
)

func TestNewRegistry(t *testing.T) {
	t.Parallel()

	registry, err := metrics.NewRegistry()

	require.NoError(t, err)
	require.NotNil(t, registry)

	// Remove the registry from the global default registerer so this test does not
	// leak state that interferes with other parallel tests.
	t.Cleanup(func() {
		prometheus.DefaultRegisterer.Unregister(registry)
	})

	// Verify the returned registry is functional and accepts metric registration
	testCounter := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "test_new_registry_counter_total",
		Help: "A counter used to verify the registry accepts metric registration.",
	})
	assert.NoError(t, registry.Register(testCounter))
}

func TestInitializeBHCEMetrics(t *testing.T) {
	t.Parallel()

	registry := prometheus.NewRegistry()
	err := metrics.InitializeBHCEMetrics(registry)

	assert.NoError(t, err)
}
