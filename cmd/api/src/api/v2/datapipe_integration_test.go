// Copyright 2024 Specter Ops, Inc.
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
	"time"

	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
)

func TestGetDatapipeStatus(t *testing.T) {
	testCtx := integration.NewFOSSContext(t)

	testCtx.WaitForDatapipeIdle(90 * time.Second)

	datapipeStatus, err := testCtx.AdminClient().GetDatapipeStatus()
	require.Nil(t, err)
	require.Equal(t, datapipeStatus.Status, model.DatapipeStatusIdle)
}
