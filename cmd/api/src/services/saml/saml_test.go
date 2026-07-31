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

package saml

import (
	"testing"

	crewjamsaml "github.com/crewjam/saml"
	"github.com/stretchr/testify/require"
)

func TestValidateRequestID(t *testing.T) {
	t.Run("allows unsolicited response when IDP initiated login is enabled", func(t *testing.T) {
		err := validateRequestID(true, crewjamsaml.Response{}, nil)
		require.NoError(t, err)
	})

	t.Run("allows matching SP initiated response", func(t *testing.T) {
		err := validateRequestID(false, crewjamsaml.Response{
			InResponseTo: "request-id",
		}, []string{"request-id"})
		require.NoError(t, err)
	})

	t.Run("rejects mismatched SP initiated response", func(t *testing.T) {
		err := validateRequestID(false, crewjamsaml.Response{
			InResponseTo: "unexpected-request-id",
		}, []string{"request-id"})
		require.Error(t, err)
	})
}
