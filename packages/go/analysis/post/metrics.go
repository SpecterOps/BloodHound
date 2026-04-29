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
package post

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	postOperationsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "analysis",
		Name:      "post_processing_ops",
		Help:      "Post-processing operation statistics.",
	}, []string{
		"kind",
		"operation",
	})
)

// RegisterPostProcessingMetrics registers the analysis post-processing metrics.
func RegisterPostProcessingMetrics(namespace string, registerer prometheus.Registerer) error {
	postOperationsVec = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: namespace,
		Name:      "post_processing_ops",
		Help:      "Post-processing operation statistics.",
	}, []string{
		"kind",
		"operation",
	})

	if err := registerer.Register(postOperationsVec); err != nil {
		return fmt.Errorf("failed to register analysis post-processing metrics: %w", err)
	}

	return nil
}
