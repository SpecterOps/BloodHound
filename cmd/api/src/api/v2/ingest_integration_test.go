// Copyright 2023 Specter Ops, Inc.
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

//go:build serial_integration
// +build serial_integration

package v2_test

import (
	"testing"

	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/test/fixtures/fixtures"
)

func Test_FileUploadWorkFlowVersion5(t *testing.T) {
	testCtx := integration.NewContext(t, integration.StartBHServer)

	testCtx.SendFileIngest([]string{
		"v5/ingest/domains.json",
		"v5/ingest/computers.json",
		"v5/ingest/containers.json",
		"v5/ingest/gpos.json",
		"v5/ingest/groups.json",
		"v5/ingest/ous.json",
		"v5/ingest/users.json",
		"v5/ingest/deleted.json",
		"v5/ingest/sessions.json",
	})

	//Assert that we created stuff we expected
	testCtx.AssertIngest(fixtures.IngestAssertions)
}

func Test_FileUploadWorkFlowVersion6(t *testing.T) {
	testCtx := integration.NewContext(t, integration.StartBHServer)

	testCtx.SendFileIngest([]string{
		"v6/ingest/domains.json",
		"v6/ingest/computers.json",
		"v6/ingest/containers.json",
		"v6/ingest/gpos.json",
		"v6/ingest/groups.json",
		"v6/ingest/ous.json",
		"v6/ingest/users.json",
		"v6/ingest/deleted.json",
		"v6/ingest/sessions.json",
	})

	//Assert that we created stuff we expected
	testCtx.AssertIngest(fixtures.IngestAssertions)
}
