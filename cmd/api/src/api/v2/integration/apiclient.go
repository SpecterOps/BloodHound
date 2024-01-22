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

package integration

import (
	"github.com/specterops/bloodhound/src/api/v2/apiclient"
	"github.com/stretchr/testify/require"
)

const (
	AdminPrincipal     = "admin"
	AdminInitialSecret = "admin"
	AdminUpdatedSecret = "adminAdmin123***"
)

func (s *Context) newAPIClient() apiclient.Client {
	authClient, err := apiclient.NewClient(s.GetRootURL().String())

	require.Nil(s.TestCtrl, err, "Unable to create auth client: %v", err)
	return authClient
}

func (s *Context) NewAPIClient(credentials apiclient.CredentialsHandler) apiclient.Client {
	newClient := s.newAPIClient()

	switch typedCredentials := credentials.(type) {
	case *apiclient.SecretCredentialsHandler:
		typedCredentials.Client = newClient
	}

	newClient.Credentials = credentials
	return newClient
}

func (s *Context) initAdminClient() {
	authClient := s.NewAPIClient(&apiclient.SecretCredentialsHandler{
		Username: AdminPrincipal,
		Secret:   AdminInitialSecret,
	})

	if user, err := authClient.GetSelf(); err != nil {
		s.TestCtrl.Fatalf("Failed looking up user details: %v", err)
	} else if err := authClient.SetUserSecret(user.ID, AdminUpdatedSecret, false); err != nil {
		s.TestCtrl.Fatalf("Failed resetting expired user password: %v", err)
	}

	s.adminClient = &authClient
}

func (s *Context) AdminClient() apiclient.Client {
	if s.adminClient == nil {
		s.initAdminClient()
	}

	return *s.adminClient
}
