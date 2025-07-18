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

package fixtures

import (
	"fmt"
	"log"

	v2 "github.com/specterops/bloodhound/cmd/api/src/api/v2"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/apiclient"
	"github.com/specterops/bloodhound/cmd/api/src/api/v2/integration"
	"github.com/specterops/bloodhound/cmd/api/src/config"
	"github.com/specterops/bloodhound/packages/go/lab"
)

const UserUpdateSecret = "userUser123***"

var BHAdminApiClientFixture = NewAdminApiClientFixture(ConfigFixture, BHApiFixture)

func NewAdminApiClientFixture(cfgFixture *lab.Fixture[config.Configuration], apiFixture *lab.Fixture[bool]) *lab.Fixture[apiclient.Client] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (apiclient.Client, error) {
		if cfg, ok := lab.Unpack(harness, cfgFixture); !ok {
			return apiclient.Client{}, fmt.Errorf("unable to unpack cfgFixture")
		} else {
			credentials := &apiclient.SecretCredentialsHandler{
				Username: integration.AdminPrincipal,
				Secret:   integration.AdminInitialSecret,
			}

			client, err := apiclient.NewClient(cfg.RootURL.String())
			if err != nil {
				return apiclient.Client{}, fmt.Errorf("unable to initialize api client: %w", err)
			}

			credentials.Client = client
			client.Credentials = credentials

			if user, err := client.GetSelf(); err != nil {
				return apiclient.Client{}, fmt.Errorf("failed looking up user details: %w", err)
			} else if err := client.SetUserSecretWithCurrentPassword(user.ID, v2.SetUserSecretRequest{
				CurrentSecret: integration.AdminInitialSecret,
				Secret:        integration.AdminUpdatedSecret,
			}); err != nil {
				return apiclient.Client{}, fmt.Errorf("failed resetting expired user password: %w", err)
			}

			return client, nil
		}
	}, nil)

	if err := lab.SetDependency(fixture, cfgFixture); err != nil {
		log.Fatalln(err)
	}
	if err := lab.SetDependency(fixture, apiFixture); err != nil {
		log.Fatalln(err)
	}

	return fixture
}

func NewUserApiClientFixture(cfgFixture *lab.Fixture[config.Configuration], adminApiFixture *lab.Fixture[apiclient.Client], roleNames ...string) *lab.Fixture[apiclient.Client] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (apiclient.Client, error) {
		if cfg, ok := lab.Unpack(harness, cfgFixture); !ok {
			return apiclient.Client{}, fmt.Errorf("unable to unpack cfgFixture")
		} else if adminClient, ok := lab.Unpack(harness, adminApiFixture); !ok {
			return apiclient.Client{}, fmt.Errorf("unable to unpack adminApiFixture")
		} else if username, err := config.GenerateSecureRandomString(7); err != nil {
			return apiclient.Client{}, fmt.Errorf("unable to generate random username")
		} else if roles, err := adminClient.ListRoles(); err != nil {
			return apiclient.Client{}, fmt.Errorf("unable to get roles")
		} else {
			var (
				roleIds []int32
				secret  = integration.UserInitialSecret
			)

			for _, r := range roleNames {
				if foundRole, found := roles.Roles.FindByName(r); !found {
					return apiclient.Client{}, fmt.Errorf("unable to find role")
				} else {
					roleIds = append(roleIds, foundRole.ID)
				}
			}

			// Create user in database
			if user, err := adminClient.CreateUser(username, "", roleIds); err != nil {
				return apiclient.Client{}, fmt.Errorf("failed to create user in db")
			} else {
				if err := adminClient.SetUserSecret(user.ID, secret, false); err != nil {
					return apiclient.Client{}, fmt.Errorf("failed resetting expired user password: %w", err)
				}
			}
			// Get api client for user
			client, err := apiclient.NewClient(cfg.RootURL.String())
			if err != nil {
				return apiclient.Client{}, fmt.Errorf("unable to initialize api client: %w", err)
			}

			credentials := &apiclient.SecretCredentialsHandler{
				Username: username,
				Secret:   secret,
			}
			credentials.Client = client
			client.Credentials = credentials

			if _, err := client.GetSelf(); err != nil {
				return apiclient.Client{}, fmt.Errorf("failed looking up user details: %w", err)
			}

			return client, nil
		}
	}, nil)

	if err := lab.SetDependency(fixture, cfgFixture); err != nil {
		log.Fatalln(err)
	}
	if err := lab.SetDependency(fixture, adminApiFixture); err != nil {
		log.Fatalln(err)
	}

	return fixture
}
