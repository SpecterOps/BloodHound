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

package modules_test

import (
	"testing"

	"github.com/specterops/bloodhound/server/modules"
	"github.com/stretchr/testify/require"
)

func TestRegister_ForwardsToAnalysis(t *testing.T) {
	// A zero-value Deps has a nil Router. The delegation chain
	// (analysis.Register → handlers.Register → routerInst.GET) panics when it
	// reaches the nil pointer, confirming that Register calls into the analysis
	// module rather than silently no-oping.
	require.Panics(t, func() {
		modules.Register(modules.Deps{})
	})
}
