// Copyright 2026 Specter Ops, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
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
package ingest_test

import (
	"testing"

	"github.com/specterops/bloodhound/packages/go/chow/ingest"
)

func TestSiteDataTypes(t *testing.T) {
	var (
		siteDataTypes     = []ingest.DataType{"sites", "siteservers", "sitesubnets"}
		originalDataTypes = ingest.AllOriginalIngestDataTypes()
	)

	for _, siteDataType := range siteDataTypes {
		t.Run(string(siteDataType), func(t *testing.T) {
			if !siteDataType.IsValidOriginalType() {
				t.Errorf("%q is not a valid original type", siteDataType)
			}

			for _, originalDataType := range originalDataTypes {
				if originalDataType == siteDataType {
					return
				}
			}

			t.Errorf("%q is missing from AllOriginalIngestDataTypes", siteDataType)
		})
	}
}
