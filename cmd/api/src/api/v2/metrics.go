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

package v2

import (
	"github.com/prometheus/client_golang/prometheus"
)

const (
	cypherQueryErrorTypeTimeout = "timeout"
	cypherQueryErrorTypeMemory  = "memory"
	cypherQueryErrorTypeDecode  = "decode"
	cypherQueryErrorTypeFitness = "fitness"
	cypherQueryErrorTypeParse   = "parse"
	cypherQueryErrorTypeExecute = "execute"
	cypherQueryErrorTypeUnknown = "unknown"
)

var cypherQueryErrors = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: "bh",
		Subsystem: "api",
		Name:      "cypher_query_errors",
	},
	[]string{"error_type"},
)

func RegisterApiEndpointMetrics(registry prometheus.Registerer) error {
	return registry.Register(cypherQueryErrors)
}
