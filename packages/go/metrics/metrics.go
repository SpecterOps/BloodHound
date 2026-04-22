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
package metrics

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
)

// NewRegistry creates a custom Prometheus registry and registers it to the
// default registerer.
//
// Returns the custom registry for metric registration, or an error if registration fails.
func NewRegistry() (*prometheus.Registry, error) {
	promRegistry := prometheus.NewRegistry()

	// Register custom registry to default registerer so promhttp.Handler() exposes it
	if err := prometheus.DefaultRegisterer.Register(promRegistry); err != nil {
		return nil, fmt.Errorf("failed to register default metrics collector: %w", err)
	}

	return promRegistry, nil
}

// InitializeBHCEMetrics registers all BHCE Prometheus metrics to the provided registerer.
func InitializeBHCEMetrics(registerer prometheus.Registerer) error {
	if err := post.InitializePostProcessingMetrics(registerer); err != nil {
		return fmt.Errorf("failed to register post-processing metrics: %w", err)
	}

	return nil
}
