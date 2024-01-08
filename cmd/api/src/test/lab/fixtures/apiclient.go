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

	"github.com/specterops/bloodhound/lab"
	"github.com/specterops/bloodhound/src/api/v2/apiclient"
	"github.com/specterops/bloodhound/src/api/v2/integration"
)

var BHApiClientFixture = NewApiClientFixture(BHApiFixture)

func NewApiClientFixture(apiFixture *lab.Fixture[bool]) *lab.Fixture[apiclient.Client] {
	fixture := lab.NewFixture(func(harness *lab.Harness) (apiclient.Client, error) {
		if config, ok := lab.Unpack(harness, ConfigFixture); !ok {
			return apiclient.Client{}, fmt.Errorf("unable to unpack ConfigFixture")
		} else {
			credentials := &apiclient.SecretCredentialsHandler{
				Username: integration.AdminPrincipal,
				Secret:   integration.AdminInitialSecret,
			}

			client, err := apiclient.NewClient(config.RootURL.String())
			if err != nil {
				return apiclient.Client{}, fmt.Errorf("unable to initialize api client: %w", err)
			}

			credentials.Client = client
			client.Credentials = credentials

			if user, err := client.GetSelf(); err != nil {
				return apiclient.Client{}, fmt.Errorf("failed looking up user details: %w", err)
			} else if err := client.SetUserSecret(user.ID, integration.AdminUpdatedSecret, false); err != nil {
				return apiclient.Client{}, fmt.Errorf("failed resetting expired user password: %w", err)
			}

			return client, nil
		}
	}, nil)
	if err := lab.SetDependency(fixture, apiFixture); err != nil {
		log.Fatalln(err)
	}
	return fixture
}
