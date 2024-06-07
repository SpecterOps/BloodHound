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

package v2_test

import (
	"testing"

	"github.com/specterops/bloodhound/src/api/v2/integration"
	"github.com/specterops/bloodhound/src/model"
	"github.com/stretchr/testify/require"
)

func TestRequestAnalysis(t *testing.T) {
	testCtx := integration.NewFOSSContext(t)

	err := testCtx.AdminClient().RequestAnalysis()
	require.Nil(t, err)

	analReq, err := testCtx.AdminClient().GetAnalysisRequest()
	require.Nil(t, err)
	require.Equal(t, analReq.RequestType, model.AnalysisRequestAnalysis)
}
