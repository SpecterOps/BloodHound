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

package handlers_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/specterops/bloodhound/server/analysis/handlers"
	"github.com/specterops/bloodhound/server/analysis/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildRequestedAnalysisView(t *testing.T) {
	var (
		requestedAt = time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
		input       = services.RequestedAnalysis{
			RequestedBy:           "analyst@example.com",
			RequestType:           services.RequestedAnalysisTypeDeletion,
			RequestedAt:           requestedAt,
			DeleteAllGraph:        true,
			DeleteSourcelessGraph: true,
			DeleteSourceKinds:     []string{"AZBase", "AZGroup"},
			DeleteRelationships:   []string{"HasSession", "MemberOf"},
		}
		view = handlers.BuildRequestedAnalysisView(input)
	)

	assert.Equal(t, input.RequestedBy, view.RequestedBy)
	assert.Equal(t, input.RequestType, view.RequestType)
	assert.Equal(t, input.RequestedAt, view.RequestedAt)
	assert.Equal(t, input.DeleteAllGraph, view.DeleteAllGraph)
	assert.Equal(t, input.DeleteSourcelessGraph, view.DeleteSourcelessGraph)
	assert.Equal(t, input.DeleteSourceKinds, view.DeleteSourceKinds)
	assert.Equal(t, input.DeleteRelationships, view.DeleteRelationships)
}

func TestRequestedAnalysisView_View(t *testing.T) {
	var (
		requestedAt = time.Date(2026, 3, 15, 12, 0, 0, 0, time.UTC)
		view        = handlers.RequestedAnalysisView{
			RequestedBy:         "analyst@example.com",
			RequestType:         services.RequestedAnalysisTypeAnalysis,
			RequestedAt:         requestedAt,
			DeleteAllGraph:      false,
			DeleteSourceKinds:   []string{"AZBase"},
			DeleteRelationships: []string{"HasSession"},
		}
	)

	rawJSON, err := view.JSONView()
	require.NoError(t, err)

	var decoded handlers.RequestedAnalysisView
	require.NoError(t, json.Unmarshal(rawJSON, &decoded))

	assert.Equal(t, view.RequestedBy, decoded.RequestedBy)
	assert.Equal(t, view.RequestType, decoded.RequestType)
	assert.Equal(t, view.RequestedAt.UTC(), decoded.RequestedAt.UTC())
	assert.Equal(t, view.DeleteAllGraph, decoded.DeleteAllGraph)
	assert.Equal(t, view.DeleteSourceKinds, decoded.DeleteSourceKinds)
	assert.Equal(t, view.DeleteRelationships, decoded.DeleteRelationships)
}
