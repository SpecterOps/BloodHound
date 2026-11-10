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
	"net/http"
	"net/url"
	"testing"

	"github.com/specterops/bloodhound/cmd/api/src/auth"
	"github.com/specterops/bloodhound/cmd/api/src/bhctx"
	"github.com/specterops/bloodhound/cmd/api/src/model"
	"github.com/stretchr/testify/require"
)

// newSessionRequest returns an *http.Request whose context carries a bhctx.Context with the
// supplied session ID and host wired in. This mirrors what the auth middleware does for real requests.
func newSessionRequest(t *testing.T, sessionID int64, host *url.URL) *http.Request {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, "/api/v2/logout", nil)
	require.NoError(t, err)
	bhCtx := &bhctx.Context{
		Host:    host,
		AuthCtx: auth.Context{Session: model.UserSession{BigSerial: model.BigSerial{ID: sessionID}}},
	}
	return bhctx.SetRequestContext(req, bhCtx)
}
