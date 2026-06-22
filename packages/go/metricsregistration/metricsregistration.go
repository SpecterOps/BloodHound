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

// Package metricsregistration aggregates Prometheus metric registrations for
// BHCE subsystems. Entrypoints call RegisterBHCEMetrics against the Prometheus
// registry they own before exposing it to prometheus.DefaultRegisterer.
package metricsregistration

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/specterops/bloodhound/cmd/api/src/api/middleware"
	"github.com/specterops/bloodhound/packages/go/analysis"
	"github.com/specterops/bloodhound/packages/go/analysis/post"
	"github.com/specterops/bloodhound/packages/go/metrics"
)

// RegisterBHCEMetrics registers all BHCE subsystem Prometheus metrics with the
// provided registerer.
func RegisterBHCEMetrics(registerer prometheus.Registerer) error {
	if err := analysis.RegisterAnalysisMetrics(registerer); err != nil {
		return fmt.Errorf("failed to register analysis metrics: %w", err)
	}

	if err := post.RegisterPostProcessingMetrics(registerer); err != nil {
		return fmt.Errorf("failed to register post-processing metrics: %w", err)
	}

	if err := middleware.RegisterApiMiddlewareMetrics(registerer); err != nil {
		return fmt.Errorf("failed to register API middleware metrics: %w", err)
	}

	if err := metrics.RegisterIngestMetrics(registerer); err != nil {
		return fmt.Errorf("failed to register ingest metrics: %w", err)
	}

	return nil
}
