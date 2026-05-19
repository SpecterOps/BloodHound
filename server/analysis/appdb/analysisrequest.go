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

package appdb

import (
	"time"

	"github.com/specterops/bloodhound/server/analysis/service"
)

// analysisRequest is the package-local representation of a row in the analysis_request_switch table.
// It exists only to hold raw scanned values; callers receive the application-level service.RequestedAnalysis.
type analysisRequest struct {
	RequestedBy           string
	RequestType           string
	RequestedAt           time.Time
	DeleteAllGraph        bool
	DeleteSourcelessGraph bool
	DeleteSourceKinds     []string
	DeleteRelationships   []string
}

// toRequestedAnalysis translates a raw DB row into the domain model.
func toRequestedAnalysis(row analysisRequest) service.RequestedAnalysis {
	return service.RequestedAnalysis{
		RequestedBy:           row.RequestedBy,
		RequestType:           service.RequestedAnalysisType(row.RequestType),
		RequestedAt:           row.RequestedAt,
		DeleteAllGraph:        row.DeleteAllGraph,
		DeleteSourcelessGraph: row.DeleteSourcelessGraph,
		DeleteSourceKinds:     row.DeleteSourceKinds,
		DeleteRelationships:   row.DeleteRelationships,
	}
}
