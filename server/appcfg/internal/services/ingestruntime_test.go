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

package services_test

import (
	"encoding/json"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/database/types"
	legacyappcfg "github.com/specterops/bloodhound/cmd/api/src/model/appcfg"
	"github.com/specterops/bloodhound/server/appcfg/internal/services"
	"github.com/specterops/bloodhound/server/appcfg/internal/services/mocks"
	"github.com/stretchr/testify/require"
)

func TestIngestRuntimeParameterValidation(t *testing.T) {
	var service = services.NewService(mocks.NewMockDatabase(t))
	require.True(t, service.IsAPIAllowedKey(services.IngestRuntimeKey))
	require.Equal(t, string(legacyappcfg.IngestRuntimeKey), string(services.IngestRuntimeKey))
	for _, test := range []struct {
		name, field string
		value       any
		valid       bool
	}{
		{name: "defaults", valid: true},
		{name: "minimum", field: "cleanup_queue_size", value: 1, valid: true},
		{name: "maximum", field: "cleanup_queue_size", value: 4096, valid: true},
		{name: "zero workers", field: "download_workers", value: 0},
		{name: "large worker pool", field: "download_workers", value: 17},
		{name: "large window", field: "download_window", value: 65},
		{name: "negative archive queue", field: "archive_ready_entries", value: -1},
		{name: "large archive queue", field: "archive_ready_entries", value: 17},
		{name: "zero cleanup workers", field: "cleanup_workers", value: 0},
		{name: "large cleanup pool", field: "cleanup_workers", value: 17},
		{name: "zero cleanup queue", field: "cleanup_queue_size", value: 0},
		{name: "large cleanup queue", field: "cleanup_queue_size", value: 4097},
		{name: "window below workers", field: "download_window", value: 1},
		{name: "fractional", field: "download_workers", value: 1.5},
		{name: "string", field: "download_workers", value: "2"},
		{name: "null", field: "download_workers", value: nil},
		{name: "unknown", field: "unknown", value: 1},
		{name: "missing", field: "download_workers", value: "remove"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var (
				fields = map[string]any{"download_workers": 2, "download_window": 4, "archive_ready_entries": 2, "cleanup_workers": 1, "cleanup_queue_size": 64}
				object types.JSONBObject
			)
			if test.field != "" {
				fields[test.field] = test.value
			}
			if test.name == "missing" {
				delete(fields, test.field)
			}
			encoded, err := json.Marshal(fields)
			require.NoError(t, err)
			require.NoError(t, json.Unmarshal(encoded, &object))
			legacyParameter := legacyappcfg.Parameter{Key: legacyappcfg.IngestRuntimeKey, Value: object}
			if test.valid {
				require.Nil(t, service.ValidateValue(services.IngestRuntimeKey, object))
				require.Nil(t, legacyParameter.Validate())
			} else {
				require.NotEmpty(t, service.ValidateValue(services.IngestRuntimeKey, object))
				require.NotEmpty(t, legacyParameter.Validate())
			}
		})
	}
}
